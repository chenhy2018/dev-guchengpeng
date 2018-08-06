package rpc

import (
	"io"
	"net/http"
	"qbox.us/api"
	"qbox.us/errors"
	"github.com/qiniu/log.v1"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"
)

// ---------------------------------------------------------------------------

type ServerCodec interface {
	ReadRequestHeader(req *http.Request) (method string, err error)
	ReadRequestBody(r io.Reader, args interface{}) (err error)
	WriteResponse(w http.ResponseWriter, body interface{}, err error)
}

// ---------------------------------------------------------------------------

// Precompute the reflect type for error.  Can't use error directly
// because Typeof takes an empty interface value.  This is annoying.
var unusedError *error
var typeOfOsError = reflect.TypeOf(unusedError).Elem()

type methodType struct {
	method    reflect.Method
	ArgType   reflect.Type
	ReplyType reflect.Type
}

type service struct {
	name   string                 // name of service
	rcvr   reflect.Value          // receiver of methods for the service
	typ    reflect.Type           // type of the receiver
	method map[string]*methodType // registered methods
}

// Server represents an RPC Server.
type Server struct {
	serviceMap map[string]*service
	codec      ServerCodec
}

// NewServer returns a new Server.
func NewServer(codec ServerCodec) *Server {
	return &Server{serviceMap: make(map[string]*service), codec: codec}
}

// ---------------------------------------------------------------------------

// ServeHTTP implements an http.Handler that answers RPC requests.
func (server *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	method, err := server.codec.ReadRequestHeader(req)
	if err != nil {
		err = errors.Info(err, "rpc.Server.ServeHTTP", "rpc.codec.ReadRequestHeader").Detail(err).Warn()
		return
	}
	resp, err := server.Call(method, req.Body)
	server.codec.WriteResponse(w, resp, err)
}

func (server *Server) Call(method string, body io.Reader) (resp interface{}, err error) {

	serviceMethod := strings.Split(method, ".")
	if len(serviceMethod) != 2 {
		err = errors.Info(api.EInvalidArgs, "rpc.Server.Call", "rpc: service/method request ill-formed:", method).Warn()
		return
	}

	service := server.serviceMap[serviceMethod[0]]
	if service == nil {
		err = errors.Info(api.EInvalidArgs, "rpc.Server.Call", "rpc: can't find service:", method).Warn()
		return
	}

	mtype := service.method[serviceMethod[1]]
	if mtype == nil {
		err = errors.Info(api.EInvalidArgs, "rpc.Server.Call", "rpc: can't find method:", method).Warn()
		return
	}

	// Decode the argument value.
	var argv reflect.Value
	argIsValue := false // if true, need to indirect before calling.
	if mtype.ArgType.Kind() == reflect.Ptr {
		argv = reflect.New(mtype.ArgType.Elem())
	} else {
		argv = reflect.New(mtype.ArgType)
		argIsValue = true
	}
	// argv guaranteed to be a pointer now.
	replyv := reflect.New(mtype.ReplyType.Elem())
	err = server.codec.ReadRequestBody(body, argv.Interface())
	if err != nil {
		err = errors.Info(err, "rpc.Server.Call", method, "err: codec.ReadRequestBody failed").Detail(err).Warn()
		return
	}
	if argIsValue {
		argv = argv.Elem()
	}

	function := mtype.method.Func

	// Invoke the method, providing a new value for the reply.
	log.Debug("rpc.Server.Call - begin invoke:", method, argv)
	returnValues := function.Call([]reflect.Value{service.rcvr, argv, replyv})

	// The return value for the method is an os.Error.
	errInter := returnValues[0].Interface()
	log.Debug("rpc.Server.Call - end invoke:", method, argv, replyv.Interface(), errInter)

	if errInter != nil {
		err = errInter.(error)
		err = errors.Info(err, "rpc.Server.Call", method, "err: function.Call failed").Detail(err).Warn()
		return
	}

	return replyv.Interface(), nil
}

// ---------------------------------------------------------------------------

// Is this an exported - upper case - name?
func isExported(name string) bool {
	rune, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(rune)
}

// Is this type exported or local to this package?
func isExportedOrLocalType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.PkgPath() == "" || isExported(t.Name())
}

// Register publishes in the server the set of methods of the
// receiver value that satisfy the following conditions:
//	- exported method
//	- two arguments, both pointers to exported structs
//	- one return value, of type os.Error
// It returns an error if the receiver is not an exported type or has no
// suitable methods.
// The client accesses each method using a string of the form "Type.Method",
// where Type is the receiver's concrete type.
func (server *Server) Register(rcvr interface{}) error {
	return server.register(rcvr, "", false)
}

// RegisterName is like Register but uses the provided name for the type
// instead of the receiver's concrete type.
func (server *Server) RegisterName(name string, rcvr interface{}) error {
	return server.register(rcvr, name, true)
}

func (server *Server) register(rcvr interface{}, name string, useName bool) error {
	if server.serviceMap == nil {
		server.serviceMap = make(map[string]*service)
	}
	s := new(service)
	s.typ = reflect.TypeOf(rcvr)
	s.rcvr = reflect.ValueOf(rcvr)
	sname := reflect.Indirect(s.rcvr).Type().Name()
	if useName {
		sname = name
	}
	if sname == "" {
		log.Fatal("rpc: no service name for type", s.typ.String())
	}
	if s.typ.PkgPath() != "" && !isExported(sname) && !useName {
		err := errors.Info(errors.EINVAL, "rpc.Server.register", "type "+sname+" is not exported").Warn()
		return err
	}
	if _, present := server.serviceMap[sname]; present {
		err := errors.Info(errors.EEXIST, "rpc.Server.register", "service already defined: "+sname).Warn()
		return err
	}
	s.name = sname
	s.method = make(map[string]*methodType)

	// Install the methods
	for m := 0; m < s.typ.NumMethod(); m++ {
		method := s.typ.Method(m)
		mtype := method.Type
		mname := method.Name
		if mtype.PkgPath() != "" || !isExported(mname) {
			continue
		}
		// Method needs three ins: receiver, *args, *reply.
		if mtype.NumIn() != 3 {
			log.Debug("method", mname, "has wrong number of ins:", mtype.NumIn())
			continue
		}
		// First arg need not be a pointer.
		argType := mtype.In(1)
		if !isExportedOrLocalType(argType) {
			log.Debug(mname, "argument type not exported or local:", argType)
			continue
		}
		// Second arg must be a pointer.
		replyType := mtype.In(2)
		if replyType.Kind() != reflect.Ptr {
			log.Debug("method", mname, "reply type not a pointer:", replyType)
			continue
		}
		if !isExportedOrLocalType(replyType) {
			log.Debug("method", mname, "reply type not exported or local:", replyType)
			continue
		}
		// Method needs one out: os.Error.
		if mtype.NumOut() != 1 {
			log.Debug("method", mname, "has wrong number of outs:", mtype.NumOut())
			continue
		}
		if returnType := mtype.Out(0); returnType != typeOfOsError {
			log.Debug("method", mname, "returns", returnType.String(), "not os.Error")
			continue
		}
		s.method[mname] = &methodType{method: method, ArgType: argType, ReplyType: replyType}
	}

	if len(s.method) == 0 {
		err := errors.Info(errors.EINVAL,
			"rpc.Server.register", "type "+sname+" has no exported methods of suitable type").Warn()
		return err
	}
	server.serviceMap[s.name] = s
	return nil
}

// ---------------------------------------------------------------------------
