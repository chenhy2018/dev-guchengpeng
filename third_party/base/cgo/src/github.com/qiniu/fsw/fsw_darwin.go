package fsw

// #include "fsw_darwin.h"
// #cgo LDFLAGS: -framework CoreServices
import "C"

import (
	"fmt"
	"strconv"
	"syscall"
	"unsafe"
)

// -------------------------------------------------------------------------

type FSEvent C.FSEvent

type Event struct {
	*FSEvent
}

func (e Event) IsDir() bool {
	panic("not implemnt")
	return false
}

// IsCreate reports whether the Event was triggered by a creation
func (e Event) IsCreate() bool {
	panic("not implemnt")
	return false
}

// IsDelete reports whether the Event was triggered by a deletion
func (e Event) IsDelete() bool {
	panic("not implemnt")
	return false
}

// IsRename reports whether the Event was triggered by a rename
func (e Event) IsRename() bool {
	panic("not implemnt")
	return false
}

// IsRename reports whether the Event was triggered by a rename
func (e Event) Path() string {
	panic("not implemnt")
	return ""
}

// ----------------------------------------------------------

const (
	FSEventStreamEventFlagNone            = 0x00000000
	FSEventStreamEventFlagMustScanSubDirs = 0x00000001
	FSEventStreamEventFlagUserDropped     = 0x00000002
	FSEventStreamEventFlagKernelDropped   = 0x00000004
	FSEventStreamEventFlagEventIdsWrapped = 0x00000008
	FSEventStreamEventFlagHistoryDone     = 0x00000010
	FSEventStreamEventFlagRootChanged     = 0x00000020
	FSEventStreamEventFlagMount           = 0x00000040
	FSEventStreamEventFlagUnmount         = 0x00000080 /* These flags are only set if you specified the FileEvents*/
	/* flags when creating the stream.*/
	FSEventStreamEventFlagItemCreated       = 0x00000100
	FSEventStreamEventFlagItemRemoved       = 0x00000200
	FSEventStreamEventFlagItemInodeMetaMod  = 0x00000400
	FSEventStreamEventFlagItemRenamed       = 0x00000800
	FSEventStreamEventFlagItemModified      = 0x00001000
	FSEventStreamEventFlagItemFinderInfoMod = 0x00002000
	FSEventStreamEventFlagItemChangeOwner   = 0x00004000
	FSEventStreamEventFlagItemXattrMod      = 0x00008000
	FSEventStreamEventFlagItemIsFile        = 0x00010000
	FSEventStreamEventFlagItemIsDir         = 0x00020000
	FSEventStreamEventFlagItemIsSymlink     = 0x00040000
)

var eventFlags = map[uint]string{
	FSEventStreamEventFlagMustScanSubDirs: "FSEventStreamEventFlagMustScanSubDirs",
	FSEventStreamEventFlagRootChanged:     "FSEventStreamEventFlagRootChanged",
	FSEventStreamEventFlagMount:           "FSEventStreamEventFlagMount",
	FSEventStreamEventFlagUnmount:         "FSEventStreamEventFlagUnmount",
	FSEventStreamEventFlagHistoryDone:     "FSEventStreamEventFlagHistoryDone",
	FSEventStreamEventFlagItemCreated:     "FSEventStreamEventFlagItemCreated",
	FSEventStreamEventFlagItemRemoved:     "FSEventStreamEventFlagItemRemoved",
	FSEventStreamEventFlagItemIsFile:      "FSEventStreamEventFlagItemIsFile",
	FSEventStreamEventFlagItemIsDir:       "FSEventStreamEventFlagItemIsDir",
	FSEventStreamEventFlagItemIsSymlink:   "FSEventStreamEventFlagItemIsSymlink",
	FSEventStreamEventFlagItemRenamed:     "FSEventStreamEventFlagItemRenamed",
	FSEventStreamEventFlagItemChangeOwner: "FSEventStreamEventFlagItemChangeOwner",
}

func EventFlags(flags uint) (s string) {
	for flags != 0 {
		flags1 := flags & (flags - 1)
		flag := flags ^ flags1
		if s != "" {
			s += "|"
		}
		if sflag, ok := eventFlags[flag]; ok {
			s += sflag
		} else {
			s += "0x" + strconv.FormatUint(uint64(flag), 16)
		}
		flags = flags1
	}
	if s == "" {
		return "FSEventStreamEventFlagDirChanged"
	}
	return
}

func (ev Event) String() string {
	return fmt.Sprintf("%s: %s", ev.Name(), ev.FlagsString())
}

func (ev Event) Flags() uint {
	return uint(ev.EventFlags)
}

func (ev Event) FlagsString() string {
	return EventFlags(uint(ev.EventFlags))
}

func (ev Event) Name() string {
	return C.GoString(&ev.CPath[0])
}

// -------------------------------------------------------------------------

type Watcher struct {
	fsw    C.FSWatcher
	dir    string
	handle func(ev Event)
}

func Open(dir string) (watcher *Watcher, err error) {

	w := &Watcher{dir: dir}
	return w, nil
}

func (w *Watcher) Start(cfg *Config, handle func(ev Event)) (err error) {

	name := C.CString(w.dir)
	defer C.free(unsafe.Pointer(name))

	w.handle = handle
	rv := C.FSWatcher_Open(&w.fsw, name, unsafe.Pointer(w))
	if rv != 0 {
		err = syscall.Errno(rv)
		return
	}

	go C.FSWatcher_Run(&w.fsw)
	return nil
}

func (w *Watcher) Close() error {

	C.FSWatcher_Close(&w.fsw)
	return nil
}

// -------------------------------------------------------------------------

//export goEventCallback
func goEventCallback(obj, param unsafe.Pointer) {

	ev := *(*FSEvent)(unsafe.Pointer(param))
	w := (*Watcher)(unsafe.Pointer(obj))
	w.handle(Event{&ev})
}

// ----------------------------------------------------------
