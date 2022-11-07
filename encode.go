package erltf

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"reflect"
)

// DefaultBufferSize is the default size used for a new buffer when creating a new [Encoder] via
// [NewEncoder]. This only works if a nil value is passed or the given byte slice's cap is below
// this value, in which case it creates a new byte slice and copies the memory.
var DefaultBufferSize = 2048

// AlwaysEncodeStringsAsBinary is whether to always encode string values using the [BinaryExt]
// identifier if they are passed to an [Encoder]'s EncodeAsETF method.
var AlwaysEncodeStringsAsBinary = false

// Encoder is an interface that can be implemented for either encoding ETF data with a custom type,
// or to act as a wrapper of an [Encoder] returned from [NewEncoder].
type Encoder interface {
	io.Writer

	EncodeAsETF(v any) (n int, err error)
	EncodeAsBinaryETF(v []byte) (n int, err error)
}

// NewEncoder returns a new [Encoder] value ready for use. Passing a nil value or []byte with a cap
// <= [DefaultBufferSize] will cause this function to create a new []byte with a cap equal to
// [DefaultBufferSize].
func NewEncoder(buf []byte) (Encoder, error) {
	if buf == nil || cap(buf) == 0 {
		buf = make([]byte, DefaultBufferSize)
	} else if cap(buf) < DefaultBufferSize {
		nBuf := make([]byte, DefaultBufferSize)
		copy(nBuf, buf)
		buf = nBuf
	}

	enc := &encoder{bytes.NewBuffer(buf)}
	if err := enc.buf.WriteByte(TermFormatVersion); err != nil {
		return nil, err
	}
	return enc, nil
}

type encoder struct {
	buf *bytes.Buffer
}

func (e *encoder) Write(p []byte) (n int, err error) { return e.EncodeAsBinaryETF(p) }

// Prevent allocating these buffers multiple times as these types *may* be common.
var (
	nilBuf   = []byte{byte(SmallAtomUTF8Ext), 3, 'n', 'i', 'l'}
	trueBuf  = []byte{byte(SmallAtomUTF8Ext), 4, 't', 'r', 'u', 'e'}
	falseBuf = []byte{byte(SmallAtomUTF8Ext), 5, 'f', 'a', 'l', 's', 'e'}
)

// ErrUnsupportedTypeEncode is returned when a value passed to an [Encoder]'s EncodeAsETF function is unsupported.
var ErrUnsupportedTypeEncode = errors.New("github.com/Gophigure/erltf: attempt to encode unsupported type (or kind)")

// ErrStringTooLong is returned when attempting to encode a string that is too large.
var ErrStringTooLong = errors.New("github.com/Gophigure/erltf: attempt to encode string with size larger than the uint32 limit")

// ErrListTooLarge is returned when attempting to encode an array, slice or string that is too large.
var ErrListTooLarge = errors.New("github.com/Gophigure/erltf: attempt to encode array, slice or string with a length larger than the uint32 limit")

// EncodeAsETF is used to write any supported value to the internal buffer.
func (e *encoder) EncodeAsETF(v any) (n int, err error) {
	val := reflect.ValueOf(v)
	// Serialize untyped nil or interface-typed nil values to 'nil', handle potentially invalid
	// kinds as well.
	if !val.IsValid() || val.Kind() == reflect.Invalid {
		return e.buf.Write(nilBuf)
	}

	switch val.Kind() {
	case reflect.Chan,
		reflect.Func,
		reflect.Complex64,
		reflect.Complex128,
		reflect.Interface,
		reflect.Uintptr,
		reflect.UnsafePointer:
		// Panic because a program attempting to serialize these kinds through erltf should be
		// considered a mistake and must be handled ASAP.
		panic("github.com/Gophigure/erltf: attempt to serialize invalid type " + val.Type().Name())
	case reflect.Bool:
		if v.(bool) {
			return e.buf.Write(trueBuf)
		}
		return e.buf.Write(falseBuf)
	case reflect.Uint8:
		return e.buf.Write([]byte{byte(SmallIntegerExt), byte(val.Uint())})
	case reflect.Int32:
		buf := make([]byte, 1, 5)
		buf[0] = byte(IntegerExt)

		return e.buf.Write(binary.BigEndian.AppendUint32(buf, uint32(val.Uint())))
	case reflect.Uint32:
		// TODO: Consider implementing this with a loop ourselves to not waste encoding unused large
		// bits.

		buf := make([]byte, 3, 7)
		buf[0] = byte(SmallBigExt)
		buf[1] = 4
		buf[2] = 0

		return e.buf.Write(binary.LittleEndian.AppendUint32(buf, uint32(val.Uint())))
	case reflect.Float32, reflect.Float64:
		buf := make([]byte, 1, 9)
		buf[0] = byte(NewFloatExt)

		return e.buf.Write(binary.BigEndian.AppendUint64(
			buf,
			math.Float64bits(val.Float()),
		))
	case reflect.String:
		str := val.String()
		if AlwaysEncodeStringsAsBinary {
			return e.EncodeAsBinaryETF([]byte(str))
		}

		length := len(str)
		if length < math.MaxUint16 {
			buf := make([]byte, 1, 3+length)
			buf[0] = byte(StringExt)
			buf = binary.LittleEndian.AppendUint16(buf, uint16(length))

			return e.buf.Write(append(buf, str...))
		}
		fallthrough
	case reflect.Array, reflect.Slice:
		length := val.Len()
		if length > math.MaxUint32 {
			return 0, ErrListTooLarge
		}

		buf := make([]byte, 1, 5)
		buf[0] = byte(ListExt)
		buf = binary.BigEndian.AppendUint32(buf, uint32((val.Len())))

		n, err = e.buf.Write(buf)
		if err != nil {
			return
		}

		for i := 0; i < length; i++ {
			elementLength, elementErr := e.EncodeAsETF(val.Index(i))
			if elementErr != nil {
				return n, elementErr
			}
			n += elementLength
		}

		return e.buf.Write(nilBuf)
	case reflect.Struct:
	case reflect.Map:
	}

	return 0, ErrUnsupportedTypeEncode
}

// EncodeAsBinaryETF encodes a slice of bytes as binary data using the [BinaryExt] identifier. This
// is useful for 'forcing' the encoding of a value as binary data.
func (e *encoder) EncodeAsBinaryETF(v []byte) (n int, err error) {
	length := len(v)
	if length > math.MaxUint32 {
		return 0, ErrStringTooLong
	}

	buf := make([]byte, 1, 5+length)
	buf[0] = byte(BinaryExt)
	buf = binary.BigEndian.AppendUint32(buf, uint32(length))

	return e.buf.Write(append(buf, v...))
}
