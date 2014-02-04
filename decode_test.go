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

// TestUnmarshal: Case 1) Extract bit array from a byte array.
// No operation of fetching byte array from arbitrary bit position.
func TestUnmarshalCase1(t *testing.T) {
	type S struct {
		F1 uint8  `bits:"1"`
		_  uint8  `bits:"7"`
		F2 uint16 `bits:"10"`
		_  byte
		_  uint8  `bits:"6"`
		F3 uint32 `bits:"22"`
		_  uint16 `bits:"10"`
		F4 uint64 `bits:"37"`
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

// TestUnmarshal: Case 2) Extract bit array from a byte array.
// No operation of fetching byte array from arbitrary bit position.
func TestUnmarsharlCase2(t *testing.T) {
	type S struct {
		F1 uint8 `bits:"2"`
		_  uint8
		F2 []byte `binary:"12"`
	}

	var data = []byte{
		0xd2, // 11|01,0010
		0x19, // 00|01,1001
		0x5b, // 01|01,1011
		0x1b, // 00|01,1011
		0x1b, // 00|01,1011
		0xcb, // 11|00,1011
		0x08, // 00|00,1000
		0x1d, // 00|01,1101
		0xdb, // 11|01,1011
		0xdc, // 11|01,1100
		0x9b, // 10|01,1011
		0x19, // 00|01,1001
		0x3f, // 00|11,1111
	}

	buf := bytes.NewBuffer(data)
	b := NewBuffer(buf)
	out := &S{}
	err := Unmarshal(b, out)
	if err != nil && err != io.EOF {
		t.Error(err)
	}

	want := &S{
		F1: uint8(0x03), // 0000,0011
		F2: []byte("Hello, world"),
	}
	if !reflect.DeepEqual(want, out) {
		t.Errorf("want=%#v, out: %#v", want, out)
	}
}
