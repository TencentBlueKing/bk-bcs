package utils

import (
	"net"
	"strconv"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-apiserver-proxy/pkg/ipvs"
)

// EndPoint wrap IP&Port
type EndPoint struct {
	IP   string
	Port uint32
}

// String trans endpoint to ip:port
func (ep EndPoint) String() string {
	port := strconv.Itoa(int(ep.Port))
	return ep.IP + ":" + port
}

// SplitServer split server to ip, port
func SplitServer(server string) (string, uint32) {
	s := strings.Split(server, ":")
	if len(s) != 2 {
		blog.Warn("SplitServer error: len(s) is not two.")
		return "", 0
	}
	blog.V(5).Infof("SplitServer debug: IP: %s, Port: %s", s[0], s[1])

	p, err := strconv.Atoi(s[1])
	if err != nil {
		blog.Warn("SplitServer error: ", err)
		return "", 0
	}
	return s[0], uint32(p)
}

// BuildVirtualServer build vip to ipvs.VirtualServer
func BuildVirtualServer(vip string) *ipvs.VirtualServer {
	ip, port := SplitServer(vip)
	virServer := &ipvs.VirtualServer{
		Address:   net.ParseIP(ip),
		Protocol:  "TCP",
		Port:      port,
		Scheduler: "rr",
		Flags:     0,
		Timeout:   0,
	}
	return virServer
}

// BuildRealServer build real to ipvs.RealServer
func BuildRealServer(real string) *ipvs.RealServer {
	ip, port := SplitServer(real)
	realServer := &ipvs.RealServer{
		Address: net.ParseIP(ip),
		Port:    port,
		Weight:  1,
	}
	return realServer
}
