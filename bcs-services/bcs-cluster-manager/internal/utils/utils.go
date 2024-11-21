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

// Package utils for utils
package utils

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net"
	"net/http"
	"os"
	"regexp"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/common/util"
	"github.com/kirito41dd/xslice"
	"github.com/robfig/cron/v3"
	"go-micro.dev/v4/registry"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	icommon "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
)

const (
	// IPV4 ipv4 flag
	IPV4 = "ipv4"
	// IPV6 ipv6 flag
	IPV6 = "ipv6"
	// DualStack dual stack flag
	DualStack = "dual"
)

const (
	intel = "intel"
	amd   = "amd"
)

var (
	// DefaultMask default mask
	DefaultMask = 24
)

// SplitAddrString split address string
func SplitAddrString(addrs string) []string {
	addrs = strings.ReplaceAll(addrs, ";", ",")
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

// SplitInt64sChunks split int64 chunk
func SplitInt64sChunks(strList []int64, limit int) [][]int64 {
	if limit <= 0 || len(strList) == 0 {
		return nil
	}
	i := xslice.SplitToChunks(strList, limit)
	ss, ok := i.([][]int64)
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
	return str == tranStr
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
	body, err := os.ReadFile(file)
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
	sList = append(sList, slice...)

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

// CheckNodeIfReady redy check node
func CheckNodeIfReady(n *corev1.Node) bool {
	if n == nil {
		return false
	}

	if len(n.Status.Conditions) == 0 {
		return false
	}

	// 检查Node是否处于Ready状态
	for _, condition := range n.Status.Conditions {
		if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
			return true
		}
	}

	return false
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
	ip := rand.Uint32() // nolint
	binary.LittleEndian.PutUint32(buf, ip)
	return string(buf)
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

// Int64PtrToInt64 ptrInt64 to int64
func Int64PtrToInt64(num *int64) int64 {
	if num == nil {
		return 0
	}

	return *num
}

// MatchPatternSubnet inner match inner subnet,
// if str == region, match node subnet; if str == region-clusterId, match cluster subnet
func MatchPatternSubnet(subnetName, str string) bool {
	var match bool
	patterns := []string{fmt.Sprintf("^%s-[1-9]-[0-9]+", str)}

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

// Split 分割字符串，支持 " ", ";", "," 分隔符
func Split(originStr string) []string {
	originStr = strings.ReplaceAll(originStr, ";", ",")
	originStr = strings.ReplaceAll(originStr, " ", ",")
	return strings.FieldsFunc(originStr, func(c rune) bool { return c == ',' })
}

// Partition 从指定分隔符的第一个位置，将字符串分为两段
func Partition(s string, sep string) (string, string) {
	parts := strings.SplitN(s, sep, 2)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}

// StringToInt string to int
func StringToInt(str string) (int, error) {
	num, err := strconv.Atoi(str)
	if err != nil {
		return 0, err
	}

	return num, nil
}

// TaintToK8sTaint convert taint to k8s taint
func TaintToK8sTaint(taint []*proto.Taint) []corev1.Taint {
	taints := make([]corev1.Taint, 0)
	for _, v := range taint {
		taints = append(taints, corev1.Taint{
			Key:    v.Key,
			Value:  v.Value,
			Effect: corev1.TaintEffect(v.Effect),
		})
	}
	return taints
}

// K8sTaintToTaint convert k8s taint to taint
func K8sTaintToTaint(taint []corev1.Taint) []*proto.Taint {
	taints := make([]*proto.Taint, 0)
	for _, v := range taint {
		taints = append(taints, &proto.Taint{
			Key:    v.Key,
			Value:  v.Value,
			Effect: string(v.Effect),
		})
	}
	return taints
}

// AllocateMachinesToAZs allocate num machines ro num zones
func AllocateMachinesToAZs(numMachines, numAZs int) [][]int {
	if numAZs <= 0 {
		return nil
	}

	allocation := make([][]int, numAZs)

	for i := 0; i < numMachines; i++ {
		azIndex := i % numAZs
		machineID := i + 1
		allocation[azIndex] = append(allocation[azIndex], machineID)
	}

	return allocation
}

// IsMasterNode check master node
func IsMasterNode(labels map[string]string) bool {
	_, ok1 := labels[icommon.MasterRole]
	_, ok2 := labels[icommon.ControlPlanRole]
	if ok1 || ok2 {
		return true
	}

	return false
}

// ExistRunningNodes check exist running nodes
func ExistRunningNodes(nodes []*proto.ClusterNode) bool {
	for i := range nodes {
		if nodes[i].GetStatus() == icommon.StatusRunning {
			return true
		}
	}

	return false
}

// FilterEmptyString filter empty string
func FilterEmptyString(strList []string) []string {
	filterStrings := make([]string, 0)

	for i := range strList {
		if len(strList[i]) > 0 {
			filterStrings = append(filterStrings, strList[i])
		}
	}

	return filterStrings
}

// CompareVersion 比较两个版本号字符串
func CompareVersion(v1, v2 string) int {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	for i := 0; i < len(parts1) || i < len(parts2); i++ {
		num1, _ := strconv.Atoi(parts1[i])
		num2, _ := strconv.Atoi(parts2[i])

		if num1 < num2 {
			return -1
		} else if num1 > num2 {
			return 1
		}
	}

	return 0
}

// DistributeMachines 平均分配机器到各个区域
func DistributeMachines(numMachines int, rZones []string) []int {
	if len(rZones) == 0 {
		return nil
	}

	numZones := len(rZones)

	base := numMachines / numZones      // 每个区域基本分配的机器数
	remainder := numMachines % numZones // 不能均匀分配的余数

	// 创建一个slice来存储每个区域分配到的机器数
	distributions := make([]int, numZones)

	// 每个区域都分配到基本的机器数
	for i := range distributions {
		distributions[i] = base
	}

	// 将余下的机器随机分散到各个区域
	for i := 0; i < remainder; i++ {
		distributions[generateRandomNumber(0, numZones)]++
	}

	return distributions
}

// generateRandomNumber 生成一个在[min, max)之间的随机整数
func generateRandomNumber(min int, max int) int {
	rand.Seed(time.Now().UnixNano()) // nolint
	return rand.Intn(max-min) + min  // nolint
}

// MergeStringIntMaps 合并两个 map[string]int
func MergeStringIntMaps(map1, map2 map[string]int) map[string]int {
	copyMap1 := make(map[string]int) // 创建一个新的 map
	for key, value := range map1 {
		copyMap1[key] = value // 复制元素
	}

	// 将 map2 的内容合并到 map1
	for key, value := range map2 {
		if existingValue, ok := copyMap1[key]; ok { // 键已存在
			copyMap1[key] = existingValue + value // 累加值
		} else { // 键不存在
			copyMap1[key] = value // 添加新键值对
		}
	}
	return copyMap1
}

// BuildAllocateVpcCidrLockKey generate vpc cidr lock
func BuildAllocateVpcCidrLockKey(cloud, region, vpcId string) string {
	return fmt.Sprintf("/bcs-services/bcs-cluster-manager/%s/%s/%s", cloud, region, vpcId)
}

// BuildClusterLockKey generate cluster lock
func BuildClusterLockKey(clusterId string) string {
	return fmt.Sprintf("/bcs-services/bcs-cluster-manager/%s", clusterId)
}

// ValidateCronExpr 验证给定的 cron 表达式是否有效
func ValidateCronExpr(expr string) error {
	_, err := cron.ParseStandard(expr)
	return err
}

// Decompose 将给定的数字分割为指定的值集合内的最大值分布
func Decompose(num int, values []int) ([]int, error) {
	if num <= 0 {
		return nil, fmt.Errorf("无效的数字：%d，必须为正整数", num)
	}
	// 对 values 按降序排序
	sort.Sort(sort.Reverse(sort.IntSlice(values)))

	var result []int
	for _, v := range values {
		for num >= v {
			result = append(result, v)
			num -= v
		}
	}

	if num > 0 {
		result = append(result, values[len(values)-1])
	}

	return result, nil
}

// Uint32ToInt uint32 to int
func Uint32ToInt(uint32Nums []uint32) ([]int, error) {
	intNums := make([]int, 0)

	for i := range uint32Nums {
		if uint32Nums[i] > math.MaxInt32 {
			return nil, fmt.Errorf("无法将 %d 转换为 int：超出范围", uint32Nums[i])
		}

		intNums = append(intNums, int(uint32Nums[i]))
	}

	return intNums, nil
}

// ConvertStringsToInterfaces xxx
func ConvertStringsToInterfaces(strings []string) []interface{} {
	interfaces := make([]interface{}, len(strings))

	for i, v := range strings {
		interfaces[i] = v
	}

	return interfaces
}
