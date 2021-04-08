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
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
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
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"bk-bscp/pkg/kit"
)

// Global constant variables.
const (
	// APPIDPREFIX is prefix of appid.
	APPIDPREFIX = "A"

	// CFGIDPREFIX is prefix of cfgid.
	CFGIDPREFIX = "F"

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

	// CONFIGTEMPLATEPREFIX is prefix of config template.
	CONFIGTEMPLATEPREFIX = "T"

	// CONFIGTEMPLATEVERSIONPREFIX is prefix of config template version.
	CONFIGTEMPLATEVERSIONPREFIX = "TV"

	// VARGROUPPREFIX is prefix of variable group.
	VARGROUPPREFIX = "VG"

	// VARPREFIX is prefix of variable.
	VARPREFIX = "V"
)

const (
	// RidHeaderKey is request id header key.
	RidHeaderKey = "X-Bkapi-Request-Id"

	// UserHeaderKey is operator name header key.
	UserHeaderKey = "X-Bkapi-User-Name"

	// AppCodeHeaderKey is blueking application code header key.
	AppCodeHeaderKey = "X-Bkapi-App-Code"

	// ContentIDHeaderKey is common content sha256 id.
	ContentIDHeaderKey = "X-Bkapi-File-Content-Id"

	// ContentOverwriteHeaderKey is common content upload overwrite flag.
	ContentOverwriteHeaderKey = "X-Bkapi-File-Content-Overwrite"

	// AuthorizationHeaderKey is common authorization flag, only for internal authorize.
	AuthorizationHeaderKey = "X-Bkapi-Authorization"
)

// Global sequence num.
var seq uint64

// SequenceNum return an uint64 as a global sequence num.
func SequenceNum() uint64 {
	return atomic.AddUint64(&seq, 1)
}

// GenUUIDV1 generates a version 1 UUID string.
func GenUUIDV1() (string, error) {
	newUUID, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	return newUUID.String(), nil
}

// GenUUIDV4 generates a version 4 UUID string.
func GenUUIDV4() (string, error) {
	newUUID, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return newUUID.String(), nil
}

// GenUUID generates an UUID string which is version 4 or version 1.
func GenUUID() string {
	// V4: one chance in 17 billion.
	newUUID, err := GenUUIDV4()
	if err != nil {
		// NOTE: V1.
		newUUID, _ = GenUUIDV1()
	}
	return newUUID
}

// Sequence return a uuid string as a global sequence id.
func Sequence() string {
	return strings.Replace(GenUUID(), "-", "", -1)
}

// GenAppID generates an app id.
func GenAppID() (string, error) {
	return fmt.Sprintf("%s-%s", APPIDPREFIX, GenUUID()), nil
}

// GenCfgID generates a config id.
func GenCfgID() (string, error) {
	return fmt.Sprintf("%s-%s", CFGIDPREFIX, GenUUID()), nil
}

// GenCommitID generates a commit id.
func GenCommitID() (string, error) {
	return fmt.Sprintf("%s-%s", COMMITIDPREFIX, GenUUID()), nil
}

// GenMultiCommitID generates a multi commit id.
func GenMultiCommitID() (string, error) {
	return fmt.Sprintf("%s-%s", MULTICOMMITIDPREFIX, GenUUID()), nil
}

// GenReleaseID generates a release id.
func GenReleaseID() (string, error) {
	return fmt.Sprintf("%s-%s", RELEASEIDPREFIX, GenUUID()), nil
}

// GenMultiReleaseID generates a multi release id.
func GenMultiReleaseID() (string, error) {
	return fmt.Sprintf("%s-%s", MULTIRELEASEIDPREFIX, GenUUID()), nil
}

// GenStrategyID generates a strategy id.
func GenStrategyID() (string, error) {
	return fmt.Sprintf("%s-%s", STRATEGYIDPREFIX, GenUUID()), nil
}

// GenTemplateID generate a config template id.
func GenTemplateID() (string, error) {
	return fmt.Sprintf("%s-%s", CONFIGTEMPLATEPREFIX, GenUUID()), nil
}

// GenTemplateVersionID generate a config template id.
func GenTemplateVersionID() (string, error) {
	return fmt.Sprintf("%s-%s", CONFIGTEMPLATEVERSIONPREFIX, GenUUID()), nil
}

// GenVariableGroupID generate a variable group id.
func GenVariableGroupID() (string, error) {
	return fmt.Sprintf("%s-%s", VARGROUPPREFIX, GenUUID()), nil
}

// GenVariableID generate a variable id.
func GenVariableID() (string, error) {
	return fmt.Sprintf("%s-%s", VARPREFIX, GenUUID()), nil
}

// SHA1 returns a sha1 string of the data string.
func SHA1(data string) string {
	hash := sha1.New()
	if _, err := io.WriteString(hash, data); err != nil {
		return ""
	}
	return fmt.Sprintf("%X", hash.Sum(nil))
}

// SHA256 returns a sha256 string of the data string.
func SHA256(data string) string {
	hash := sha256.New()
	if _, err := io.WriteString(hash, data); err != nil {
		return ""
	}
	return fmt.Sprintf("%X", hash.Sum(nil))
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

	return "", errors.New("unknown target eth address")
}

// EmptyStr returns an empty string.
func EmptyStr() string {
	return ""
}

// EmptyJSONStr returns an empty json string.
func EmptyJSONStr() string {
	return "{}"
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

// ToFileMode converts string to FileMode type.
func ToFileMode(s string) os.FileMode {
	t, err := strconv.ParseUint(s, 0, 32)
	if err != nil {
		return 0
	}
	return os.FileMode(t)
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

// ParseFpath parses config fpath and return final path string.
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
	// config fpath is relative path, add root dir and parse by filepath Clean.
	return filepath.Clean(fmt.Sprintf("/%s", fpath))
}

// CreateFile creates the named file. If the file already exists,
// would not make it truncated. If the file does not exist, it is created with mode 0666.
// Support multi level dirs, auto create dirs if it is not exist.
// The permission bits perm (before umask) are used for all
// directories that MkdirAll creates. If path is already a directory, does nothing.
func CreateFile(fileName string) (*os.File, error) {
	// target file.
	fileName = filepath.Clean(fileName)

	// get directory cleaned.
	dir, _ := filepath.Split(fileName)

	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, err
	}
	return os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0666)
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

// DelayRandomMS delaies a random millisecond time.
func DelayRandomMS(max int) {
	rand.Seed(time.Now().UnixNano())
	time.Sleep(time.Duration(rand.Intn(max)) * time.Millisecond)
}

// RandomMS returns a random millisecond time.
func RandomMS(max int) time.Duration {
	rand.Seed(time.Now().UnixNano())
	return time.Duration(rand.Intn(max)) * time.Millisecond
}

// GRPCMethod return method name of target gRPC call from context.
func GRPCMethod(ctx context.Context) string {
	method, ok := grpc.Method(ctx)
	if !ok {
		return "unknown"
	}

	_, name := filepath.Split(method)
	if len(name) != 0 {
		return name
	}
	return "unknown"
}

// GRPCMetadata returns MD which is a mapping from metadata keys to values.
func GRPCMetadata(ctx context.Context) map[string]string {
	m := make(map[string]string, 0)

	// incoming context.
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok || len(md) == 0 {
		// context metadata empty.
		return m
	}

	for key, value := range md {
		if len(value) != 0 {
			// NOTE: only first value item.
			m[key] = value[0]
		}
	}
	return m
}

// RequestKit returns method/metadata KVs in grpc request.
func RequestKit(ctx context.Context) kit.Kit {
	method := GRPCMethod(ctx)
	metadata := GRPCMetadata(ctx)

	return kit.Kit{
		Ctx:     ctx,
		User:    metadata[strings.ToLower(UserHeaderKey)],
		Rid:     metadata[strings.ToLower(RidHeaderKey)],
		Method:  method,
		AppCode: metadata[strings.ToLower(AppCodeHeaderKey)],

		// NOTE: grpc metadata is only used for internal modules.
		Authorization: metadata[strings.ToLower(AuthorizationHeaderKey)],
	}
}

// HTTPRequestKit returns method/metadata KVs in http request.
func HTTPRequestKit(req *http.Request) kit.Kit {
	return kit.Kit{
		Ctx:     context.Background(),
		User:    req.Header.Get(UserHeaderKey),
		Rid:     req.Header.Get(RidHeaderKey),
		Method:  fmt.Sprintf("%s-%s", req.RequestURI, req.Method),
		AppCode: req.Header.Get(AppCodeHeaderKey),
	}
}

// DiskStatus is disk usage status.
type DiskStatus struct {
	// All is all disk capacity.
	All uint64 `json:"all"`

	// Used is used disk capacity.
	Used uint64 `json:"used"`

	// Free is free disk capacity.
	Free uint64 `json:"free"`
}

// DiskUsage return current disk usage of path/disk.
func DiskUsage(path string) (*DiskStatus, error) {
	fs := syscall.Statfs_t{}
	if err := syscall.Statfs(path, &fs); err != nil {
		return nil, err
	}

	status := &DiskStatus{}

	// disk bytes capacity.
	status.All = fs.Blocks * uint64(fs.Bsize)
	status.Free = fs.Bfree * uint64(fs.Bsize)
	status.Used = status.All - status.Free

	return status, nil
}

// StatDirectoryFileSize stats files size under target dir.
func StatDirectoryFileSize(dir string) (int64, error) {
	fList, err := ioutil.ReadDir(dir)
	if err != nil {
		return 0, err
	}

	var size int64

	for _, f := range fList {
		if f.IsDir() {
			subSize, err := StatDirectoryFileSize(dir + "/" + f.Name())
			if err != nil {
				return 0, err
			}
			size += subSize
		} else {
			size += f.Size()
		}
	}
	return size, nil
}
