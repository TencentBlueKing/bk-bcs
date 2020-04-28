/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package common

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
)

// Global constant variables.
const (
	// BIDPREFIX is prefix of bid.
	BIDPREFIX = "B"

	// APPIDPREFIX is prefix of appid.
	APPIDPREFIX = "A"

	// CLUSTERIDPREFIX is prefix of clusterid.
	CLUSTERIDPREFIX = "C"

	// ZONEIDPREFIX is prefix of zoneid.
	ZONEIDPREFIX = "Z"

	// CFGSETIDPREFIX is prefix of cfgsetid.
	CFGSETIDPREFIX = "F"

	// COMMITIDPREFIX is prefix of commitid.
	COMMITIDPREFIX = "M"

	// MULTICOMMITIDPREFIX is prefix of multi commitid.
	MULTICOMMITIDPREFIX = "MM"

	// RELEASEIDPREFIX is prefix of releaseid.
	RELEASEIDPREFIX = "R"

	// MULTIRELEASEIDPREFIX is prefix of multi releaseid.
	MULTIRELEASEIDPREFIX = "MR"

	// STRATEGYIDPREFIX is prefix of strategyid.
	STRATEGYIDPREFIX = "S"
)

// Global sequence num.
var seq uint64

// Sequence return an uint64 as a global sequence num.
func Sequence() uint64 {
	return atomic.AddUint64(&seq, 1)
}

// GenUUID generates an UUID string.
func GenUUID() (string, error) {
	uuid, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	return uuid.String(), nil
}

// GenBid generates a business id.
func GenBid() (string, error) {
	uuid, err := GenUUID()
	if err != nil {
		return "", err
	}
	id := fmt.Sprintf("%s-%s", BIDPREFIX, uuid)
	return id, nil
}

// GenAppid generates an app id.
func GenAppid() (string, error) {
	uuid, err := GenUUID()
	if err != nil {
		return "", err
	}
	id := fmt.Sprintf("%s-%s", APPIDPREFIX, uuid)
	return id, nil
}

// GenClusterid generates a cluster id.
func GenClusterid() (string, error) {
	uuid, err := GenUUID()
	if err != nil {
		return "", err
	}
	id := fmt.Sprintf("%s-%s", CLUSTERIDPREFIX, uuid)
	return id, nil
}

// GenZoneid generates a zone id.
func GenZoneid() (string, error) {
	uuid, err := GenUUID()
	if err != nil {
		return "", err
	}
	id := fmt.Sprintf("%s-%s", ZONEIDPREFIX, uuid)
	return id, nil
}

// GenCfgsetid generates a config set id.
func GenCfgsetid() (string, error) {
	uuid, err := GenUUID()
	if err != nil {
		return "", err
	}
	id := fmt.Sprintf("%s-%s", CFGSETIDPREFIX, uuid)
	return id, nil
}

// GenCommitid generates a commit id.
func GenCommitid() (string, error) {
	uuid, err := GenUUID()
	if err != nil {
		return "", err
	}
	id := fmt.Sprintf("%s-%s", COMMITIDPREFIX, uuid)
	return id, nil
}

// GenMultiCommitid generates a multi commit id.
func GenMultiCommitid() (string, error) {
	uuid, err := GenUUID()
	if err != nil {
		return "", err
	}
	id := fmt.Sprintf("%s-%s", MULTICOMMITIDPREFIX, uuid)
	return id, nil
}

// GenReleaseid generates a release id.
func GenReleaseid() (string, error) {
	uuid, err := GenUUID()
	if err != nil {
		return "", err
	}
	id := fmt.Sprintf("%s-%s", RELEASEIDPREFIX, uuid)
	return id, nil
}

// GenMultiReleaseid generates a multi release id.
func GenMultiReleaseid() (string, error) {
	uuid, err := GenUUID()
	if err != nil {
		return "", err
	}
	id := fmt.Sprintf("%s-%s", MULTIRELEASEIDPREFIX, uuid)
	return id, nil
}

// GenStrategyid generates a strategy id.
func GenStrategyid() (string, error) {
	uuid, err := GenUUID()
	if err != nil {
		return "", err
	}
	id := fmt.Sprintf("%s-%s", STRATEGYIDPREFIX, uuid)
	return id, nil
}

// SHA256 returns a sha256 string of the data string.
func SHA256(data string) string {
	t := sha256.New()
	if _, err := io.WriteString(t, data); err != nil {
		return ""
	}
	return fmt.Sprintf("%X", t.Sum(nil))
}

// FileSHA256 returns sha256 string of the file.
func FileSHA256(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer f.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%X", hash.Sum(nil)), nil
}

// GetenvCfg gets the env value of target key, returns
// default value if not exist or empty.
func GetenvCfg(key, defaultVal string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return defaultVal
	}
	return value
}

// GetenvCfgDuration gets the env time value of target key, returns
// default value if not exist or empty or any error happened.
func GetenvCfgDuration(key string, defaultVal time.Duration) time.Duration {
	value := os.Getenv(key)
	if len(value) == 0 {
		return defaultVal
	}

	t, err := time.ParseDuration(value)
	if err != nil {
		return defaultVal
	}

	return t
}

// GetenvCfgInt gets the env int value of target key, returns
// default value if not exist or empty or any error happened.
func GetenvCfgInt(key string, defaultVal int) int {
	value := os.Getenv(key)
	if len(value) == 0 {
		return defaultVal
	}

	i, err := strconv.Atoi(value)
	if err != nil {
		return defaultVal
	}

	return i
}

// GetEthAddr returns local address string of target eth.
func GetEthAddr(key string) (string, error) {
	ifis, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	if len(key) == 0 {
		key = "eth1"
	}

	for _, ifi := range ifis {
		if ifi.Name != key {
			continue
		}

		eth, err := net.InterfaceByName(ifi.Name)
		if err != nil {
			return "", err
		}

		addrs, err := eth.Addrs()
		if err != nil {
			return "", err
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				return ipnet.IP.String(), nil
			}
		}
	}

	return "", errors.New("unknow target eth address")
}

// ToStr converts int to string.
func ToStr(i int) string {
	return strconv.Itoa(i)
}

// ToInt converts string to int, returns 0 if any error happened.
func ToInt(str string) int {
	i, err := strconv.Atoi(str)
	if err != nil {
		return 0
	}
	return i
}

// Endpoint returns endpoint format string.
func Endpoint(ip string, port int) string {
	return fmt.Sprintf("%s:%d", ip, port)
}

// TimeNowMS returns millisecond timestamp.
func TimeNowMS() int64 {
	return time.Now().UnixNano() / 1e6
}

// ToMSTimestamp converts time.Time to millisecond timestamp.
func ToMSTimestamp(t time.Time) int64 {
	return t.UnixNano() / 1e6
}

// HandleSignals handles the OS signals.
func HandleSignals(exitFunc func()) {
	var onece sync.Once

	// recvice syscall.SIGINT and syscall.SIGTERM.
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)

	// waitting for signals.
	go func() {
		<-sigc
		onece.Do(exitFunc)
		os.Exit(1)
	}()
}

// SetupHTTPPprof setup the httpserver pprof.
func SetupHTTPPprof(addr string) {
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/debug/pprof/", func(w http.ResponseWriter, r *http.Request) {
			http.DefaultServeMux.ServeHTTP(w, r)
		})

		http.ListenAndServe(addr, mux)
	}()
}

// SetupCPUPprof setup the cpu pprof.
func SetupCPUPprof(filepath string, cpuprofileOut **os.File) {
	if cpuprofileOut == nil {
		return
	}
	dirs := strings.Split(filepath, "/")
	if err := os.MkdirAll(strings.Join(dirs[:len(dirs)-1], "/"), os.ModePerm); err != nil {
		log.Fatal(err)
		return
	}

	out, err := os.Create(filepath)
	if err != nil {
		log.Fatal(err)
		return
	}

	pprof.StartCPUProfile(out)
	*cpuprofileOut = out
}

// CollectCPUPprofData collects and finishes the cpu pprof.
func CollectCPUPprofData(file io.Closer) {
	pprof.StopCPUProfile()

	if file != nil {
		file.Close()
	}
}

// CollectMemPprofData collects and finishes the memory pprof.
func CollectMemPprofData(filepath string) {
	runtime.GC()

	dirs := strings.Split(filepath, "/")
	if err := os.MkdirAll(strings.Join(dirs[:len(dirs)-1], "/"), os.ModePerm); err != nil {
		log.Fatal(err)
		return
	}

	out, err := os.Create(filepath)
	if err != nil {
		log.Fatal(err)
		return
	}

	pprof.WriteHeapProfile(out)
	out.Close()

	var memStat runtime.MemStats
	runtime.ReadMemStats(&memStat)
	log.Print("mem", "Memory pprof data: Total: %d  Used: %d  System: %d",
		memStat.TotalAlloc, memStat.Alloc, memStat.Sys)
}

// ParseFpath parses configset fpath and return final path string.
// 				a/b/c         -> /a/b/c
// 				./a/b/c       -> /a/b/c
// 				/a/b/c        -> /a/b/c
// 				.//a/b/c      -> /a/b/c
// 				/a/b/../c     -> /a/c
// 				.//a/b/../c   -> /a/c
//
//              ""            -> /
//              "./"          -> /
//              "/"           -> /
//              "."           -> /
//              "./////"      -> /
//              ".//a/..//"   -> /
func ParseFpath(fpath string) string {
	// configset fpath is relative path, add root dir and parse by filepath Clean.
	return filepath.Clean(fmt.Sprintf("/%s", fpath))
}

// ParseHTTPBasicAuth parses http basic authorization, and return auth token.
func ParseHTTPBasicAuth(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok || len(md) == 0 {
		// context metadata empty.
		return "", nil
	}

	// Authorization: <type> <credentials>.
	// eg: grpcgateway-authorization: Basic YWRtaW46cGFzc3dvcmQ=
	authMD, ok := md["grpcgateway-authorization"]
	if !ok || len(authMD) == 0 {
		// auth info metadata empty.
		return "", nil
	}
	authInfo := strings.Split(authMD[0], " ")

	if len(authInfo) != 2 {
		// not std http authorization.
		return "", nil
	}

	// NOTE only handle Basic type.
	if authInfo[0] != "Basic" {
		// not basic auth type.
		return "", nil
	}

	// decode base64 credentials.
	auth, err := base64.StdEncoding.DecodeString(authInfo[1])
	if err != nil {
		return "", fmt.Errorf("decode auth info, %+v, %+v", authInfo, err)
	}
	return string(auth), nil
}

// VerifyUserPWD verify USER:PASSWORD.
func VerifyUserPWD(input, setting string) bool {
	// no PASSWORD setting.
	if len(setting) == 0 {
		return true
	}

	// no input PASSWORD.
	if len(input) == 0 {
		return false
	}

	// verify input and setting PASSWORD.
	inputUserPWD := strings.Split(input, ":")
	settingUserPWD := strings.Split(setting, ":")

	// USER and PASSWORD.
	if len(inputUserPWD) != len(settingUserPWD) {
		return false
	}
	if len(inputUserPWD) != 2 {
		return false
	}

	if inputUserPWD[0] != settingUserPWD[0] {
		return false
	}
	if inputUserPWD[1] != settingUserPWD[1] {
		return false
	}
	return true
}
