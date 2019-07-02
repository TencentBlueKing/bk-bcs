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

package mongodb

import (
	"reflect"
	"testing"
)

func TestDotHandler_1(t *testing.T) {
	before := map[string]interface{}{
		"foo1":         "bar1",
		"foo.foo.foo2": "bar.bar.bar2",
		"foo3": map[string]interface{}{
			"foo.foo.foo4": "bar4",
		},
	}
	after := map[string]interface{}{
		"foo1":                   "bar1",
		"foo\uff0efoo\uff0efoo2": "bar.bar.bar2",
		"foo3": map[string]interface{}{
			"foo\uff0efoo\uff0efoo4": "bar4",
		},
	}

	if !reflect.DeepEqual(dotHandler(before), after) {
		t.Errorf("dotHandler do not work as expected! before:\n%v\nafter:\n%v", before, after)
	}
}

func TestDotHandler_2(t *testing.T) {
	before := []interface{}{
		map[string]interface{}{
			"foo.foo.foo1": "bar1",
			"foo.foo.foo2": []interface{}{
				"foo.foo3",
				123,
				4.5,
				map[string]interface{}{
					"foo.foo4": "bar.bar.bar4",
				},
			},
		},
		map[string]interface{}{
			"foo.5": true,
		},
	}
	after := []interface{}{
		map[string]interface{}{
			"foo\uff0efoo\uff0efoo1": "bar1",
			"foo\uff0efoo\uff0efoo2": []interface{}{
				"foo.foo3",
				123,
				4.5,
				map[string]interface{}{
					"foo\uff0efoo4": "bar.bar.bar4",
				},
			},
		},
		map[string]interface{}{
			"foo\uff0e5": true,
		},
	}

	if !reflect.DeepEqual(dotHandler(before), after) {
		t.Errorf("dotHandler do not work as expected! \nbefore:\n%v\nafter:\n%v", before, after)
	}
}

func TestDotRecover_1(t *testing.T) {
	before := []interface{}{
		map[string]interface{}{
			"foo1":                   "bar1",
			"foo\uff0efoo\uff0efoo2": "bar.bar.bar2",
			"foo3": map[string]interface{}{
				"foo\uff0efoo\uff0efoo4": "bar4",
			},
		},
	}
	after := []interface{}{
		map[string]interface{}{
			"foo1":         "bar1",
			"foo.foo.foo2": "bar.bar.bar2",
			"foo3": map[string]interface{}{
				"foo.foo.foo4": "bar4",
			},
		},
	}

	if !reflect.DeepEqual(dotRecover(before), after) {
		t.Errorf("dotRecover do not work as expected! \nbefore:\n%v\nafter:\n%v", before, after)
	}
}

func TestDotRecover_2(t *testing.T) {
	before := []interface{}{
		map[string]interface{}{
			"foo\uff0efoo\uff0efoo1": "bar1",
			"foo\uff0efoo\uff0efoo2": []interface{}{
				"foo.foo3",
				4.5,
				123,
				map[string]interface{}{
					"foo\uff0efoo4": "bar.bar.bar4",
				},
			},
		},
		map[string]interface{}{
			"foo\uff0e5": true,
		},
	}
	expect := []interface{}{
		map[string]interface{}{
			"foo.foo.foo1": "bar1",
			"foo.foo.foo2": []interface{}{
				"foo.foo3",
				float64(4.5),
				uint64(123),
				map[string]interface{}{
					"foo.foo4": "bar.bar.bar4",
				},
			},
		},
		map[string]interface{}{
			"foo.5": true,
		},
	}

	if after := dotRecover(before); !reflect.DeepEqual(after, expect) {
		t.Errorf("dotRecover do not work as expected! \nbefore:\n%v\nafter:\n%v\nexpect:\n%v\n", before, after, expect)
	}
}
