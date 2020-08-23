package printer

import (
	"strings"
)

const (
	base64 = string("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/")
)

// VLQEncode encodes a signed int into base64 byte slice.
// According to https://sourcemaps.info/spec.html, 32-bit
// is long enough for sourcemapping.
func VLQEncode(in int32) []byte {
	// Move the sign bit to the least significant bit and
	// represent the magnitude as if positive. This lets us
	// treat at sextets the same instead of special casing the
	// sign bit.
	if in < 0 {
		// For negative, this means we undo the 2s complement.
		in = (-in << 1) + 1
	} else {
		in <<= 1
	}

	// XXX: allocation?
	var rv []byte
	for {
		// Content bits.
		cur := in & 0b11111

		// Note: signed shift is ok because we've already switch over
		// to positive representation.
		in >>= 5

		if in != 0 {
			// Continuation bit.
			cur |= 0b100000
		}

		rv = append(rv, base64[cur])

		if in <= 0 {
			break
		}
	}

	return rv
}

// VLQDecode decodes the input slice into a signed bit.
// It is complementary to VLQEncode. It returns the read
// length because the ending is unknown to the caller
// before decoding.
func VLQDecode(in []byte) (value int32, len int32) {
	var rv int32
	var shift int32
	var read int32

	for _, b := range in {
		read++

		cur := int32(strings.IndexByte(base64, b))

		rv += (cur & 0b11111) << shift

		shift += 5

		if cur&0b100000 == 0 {
			break
		}
	}

	if rv&1 == 1 {
		return -((rv - 1) >> 1), read
	}

	return rv >> 1, read
}
