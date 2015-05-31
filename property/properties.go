package property

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/gsdocker/gserrors"
)

// Errors .
var (
	ErrLoad     = errors.New("load package error")
	ErrNotFound = errors.New("property not found")
)

// Properties .
type Properties map[string]interface{}

// Expand rewrites content to replace ${k} with properties[k] for each key k in match.
func (properties Properties) Expand(content string) string {
	for k, v := range properties {

		if stringer, ok := v.(fmt.Stringer); ok {
			fmt.Println(stringer.String())
			content = strings.Replace(content, "${"+k+"}", stringer.String(), -1)
		} else {
			content = strings.Replace(content, "${"+k+"}", fmt.Sprintf("%v", v), -1)
		}

	}
	return content
}

// Query query property value
func (properties Properties) Query(name string, val interface{}) error {
	if property, ok := properties[name]; ok {
		content, err := json.Marshal(property)

		if err != nil {
			return err
		}

		return json.Unmarshal(content, val)
	}

	return ErrNotFound
}

// NotFound property not found
func NotFound(err error) bool {
	for {
		if gserror, ok := err.(gserrors.GSError); ok {
			err = gserror.Origin()
			continue
		}

		break
	}

	if err == ErrNotFound {
		return true
	}

	return false
}
