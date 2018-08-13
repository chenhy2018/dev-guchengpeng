package config

import (
	"reflect"

	"github.com/teapots/teapot"
)

func Decode(config *teapot.Config, env interface{}) {
	val := reflect.ValueOf(env)

	if val.Kind() != reflect.Ptr {
		panic("env instance must be ptr struct")
	}
	val = val.Elem()
	if val.Kind() != reflect.Struct {
		panic("env instance must be ptr struct")
	}

	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		fVal := val.Field(i)
		fType := typ.Field(i)

		name := fType.Tag.Get("conf")
		if len(name) == 0 || name == "-" {
			continue
		}

		if fVal.Kind() == reflect.Struct {
			parseSection(config, fVal, name)
			continue
		}

		config.Bind(fVal.Addr().Interface(), name)
	}
}

func parseSection(config *teapot.Config, val reflect.Value, sec string) {
	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		fVal := val.Field(i)
		fType := typ.Field(i)

		name := fType.Tag.Get("conf")
		if len(name) == 0 || name == "-" {
			continue
		}

		config.Bind(fVal.Addr().Interface(), sec+"::"+name)
	}
}