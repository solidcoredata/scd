package scdhandler

import (
	"bytes"
)

type ErrorList []error

func (list ErrorList) Error() string {
	buf := bytes.Buffer{}
	for i, item := range list {
		if i != 0 {
			buf.WriteString("\n")
		}
		buf.WriteString(item.Error())
	}
	return buf.String()
}
