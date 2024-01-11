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

package uuid

import "testing"

func BenchmarkUUID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = UUID()
	}
	// Benchmark Result
	// 1: pborman/uuid
	// 2: github.com/google/uuid
	// 3: github.com/google/uuid with uuid.EnableRandPool()

	// BenchmarkUUID-16         5305968               223.3 ns/op            64 B/op          2 allocs/op
	// BenchmarkUUID-16          815874               1463 ns/op             64 B/op          2 allocs/op
	// BenchmarkUUID-16         3996272               279.6 ns/op            48 B/op          1 allocs/op
}
