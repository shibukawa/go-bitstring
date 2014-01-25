package bitarray

import (
	_ "bytes"
	"errors"
	"io"
)

const (
	Uint8Size  = uint8(8)
	Uint16Size = uint8(16)
	Uint32Size = uint8(32)
	Uint64Size = uint8(64)
	Int8Size   = uint8(8)
	Int16Size  = uint8(16)
	Int32Size  = uint8(32)
	Int64Size  = uint8(64)
)

var ErrSizeTooLarge = errors.New("bitarray: Specified is too large.")

type Buffer struct {
	buf   io.ByteReader
	n     uint8 // Index of current bit position in byte segmentation.
	extra uint8 // Extra byte unmanupilated in last operation.
}

func NewBuffer(b io.ByteReader) *Buffer {
	return &Buffer{
		buf: b,
	}
}

// PopUint8 extract first `size` bits from Buffer. If buffer reaches tail of buffer,
// it returns bits left in the buffer and io.EOF
func (b *Buffer) PopUint8(size uint8) (uint8, error) {
	if size > Uint8Size {
		return 0, ErrSizeTooLarge
	}

	var bin uint8
	if b.n == 0 {
		c, err := b.buf.ReadByte()
		if err != nil {
			return 0, err
		}
		bin = uint8(c) >> (Uint8Size - size)
		b.n += size
		b.extra = uint8(c) << size
	} else {
		if b.n+size < Uint8Size {
			bin = b.extra >> (Uint8Size - size)
			b.n += size
			b.extra = b.extra << size
		} else {
			c, err := b.buf.ReadByte()
			if err == io.EOF {
				bin = b.extra >> b.n
				b.n = 0
				return bin, io.EOF
			}
			if err != nil {
				return 0, err
			}
			n := (b.n + size) % Uint8Size
			bin = uint8(c) >> (Uint8Size - n)
			bin += b.extra >> (size - n)
			b.extra = uint8(c) << n
			b.n = n
		}
	}
	return bin, nil
}
