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
	InitIPv6Address(&ip)
	t.Log(ip)
}
