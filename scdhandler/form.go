package scdhandler

import (
	"bytes"
	"fmt"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
)

type Form struct {
	Value url.Values
	File  map[string][]*multipart.FileHeader
}

func (r *Request) FormValues() (*Form, error) {
	ct, params, err := mime.ParseMediaType(r.ContentType)
	if err != nil {
		return nil, fmt.Errorf("scdhandler: unable to parse Content-Type: %v", err)
	}
	switch ct {
	default:
		return nil, fmt.Errorf("scdhandler: unsupported Content-Type %q", ct)
	case "application/x-www-form-urlencoded":
		values, err := url.ParseQuery(string(r.Body))
		if err != nil {
			return nil, err
		}
		return &Form{
			Value: values,
		}, nil
	case "multipart/form-data":
		boundary, ok := params["boundary"]
		if !ok {
			return nil, http.ErrMissingBoundary
		}
		maxFormSize := int64(10 << 30)
		r := multipart.NewReader(bytes.NewBuffer(r.Body), boundary)
		form, err := r.ReadForm(maxFormSize)
		if err != nil {
			return nil, err
		}
		return &Form{
			Value: url.Values(form.Value),
			File:  form.File,
		}, nil
	}
}
