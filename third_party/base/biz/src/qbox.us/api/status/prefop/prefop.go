package prefop

type StatusCode int

const (
	Success StatusCode = iota
	Waiting
	Doing
	ExecFailed
	CallbackFailed
)

var descs = [...]string{
	"The fop was completed successfully",
	"The fop is waiting for execution",
	"The fop is executing now",
	"The fop is failed",
	"Callback failed",
}

func (s StatusCode) String() string {
	if 0 <= int(s) && int(s) < len(descs) {
		return descs[s]
	}
	return ""
}

type Status struct {
	Id          string  `json:"id"`
	Pipeline    string  `json:"pipeline"`
	Code        int     `json:"code"`
	Desc        string  `json:"desc"`
	Reqid       string  `json:"reqid"`
	InputBucket string  `json:"inputBucket"`
	InputKey    string  `json:"inputKey"`
	Items       []*Item `json:"items"`
}

type Item struct {
	Cmd       string      `json:"cmd"`
	Code      int         `json:"code"`
	Desc      string      `json:"desc"`
	Error     string      `json:"error,omitempty"`
	Hash      string      `json:"hash,omitempty"`
	Key       string      `json:"key,omitempty"`
	Keys      []string    `json:"keys,omitempty"`
	Result    interface{} `json:"result,omitempty"`
	ReturnOld int         `json:"returnOld"`
}

func NewStatus(id, pipelineId, reqid, inputBucket, inputKey string, cmds []string, code StatusCode) *Status {

	status := &Status{
		Id:          id,
		Pipeline:    pipelineId,
		Code:        int(code),
		Reqid:       reqid,
		InputBucket: inputBucket,
		InputKey:    inputKey,
		Desc:        code.String(),
		Items:       make([]*Item, len(cmds)),
	}
	for i, cmd := range cmds {
		status.Items[i] = &Item{Cmd: cmd, Code: int(code), Desc: code.String()}
	}
	return status
}

func (s *Status) SetStatusCode(code StatusCode) {
	s.Code = int(code)
	s.Desc = code.String()
}

func (s *Status) SetStatusDesc(desc string) {
	s.Desc = desc
}
