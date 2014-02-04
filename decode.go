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
	"errors"
	"io"
	"log"
	"reflect"
	"strconv"
)

var (
	ErrFieldSizeTooLarge = errors.New("bitarray: Specified bit size is too large for field")
)

// Decoder reads and decodes bit array objects from an input stream
type Decoder struct {
	buf *Buffer
}

// NewDecoder returns a new Decoder that reads from b.
func NewDecoder(b *Buffer) *Decoder {
	return &Decoder{
		buf: b,
	}
}

func Unmarshal(b *Buffer, v interface{}) error {
	d := NewDecoder(b)
	return d.Unmarshal(v)
}

func (d *Decoder) Unmarshal(v interface{}) error {
	kind := reflect.ValueOf(v).Kind()
	if kind != reflect.Ptr {
		return errors.New("bitarray.Unmarshal: invalid type " + kind.String())
	}

	st := reflect.ValueOf(v).Elem()
	typ := st.Type()
	if typ.Kind() != reflect.Struct {
		return errors.New("bitarray.Unmarshal: invalid type " + st.String())

	}

	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)
		sizeStr := typ.Field(i).Tag.Get("bit")
		if len(sizeStr) == 0 {
			continue
		}
		size, err := strconv.ParseUint(sizeStr, 0, 64)
		if err != nil {
			return err
		}

		//log.Printf("%v", typ.Field(i).Name)

		// TODO(ymotongpoo): Require refactoring
		switch field.Kind() {
		case reflect.Uint8:
			if size > Uint8Size {
				return ErrFieldSizeTooLarge
			}
			bit, err := d.buf.PopUint8(size)
			if err != nil && err != io.EOF {
				return err
			}
			if typ.Field(i).Name == "_" {
				continue
			}
			st.Field(i).SetUint(uint64(bit))
		case reflect.Uint16:
			if size > Uint16Size {
				return ErrFieldSizeTooLarge
			}
			bit, err := d.buf.PopUint16(size)
			if err != nil && err != io.EOF {
				return err
			}
			if typ.Field(i).Name == "_" {
				continue
			}
			st.Field(i).SetUint(uint64(bit))
		case reflect.Uint32:
			if size > Uint32Size {
				return ErrFieldSizeTooLarge
			}
			bit, err := d.buf.PopUint32(size)
			if err != nil && err != io.EOF {
				return err
			}
			if typ.Field(i).Name == "_" {
				continue
			}
			st.Field(i).SetUint(uint64(bit))
		case reflect.Uint64:
			if size > Uint64Size {
				return ErrFieldSizeTooLarge
			}
			bit, err := d.buf.PopUint64(size)
			if err != nil && err != io.EOF {
				return err
			}
			if typ.Field(i).Name == "_" {
				continue
			}
			st.Field(i).SetUint(bit)
		default:
			log.Printf("%s: Failed to get type (%s)", typ.Field(i).Name, typ.Kind())
			// TODO(ymotongpoo): Add exceptional process
			return nil
		}
	}
	return nil
}
