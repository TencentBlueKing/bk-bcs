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

package cases

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/uuid"
	pbbase "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/base"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/jsoni"
)

// RandName generate rand resource name.
func RandName(prefix string) string {
	return fmt.Sprintf("%s-%s-%s", prefix, time.Now().Format("2006-01-02-15_04_05"), uuid.UUID())
}

// RandNameN generate rand resource name of n length.
func RandNameN(prefix string, n int) string {
	uid := uuid.UUID()
	result := fmt.Sprintf("%s-%s-%s", uid, prefix, RandString(n-2-len(prefix)-len(uid)))
	return result[:n]
}

// RandString generate a random string of n length.
func RandString(n int) string {
	if n <= 0 {
		return ""
	}

	var chars = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l",
		"m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"}

	str := strings.Builder{}
	length := len(chars)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < n; i++ {
		str.WriteString(chars[r.Intn(length)])
	}

	return str.String()
}

// SoShouldJsonEqual a func for So() of convey, it is aimed to verify that
// the JSON of the two values are equal.
func SoShouldJsonEqual(actual interface{}, expected ...interface{}) string {
	if len(expected) != 1 {
		return "SoShouldJsonEqual only need one expected value"
	}

	jActual, err := jsoni.Marshal(actual)
	if err != nil {
		errS := fmt.Sprintf("marshal json error in actual: %v", actual)
		log.Println(errS)
		return errS
	}

	jExpected, err := jsoni.Marshal(expected[0])
	if err != nil {
		return fmt.Sprintf("marshal json error in expected: %v", expected)
	}
	if string(jActual) == string(jExpected) {
		return ""
	}
	return fmt.Sprintf("Expected: %v, but actual: %v", expected, actual)
}

// SoRevision a func for So() of convey, it is aimed to verify revision
func SoRevision(actual interface{}, expected ...interface{}) string {
	if len(expected) > 0 {
		return "only need revision"
	}
	revision := actual.(*pbbase.Revision)
	if revision == nil {
		return "revision is null"
	} else if len(revision.Reviser) == 0 {
		return "revision's Reviser is null"
	} else if len(revision.CreateAt) == 0 {
		return "revision's CreateAt is null"
	} else if len(revision.UpdateAt) == 0 {
		return "revision's UpdateAt is null"
	} else if len(revision.Creator) == 0 {
		return "revision's Creator is null"
	}
	return ""
}

// SoCreateRevision a func for So() of convey, it is aimed to verify create revision
func SoCreateRevision(actual interface{}, expected ...interface{}) string {
	if len(expected) > 0 {
		return "only need revision"
	}
	revision, ok := actual.(*pbbase.CreatedRevision)
	if !ok {
		return "the value of actual must be *pbbase.CreatedRevision"
	}

	if revision == nil {
		return "revision is null"
	} else if len(revision.CreateAt) == 0 {
		return "revision's CreateAt is null"
	} else if len(revision.Creator) == 0 {
		return "revision's Creator is null"
	}
	return ""
}
