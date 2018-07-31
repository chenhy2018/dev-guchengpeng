package fsw

import (
	"errors"
	"fmt"
	"github.com/qiniu/log.v1"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

// ----------------------------------------------------------

type EventInst struct {
	Mask   uint32 // Mask of events
	Cookie uint32 // Unique cookie associating related events (for rename(2))
	w      *Watcher
	Name   string // File name (optional)
	Wd     int    // Watch descriptor
}

type Event struct {
	*EventInst
}

func (e Event) IsDir() bool {
	return (e.Mask & IN_ISDIR) != 0
}

// IsCreate reports whether the Event was triggered by a creation
func (e Event) IsCreate() bool {
	return (e.Mask & (IN_CREATE|IN_MOVED_TO)) != 0
}

// IsDelete reports whether the Event was triggered by a deletion
func (e Event) IsDelete() bool {
	return (e.Mask & (IN_DELETE_SELF|IN_DELETE)) != 0
}

// IsRename reports whether the Event was triggered by a rename
func (e Event) IsRename() bool {
	return (e.Mask & (IN_MOVE_SELF|IN_MOVED_FROM)) != 0
}

// ----------------------------------------------------------

type entry struct {
	name     string
	parentWd int
}

type Watcher struct {
	paths  map[int]entry // Watch descriptor => entry
	done   chan bool     // Channel for sending a "quit message" to the reader goroutine
	Root   string
	fd     int
	RootWd int
}

func Open(path1 string) (w *Watcher, err error) {

	fd, errno := syscall.InotifyInit()
	if fd == -1 {
		return nil, os.NewSyscallError("inotify_init", errno)
	}

	root, name := filepath.Split(path1)
	if root != "" {
		root = root[:len(root)-1]
	} else {
		root = "."
	}
	w = &Watcher{make(map[int]entry), nil, root, fd, -1}

	err = w.addWatch(fd, -1, name, path1, logError)
	if err != nil {
		w.Close()
	}
	return
}

func (w *Watcher) Close() error {

	if w.fd != -1 {
		syscall.Close(w.fd)
		w.fd = -1
		if w.done != nil {
			log.Debug("Terminating fswatcher ...")
			w.done <- true
			w.done = nil
		}
	}
	return nil
}

func (w *Watcher) Start(cfg *Config, onEvent func(ev Event)) {

	validateCfg(&cfg)

	if cfg.WaitingToClose != 0 {
		w.done = make(chan bool, 0)
	}

	go w.monitor(w.fd, w.done, onEvent, cfg.OnError, cfg.EventsLimit)
}

// AddWatch adds path to the watched file set.
// The flags are interpreted as described in inotify_add_watch(2).
func (w *Watcher) addWatch(fd int, parentWd int, name, path string, onError func(err error)) error {

	wd, err := syscall.InotifyAddWatch(fd, path, watchFLAGS)
	if wd == -1 {
		onError(err)
		return err
	}

	w.paths[wd] = entry{name, parentWd}
	if parentWd == -1 {
		w.RootWd = wd
	}

	f, err := os.Open(path)
	if err != nil {
		onError(err)
		return err
	}
	defer f.Close()

	list, err := f.Readdir(-1)
	if err != nil {
		onError(err)
		return err
	}

	path1 := path + "/"
	for i := range list {
		fi := list[i]
		if fi.IsDir() {
			name := fi.Name()
			w.addWatch(fd, wd, name, path1+name, onError)
		}
	}

	return nil
}

func (w *Watcher) removeWatch(fd int, wd int) {

	syscall.InotifyRmWatch(fd, uint32(wd))
	delete(w.paths, wd)
}

func (w *Watcher) resolve(wd int) string {

	if wd != -1 {
		e := w.paths[wd]
		return w.resolve(e.parentWd) + "/" + e.name
	}
	return w.Root
}

func (w *Watcher) monitor(
	fd int, done chan bool, onEvent func(ev Event), onError func(err error), eventsLimit int) {

	var n int
	var oldev *EventInst
	var newpath entry
	var err error

	buf := make([]byte, syscall.SizeofInotifyEvent*eventsLimit) // Buffer for a maximum of 1024 raw events

	readEvents := func() {

		var offset uint32 = 0
		// We don't know how many events we just read into the buffer
		// While the offset points to at least one whole event...
		for offset <= uint32(n-syscall.SizeofInotifyEvent) {
			// Point "raw" to the event in the buffer
			raw := (*syscall.InotifyEvent)(unsafe.Pointer(&buf[offset]))

			var event EventInst
			event.Mask = uint32(raw.Mask)
			event.Cookie = uint32(raw.Cookie)
			event.Wd = int(raw.Wd)
			event.w = w
			nameLen := uint32(raw.Len)
			// If the event happened to the watched directory or the watched file, the kernel
			// doesn't append the filename to the event, but we would like to always fill the
			// the "Name" field with a valid filename. We retrieve the path of the watch from
			// the "paths" map.
			if nameLen > 0 {
				// Point "bytes" at the first byte of the filename
				bytes := (*[syscall.PathMax]byte)(unsafe.Pointer(&buf[offset+syscall.SizeofInotifyEvent]))
				// The filename is padded with NUL bytes. TrimRight() gets rid of those.
				event.Name = strings.TrimRight(string(bytes[0:nameLen]), "\000")
			}

			// Move to the next event in the buffer
			offset += syscall.SizeofInotifyEvent + nameLen

			// Send the event on the events channel
			switch event.Mask {
			case IN_ISDIR | IN_CREATE:
				w.addWatch(fd, event.Wd, event.Name, event.Path(), onError)
			case IN_MOVED_FROM | IN_ISDIR:
				oldev = &event
				newpath = entry{}
			case IN_MOVED_TO | IN_ISDIR:
				if oldev != nil && oldev.Cookie == event.Cookie { // move
					newpath = entry{event.Name, event.Wd}
				} else {
					w.addWatch(fd, event.Wd, event.Name, event.Path(), onError)
				}
			case IN_MOVE_SELF:
				if oldev != nil {
					olde := w.paths[event.Wd]
					if oldev.Wd == olde.parentWd && oldev.Name == olde.name {
						if newpath.name != "" {
							w.paths[event.Wd] = newpath
						} else {
							w.removeWatch(fd, event.Wd)
						}
					}
					oldev = nil
				}
				continue
			case IN_ISDIR | IN_DELETE_SELF:
				w.removeWatch(fd, event.Wd)
				continue
			case IN_IGNORED:
				continue
			}
			onEvent(Event{&event})
		}
	}

	for {
		n, err = syscall.Read(fd, buf)

		// If EOF message is received
		if n <= 0 {
			break
		}
		if n < syscall.SizeofInotifyEvent {
			onError(errors.New("inotify: short read in readEvents()"))
			continue
		}

		readEvents()
	}

	log.Debug("FSMon: terminated -", err)
	if done != nil {
		<-done
	}
	return
}

func (e *EventInst) Path() string {
	return e.w.resolve(e.Wd) + "/" + e.Name
}

// String formats the event e in the form
// "filename: 0xEventMask = IN_ACCESS|IN_ATTRIB_|..."
func (e *EventInst) String() string {
	var events string = ""

	m := e.Mask
	for _, b := range eventBits {
		if m&b.Value != 0 {
			m &^= b.Value
			events += "|" + b.Name
		}
	}

	if m != 0 {
		events += fmt.Sprintf("|%#x", m)
	}
	if len(events) > 0 {
		events = " == " + events[1:]
	}

	return fmt.Sprintf("%q: %#x%s %d", e.Path(), e.Mask, events, e.Cookie)
}

const (
	watchFLAGS = IN_ATTRIB | IN_CREATE | IN_DELETE | IN_DELETE_SELF |
		IN_MODIFY | IN_MOVE | IN_MOVED_FROM | IN_MOVED_TO | IN_MOVE_SELF |
		IN_UNMOUNT | IN_Q_OVERFLOW
)

const (
	// Options for inotify_init() are not exported
	// IN_CLOEXEC    uint32 = syscall.IN_CLOEXEC
	// IN_NONBLOCK   uint32 = syscall.IN_NONBLOCK

	// Options for AddWatch
	IN_DONT_FOLLOW uint32 = syscall.IN_DONT_FOLLOW
	IN_ONESHOT     uint32 = syscall.IN_ONESHOT
	IN_ONLYDIR     uint32 = syscall.IN_ONLYDIR

	// The "IN_MASK_ADD" option is not exported, as AddWatch
	// adds it automatically, if there is already a watch for the given path
	// IN_MASK_ADD      uint32 = syscall.IN_MASK_ADD

	// Events
	IN_ACCESS        uint32 = syscall.IN_ACCESS
	IN_ALL_EVENTS    uint32 = syscall.IN_ALL_EVENTS
	IN_ATTRIB        uint32 = syscall.IN_ATTRIB
	IN_CLOSE         uint32 = syscall.IN_CLOSE
	IN_CLOSE_NOWRITE uint32 = syscall.IN_CLOSE_NOWRITE
	IN_CLOSE_WRITE   uint32 = syscall.IN_CLOSE_WRITE
	IN_CREATE        uint32 = syscall.IN_CREATE
	IN_DELETE        uint32 = syscall.IN_DELETE
	IN_DELETE_SELF   uint32 = syscall.IN_DELETE_SELF
	IN_MODIFY        uint32 = syscall.IN_MODIFY
	IN_MOVE          uint32 = syscall.IN_MOVE
	IN_MOVED_FROM    uint32 = syscall.IN_MOVED_FROM
	IN_MOVED_TO      uint32 = syscall.IN_MOVED_TO
	IN_MOVE_SELF     uint32 = syscall.IN_MOVE_SELF
	IN_OPEN          uint32 = syscall.IN_OPEN

	// Special events
	IN_ISDIR      uint32 = syscall.IN_ISDIR
	IN_IGNORED    uint32 = syscall.IN_IGNORED
	IN_Q_OVERFLOW uint32 = syscall.IN_Q_OVERFLOW
	IN_UNMOUNT    uint32 = syscall.IN_UNMOUNT
)

var eventBits = []struct {
	Value uint32
	Name  string
}{
	{IN_ACCESS, "IN_ACCESS"},
	{IN_ATTRIB, "IN_ATTRIB"},
	//	{IN_CLOSE, "IN_CLOSE"},
	{IN_CLOSE_NOWRITE, "IN_CLOSE_NOWRITE"},
	{IN_CLOSE_WRITE, "IN_CLOSE_WRITE"},
	{IN_CREATE, "IN_CREATE"},
	{IN_DELETE, "IN_DELETE"},
	{IN_DELETE_SELF, "IN_DELETE_SELF"},
	{IN_MODIFY, "IN_MODIFY"},
	//	{IN_MOVE, "IN_MOVE"},
	{IN_MOVED_FROM, "IN_MOVED_FROM"},
	{IN_MOVED_TO, "IN_MOVED_TO"},
	{IN_MOVE_SELF, "IN_MOVE_SELF"},
	{IN_OPEN, "IN_OPEN"},
	{IN_ISDIR, "IN_ISDIR"},
	{IN_IGNORED, "IN_IGNORED"},
	{IN_Q_OVERFLOW, "IN_Q_OVERFLOW"},
	{IN_UNMOUNT, "IN_UNMOUNT"},
}
