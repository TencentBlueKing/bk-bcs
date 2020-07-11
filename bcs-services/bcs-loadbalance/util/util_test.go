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

package util_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/Tencent/bk-bcs/bcs-services/bcs-loadbalance/util"
)

var _ = Describe("Util", func() {
	Describe("Execute command", func() {
		It("Normal test", func() {
			result, flag := ExeCommand("echo hello")
			Expect(result).To(Equal("hello\n"))
			Expect(flag).To(Equal(true))
		})
		It("Execute error test", func() {
			_, flag := ExeCommand("./unknown-command")
			Expect(flag).To(Equal(false))
		})
	})

	Describe("GetSubsection", func() {
		It("[a,b,c] - [b,c,d] -> [a]", func() {
			subs := GetSubsection([]string{"a", "b", "c"}, []string{"b", "c", "d"})
			Expect(subs).To(Equal([]string{"a"}))
		})
		It("[a,b,c] - [] -> [a,b,c]", func() {
			subs := GetSubsection([]string{"a", "b", "c"}, []string{})
			Expect(subs).To(Equal([]string{"a", "b", "c"}))
		})
		It("[] - [a,b,c] -> [] ", func() {
			subs := GetSubsection([]string{}, []string{"a", "b", "c"})
			Expect(subs).To(Equal([]string{}))
		})
	})

	Describe("TrimSpecialChar", func() {
		It("a/b c~d*e.f\\g -> abcdefd", func() {
			Expect(TrimSpecialChar("a/b c~d*e.f\\g")).To(Equal("abcdefg"))
		})
	})

	Describe("GetValidZookeeperPath", func() {
		It("a/b c~d*e.f\\g -> a_bcdef", func() {
			Expect(GetValidZookeeperPath("a/b c~d*e.f\\g")).To(Equal("a_bcdefg"))
		})
		It("/ -> ", func() {
			Expect(GetValidZookeeperPath("/")).To(Equal(""))
		})
	})

	Describe("GetValidTargetGroupSub", func() {
		It("a/b c~d*e.f\\g -> a-bcdef", func() {
			Expect(GetValidTargetGroupSub("a/b c~d*e.f\\g")).To(Equal("a-bcdefg"))
		})
		It("/ -> ", func() {
			Expect(GetValidTargetGroupSub("/")).To(Equal(""))
		})
	})
})
