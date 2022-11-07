package erltf

import (
	"bytes"
	"errors"
	"io"
	"reflect"
)

var DefaultBufferSize = 2048

type Encoder interface {
	io.Writer

	Encode(v any) (n int, err error)
}

func NewEncoder(buf []byte) (Encoder, error) {
	if buf == nil {
		buf = make([]byte, DefaultBufferSize)
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

var ErrUnsupportedTypeEncode = errors.New("github.com/Gophigure/erltf: attempt to encode unsupported type (or kind)")

func (e *encoder) Encode(v any) (n int, err error) {
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
