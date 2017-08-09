// Copyright 2017 The Solid Core Data Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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
