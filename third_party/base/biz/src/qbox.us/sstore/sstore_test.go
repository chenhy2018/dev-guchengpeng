package sstore

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/qiniu/log.v1"
	"github.com/qiniu/ts"
	"github.com/stretchr/testify.v1/require"
	"qbox.us/cc/time"
	"qbox.us/objid"
)

func init() {
	log.SetOutputLevel(0)
}

func TestLenPutTime(t *testing.T) {
	if LenPutTime != 8 {
		t.Fatal(LenPutTime)
	}
}

const keyHint = 113

var key = []byte("foo")

type keyFinder struct{}

func (r keyFinder) Find(i uint32) []byte {
	if i != 113 {
		return nil
	}
	fmt.Println("keyFinder.Find ok!")
	return key
}

func TestEncodeFhandle(t *testing.T) {

	fh := make([]byte, 21)
	copy(fh, "test")

	fi := &FhandleInfo{
		Fhandle:  fh,
		MimeType: "image/jpg",
		AttName:  "foo.jpg",
		Fsize:    123,
		Deadline: time.Nanoseconds() + 10e9,
		KeyHint:  keyHint,
		Uid:      1234,
		PutTime:  time.Nanoseconds(),
	}

	oid := objid.Encode(1234, uint64(time.Nanoseconds()))
	/*	err := DecodeOid(fi, oid)
		if err != nil {
			ts.Fatal(t, "EncodeFhandle failed:", err)
		}
	*/

	efh := EncodeFhandle_11(fi, key)
	{
		fi2 := DecodeFhandle(efh, oid, keyFinder{})
		if fi2 == nil || !bytes.Equal(fi2.Fhandle, fi.Fhandle) || fi2.Fsize != fi.Fsize || fi2.OidLow != fi.OidLow || fi2.OidHigh != fi.OidHigh ||
			fi2.Deadline != fi.Deadline || fi2.MimeType != fi.MimeType || fi2.Ver != FhandleVer_10 {
			ts.Fatal(t, "DecodeFhandle failed")
		}
		require.Equal(t, efh, EncodeFhandle_11(fi2, key))
		fmt.Println(fi2)
	}
	{
		fi2 := DecodeFhandle_10(efh, oid, keyFinder{})
		if fi2 == nil || !bytes.Equal(fi2.Fhandle, fi.Fhandle) || fi2.Fsize != fi.Fsize || fi2.OidLow != fi.OidLow || fi2.OidHigh != fi.OidHigh ||
			fi2.Deadline != fi.Deadline || fi2.MimeType != fi.MimeType || fi2.Ver != FhandleVer_10 {
			ts.Fatal(t, "DecodeFhandle failed")
		}
		require.Equal(t, efh, EncodeFhandle_11(fi2, key))
		fmt.Println(fi2)
	}

	fi.Public = 0x23
	fi.Utype = 0x10
	efh2 := EncodeFhandle_11(fi, key)
	{
		fi2 := DecodeFhandle(efh2, oid, keyFinder{})
		if fi2 == nil || !bytes.Equal(fi2.Fhandle, fi.Fhandle) || fi2.Fsize != fi.Fsize || fi2.OidLow != fi.OidLow || fi2.OidHigh != fi.OidHigh ||
			fi2.Deadline != fi.Deadline || fi2.MimeType != fi.MimeType || fi2.Utype != fi.Utype ||
			fi2.Ver != FhandleVer_11 || fi2.Public != fi.Public || fi2.PutTime != 0 {
			ts.Fatal(t, "DecodeFhandle failed")
		}
		require.Equal(t, efh2, EncodeFhandle_11(fi2, key))
		fmt.Println(fi2)
	}

	efh3 := EncodeFhandle_10(fi, key)
	{
		fi2 := DecodeFhandle(efh3, oid, keyFinder{})
		if fi2 == nil || !bytes.Equal(fi2.Fhandle, fi.Fhandle) || fi2.Fsize != fi.Fsize || fi2.OidLow != fi.OidLow || fi2.OidHigh != fi.OidHigh ||
			fi2.Deadline != fi.Deadline || fi2.MimeType != fi.MimeType || fi2.Ver != FhandleVer_10 || fi2.Public != 0 {
			ts.Fatal(t, "DecodeFhandle failed")
		}
		require.Equal(t, efh3, EncodeFhandle_10(fi2, key))
		fmt.Println(fi2)
	}
	{
		fi2 := DecodeFhandle_10(efh3, oid, keyFinder{})
		if fi2 == nil || !bytes.Equal(fi2.Fhandle, fi.Fhandle) || fi2.Fsize != fi.Fsize || fi2.OidLow != fi.OidLow || fi2.OidHigh != fi.OidHigh ||
			fi2.Deadline != fi.Deadline || fi2.MimeType != fi.MimeType || fi2.Ver != FhandleVer_10 || fi2.Public != 0 {
			ts.Fatal(t, "DecodeFhandle failed")
		}
		require.Equal(t, efh3, EncodeFhandle_10(fi2, key))
		fmt.Println(fi2)
	}

	fi.Utype = 0
	efh4 := EncodeFhandle_12(fi, key)
	{
		fi2 := DecodeFhandle(efh4, oid, keyFinder{})
		if fi2 == nil || !bytes.Equal(fi2.Fhandle, fi.Fhandle) || fi2.Fsize != fi.Fsize || fi2.OidLow != fi.OidLow || fi2.OidHigh != fi.OidHigh ||
			fi2.Deadline != fi.Deadline || fi2.MimeType != fi.MimeType || fi2.Utype != fi.Utype ||
			fi2.Ver != FhandleVer_12 || fi2.Public != fi.Public || fi2.PutTime != fi.PutTime {
			ts.Fatal(t, "DecodeFhandle failed")
		}
		require.Equal(t, efh4, EncodeFhandle_12(fi2, key))
		fmt.Println(fi2)
	}

	fi.Utype = 0x8384
	efh4 = EncodeFhandle_12(fi, key)
	{
		fi2 := DecodeFhandle(efh4, oid, keyFinder{})
		if fi2 == nil || !bytes.Equal(fi2.Fhandle, fi.Fhandle) || fi2.Fsize != fi.Fsize || fi2.OidLow != fi.OidLow || fi2.OidHigh != fi.OidHigh ||
			fi2.Deadline != fi.Deadline || fi2.MimeType != fi.MimeType || fi2.Utype != fi.Utype ||
			fi2.Ver != FhandleVer_12 || fi2.Public != fi.Public || fi2.PutTime != fi.PutTime {
			ts.Fatal(t, "DecodeFhandle failed")
		}
		require.Equal(t, efh4, EncodeFhandle_12(fi2, key))
		fmt.Println(fi2)
	}

	fi.AttName = "6aa318717e0a5e8145e3981467bd80f4839814724481472c98125450813c6627fe560e80ff8c16b36c467b339a87fc0bff7d5e6e37fed50380b001f810b6b671a162f814979780499158139bc1814a47f7e667b080794e78135979810590f7bdeb2d7e8ad9775c9254814ed9d814f40a81218138000e668150b487fe3b5e.apk"
	efh5 := EncodeFhandle_13(fi, key)
	{
		fi2 := DecodeFhandle(efh5, oid, keyFinder{})
		if fi2 == nil || !bytes.Equal(fi2.Fhandle, fi.Fhandle) || fi2.Fsize != fi.Fsize || fi2.OidLow != fi.OidLow || fi2.OidHigh != fi.OidHigh ||
			fi2.Deadline != fi.Deadline || fi2.MimeType != fi.MimeType || fi2.Utype != fi.Utype || fi2.AttName != fi.AttName ||
			fi2.Ver != FhandleVer_13 || fi2.Public != fi.Public || fi2.PutTime != fi.PutTime {
			ts.Fatal(t, "DecodeFhandle failed")
		}
		require.Equal(t, efh5, EncodeFhandle_13(fi2, key))
		fmt.Println(fi2)
	}

	// newest
	fi.Utype = 0
	fi.FileType = 1
	efh99 := EncodeFhandle(fi, key)
	{
		fi2 := DecodeFhandle(efh99, oid, keyFinder{})
		if fi2 == nil || !bytes.Equal(fi2.Fhandle, fi.Fhandle) || fi2.Fsize != fi.Fsize || fi2.OidLow != fi.OidLow || fi2.OidHigh != fi.OidHigh ||
			fi2.Deadline != fi.Deadline || fi2.MimeType != fi.MimeType || fi2.Utype != fi.Utype || fi2.AttName != fi.AttName ||
			fi2.Ver != FhandleVer || fi2.Public != fi.Public || fi2.PutTime != fi.PutTime || fi.FileType != fi2.FileType {
			ts.Fatal(t, "DecodeFhandle failed")
		}
		require.Equal(t, efh99, EncodeFhandle(fi2, key))
		fmt.Println(fi2)
	}
	fi.Utype = 0x8384
	fi.FileType = 0x6174
	efh99 = EncodeFhandle(fi, key)
	{
		fi2 := DecodeFhandle(efh99, oid, keyFinder{})
		if fi2 == nil || !bytes.Equal(fi2.Fhandle, fi.Fhandle) || fi2.Fsize != fi.Fsize || fi2.OidLow != fi.OidLow || fi2.OidHigh != fi.OidHigh ||
			fi2.Deadline != fi.Deadline || fi2.MimeType != fi.MimeType || fi2.Utype != fi.Utype || fi2.AttName != fi.AttName ||
			fi2.Ver != FhandleVer || fi2.Public != fi.Public || fi2.PutTime != fi.PutTime || fi.FileType != fi2.FileType {
			ts.Fatal(t, "DecodeFhandle failed")
		}
		require.Equal(t, efh99, EncodeFhandle(fi2, key))
		fmt.Println(fi2)
	}
}
