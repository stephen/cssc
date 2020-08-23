package printer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVLQEncode(t *testing.T) {
	assert.Equal(t, []byte("A"), VLQEncode(0))
	assert.Equal(t, []byte("2H"), VLQEncode(123))
	assert.Equal(t, []byte("gkxH"), VLQEncode(123456))
	assert.Equal(t, []byte("qxmvrH"), VLQEncode(123456789))
}

func TestVLQDecode(t *testing.T) {
	val, len := VLQDecode([]byte("A"))
	assert.Equal(t, int32(0), val)
	assert.Equal(t, int32(1), len)

	val, len = VLQDecode([]byte("2H"))
	assert.Equal(t, int32(123), val)
	assert.Equal(t, int32(2), len)

	val, len = VLQDecode([]byte("gkxH"))
	assert.Equal(t, int32(123456), val)
	assert.Equal(t, int32(4), len)

	val, len = VLQDecode([]byte("qxmvrH"))
	assert.Equal(t, int32(123456789), val)
	assert.Equal(t, int32(6), len)

}

func TestVLQDecode_Continuation(t *testing.T) {
	var values []int32
	in := []byte("A2HgkxHqxmvrH")

	for {
		val, len := VLQDecode(in)
		if len == 0 {
			break
		}
		in = in[len:]

		values = append(values, val)
	}

	assert.Equal(t, []int32{0, 123, 123456, 123456789}, values)
}

func BenchmarkVLQEncode(b *testing.B) {
	b.Run("short encode", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			VLQEncode(1)
		}
	})

	b.Run("long encode", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			VLQEncode(123456789)
		}
	})

	b.Run("short decode", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			VLQDecode([]byte("B"))
		}
	})

	b.Run("long decode", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			VLQDecode([]byte("A2HgkxHqxmvrH"))
		}
	})
}
