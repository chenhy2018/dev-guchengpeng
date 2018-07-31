package model

import "time"

type Point struct {
	Measurement string
	Time        time.Time
	TagK        []string
	TagV        []string
	FieldK      []string
	FieldV      []interface{}
}

type PointInM struct {
	Measurement string
	Time        time.Time
	Tags        map[string]string
	Fields      map[string]interface{}
}
