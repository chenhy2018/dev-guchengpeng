package ufopcc

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/qiniu/http/httputil.v1"
)

var (
	ErrInvalidTarName = httputil.NewError(400, "invalid tar name")
)

// Common type.
type InstanceState int

const (
	STATE_NOEXIST InstanceState = iota
	STATE_RUNNING InstanceState = iota
	STATE_STOPPED InstanceState = iota
	STATE_UNKNOWN InstanceState = iota
)

func (is InstanceState) String() string {
	switch is {
	case STATE_NOEXIST:
		return "NoExist"
	case STATE_RUNNING:
		return "Running"
	case STATE_STOPPED:
		return "Stopped"
	case STATE_UNKNOWN:
		return "Unknown"
	}
	return "bad state"
}

type ImageState int

const (
	STATE_BUILDING      ImageState = iota
	STATE_BUILD_FAIL    ImageState = iota
	STATE_BUILD_SUCCESS ImageState = iota
)

func (is ImageState) String() string {

	switch is {
	case STATE_BUILDING:
		return "building"
	case STATE_BUILD_FAIL:
		return "build failed"
	case STATE_BUILD_SUCCESS:
		return "build success"
	}
	return "unknown"
}

type UserImage struct {
	State    ImageState `json:"state" bson:"state"`
	Version  int        `json:"version" bson:"v"`
	Uapp     string     `json:"uapp" bson:"uapp"`
	Uid      uint32     `json:"uid" bson:"uid"`
	Registry string     `json:"registry_host" bson:"res_h"`
	CreateAt int64      `json:"create_at" bson:"ct"`
	Desc     string     `json:"desc" bson:"desc"`
}

func (ui *UserImage) LogName() string {

	// [uid].[uapp].v[version].log
	return fmt.Sprintf("%d.%s.v%d.log", ui.Uid, ui.Uapp, ui.Version)
}

func ParseImageTarName(name string) (uid uint32, uapp string, v int, err error) {

	l := strings.Split(name, ".")
	if len(l) != 4 {
		err = ErrInvalidTarName
		return
	}

	uid_, err := strconv.ParseUint(l[0], 10, 32)
	if err != nil {
		err = ErrInvalidTarName
		return
	}
	uid = uint32(uid_)

	uapp = l[1]

	n, err := fmt.Sscanf(l[2], "v%d", &v)
	if err != nil || n != 1 {
		err = ErrInvalidTarName
		return
	}
	return
}
