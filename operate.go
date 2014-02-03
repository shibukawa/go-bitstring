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
package bitarray

import (
	"errors"
	"io"
	"log"
)

const (
	Uint8Size  = uint64(8)
	Uint16Size = uint64(16)
	Uint32Size = uint64(32)
	Uint64Size = uint64(64)
	Int8Size   = uint64(8)
	Int16Size  = uint64(16)
	Int32Size  = uint64(32)
	Int64Size  = uint64(64)
)

// ErrSizeTooLarge is passed to panic if specified size for operations is too large
// for its return type.
var ErrSizeTooLarge = errors.New("bitarray: Specified is too large.")

// A Buffer is a variable-sized buffer of bytes with basic bit extract operations.
type Buffer struct {
	buf    io.ByteReader // contents should be io.ByteReader ready type.
	n      uint8         // index of current bit position in byte segmentation.
	extra  uint8         // extra byte unmanupilated in last operation.
	unread bool          // flag if this buffer is unread or not.
}

func NewBuffer(b io.ByteReader) *Buffer {
	return &Buffer{
		buf:    b,
		unread: true,
	}
}

// PopUint8 extract next `size` bits from Buffer. If buffer reaches tail of buffer,
// it returns bits left in the buffer and io.EOF
func (b *Buffer) PopUint8(size uint64) (uint8, error) {
	if size > Uint8Size {
		return 0, ErrSizeTooLarge
	}

	var bin uint8
	if b.unread || uint64(b.n)+size >= Uint8Size {
		c, err := b.buf.ReadByte()
		if err == io.EOF {
			bin = b.extra >> b.n
			b.n += uint8(size) - uint8(Uint8Size) // Add overflowed bit size
			b.extra = 0x00
			return bin, io.EOF
		}
		if err != nil {
			return 0, err
		}
		n := (uint64(b.n) + size) % Uint8Size
		bin = uint8(c) >> (Uint8Size - n)
		bin += b.extra >> b.n << n
		b.extra = uint8(c) << n
		b.n = uint8(n)
		if b.unread {
			b.unread = false
		}
	} else {
		bin = b.extra >> (Uint8Size - size)
		b.n += uint8(size)
		b.extra = b.extra << size
	}
	return bin, nil
}

// PopUint16 extract next `size` bits from Buffer. If buffer reaches tail of buffer,
// it returns bits left in the buffer and io.EOF
func (b *Buffer) PopUint16(size uint64) (uint16, error) {
	if size > Uint16Size {
		return 0, ErrSizeTooLarge
	}

	if size <= Uint8Size {
		bin, err := b.PopUint8(size)
		return uint16(bin), err
	} else {
		bin, err := b.PopUint8(Uint8Size)
		if err != nil {
			if err == io.EOF {
				b.n += uint8(size - Uint8Size) // Add overflowed bit size
			}
			return uint16(bin), err
		}
		leftSize := size - Uint8Size
		bin16 := uint16(bin) << leftSize

		bin8, err := b.PopUint8(leftSize)
		if err == io.EOF {
			bin16 = bin16 >> b.n
		}
		return bin16 + uint16(bin8), err
	}
}

// PopUint32 extract next `size` bits from Buffer. If buffer reaches tail of buffer,
// it returns bits left in the buffer and io.EOF
func (b *Buffer) PopUint32(size uint64) (uint32, error) {
	if size > Uint32Size {
		return 0, ErrSizeTooLarge
	}

	if size <= Uint8Size {
		bin, err := b.PopUint8(size)
		return uint32(bin), err
	} else if size <= Uint16Size {
		bin, err := b.PopUint16(size)
		return uint32(bin), err
	}

	bin, err := b.PopUint16(Uint16Size)
	if err != nil {
		if err == io.EOF {
			b.n += uint8(size - Uint16Size)
		}
		return uint32(bin), err
	}
	leftSize := size - Uint16Size

	if leftSize <= Uint8Size {
		bin32 := uint32(bin) << leftSize
		bin8, err := b.PopUint8(leftSize)
		if err == io.EOF {
			bin32 = bin32 >> b.n
		}
		return bin32 + uint32(bin8), err
	}

	bin32 := uint32(bin) << leftSize
	bin16, err := b.PopUint16(leftSize)
	if err == io.EOF {
		bin32 = bin32 >> b.n
	}
	return bin32 + uint32(bin16), err
}

// PopUint64 extract next `size` bits from Buffer. If buffer reaches tail of buffer,
// it returns bits left in the buffer and io.EOF
func (b *Buffer) PopUint64(size uint64) (uint64, error) {
	if size > Uint64Size {
		return 0, ErrSizeTooLarge
	}

	if size <= Uint8Size {
		bin, err := b.PopUint8(size)
		return uint64(bin), err
	} else if size <= Uint16Size {
		bin, err := b.PopUint16(size)
		return uint64(bin), err
	} else if size <= Uint32Size {
		bin, err := b.PopUint32(size)
		return uint64(bin), err
	}
	bin, err := b.PopUint32(Uint32Size)
	if err != nil {
		if err == io.EOF {
			b.n += uint8(size - Uint32Size)
		}
		return uint64(bin), err
	}
	leftSize := size - Uint32Size

	if leftSize <= Uint8Size {
		bin64 := uint64(bin) << leftSize
		bin8, err := b.PopUint8(leftSize)
		if err == io.EOF {
			bin64 = bin64 >> b.n
		}
		return bin64 + uint64(bin8), err
	} else if leftSize <= Uint16Size {
		bin64 := uint64(bin) << leftSize
		bin16, err := b.PopUint16(leftSize)
		if err == io.EOF {
			bin64 = bin64 >> b.n
		}
		return bin64 + uint64(bin16), err
	}

	bin64 := uint64(bin) << leftSize
	bin32, err := b.PopUint32(leftSize)
	if err == io.EOF {
		bin64 = bin64 >> b.n
	}
	return bin64 + uint64(bin32), err
}

// PopBytes extract next `size` bytes from Buffer. If buffer reaches tail of buffer,
// it returns bits left in the buffer and io.EOF
func (b *Buffer) PopBytes(size uint64) ([]byte, error) {
	bytes := []byte{}
	log.Printf("b: %#v, start: %#v", b.buf, bytes)
	for i := uint64(0); i < size; i++ {
		byt, err := b.PopUint8(Uint8Size)
		if err != nil {
			bytes = append(bytes, byt)
			return bytes, err
		}
		bytes = append(bytes, byt)
	}
	return bytes, nil
}
