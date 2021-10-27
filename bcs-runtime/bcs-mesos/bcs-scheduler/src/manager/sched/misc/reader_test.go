/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package misc

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func Example() {
	r := NewReader(strings.NewReader("6\nhello 0\n6\nworld!"))
	records, err := ioutil.ReadAll(r)
	fmt.Println(string(records), err)
	// Output:
	// hello world! <nil>
}

func TestReader(t *testing.T) {
	for i, tt := range []struct {
		in     string
		out    []byte
		fwd, n int
		err    error
	}{
		{"1\na0\n1\nb", []byte("a"), 0, 1, nil},
		{"1\na0\n1\nb", []byte("b"), 1, 1, nil},
		{"1\na", []byte{}, 0, 0, nil},
		{"2\nab", []byte("a"), 0, 1, nil},
		{"2\nab", []byte("ab"), 0, 2, nil},
		{"2\nab", []byte("b"), 1, 1, nil},
		{"2\nab", []byte{'b', 0}, 1, 1, nil},
		{"2\nab", []byte{0}, 2, 0, io.EOF},
		{ // size = (2 << 63) + 1
			"18446744073709551616\n", []byte{0}, 0, 0, &strconv.NumError{
				Func: "ParseUint",
				Num:  "18446744073709551616",
				Err:  strconv.ErrRange,
			},
		},
	} {
		type expect struct {
			p   []byte
			n   int
			err error
		}
		r := NewReader(strings.NewReader(tt.in))
		if n, err := r.Read(make([]byte, tt.fwd)); err != nil || n != tt.fwd {
			t.Fatalf("test #%d: failed to read forward %d bytes: %v", i, n, err)
		}
		want := expect{[]byte(tt.out), tt.n, tt.err}
		got := expect{p: make([]byte, len(tt.out))}
		if got.n, got.err = r.Read(got.p); !reflect.DeepEqual(got, want) {
			t.Errorf("test #%d: got: %+v, want: %+v", i, got, want)
		}
	}
}

func BenchmarkReader(b *testing.B) {
	var buf bytes.Buffer
	genRecords(b, &buf)

	r := NewReader(&buf)
	p := make([]byte, 256)

	b.StopTimer()
	b.ResetTimer()
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		if n, err := r.Read(p); err != nil && err != io.EOF {
			b.Fatal(err)
		} else {
			b.SetBytes(int64(n))
		}
	}
}

func genRecords(tb testing.TB, w io.Writer) {
	rnd := rng{rand.New(rand.NewSource(0xdeadbeef))}
	buf := make([]byte, 2<<12)
	for i := 0; i < cap(buf); i++ {
		sz := rnd.Intn(cap(buf))
		n, err := rnd.Read(buf[:sz])
		if err != nil {
			tb.Fatal(err)
		}
		header := strconv.FormatInt(int64(n), 10) + "\n"
		if _, err = io.WriteString(w, header); err != nil {
			tb.Fatal(err)
		} else if _, err = w.Write(buf[:n]); err != nil {
			tb.Fatal(err)
		}
	}
}

type rng struct{ *rand.Rand }

func (r rng) Read(p []byte) (n int, err error) {
	for i := 0; i < len(p); i += 7 {
		val := r.Int63()
		for j := 0; i+j < len(p) && j < 7; j++ {
			p[i+j] = byte(val)
			val >>= 8
		}
	}
	return len(p), nil
}
