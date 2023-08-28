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
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"regexp"
	"runtime/debug"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/common/util"

	"github.com/kirito41dd/xslice"
	"github.com/micro/go-micro/v2/registry"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	// IPV4 ipv4 flag
	IPV4 = "ipv4"
	// IPV6 ipv6 flag
	IPV6 = "ipv6"
)

const (
	intel = "intel"
	amd   = "amd"
)

// SplitAddrString split address string
func SplitAddrString(addrs string) []string {
	addrs = strings.Replace(addrs, ";", ",", -1)
	addrArray := strings.Split(addrs, ",")
	return addrArray
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

// SliceContainInString return true if slice contain in string
func SliceContainInString(l []string, s string) bool {
	for _, objStr := range l {
		if strings.Contains(s, objStr) {
			return true
		}
	}
	return false
}

// StringContainInSlice returns true if given string contain in slice
func StringContainInSlice(s string, l []string) bool {
	for _, objStr := range l {
		if strings.Contains(objStr, s) {
			return true
		}
	}
	return false
}

// StringContainInMap returns true if given string contain in map
func StringContainInMap(s string, m map[string]string) (bool, string) {
	ele, exist := m[s]

	return exist, ele
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

// ToStringObject convert data string to object
func ToStringObject(data []byte, object interface{}) error {
	return json.Unmarshal(data, object)
}

// SplitExistString split str1List exist str0List
func SplitExistString(str0List []string, str1List []string) ([]string, []string) {
	str0Map := sets.NewString(str0List...)
	var (
		existStr, notExistStr = make([]string, 0), make([]string, 0)
	)
	for i := range str1List {
		if str0Map.Has(str1List[i]) {
			existStr = append(existStr, str1List[i])
			continue
		}

		notExistStr = append(notExistStr, str1List[i])
	}

	return existStr, notExistStr
}

// JudgeBase64 check str if is base64 string
func JudgeBase64(str string) bool {
	pattern := "^([A-Za-z0-9+/]{4})*([A-Za-z0-9+/]{4}|[A-Za-z0-9+/]{3}=|[A-Za-z0-9+/]{2}==)$"
	matched, err := regexp.MatchString(pattern, str)
	if err != nil {
		return false
	}
	if !(len(str)%4 == 0 && matched) {
		return false
	}
	unCodeStr, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return false
	}
	tranStr := base64.StdEncoding.EncodeToString(unCodeStr)
	if str == tranStr {
		return true
	}
	return false
}

// MergeMap merge map
func MergeMap(mObj ...map[string]string) map[string]string {
	newObj := map[string]string{}
	for _, m := range mObj {
		for k, v := range m {
			newObj[k] = v
		}
	}
	return newObj
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

// GetFileContent get file content
func GetFileContent(file string) (string, error) {
	body, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}

	return string(body), nil
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

// Base64Encode encode src to base64
func Base64Encode(src string) string {
	return base64.StdEncoding.EncodeToString([]byte(src))
}

// Base64Decode encode src to base64
func Base64Decode(src string) (string, error) {
	if len(src) == 0 {
		return src, nil
	}

	dst, err := base64.StdEncoding.DecodeString(src)
	if err != nil {
		return src, err
	}

	return string(dst), nil
}

// GetNodeIPAddress get node address
func GetNodeIPAddress(node *corev1.Node) ([]string, []string) {
	ipv4Address := make([]string, 0)
	ipv6Address := make([]string, 0)

	for _, address := range node.Status.Addresses {
		if address.Type == corev1.NodeInternalIP {
			switch {
			case util.IsIPv6(address.Address):
				ipv6Address = append(ipv6Address, address.Address)
			case util.IsIPv4(address.Address):
				ipv4Address = append(ipv4Address, address.Address)
			default:
				blog.Errorf("unsupported ip type")
			}
		}
	}

	return ipv4Address, ipv6Address
}

// GetValueFromMap get value from map
func GetValueFromMap(m map[string]string, key string) string {
	v, ok := m[key]
	if !ok || v == "" {
		return ""
	}

	return v
}

// StringsToMap strings to map, like k1=v1;k2=v2 to {k1:v1, k2:v2}
func StringsToMap(str string) map[string]string {
	desMap := make(map[string]string)
	if len(str) == 0 {
		return desMap
	}

	strSlice := strings.Split(str, ";")
	if len(strSlice) == 0 {
		return desMap
	}

	for _, sub := range strSlice {
		if sub == "" {
			continue
		}
		ss := strings.Split(sub, "=")
		if len(ss) >= 1 {
			desMap[ss[0]] = ss[1]
		}
	}
	return desMap
}

// MapToStrings map to strings, like {k1:v1, k2:v2} to k1=v1;k2=v2
func MapToStrings(m map[string]string) string {
	strs := ""
	for k, v := range m {
		s := fmt.Sprintf("%s=%s;", k, v)
		strs += s
	}

	return strs
}

// FakeIPV4Addr generate ipv4 address
func FakeIPV4Addr() string {
	buf := make([]byte, 4)
	ip := rand.Uint32()
	binary.LittleEndian.PutUint32(buf, ip)
	return fmt.Sprintf("%s", net.IP(buf))
}

// GetCpuModuleType get cpuType label
func GetCpuModuleType(cpu string) string {
	lower := strings.ToLower(cpu)

	if strings.Contains(lower, intel) {
		return intel
	}

	if strings.Contains(lower, amd) {
		return amd
	}

	return ""
}

// StringPtrToString ptrString to string
func StringPtrToString(str *string) string {
	if str == nil {
		return ""
	}

	return *str
}

// MatchSubnet inner match subnet
func MatchSubnet(subnetName, region string) bool {
	var match bool
	patterns := []string{fmt.Sprintf("^%s-[1-9]-[0-9]+", region)}

	for _, pattern := range patterns {
		m, _ := regexp.MatchString(pattern, subnetName)
		match = match || m
	}
	return match
}

// GenerateNamespaceName generate vcluster namespace name
func GenerateNamespaceName(prefix, projectCode string, clusterID string) string {
	if prefix == "" {
		prefix = "vcluster"
	}

	return fmt.Sprintf("%s-%s-%s", prefix, projectCode, strings.ToLower(clusterID))
}
