package fsw

/*
#include <stdlib.h>
#include <windows.h>

typedef struct tagMY_FILE_NOTIFY_INFORMATION_HEADER {
  DWORD NextEntryOffset;
  DWORD Action;
  DWORD FileNameLength;
} MY_FILE_NOTIFY_INFORMATION_HEADER;

static size_t my_sizeof_event() {
	return sizeof(MY_FILE_NOTIFY_INFORMATION_HEADER) + MAX_PATH;
}
*/
import "C"

import (
	"fmt"
	"io"
	"os"
	"syscall"
	"unicode/utf16"
	"unsafe"
)

// -------------------------------------------------------------------------

const (
	FILE_ACTION_ADDED            = C.FILE_ACTION_ADDED
	FILE_ACTION_REMOVED          = C.FILE_ACTION_REMOVED
	FILE_ACTION_MODIFIED         = C.FILE_ACTION_MODIFIED
	FILE_ACTION_RENAMED_OLD_NAME = C.FILE_ACTION_RENAMED_OLD_NAME
	FILE_ACTION_RENAMED_NEW_NAME = C.FILE_ACTION_RENAMED_NEW_NAME
)

type Event struct {
	Action int
	Name   string
}

var eventActionNames = []string{
	"FILE_ACTION_ADDED",
	"FILE_ACTION_REMOVED",
	"FILE_ACTION_MODIFIED",
	"FILE_ACTION_RENAMED_OLD_NAME",
	"FILE_ACTION_RENAMED_NEW_NAME",
}

func (e *Event) String() string {
	idx := uint(e.Action - 1)
	if idx < 5 {
		return fmt.Sprintf("%s: %s", eventActionNames[idx], e.Name)
	}
	return fmt.Sprintf("%d: %s", e.Action, e.Name)
}

// -------------------------------------------------------------------------

type Watcher struct {
	done chan bool // Channel for sending a "quit message" to the monitor goroutine
	hio  syscall.Handle
	hdir syscall.Handle
}

func Open(root string) (w *Watcher, err error) {

	hdir, err := syscall.CreateFile(
		syscall.StringToUTF16Ptr(root),
		C.FILE_LIST_DIRECTORY,
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE,
		nil, C.OPEN_EXISTING,
		C.FILE_FLAG_BACKUP_SEMANTICS|C.FILE_FLAG_OVERLAPPED, 0)
	if err != nil {
		return
	}
	hio, err := syscall.CreateIoCompletionPort(hdir, 0, 0, 0)
	if err != nil {
		syscall.CloseHandle(hdir)
		return
	}

	w = &Watcher{nil, hio, hdir}
	return
}

func (w *Watcher) Close() error {
	if w.hio != syscall.InvalidHandle {
		C.PostQueuedCompletionStatus(C.HANDLE(uintptr(w.hio)),
			C.DWORD(0),
			C.ULONG_PTR(uintptr(0)),
			C.LPOVERLAPPED(unsafe.Pointer(uintptr(0))))
		syscall.CloseHandle(w.hio)
		w.hio = syscall.InvalidHandle
		if w.done != nil {
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

	go monitor(w.hdir, w.hio, w.done, onEvent, cfg.OnError, cfg.EventsLimit)
}

func monitor(
	hdir syscall.Handle, hio syscall.Handle, done chan bool, onEvent func(ev Event), onError func(err error), eventsLimit int) {

	buffer := make([]byte, eventsLimit*int(C.my_sizeof_event()))

	for {
		err := readEvents(hdir, hio, done, onEvent, buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			onError(err)
		}
	}

	syscall.CloseHandle(hio)
	syscall.CloseHandle(hdir)
	fmt.Println("FSMon: terminated")
	if done != nil {
		<-done
	}
}

func readEvents(
	hdir syscall.Handle, hio syscall.Handle, done chan bool,
	onEvent func(ev Event), buffer []byte) (err error) {

	var ov C.OVERLAPPED
	var bytes C.DWORD = 0
	var flag C.DWORD = C.FILE_NOTIFY_CHANGE_FILE_NAME |
		C.FILE_NOTIFY_CHANGE_DIR_NAME | C.FILE_NOTIFY_CHANGE_ATTRIBUTES |
		C.FILE_NOTIFY_CHANGE_SIZE | C.FILE_NOTIFY_CHANGE_LAST_WRITE | C.FILE_NOTIFY_CHANGE_CREATION

	ok := C.ReadDirectoryChangesW(
		C.HANDLE(uintptr(hdir)),
		C.PVOID(&buffer[0]), C.DWORD(len(buffer)), C.BOOL(1), flag,
		C.LPDWORD(&bytes), C.LPOVERLAPPED(unsafe.Pointer(&ov)),
		C.LPOVERLAPPED_COMPLETION_ROUTINE(unsafe.Pointer(uintptr(0))))
	if ok == C.FALSE {
		err = os.NewSyscallError("ReadDirectoryChangesW", syscall.GetLastError())
		return
	}

	var recvedBytes uint32
	var exitFlag uint32
	var lpov *syscall.Overlapped
	err = syscall.GetQueuedCompletionStatus(hio, &recvedBytes, &exitFlag, &lpov, syscall.INFINITE)
	if err != nil {
		return
	}
	if lpov == nil { // quit
		err = io.EOF
		return
	}

	offset := uintptr(0)
	for {
		nih := (*C.MY_FILE_NOTIFY_INFORMATION_HEADER)(unsafe.Pointer(&buffer[offset]))
		fname := (*[1 << 29]uint16)(unsafe.Pointer(&buffer[offset+unsafe.Sizeof(*nih)]))[:nih.FileNameLength>>1]
		onEvent(Event{int(nih.Action), string(utf16.Decode(fname))})
		if nih.NextEntryOffset == 0 {
			break
		}
		offset += uintptr(nih.NextEntryOffset)
	}
	return
}

// -------------------------------------------------------------------------
