package line

import (
	"bytes"
	"reflect"
	"strconv"
)

func MarshalJson(v interface{}) []byte {
	buf := bytes.NewBuffer(nil)
	MarshalJsonWrite(buf, v)
	return buf.Bytes()
}

func MarshalJsonWrite(w *bytes.Buffer, v interface{}) {
	var first bool = true
	encodeJson(w, &first, []byte{}, reflect.ValueOf(v))
}

func comma(w *bytes.Buffer, first *bool) {
	if !*first {
		w.WriteByte(',')
	} else {
		*first = false
	}
}

func newPrefix(prefix []byte) []byte {
	if prefix == nil || len(prefix) == 0 {
		return []byte{}
	}
	return append(prefix, '.')
}

func EncodeJson(w *bytes.Buffer, first *bool, prefix []byte, v reflect.Value) {
	encodeJson(w, first, prefix, v)
}

func encodeJson(w *bytes.Buffer, first *bool, prefix []byte, v reflect.Value) {
	if !v.IsValid() {
		return
	}

	switch v.Kind() {
	case reflect.Bool:
		comma(w, first)
		w.Write(escape(prefix, tagEscapeCodeK, tagEscapeCodeV))
		w.WriteByte('=')
		w.Write([]byte(strconv.FormatBool(v.Bool())))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		comma(w, first)
		w.Write(escape(prefix, tagEscapeCodeK, tagEscapeCodeV))
		w.WriteByte('=')
		w.Write([]byte(strconv.FormatInt(v.Int(), 10)))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		comma(w, first)
		w.Write(escape(prefix, tagEscapeCodeK, tagEscapeCodeV))
		w.WriteByte('=')
		w.Write([]byte(strconv.FormatUint(v.Uint(), 10)))
	case reflect.Float32, reflect.Float64:
		comma(w, first)
		w.Write(escape(prefix, tagEscapeCodeK, tagEscapeCodeV))
		w.WriteByte('=')
		w.Write([]byte(strconv.FormatFloat(v.Float(), 'f', -1, 64)))
	case reflect.String:
		comma(w, first)
		w.Write(escape(prefix, tagEscapeCodeK, tagEscapeCodeV))
		w.WriteByte('=')
		w.WriteByte('"')
		w.Write(escape([]byte(v.String()), escapeCodeK, escapeCodeV))
		w.WriteByte('"')
	case reflect.Struct:
		// TODO
	case reflect.Map:
		if v.IsNil() {
			return
		}
		prefix = newPrefix(prefix)
		vs := v.MapKeys()
		for _, k := range vs {
			key := ""
			if k.Kind() == reflect.String {
				key = k.String()
			} else if k.Kind() == reflect.Interface && k.Elem().Kind() == reflect.String {
				key = k.Elem().String()
			} else {
				continue
			}
			encodeJson(w, first,
				append(prefix, []byte(key)...), v.MapIndex(k))
		}
		return
	case reflect.Slice:
		if v.IsNil() {
			return
		}
		fallthrough
	case reflect.Array:
		n := v.Len()
		prefix = newPrefix(prefix)
		for i := 0; i < n; i++ {
			encodeJson(w, first,
				append(prefix, []byte(strconv.FormatInt(int64(i), 10))...),
				v.Index(i))
		}
	case reflect.Interface, reflect.Ptr:
		if v.IsNil() {
			return
		}
		encodeJson(w, first, prefix, v.Elem())
	default:
	}

	return
}
