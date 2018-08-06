package api

import (
	"os"
	"runtime"
	"strings"
	"syscall"

	fstab "github.com/deniswernert/go-fstab"
	"github.com/qiniu/log.v1"
)

var DiskTest bool

func GetDiskSpace(path string) (avail int64, total int64, err error) {
	fs := syscall.Statfs_t{}
	err = syscall.Statfs(path, &fs)
	if err != nil {
		return
	}

	avail = int64(fs.Bavail * uint64(fs.Bsize))
	total = int64(fs.Blocks * uint64(fs.Bsize))
	return
}

func GetEtcFstab(path string) (ok bool, err error) {
	if DiskTest {
		return ok, nil
	}
	mounts, err := fstab.ParseSystem()
	if err != nil {
		return
	}
	for _, mount := range mounts {
		if mount.File != "/" && strings.HasPrefix(path, mount.File+"/") {
			ok = true
			return
		}
	}
	log.Println("can not found disk...")
	log.Println(mounts.String())
	return
}

func GetProcMounts(path string) (ok bool, err error) {
	if DiskTest {
		return ok, nil
	}
	mounts, err := fstab.ParseProc()
	if err != nil {
		return
	}
	for _, mount := range mounts {
		if mount.File != "/" && strings.HasPrefix(path, mount.File+"/") {
			ok = true
			return
		}
	}
	log.Println("can not found disk...")
	log.Println(mounts.String())
	return
}

func DiskMounted(path string) bool {
	if os.Getenv("TRAVIS_BUILD_DIR") != "" || runtime.GOOS != "linux" {
		return true
	}
	if ok, err2 := GetEtcFstab(path); !ok || err2 != nil {
		return false
	}
	if ok, err2 := GetProcMounts(path); !ok || err2 != nil {
		return false
	}
	return true
}
