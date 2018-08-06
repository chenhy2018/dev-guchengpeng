package userinfo

import (
	"github.com/qiniu/rpc.v1"
	"qbox.us/qconf/qconfapi"
)

// ------------------------------------------------------------------------

const IOGZipGroupPrefix = GroupPrefix + "gzip_mime_types:"

func init() {
	prefixes = append(prefixes, IOGZipGroupPrefix)
}

// ------------------------------------------------------------------------

type GZipMimeTypeInfo struct {
	Mimes map[string]bool `bson:"gzip_mime_types"`
}

func (r Client) GetGzipMimeTypes(l rpc.Logger, uid uint32) (ret GZipMimeTypeInfo, err error) {

	err = r.Conn.Get(l, &ret, MakeId(IOGZipGroupPrefix, uid), qconfapi.Cache_Normal)
	return
}

// ------------------------------------------------------------------------
