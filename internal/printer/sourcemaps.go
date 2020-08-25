package printer

var (
	base64Forward = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/")

	base64Reverse = [127]byte{
		65:  0,
		66:  1,
		67:  2,
		68:  3,
		69:  4,
		70:  5,
		71:  6,
		72:  7,
		73:  8,
		74:  9,
		75:  10,
		76:  11,
		77:  12,
		78:  13,
		79:  14,
		80:  15,
		81:  16,
		82:  17,
		83:  18,
		84:  19,
		85:  20,
		86:  21,
		87:  22,
		88:  23,
		89:  24,
		90:  25,
		97:  26,
		98:  27,
		99:  28,
		100: 29,
		101: 30,
		102: 31,
		103: 32,
		104: 33,
		105: 34,
		106: 35,
		107: 36,
		108: 37,
		109: 38,
		110: 39,
		111: 40,
		112: 41,
		113: 42,
		114: 43,
		115: 44,
		116: 45,
		117: 46,
		118: 47,
		119: 48,
		120: 49,
		121: 50,
		122: 51,
		48:  52,
		49:  53,
		50:  54,
		51:  55,
		52:  56,
		53:  57,
		54:  58,
		55:  59,
		56:  60,
		57:  61,
		43:  62,
		47:  63,
	}
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

	if in>>5 == 0 {
		offset := in & 0b11111
		return []byte(base64Forward[offset : offset+1])
	}

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

		rv = append(rv, base64Forward[cur])

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

		cur := int32(base64Reverse[b])

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
