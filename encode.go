package erltf

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"reflect"
)

// EncodeETF can be implemented by custom types to overwrite default behavior when encoding into
// ETF data.
//
// This is not yet respected as there is no way to currently validate the output.
type EncodeETF interface {
	EncodeToETF() ([]byte, error)
}

// DefaultBufferSize is the default size used for a new buffer when creating a new [Encoder] via
// [NewEncoder]. This only works if a nil value is passed or the given byte slice's cap is below
// this value, in which case it creates a new byte slice and copies the memory.
var DefaultBufferSize = 1024 * 1024

// DefaultEncodeRecursionDepth is the default maximum depth an encoder should travel before
// returning an error.
var DefaultEncodeRecursionDepth = 256

// AlwaysEncodeStringsAsBinary is whether to always encode string values using the [BinaryExt]
// identifier if they are passed to an [Encoder]'s EncodeAsETF method.
var AlwaysEncodeStringsAsBinary = true

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

// ErrUnsupportedTypeEncode is returned when a value passed to an [Encoder]'s EncodeAsETF function
// is unsupported.
var ErrUnsupportedTypeEncode = errors.New("github.com/Gophigure/erltf: attempt to encode unsupported type (or kind)")

// ErrStringTooLong is returned when attempting to encode a string that is too large.
var ErrStringTooLong = errors.New("github.com/Gophigure/erltf: attempt to encode string with size larger than the uint32 limit")

// ErrListTooLarge is returned when attempting to encode an array, slice or string that is too large.
var ErrListTooLarge = errors.New("github.com/Gophigure/erltf: attempt to encode array, slice or string with a length larger than the uint32 limit")

// ErrEncodeRecursionDepthExceeded is returned when an [Encoder] attempts to encode a value past the
// maximum recursion depth, it is expected that other implementations of the [Encoder] interface
// implement this behavior for consistency.
var ErrEncodeRecursionDepthExceeded = errors.New("github.com/Gophigure/erltf: maximum recursion depth for encoding exceeded")

// ErrTooManyPairs is returned if a map has more pairs than the uint32 limit.
var ErrTooManyPairs = errors.New("github.com/Gophigure/erltf: map has to many key -> value pairs")

// EncodeAsETF is used to write any supported value to the internal buffer. Passing a pointer or
// interface will effectively reduce the nest recursion depth by 1.
func (e *encoder) EncodeAsETF(v any) (n int, err error) { return e.encode(v, DefaultBufferSize) }

func (e *encoder) encode(v any, nest int) (n int, err error) {
	val := reflect.ValueOf(v)
	// Serialize untyped nil or interface-typed nil values to 'nil', handle potentially invalid
	// kinds as well.
	if !val.IsValid() || val.Kind() == reflect.Invalid {
		return e.buf.Write(nilBuf)
	}

	switch val.Kind() {
	// TODO: Handle types that implement EncodeETF.
	case
		reflect.Chan,
		reflect.Func,
		reflect.Complex64,
		reflect.Complex128,
		reflect.Uintptr,
		reflect.UnsafePointer:
		// Panic because a program attempting to serialize these kinds through erltf should be
		// considered a mistake and must be handled ASAP.
		panic("github.com/Gophigure/erltf: attempt to serialize invalid type " + val.Type().Name())
	case reflect.Pointer, reflect.Interface:
		// This has the benefit of serializing nil values to 'nil'
		return e.encode(reflect.ValueOf(v).Elem().Interface(), nest-1)
	case reflect.Bool:
		if v.(bool) {
			return e.buf.Write(trueBuf)
		}
		return e.buf.Write(falseBuf)
	case reflect.Uint8:
		return e.buf.Write([]byte{byte(SmallIntegerExt), byte(val.Uint())})
	case reflect.Int16, reflect.Int32, reflect.Int64:
		buf := make([]byte, 3, 3+val.Type().Size())
		if cap(buf) > 7 {
			buf[0] = byte(LargeBigExt)
		} else {
			buf[0] = byte(SmallBigExt)
		}

		d := val.Int()
		if d >= 0 {
			buf[2] = 0
		} else {
			d = -d
			buf[2] = 1
		}

		enc := byte(0)
		for i := d; i > 0; enc++ {
			buf[3+enc] = byte(i & 0xFF)
			i >>= 8
		}
		buf[1] = enc

		if buf[0] == byte(SmallBigExt) {
			return e.buf.Write(binary.LittleEndian.AppendUint32(buf, uint32(d)))
		}
		return e.buf.Write(binary.LittleEndian.AppendUint64(buf, uint64(d)))
	case reflect.Uint16, reflect.Uint32, reflect.Uint64:
		buf := make([]byte, 3, 3+val.Type().Size())
		if cap(buf) > 7 {
			buf[0] = byte(LargeBigExt)
		} else {
			buf[0] = byte(SmallBigExt)
		}

		d, enc := val.Uint(), 0
		for i := d; i > 0; enc++ {
			buf[3+enc] = byte(i & 0xFF)
			i >>= 8
		}
		buf[1] = byte(enc)
		buf[2] = 0

		if buf[0] == byte(SmallBigExt) {
			return e.buf.Write(binary.LittleEndian.AppendUint32(buf, uint32(d)))
		}
		return e.buf.Write(binary.LittleEndian.AppendUint64(buf, d))
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
			elementLength, elementErr := e.encode(val.Index(i), nest-1)
			if elementErr != nil {
				return n, elementErr
			}
			n += elementLength
		}

		return e.buf.Write(nilBuf)
	case reflect.Struct:
		length := val.NumField()
		if length >= math.MaxInt32 {
			// Compile-time issue, should not be handled at during runtime.
			panic("github.com/Gophigure/erltf: struct has too many fields")
		}

		typ := val.Type()
		mapped := make(map[string]any, length)
		for i := 0; i < length; i++ {
			fieldTyp := typ.Field(i)
			tag := fieldTyp.Tag.Get("erltf")
			if tag == "" {
				tag = fieldTyp.Name
			} else if tag == "-" {
				continue
			}

			mapped[tag] = val.Field(i).Interface()
		}
		val = reflect.ValueOf(mapped)
		fallthrough
	case reflect.Map:
		if val.Type().Key().Kind() != reflect.String {
			// Compile-time issue, should not be handled during runtime.
			panic("github.com/Gophigure/erltf: map's key type must be of kind string")
		}

		length := val.Len()
		if length > math.MaxUint32 {
			return 0, ErrTooManyPairs
		}

		buf := make([]byte, 1, 5)
		buf[0] = byte(MapExt)
		n, err = e.buf.Write(binary.BigEndian.AppendUint32(buf, uint32(length)))
		if err != nil {
			return
		}

		iter := val.MapRange()
		for iter.Next() {
			k, v := iter.Key(), iter.Value()
			keyLen, keyErr := e.encode(k.Interface(), nest-1)
			if keyErr != nil {
				return n, keyErr
			}
			n += keyLen

			valLen, valErr := e.encode(v.Interface(), nest-1)
			if valErr != nil {
				return n, valErr
			}
			n += valLen
		}

		return n, nil
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
