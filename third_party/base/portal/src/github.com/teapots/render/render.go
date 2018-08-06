package render

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/teapots/teapot"
)

const (
	ContentType    = "Content-Type"
	ContentLength  = "Content-Length"
	ContentBinary  = "application/octet-stream"
	ContentJSON    = "application/json"
	ContentHTML    = "text/html"
	ContentXHTML   = "application/xhtml+xml"
	ContentXML     = "text/xml"
	defaultCharset = "UTF-8"
)

// Included helper functions for use when rendering html
var helperFuncs = template.FuncMap{
	"yield": func() (string, error) {
		return "", fmt.Errorf("yield called with no layout defined")
	},
	"current": func() (string, error) {
		return "", nil
	},
}

// Render is a service that can be injected into a Teapot handler. Render provides functions for easily writing JSON and
// HTML templates out to a http Response.
type Render interface {
	// HTML renders a html template specified by the name and writes the result and given status to the http.ResponseWriter.
	HTML(name string, v interface{}, htmlOpt ...HTMLOptions) error
	// Render Template to string
	HTMLString(name string, v interface{}, htmlOpt ...HTMLOptions) (string, error)
	// Data writes the raw byte array to the http.ResponseWriter.
	Data(v []byte, status ...int) error
	// JSON writes the given status and JSON serialized version of the given value to the http.ResponseWriter.
	JSON(v interface{}, status ...int) error
	// XML writes the given status and XML serialized version of the given value to the http.ResponseWriter.
	XML(v interface{}, status ...int) error
	// Status is an alias for Error (writes an http status to the http.ResponseWriter)
	Status(status int)
	// Redirect is a convienience function that sends an HTTP redirect. If status is omitted, uses 302 (Found)
	Redirect(location string, status ...int)
	// Template returns the internal *template.Template used to render the HTML
	Template() *template.Template
	// Header exposes the header struct from http.ResponseWriter.
	Header() http.Header
}

// Delims represents a set of Left and Right delimiters for HTML template rendering
type Delims struct {
	// Left delimiter, defaults to {{
	Left string
	// Right delimiter, defaults to }}
	Right string
}

// Options is a struct for specifying configuration options for the render.Renderer middleware
type Options struct {
	// Directory to load templates. Default is "templates"
	Directory string
	// If Directory not exists then panic
	DirectoryMustExists bool
	// Layout template name. Will not render a layout if "". Defaults to "".
	Layout string
	// Extensions to parse template files from. Defaults to [".tmpl"]
	Extensions []string
	// Funcs is a slice of FuncMaps to apply to the template upon compilation. This is useful for helper functions. Defaults to [].
	Funcs []template.FuncMap
	// Delims sets the action delimiters to the specified strings in the Delims struct.
	Delims Delims
	// Appends the given charset to the Content-Type header. Default is "UTF-8".
	Charset string
	// Outputs human readable JSON
	IndentJSON bool
	// Outputs human readable XML
	IndentXML bool
	// Prefixes the JSON output with the given bytes.
	PrefixJSON []byte
	// Prefixes the XML output with the given bytes.
	PrefixXML []byte
	// Allows changing of output to XHTML instead of HTML. Default is "text/html"
	HTMLContentType string
}

func (o *Options) Valid() error {
	stat, err := os.Lstat(o.Directory)
	if err != nil {
		return err
	}

	mode := stat.Mode()
	if mode&os.ModeSymlink != 0 {
		link, err := filepath.EvalSymlinks(o.Directory)
		if err != nil {
			return err
		}
		o.Directory = link
	}
	return nil
}

// HTMLOptions is a struct for overriding some rendering Options for specific HTML call
type HTMLOptions struct {
	// Layout template name. Overrides Options.Layout.
	Layout string
}

// Renderer is a Middleware that maps a render.Render service into the Teapot handler chain. An single variadic render.Options
// struct can be optionally provided to configure HTML rendering. The default directory for templates is "templates" and the default
// file extension is ".tmpl".
//
// If RunMode is ModeDev then templates will be recompiled on every request. For more performance, set the
func Renderer(options ...Options) interface{} {
	opt := prepareOptions(options)
	cs := prepareCharset(opt.Charset)

	dirErr := opt.Valid()
	if opt.DirectoryMustExists && dirErr != nil {
		panic(dirErr)
	}

	// cached temaplte in prod mode
	var (
		tc  *template.Template
		mux sync.Mutex
	)

	return func(res http.ResponseWriter, req *http.Request, config *teapot.Config, log teapot.Logger) Render {
		uc := config.RunMode != teapot.ModeDev
		return &renderer{res, req, opt, cs, uc, &mux, &tc, nil}
	}
}

func prepareCharset(charset string) string {
	if len(charset) != 0 {
		return "; charset=" + charset
	}

	return "; charset=" + defaultCharset
}

func prepareOptions(options []Options) Options {
	var opt Options
	if len(options) > 0 {
		opt = options[0]
	}

	// Defaults
	if len(opt.Directory) == 0 {
		opt.Directory = "templates"
	}
	if len(opt.Extensions) == 0 {
		opt.Extensions = []string{".tmpl", ".html"}
	}
	if len(opt.HTMLContentType) == 0 {
		opt.HTMLContentType = ContentHTML
	}

	return opt
}

func compile(options Options) *template.Template {
	dir := options.Directory
	t := template.New(dir)
	t.Delims(options.Delims.Left, options.Delims.Right)
	// parse an initial template in case we don't have any
	template.Must(t.Parse("Teapot"))

	walkErr := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		r, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		ext := getExt(r)

		for _, extension := range options.Extensions {
			if ext == extension {

				buf, err := ioutil.ReadFile(path)
				if err != nil {
					panic(err)
				}

				name := (r[0 : len(r)-len(ext)])
				tmpl := t.New(filepath.ToSlash(name))

				// add our funcmaps
				for _, funcs := range options.Funcs {
					tmpl.Funcs(funcs)
				}

				// Bomb out if parse fails. We don't want any silent server starts.
				template.Must(tmpl.Funcs(helperFuncs).Parse(string(buf)))
				break
			}
		}

		return nil
	})
	if walkErr != nil {
		panic(walkErr)
	}

	return t
}

func getExt(s string) string {
	if strings.Index(s, ".") == -1 {
		return ""
	}
	return "." + strings.Join(strings.Split(s, ".")[1:], ".")
}

type renderer struct {
	http.ResponseWriter
	req             *http.Request
	opt             Options
	compiledCharset string
	useCache        bool

	mux *sync.Mutex
	tc  **template.Template
	t   *template.Template
}

func (r *renderer) JSON(v interface{}, status ...int) error {
	var result []byte
	var err error
	if r.opt.IndentJSON {
		result, err = json.MarshalIndent(v, "", "  ")
	} else {
		result, err = json.Marshal(v)
	}
	if err != nil {
		return err
	}

	// json rendered fine, write out the result
	r.Header().Set(ContentType, ContentJSON+r.compiledCharset)

	code := getCode(status)
	if code > 0 {
		r.WriteHeader(code)
	}

	if len(r.opt.PrefixJSON) > 0 {
		r.Write(r.opt.PrefixJSON)
	}
	_, err = r.Write(result)
	return err
}

func (r *renderer) HTML(name string, binding interface{}, htmlOpt ...HTMLOptions) error {
	r.compile()

	opt := r.prepareHTMLOptions(htmlOpt)
	// assign a layout if there is one
	if len(opt.Layout) > 0 {
		r.addYield(name, binding)
		name = opt.Layout
	}

	buf, err := r.execute(name, binding)
	if err != nil {
		return err
	}

	// template rendered fine, write out the result
	r.Header().Set(ContentType, r.opt.HTMLContentType+r.compiledCharset)
	_, err = io.Copy(r, buf)
	return err
}

func (r *renderer) HTMLString(name string, binding interface{}, htmlOpt ...HTMLOptions) (string, error) {
	r.compile()

	opt := r.prepareHTMLOptions(htmlOpt)
	// assign a layout if there is one
	if len(opt.Layout) > 0 {
		r.addYield(name, binding)
		name = opt.Layout
	}

	buf, err := r.execute(name, binding)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (r *renderer) XML(v interface{}, status ...int) error {
	var result []byte
	var err error
	if r.opt.IndentXML {
		result, err = xml.MarshalIndent(v, "", "  ")
	} else {
		result, err = xml.Marshal(v)
	}
	if err != nil {
		return err
	}

	// XML rendered fine, write out the result
	r.Header().Set(ContentType, ContentXML+r.compiledCharset)

	code := getCode(status)
	if code > 0 {
		r.WriteHeader(code)
	}

	if len(r.opt.PrefixXML) > 0 {
		r.Write(r.opt.PrefixXML)
	}
	_, err = r.Write(result)
	return err
}

func (r *renderer) Data(v []byte, status ...int) error {
	if r.Header().Get(ContentType) == "" {
		r.Header().Set(ContentType, ContentBinary)
	}

	code := getCode(status)
	if code > 0 {
		r.WriteHeader(code)
	}

	_, err := r.Write(v)
	return err
}

func (r *renderer) Status(status int) {
	r.WriteHeader(status)
}

func (r *renderer) Redirect(location string, status ...int) {
	code := getCode(status)
	if code == 0 {
		code = http.StatusFound
	}

	http.Redirect(r, r.req, location, code)
}

func (r *renderer) Template() *template.Template {
	return r.t
}

func (r *renderer) execute(name string, binding interface{}) (*bytes.Buffer, error) {
	buf := bytes.NewBuffer(nil)
	return buf, r.t.ExecuteTemplate(buf, name, binding)
}

func (r *renderer) addYield(name string, binding interface{}) {
	funcs := template.FuncMap{
		"yield": func() (template.HTML, error) {
			buf, err := r.execute(name, binding)
			// return safe html here since we are rendering our own template
			return template.HTML(buf.String()), err
		},
		"current": func() (string, error) {
			return name, nil
		},
	}
	r.t.Funcs(funcs)
}

func (r *renderer) prepareHTMLOptions(htmlOpt []HTMLOptions) HTMLOptions {
	if len(htmlOpt) > 0 {
		return htmlOpt[0]
	}

	return HTMLOptions{
		Layout: r.opt.Layout,
	}
}

func (r *renderer) compile() {
	if r.useCache {
		if *r.tc == nil {
			func() {
				r.mux.Lock()
				defer r.mux.Unlock()
				if *r.tc != nil {
					return
				}
				*r.tc = compile(r.opt)
			}()
		}
		// use a clone of the initial template
		r.t, _ = (*r.tc).Clone()

	} else {
		func() {
			r.mux.Lock()
			defer r.mux.Unlock()
			r.t = compile(r.opt)
		}()
	}
}

func getCode(codes []int) int {
	if len(codes) > 0 {
		return codes[0]
	}
	return 0
}
