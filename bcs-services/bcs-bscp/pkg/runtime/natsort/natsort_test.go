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
 */

package natsort

import (
	"math/rand"
	"reflect"
	"sort"
	"strconv"
	"testing"
)

func TestStringSort(t *testing.T) {
	want := []string{
		"ab", "abc1",
		"abc01", "abc2",
		"abc5", "abc10",
	}
	got := []string{
		"abc5", "abc1",
		"abc01", "ab",
		"abc10", "abc2",
	}
	sort.Sort(Natural(got))
	if !reflect.DeepEqual(want, got) {
		t.Errorf("Error: sort failed, expected: %#q, got: %#q", want, got)
	}
}

func TestNaturalLess(t *testing.T) {
	testset := []struct {
		s1, s2 string
		less   bool
	}{
		{"0", "00", true},
		{"aa", "ab", true},
		{"ab", "abc", true},
		{"abc", "ad", true},
		{"ab1", "ab2", true},
		{"ab1c", "ab1c", false},
		{"ab12", "abc", true},
		{"ab2a", "ab10", true},
		{"a0001", "a0000001", true},
		{"a10", "abcdefgh2", true},
		{"аб2аб", "аб10аб", true},
		{"2аб", "3аб", true},
		//
		{"a1b", "a01b", true},
		{"ab01b", "ab010b", true},
		{"a01b001", "a001b01", true},
		{"a1", "a1x", true},
		{"1ax", "1b", true},
		//
		{"082", "83", true},
		{"9a", "083a", true},
	}
	for _, v := range testset {
		if got := NaturalLess(v.s1, v.s2); got != v.less {
			t.Errorf("Compared %#q to %#q: expected %v, got %v",
				v.s1, v.s2, v.less, got)
		}
		// If A < B, then B < A must be false.
		// The same cannot be said if !(A < B),
		// because A might be equal to B
		if v.less {
			if v.s1 != v.s2 && NaturalLess(v.s2, v.s1) {
				t.Errorf("Reverse-compared %#q to %#q: expected false, got true",
					v.s2, v.s1)
			}
		}
	}
}

// Use a regular string sort.
// As this does not perform a natural sort,
// this is not directly comparable with the other sorts.
// It is only here for a sense of scale.
func BenchmarkStdStringSort(b *testing.B) {
	set := testSet(300)
	arr := make([]string, len(set[0]))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, list := range set {
			b.StopTimer()
			copy(arr, list)
			b.StartTimer()

			sort.Strings(arr)
		}
	}
}

// Natural sort order.
func BenchmarkNaturalStringSort(b *testing.B) {
	set := testSet(300)
	arr := make([]string, len(set[0]))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, list := range set {
			// Resetting the test set to be unsorted does not count.
			b.StopTimer()
			copy(arr, list)
			b.StartTimer()

			sort.Sort(Natural(arr))
		}
	}
}

func BenchmarkStdStringLess(b *testing.B) {
	set := testSet(300)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := range set[0] {
			k := (j + 1) % len(set[0])
			_ = set[0][j] < set[0][k]
		}
	}
}

func BenchmarkNaturalLess(b *testing.B) {
	set := testSet(300)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := range set[0] {
			k := (j + 1) % len(set[0])
			_ = NaturalLess(set[0][j], set[0][k])
		}
	}
}

// Get 1000 arrays of 10000-string-arrays (less if -short is specified).
func testSet(seed int) [][]string {
	gen := &generator{
		src: rand.New(rand.NewSource(
			int64(seed),
		)),
	}
	n := 1000
	if testing.Short() {
		n = 1
	}
	set := make([][]string, n)
	for i := range set {
		strings := make([]string, 10000)
		for idx := range strings {
			// Generate a random string
			strings[idx] = gen.NextString()
		}
		set[i] = strings
	}
	return set
}

type generator struct {
	src *rand.Rand
}

func (g *generator) NextInt(max int) int {
	return g.src.Intn(max)
}

// Gets random random-length alphanumeric string.
func (g *generator) NextString() (str string) {
	// Random-length 3-8 chars part
	strlen := g.src.Intn(6) + 3
	// Random-length 1-3 num
	numlen := g.src.Intn(3) + 1
	// Random position for num in string
	numpos := g.src.Intn(strlen + 1)
	// Generate the number
	var num string
	for i := 0; i < numlen; i++ {
		num += strconv.Itoa(g.src.Intn(10))
	}
	// Put it all together
	for i := 0; i < strlen+1; i++ {
		if i == numpos {
			str += num
		} else {
			str += string('a' + rune(g.src.Intn(16)))
		}
	}
	return str
}
