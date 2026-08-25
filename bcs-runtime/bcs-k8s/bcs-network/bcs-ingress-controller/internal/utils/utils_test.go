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

package utils

import (
	"crypto/md5"
	"fmt"
	"testing"

	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8slabels "k8s.io/apimachinery/pkg/labels"
)

func TestGenIngressLabelKey(t *testing.T) {
	// synthetic 67-char name (no real user/ingress identifier)
	longName := "abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz01234"
	if len(longName) != 67 {
		t.Fatalf("fixture name length = %d, want 67", len(longName))
	}

	testCases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "short unchanged",
			in:   "ingress1",
			want: "ingress1",
		},
		{
			name: "exact 63 unchanged",
			in:   "abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz0",
			want: "abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz0",
		},
		{
			name: "incident 67-char name truncated",
			in:   longName,
			want: longName[:50] + fmt.Sprintf("%x", md5.Sum([]byte(longName)))[:13],
		},
	}

	for _, tc := range testCases {
		got := GenIngressLabelKey(tc.in)
		if got != tc.want {
			t.Errorf("%s: got %q (len=%d), want %q (len=%d)",
				tc.name, got, len(got), tc.want, len(tc.want))
		}
		if len(got) > MaxK8sLabelLen {
			t.Errorf("%s: result len %d exceeds MaxK8sLabelLen", tc.name, len(got))
		}
		// hashed/truncated key must be valid for LabelSelectorAsSelector
		_, err := k8smetav1.LabelSelectorAsSelector(k8smetav1.SetAsLabelSelector(k8slabels.Set(map[string]string{
			got: "ingress-name",
		})))
		if err != nil {
			t.Errorf("%s: LabelSelectorAsSelector failed for key %q: %v", tc.name, got, err)
		}
	}
}

func TestGenIngressLabelKeyStable(t *testing.T) {
	name := "abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz01234"
	a := GenIngressLabelKey(name)
	b := GenIngressLabelKey(name)
	if a != b {
		t.Errorf("GenIngressLabelKey not stable: %q != %q", a, b)
	}
}
