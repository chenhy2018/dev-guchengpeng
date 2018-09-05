package line

import (
	"bytes"
	// "fmt"
	"strconv"
	"time"
	// "io"
)

var (
	escapeCodes = map[byte][]byte{
		',': []byte(`\,`),
		'"': []byte(`\"`),
		' ': []byte(`\ `),
		'=': []byte(`\=`),
	}

	escapeCodeK = []byte{',', '"', ' ', '='}
	escapeCodeV = [][]byte{
		[]byte(`\,`),
		[]byte(`\"`),
		[]byte(`\ `),
		[]byte(`\=`),
	}

	escapeCodesStr = map[string]string{}

	codeArray  = []byte{',', '"', ' ', '='}
	code2Array = [][]byte{
		[]byte(`\,`),
		[]byte(`\"`),
		[]byte(`\ `),
		[]byte(`\=`),
	}

	measurementEscapeCodes = map[byte][]byte{
		',': []byte(`\,`),
		' ': []byte(`\ `),
	}
	measurementEscapeCodeK = []byte{',', ' '}
	measurementEscapeCodeV = [][]byte{
		[]byte(`\,`),
		[]byte(`\ `),
	}

	tagEscapeCodes = map[byte][]byte{
		',': []byte(`\,`),
		' ': []byte(`\ `),
		'=': []byte(`\=`),
	}
	tagEscapeCodeK = []byte{',', ' ', '='}
	tagEscapeCodeV = [][]byte{
		[]byte(`\,`),
		[]byte(`\ `),
		[]byte(`\=`),
	}
)

func init() {
	for k, v := range escapeCodes {
		escapeCodesStr[string(k)] = string(v)
	}
}

func MarshalMap(
	measurement string, tim time.Time,
	tags map[string]string,
	fields map[string]interface{}) []byte {

	buf := bytes.NewBuffer(nil)
	MarshalMapWrite(buf, measurement, tim, tags, fields)
	return buf.Bytes()
}

func MarshalMapWrite(w *bytes.Buffer,
	measurement string, tim time.Time,
	tags map[string]string,
	fields map[string]interface{}) {

	w.Write(escape([]byte(measurement), measurementEscapeCodeK, measurementEscapeCodeV))
	encodeTagMapWrite(w, tags)
	w.WriteByte(' ')
	encodeFieldMapWrite(w, fields)
	if !tim.IsZero() {
		w.WriteByte(' ')
		w.Write([]byte(strconv.FormatInt(tim.UnixNano(), 10)))
	}
}

func MarshalArray(
	measurement string, tim time.Time,
	tagK, tagV []string,
	fieldK []string, fieldV []interface{}) []byte {

	buf := bytes.NewBuffer(nil)
	MarshalArrayWrite(buf, measurement, tim, tagK, tagV, fieldK, fieldV)
	return buf.Bytes()
}

func MarshalArrayWrite(w *bytes.Buffer,
	measurement string, tim time.Time,
	tagK, tagV []string,
	fieldK []string, fieldV []interface{}) {

	w.Write(escape([]byte(measurement), measurementEscapeCodeK, measurementEscapeCodeV))
	encodeTagArrayWrite(w, tagK, tagV)
	w.WriteByte(' ')
	encodeFieldArrayWrite(w, fieldK, fieldV)
	if !tim.IsZero() {
		w.WriteByte(' ')
		w.Write([]byte(strconv.FormatInt(tim.UnixNano(), 10)))
	}
}

func encodeTagMap(tags map[string]string) []byte {
	buf := bytes.NewBuffer(nil)
	encodeTagMapWrite(buf, tags)
	return buf.Bytes()
}

func encodeTagMapWrite(w *bytes.Buffer, tags map[string]string) {
	for key, value := range tags {
		w.WriteByte(',')
		w.Write(escape([]byte(key), tagEscapeCodeK, tagEscapeCodeV))
		w.WriteByte('=')
		w.Write(escape([]byte(value), escapeCodeK, escapeCodeV))
	}
}

func encodeTagArray(keys, values []string) []byte {
	buf := bytes.NewBuffer(nil)
	encodeTagArrayWrite(buf, keys, values)
	return buf.Bytes()
}

func encodeTagArrayWrite(w *bytes.Buffer, keys, values []string) {
	for i, key := range keys {
		value := values[i]
		w.WriteByte(',')
		w.Write(escape([]byte(key), tagEscapeCodeK, tagEscapeCodeV))
		w.WriteByte('=')
		w.Write(escape([]byte(value), escapeCodeK, escapeCodeV))
	}
}

func encodeFiledMap(fields map[string]interface{}) []byte {
	buf := bytes.NewBuffer(nil)
	encodeFieldMapWrite(buf, fields)
	return buf.Bytes()
}

func encodeFieldMapWrite(w *bytes.Buffer, fields map[string]interface{}) {
	first := true
	for key, value := range fields {
		if !first {
			w.WriteByte(',')
		} else {
			first = false
		}
		w.Write(escape([]byte(key), tagEscapeCodeK, tagEscapeCodeV))
		w.WriteByte('=')
		if bs, isString := encodeFieldValue(value); isString {
			w.WriteByte('"')
			w.Write(bs)
			w.WriteByte('"')
		} else {
			w.Write(bs)
		}
	}
}

func encodeFieldArray(keys []string, values []interface{}) []byte {
	buf := bytes.NewBuffer(nil)
	encodeFieldArrayWrite(buf, keys, values)
	return buf.Bytes()
}

func encodeFieldArrayWrite(w *bytes.Buffer, keys []string, values []interface{}) {
	first := true
	for i, key := range keys {
		if !first {
			w.WriteByte(',')
		} else {
			first = false
		}
		value := values[i]
		w.Write(escape([]byte(key), tagEscapeCodeK, tagEscapeCodeV))
		w.WriteByte('=')
		if bs, isString := encodeFieldValue(value); isString {
			w.WriteByte('"')
			w.Write(bs)
			w.WriteByte('"')
		} else {
			w.Write(bs)
		}
	}
}

func encodeFieldValue(v interface{}) ([]byte, bool) { // bool: isString
	switch t := v.(type) {
	case int8:
		return []byte(strconv.FormatInt(int64(t), 10)), false
	case int16:
		return []byte(strconv.FormatInt(int64(t), 10)), false
	case int:
		return []byte(strconv.FormatInt(int64(t), 10)), false
	case int32:
		return []byte(strconv.FormatInt(int64(t), 10)), false
	case int64:
		return []byte(strconv.FormatInt(t, 10)), false
	case uint8:
		return []byte(strconv.FormatUint(uint64(t), 10)), false
	case uint16:
		return []byte(strconv.FormatUint(uint64(t), 10)), false
	case uint:
		return []byte(strconv.FormatUint(uint64(t), 10)), false
	case uint32:
		return []byte(strconv.FormatUint(uint64(t), 10)), false
	case uint64:
		return []byte(strconv.FormatUint(t, 10)), false
	case float32:
		return []byte(strconv.FormatFloat(float64(t), 'f', -1, 64)), false
	case float64:
		return []byte(strconv.FormatFloat(t, 'f', -1, 64)), false
	case bool: // TODO
		return []byte(strconv.FormatBool(t)), false
	case []byte:
		return escape(t, escapeCodeK, escapeCodeV), true
	case string:
		return escape([]byte(t), escapeCodeK, escapeCodeV), true
	case nil:
		// skip
	default:
		// skip
	}
	return []byte{}, false
}

func escape(in []byte, codeK []byte, codeV [][]byte) []byte {
	t := make([]byte, 0, len(in))
	start := 0
	for i, b := range in {
		j := 0
		for ; j < len(codeK); j++ {
			if codeK[j] == b {
				break
			}
		}
		if j >= len(codeK) {
			continue
		}
		if i > start {
			t = append(t, in[start:i]...)
		}
		t = append(t, codeV[j]...)
		start = i + 1
	}
	t = append(t, in[start:]...)
	return t
}
