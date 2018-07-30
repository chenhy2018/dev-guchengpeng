package urlutil

import (
	"net/url"
	"github.com/qiniu/log.v1"
	"reflect"
	"strconv"
	"strings"
)

// Encode encodes the values into ``URL encoded'' form.
// e.g. "foo=bar&bar=baz"
func EncodeQuery(v map[string]interface{}) string {

	if v == nil {
		return ""
	}
	parts := make([]string, 0, len(v)) // will be large enough for most uses
	for k, vs := range v {
		if vstr, ok := vs.(string); ok {
			part := k + "=" + url.QueryEscape(vstr)
			parts = append(parts, part)
		} else {
			parts = appendPart(parts, k, reflect.ValueOf(vs))
		}
	}
	return strings.Join(parts, "&")
}

func appendPart(parts []string, k string, data reflect.Value) []string {

retry:

	kind := data.Kind()
	switch kind {
	case reflect.String:
		part := k + "=" + url.QueryEscape(data.String())
		parts = append(parts, part)
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		part := k + "=" + strconv.FormatInt(data.Int(), 10)
		parts = append(parts, part)
	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uintptr, reflect.Uint16, reflect.Uint8:
		part := k + "=" + strconv.FormatUint(data.Uint(), 10)
		parts = append(parts, part)
	case reflect.Bool:
		v := "=false"
		if data.Bool() {
			v = "=true"
		}
		parts = append(parts, k+v)
	case reflect.Ptr, reflect.Interface:
		data = data.Elem()
		goto retry
	case reflect.Struct:
		parts = appendStructPart(parts, k, data)
	case reflect.Map:
		parts = appendMapPart(parts, k, data)
	case reflect.Array, reflect.Slice:
		for i := 0; i < data.Len(); i++ {
			v := data.Index(i)
			parts = appendPart(parts, k, v)
		}
	case reflect.Float64, reflect.Float32:
		part := k + "=" + strconv.FormatFloat(data.Float(), 'g', -1, 64)
		parts = append(parts, part)
	default:
		log.Warn("appendPart failed: invalid value type -", kind)
	}
	return parts
}

func appendStructPart(parts []string, k string, sv reflect.Value) []string {

	st := sv.Type()
	for i := 0; i < sv.NumField(); i++ {
		sf := st.Field(i)
		tag := sf.Tag.Get("json")
		v := sv.Field(i)
		parts = appendPart(parts, k+"."+tag, v)
	}
	return parts
}

func appendMapPart(parts []string, k string, mv reflect.Value) []string {

	keys := mv.MapKeys()
	for _, key := range keys {
		kind := key.Kind()
		if kind != reflect.String {
			log.Warn("appendMapPart failed: invalid key type -", kind)
			continue
		}
		v := mv.MapIndex(key)
		parts = appendPart(parts, k+"."+key.String(), v)
	}
	return parts
}
