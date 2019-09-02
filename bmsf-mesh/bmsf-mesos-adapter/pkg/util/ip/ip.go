package ip

import (
	"net"
)

var (
	_, classA, _  = net.ParseCIDR("10.0.0.0/8")
	_, classA2, _ = net.ParseCIDR("9.0.0.0/8")
	_, classAa, _ = net.ParseCIDR("100.64.0.0/10")
	_, classB, _  = net.ParseCIDR("172.16.0.0/12")
	_, classC, _  = net.ParseCIDR("192.168.0.0/16")
)

//GetAvailableIP get local host first available ip address
//return empty string if error accurs
func GetAvailableIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, addr := range addrs {
		if ip, ok := addr.(*net.IPNet); ok && !ip.IP.IsLoopback() && ip.IP.To4() != nil {
			if classA.Contains(ip.IP) {
				return ip.IP.String()
			}
			if classA2.Contains(ip.IP) {
				return ip.IP.String()
			}
			if classAa.Contains(ip.IP) {
				return ip.IP.String()
			}
			if classB.Contains(ip.IP) {
				return ip.IP.String()
			}
			if classC.Contains(ip.IP) {
				return ip.IP.String()
			}
		}
	}
	return ""
}
