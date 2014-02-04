/*
Copyright 2014 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package bitstring

import (
	"bytes"
	"io"
	"reflect"
	"testing"
)

type S struct {
	F1 uint8  `bit:"1"`
	_  uint8  `bit:"7"`
	F2 uint16 `bit:"10"`
	_  byte
	_  uint8  `bit:"6"`
	F3 uint32 `bit:"22"`
	_  uint16 `bit:"10"`
	F4 uint64 `bit:"37"`
}

var data = []byte{
	0x80, // 1000,0000
	0xff, // 1111,1111
	0xc0, // 1100,0000
	0xff, // 1111,1111
	0xff, // 1111,1111
	0xfc, // 1111,1100
	0x00, // 0000,0000
	0xff, // 1111,1111
	0xff, // 1111,1111
	0xff, // 1111,1111
	0xff, // 1111,1111
	0xf8, // 1111,1000
}

func TestUnmarshal(t *testing.T) {
	buf := bytes.NewBuffer(data)
	b := NewBuffer(buf)
	out := &S{}
	err := Unmarshal(b, out)
	if err != nil && err != io.EOF {
		t.Error(err)
	}

	want := &S{
		F1: uint8(0x01),                // 0000,0001
		F2: uint16(0x03ff),             // 0000,0011|1111,1111
		F3: uint32(0x003fffff),         // 0000,0000|0011,1111|1111,1111|1111,1111
		F4: uint64(0x0000001fffffffff), // 0000,0000|0000,0000|0000,0000|0001,1111|1111,1111|1111,1111|1111,1111|1111,1111
	}

	if !reflect.DeepEqual(want, out) {
		t.Errorf("want=%#v, out: %#v", want, out)
	}
}
