/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/kirito41dd/xslice"
	"github.com/micro/go-micro/v2/registry"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
)

const (
	// IPV4 ipv4 flag
	IPV4 = "ipv4"
	// IPV6 ipv6 flag
	IPV6 = "ipv6"
)

// SplitAddrString split address string
func SplitAddrString(addrs string) []string {
	addrs = strings.Replace(addrs, ";", ",", -1)
	addrArray := strings.Split(addrs, ",")
	return addrArray
}

// SlicePtrToString to string by ","
func SlicePtrToString(ips []*string) string {
	if len(ips) == 0 {
		return ""
	}

	ipList := make([]string, 0)
	for _, ip := range ips {
		ipList = append(ipList, *ip)
	}

	return strings.Join(ipList, ",")
}

// SliceToString to string by ","
func SliceToString(slice []string) string {
	if len(slice) == 0 {
		return ""
	}
	if len(slice) == 1 {
		return slice[0]
	}

	sList := make([]string, 0)
	for _, s := range slice {
		sList = append(sList, s)
	}

	return strings.Join(sList, ",")
}

// GetXRequestIDFromHTTPRequest get X-Request-Id from http request
func GetXRequestIDFromHTTPRequest(req *http.Request) string {
	if req == nil {
		return ""
	}
	return req.Header.Get("X-Request-Id")
}

// RecoverPrintStack capture panic and print stack
func RecoverPrintStack(proc string) {
	if r := recover(); r != nil {
		blog.Errorf("[%s][recover] panic: %v, stack %v\n", proc, r, string(debug.Stack()))
		return
	}

	return
}

// StringInSlice returns true if given string in slice
func StringInSlice(s string, l []string) bool {
	for _, objStr := range l {
		if s == objStr {
			return true
		}
	}
	return false
}

// StringContainInSlice returns true if given string contain in slice
func StringContainInSlice(s string, l []string) bool {
	for _, objStr := range l {
		if strings.Contains(s, objStr) {
			return true
		}
	}
	return false
}

// IntInSlice return true if i in l
func IntInSlice(i int, l []int) bool {
	for _, obj := range l {
		if i == obj {
			return true
		}
	}
	return false
}

// SplitStringsChunks split strings chunk
func SplitStringsChunks(strList []string, limit int) [][]string {
	if limit <= 0 || len(strList) == 0 {
		return nil
	}
	i := xslice.SplitToChunks(strList, limit)
	ss, ok := i.([][]string)
	if !ok {
		return nil
	}

	return ss
}

// ToJSONString convert data struct to json string
func ToJSONString(data interface{}) string {
	b, _ := json.Marshal(data)
	return string(b)
}

// GetFileContent get file content
func GetFileContent(file string) (string, error) {
	body, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// GetServerEndpointsFromRegistryNode get dual address
func GetServerEndpointsFromRegistryNode(nodeServer *registry.Node) []string {
	// ipv4 server address
	endpoints := []string{nodeServer.Address}
	// ipv6 server address
	if ipv6Address := nodeServer.Metadata[types.IPV6]; ipv6Address != "" {
		endpoints = append(endpoints, ipv6Address)
	}

	return endpoints
}

// CheckIPAddressType check ip address type
func CheckIPAddressType(ip string) (string, error) {
	if net.ParseIP(ip) == nil {
		errMsg := fmt.Sprintf("Invalid IP Address: %s", ip)
		blog.Errorf(errMsg)
		return "", errors.New(errMsg)
	}

	for i := 0; i < len(ip); i++ {
		switch ip[i] {
		case '.':
			return IPV4, nil
		case ':':
			fmt.Printf("Given IP Address %s is IPV6 type\n", ip)
			return IPV6, nil
		}
	}

	return "", fmt.Errorf("not supported ip type")
}