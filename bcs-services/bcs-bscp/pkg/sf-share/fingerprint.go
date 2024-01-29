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

// Package sfs NOTES
package sfs

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/constant"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
)

var (
	// errFileNotFound is open not exist file error.
	errFileNotFound = "no such file or directory"
	// errNetworkInterfaceNotFound is get not exist network interface.
	errNetworkInterfaceNotFound = "no such network interface"
)

// FingerPrint define sidecar runtime fingerprint interface.
type FingerPrint interface {
	Encode() string
}

// ipType defines fingerprint type, used to customize the format of the transport IP format.
type ipType string

// Encode defines the operations to encode ip.
func (i ipType) Encode() string {
	ip := string(i)

	// ipv4
	if strings.Contains(ip, ".") {
		return strings.ReplaceAll(ip, ".", "-")
	}

	// ipv6
	if strings.Contains(ip, ":") {
		return strings.ReplaceAll(ip, ":", "-")
	}

	return ip
}

// containerFP define container fingerprint info.
type containerFP struct {
	IP          ipType `json:"ip"`
	ContainerID string `json:"cid"`
	Uid         string `json:"uid"`
}

// Encode defines the operations to encode containerFP.
func (c *containerFP) Encode() string {
	return fmt.Sprintf("%s:%s:%s", c.IP.Encode(), c.ContainerID, c.Uid)
}

// hostFP define virtual machine and physical machine fingerprints.
type hostFP struct {
	IP  ipType `json:"ip"`
	Uid string `json:"uid"`
}

// Encode defines the operations to encode hostFP.
func (c *hostFP) Encode() string {
	return fmt.Sprintf("%s:%s", c.IP.Encode(), c.Uid)
}

// GetFingerPrint get current sidecar runtime fingerprint.
func GetFingerPrint() (FingerPrint, error) {
	// determine whether the current running environment is a container by determining
	// whether "/proc/1/cpuset" file exists and "/proc/1/cpuset" content size > 32.
	file, err := os.Open("/proc/1/cpuset")
	if err != nil {
		if !strings.Contains(err.Error(), errFileNotFound) {
			return nil, err
		}

		return getHostFP()
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	var isContainer bool
	if stat.Size() > 32 {
		isContainer = true
	}

	if isContainer {
		return getContainerFP(file)
	}

	return getHostFP()
}

// minor may be multiple sidecar in the same container, which needs to be distinguished by minor.
func minor() string {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec
	return fmt.Sprintf("%06v", rnd.Int31n(1000000))
}

// getContainerFP get sidecar in container env fingerprint.
func getContainerFP(uidFile *os.File) (*containerFP, error) {

	// step1: get container id.
	br := bufio.NewReader(uidFile)
	line, _, err := br.ReadLine()
	if err != nil {
		return nil, err
	}

	element := strings.Split(string(line), "/")
	containerID := strings.TrimSpace(element[len(element)-1])
	// containerID container id min length.
	if len(containerID) < 32 {
		logs.Errorf("container id length < 32, containerID: %s", containerID)
	}

	if strings.Contains(containerID, "-") {
		containerID = strings.Split(containerID, "-")[1]
	}

	if len(containerID) > 12 {
		containerID = containerID[0:12]
	}

	// step2: get container or host ip.
	var ip string
	ip = os.Getenv(constant.EnvHostIP)
	if len(ip) == 0 {
		ethIP, err := getIPFromNetworkInterface()
		if err != nil {
			return nil, err
		}

		ip = ethIP
	}

	if ip := net.ParseIP(ip); ip == nil {
		return nil, fmt.Errorf("invalid ip %s", ip)
	}

	return &containerFP{
		IP:          ipType(ip),
		ContainerID: containerID,
		Uid:         minor(),
	}, nil
}

// getHostFP get sidecar in host env fingerprint.
func getHostFP() (*hostFP, error) {

	ip, err := getIPFromNetworkInterface()
	if err != nil {
		return nil, err
	}

	if ip := net.ParseIP(ip); ip == nil {
		return nil, fmt.Errorf("invalid ip %s", ip)
	}

	return &hostFP{
		IP:  ipType(ip),
		Uid: minor(),
	}, nil
}

// getIPFromNetworkInterface first, get it from eth0 and eth1. If you cannot get it, randomly select an ip from
// other network cards that are not lo.
func getIPFromNetworkInterface() (string, error) {

	itfsName := []string{"eth0", "eth1"}
	for _, name := range itfsName {
		itfs, err := net.InterfaceByName(name)
		if err != nil {
			if strings.Contains(err.Error(), errNetworkInterfaceNotFound) {
				continue
			}

			return "", err
		}

		addrs, err := itfs.Addrs()
		if err != nil {
			return "", err
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				return ipnet.IP.String(), nil
			}
		}
	}

	itfs, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, one := range itfs {
		if strings.Contains(one.Name, "lo") {
			continue
		}

		addrs, err := one.Addrs()
		if err != nil {
			return "", err
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				return ipnet.IP.String(), nil
			}
		}
	}

	return "", fmt.Errorf("unable to get any network IP except lo")
}
