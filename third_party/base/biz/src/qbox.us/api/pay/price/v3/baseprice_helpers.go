package v3

import (
	"encoding/json"
	"fmt"
	"reflect"
)

func (l ModelResourceGroupList) UnmarshalGroup(idx int) (interface{}, error) {
	rawGroup := l.Groups[idx]

	// Unmarshal content
	groupPtr := NewResourceGroupForType(l.Type)
	if groupPtr == nil {
		return nil, fmt.Errorf("unknown group type: %d", l.Type)
	}

	err := json.Unmarshal(rawGroup, groupPtr)
	if err != nil {
		return nil, err
	}

	// Dereference pointer
	group := reflect.ValueOf(groupPtr).Elem().Interface()

	return group, nil
}
