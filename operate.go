package bitarray

import (
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

// PopUint8 extract first `size` bits from Buffer. If buffer reaches tail of buffer,
// it returns bits left in the buffer and io.EOF
func (b *Buffer) PopUint8(size uint8) (uint8, error) {
	if size > Uint8Size {
		return 0, ErrSizeTooLarge
	}

	var bin uint8
	if b.unread || b.n+size >= Uint8Size {
		c, err := b.buf.ReadByte()
		if err == io.EOF {
			bin = b.extra >> b.n
			b.n += size - Uint8Size // Add overflowed bit size
			b.extra = 0x00
			return bin, io.EOF
		}
		if err != nil {
			return 0, err
		}
		n := (b.n + size) % Uint8Size
		bin = uint8(c) >> (Uint8Size - n)
		bin += b.extra >> b.n << n
		b.extra = uint8(c) << n
		b.n = n
		if b.unread {
			b.unread = false
		}
	} else {
		bin = b.extra >> (Uint8Size - size)
		b.n += size
		b.extra = b.extra << size
	}
	return bin, nil
}

// PopUint16 extract first `size` bits from Buffer. If buffer reaches tail of buffer,
// it returns bits left in the buffer and io.EOF
func (b *Buffer) PopUint16(size uint8) (uint16, error) {
	if size > Uint16Size {
		return 0, ErrSizeTooLarge
	}

	if size <= Uint8Size {
		bin, err := b.PopUint8(size)
		return uint16(bin), err
	} else {
		bin, err := b.PopUint8(Uint8Size)
		if err == io.EOF {
			b.n += size - Uint8Size // Add overflowed bit size
			return uint16(bin), err
		}
		if err != nil {
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

// PopUint32 extract first `size` bits from Buffer. If buffer reaches tail of buffer,
// it returns bits left in the buffer and io.EOF
// TODO(ymotongpoo): How about calling PopUint8 4 times?
func (b *Buffer) PopUint32(size uint8) (uint32, error) {
	if size > Uint32Size {
		return 0, ErrSizeTooLarge
	}

	var bin uint32
	switch {
	case size <= Uint8Size:
		bin, err := b.PopUint8(size)
		return uint32(bin), err
	case size <= Uint16Size:
		bin, err := b.PopUint16(size)
		return uint32(bin), err
	}
	return bin, nil
}

// PopUint64 extract first `size` bits from Buffer. If buffer reaches tail of buffer,
// it returns bits left in the buffer and io.EOF
// TODO(ymotongpoo): How about calling PopUint8 8 times?
func (b *Buffer) PopUint64(size uint8) (uint64, error) {
	if size > Uint32Size {
		return 0, ErrSizeTooLarge
	}

	var bin uint64
	switch {
	case size <= Uint8Size:
		bin, err := b.PopUint8(size)
		return uint64(bin), err
	case size <= Uint16Size:
		bin, err := b.PopUint16(size)
		return uint64(bin), err
	case size <= Uint32Size:
		bin, err := b.PopUint32(size)
		return uint64(bin), err
	}
	return bin, nil
}
