// Copyright 2017 The Solid Core Data Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

func NewKeyValueList(from map[string][]string) *KeyValueList {
	list := &KeyValueList{
		Values: make(map[string]*RepeatedString),
	}
	for key, value := range from {
		list.SetSlice(key, value)
	}
	return list
}

func (h *KeyValueList) Add(key, value string) {
	list, ok := h.Values[key]
	if !ok {
		list = &RepeatedString{
			Value: make([]string, 0, 1),
		}
		h.Values[key] = list
	}
	list.Value = append(list.Value, value)
}

func (h *KeyValueList) Set(key, value string) {
	h.Values[key] = &RepeatedString{
		Value: []string{value},
	}
}

func (h *KeyValueList) SetSlice(key string, value []string) {
	h.Values[key] = &RepeatedString{
		Value: value,
	}
}

func (h *KeyValueList) Del(key string) {
	delete(h.Values, key)
}

func (h *KeyValueList) Get(key string) string {
	list, ok := h.Values[key]
	if !ok {
		return ""
	}
	if len(list.Value) == 0 {
		return ""
	}
	return list.Value[0]
}
