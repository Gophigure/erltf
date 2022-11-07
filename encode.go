package erltf

import (
	"bytes"
	"errors"
	"io"
	"reflect"
)

// DefaultBufferSize is the default size used for a new buffer when creating a new [Encoder] via
// [NewEncoder]. This only works if a nil value is passed or the given byte slice's cap is below
// this value, in which case it creates a new byte slice and copies the memory.
var DefaultBufferSize = 2048

// Encoder is an interface that can be implemented for either encoding ETF data with a custom type,
// or to act as a wrapper of an [Encoder] returned from [NewEncoder].
type Encoder interface {
	io.Writer

	EncodeAsETF(v any) (n int, err error)
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

func (e *encoder) Write(p []byte) (n int, err error) {
	panic("github.com/Gophigure/erltf: encoder.Write is not yet implemented!!!")
}

// Prevent allocating these buffers multiple times as these types *may* be common.
var (
	nilBuf   = []byte{byte(SmallAtomUTF8Ext), 3, 'n', 'i', 'l'}
	trueBuf  = []byte{byte(SmallAtomUTF8Ext), 4, 't', 'r', 'u', 'e'}
	falseBuf = []byte{byte(SmallAtomUTF8Ext), 5, 'f', 'a', 'l', 's', 'e'}
)

// ErrUnsupportedTypeEncode is returned when a value passed to an [Encoder]'s EncodeAsETF function is unsupported.
var ErrUnsupportedTypeEncode = errors.New("github.com/Gophigure/erltf: attempt to encode unsupported type (or kind)")

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
	default:
		return 0, ErrUnsupportedTypeEncode
	}
}
