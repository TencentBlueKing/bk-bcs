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

package util

import (
	"net"
	"strings"
	"testing"
)

func TestIsIpv6(t *testing.T) {
	tests := []struct {
		ip   string
		want bool
	}{
		{net.IPv6loopback.String(), true},
		{net.IPv6linklocalallrouters.String(), true},
		{net.IPv6unspecified.String(), true},
		{"10.20.30.40", false},
	}
	for _, test := range tests {
		t.Run(test.ip, func(t *testing.T) {
			if got := IsIPv6(test.ip); got != test.want {
				t.Errorf("ip = %v ,want = %v", test.ip, test.want)
			}
		})
	}
}

func TestVerifyIpv6(t *testing.T) {
	tests := []struct {
		ip   string
		want bool
	}{
		{"fd08:1:2::3", true},
		{net.IPv6loopback.String(), true},
		{net.IPv6linklocalallrouters.String(), true},
		{net.IPv6unspecified.String(), true},
		{"10.20.30.40", false},
	}
	for _, test := range tests {
		t.Run(test.ip, func(t *testing.T) {
			if got := VerifyIPv6(test.ip); got != test.want {
				t.Errorf("ip = %v ,want = %v", test.ip, test.want)
			}
		})
	}
}

func TestGetIpv6Address(t *testing.T) {
	tests := []struct {
		ip   string
		want string
	}{
		{
			ip:   strings.Join([]string{net.IPv4allsys.String(), net.IPv6loopback.String()}, ","),
			want: net.IPv6loopback.String(),
		},
		{
			ip:   strings.Join([]string{net.IPv4allsys.String(), net.IPv6linklocalallrouters.String()}, ","),
			want: net.IPv6linklocalallrouters.String(),
		},
		{
			ip:   strings.Join([]string{net.IPv4allsys.String(), net.IPv6unspecified.String()}, ","),
			want: net.IPv6unspecified.String(),
		},
		{
			ip:   strings.Join([]string{net.IPv4allsys.String(), ""}, ","),
			want: "",
		},
		{
			ip:   ",",
			want: "",
		},
		{
			ip:   "",
			want: "",
		},
	}
	for _, test := range tests {
		t.Run(test.ip, func(t *testing.T) {
			if got := GetIPv6Address(test.ip); got != test.want {
				t.Errorf("ip = %v ,want = %v", test.ip, test.want)
			}
		})
	}
}

func TestInitIPv6Address(t *testing.T) {
	ip := "fd08:1:2::3"
	InitIPv6Address(ip)
	t.Log(ip)
}
