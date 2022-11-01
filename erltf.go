// Package erltf implements the ETF encoding and decoding format, the implemented version can be
// located at [TermFormatVersion].
package erltf

// TermFormatVersion is the ETF version this package implements.
const TermFormatVersion byte = 131

// TermIdentifier is used for identifying the data type of a payload.
type TermIdentifier byte

const (
	NewFloatExt TermIdentifier = iota + 70
	_
	_
	_
	_
	_
	_
	BitBinaryExt
	_
	_
	_
	_
	AtomCacheRef
	_
	_
	_
	_
	_
	_
	_
	_
	_
	_
	_
	_
	_
	_
	SmallIntegerExt
	IntegerExt
	FloatExt
	_
	_
	_
	_
	SmallTupleExt
	LargeTupleExt
	NilExt
	StringExt
	ListExt
	BinaryExt
	SmallBigExt
	LargeBigExt
	_
	_
	_
	_
	MapExt
	_
	AtomUTF8Ext
	SmallAtomUTF8Ext
)
