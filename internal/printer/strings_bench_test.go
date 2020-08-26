package printer

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"unsafe"
)

var out []byte

const (
	iters = 100
	str   = "test"
	size  = iters * len(str)
)

func BenchmarkStringBuilding(b *testing.B) {
	b.Run("strings.Builder", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			s := strings.Builder{}
			for j := 0; j < iters; j++ {
				s.WriteString(str)
			}
		}
	})

	b.Run("strings.Builder with cap", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			s := strings.Builder{}
			s.Grow(size)
			for j := 0; j < iters; j++ {
				s.WriteString(str)
			}
		}
	})

	b.Run("[]byte", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			var s []byte
			for j := 0; j < iters; j++ {
				s = append(s, str...)
			}
		}
	})

	b.Run("[]byte with cap", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			s := make([]byte, 0, size)
			for j := 0; j < iters; j++ {
				s = append(s, str...)
			}

			// Force value to escape to heap.
			out = s
		}
	})

	b.Run("bytes.Buffer", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			s := bytes.Buffer{}
			for j := 0; j < iters; j++ {
				s.WriteString(str)
			}
		}
	})

	b.Run("bytes.Buffer with cap", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			s := bytes.Buffer{}
			s.Grow(size)
			for j := 0; j < iters; j++ {
				s.WriteString(str)
			}
		}
	})
}

var strOut string

func BenchmarkString_Writes(b *testing.B) {
	b.Run("strings.Builder", func(b *testing.B) {
		b.ReportAllocs()
		s := strings.Builder{}
		for j := 0; j < iters; j++ {
			s.WriteString(str)
		}
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			strOut = s.String()
		}
	})

	b.Run("[]byte", func(b *testing.B) {
		b.ReportAllocs()

		var s []byte
		for j := 0; j < iters; j++ {
			s = append(s, str...)
		}
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			strOut = string(s)
		}
	})

	b.Run("[]byte with unsafe", func(b *testing.B) {
		b.ReportAllocs()

		var s []byte
		for j := 0; j < iters; j++ {
			s = append(s, str...)
		}
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			// Taken from strings.Builder.String().
			strOut = *(*string)(unsafe.Pointer(&s))
		}
	})

	b.Run("bytes.Buffer", func(b *testing.B) {
		b.ReportAllocs()

		s := bytes.Buffer{}
		for j := 0; j < iters; j++ {
			s.WriteString(str)
		}
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			strOut = s.String()
		}
	})
}

func BenchmarkStringBuilder_CapacityPlanning(b *testing.B) {
	for _, cap := range []int{
		0, 1, 10, 50, 100, 200, 500,
	} {
		b.Run(fmt.Sprint(cap), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				s := strings.Builder{}
				s.Grow(cap * len(str))
				for j := 0; j < iters; j++ {
					s.WriteString(str)
				}
			}
		})
	}
}
