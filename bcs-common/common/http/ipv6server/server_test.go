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
 *
 */

package ipv6server

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/types"
)

// Result Result
type Result struct {
	Result  bool        `json:"result"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

const (
	// PORT 端口号
	PORT = "8080"
	// HTTP http
	HTTP = "http://"
	// HTTPS https
	HTTPS = "https://"
	// ipv4 回环地址
	IPv4LoopBack = "127.0.0.1"
	// ipv6 回环地址
	IPv6LoopBack = "::1"
)

//  tt 全局测试 *testing.T 便于输出测试结果
var tt *testing.T

// address 监听地址
var address = []string{IPv4LoopBack, IPv6LoopBack}

// index 主页
func index(w http.ResponseWriter, r *http.Request) {
	message := r.Host
	tt.Log(r.Host, r.URL)
	res, _ := json.Marshal(&Result{Result: true, Message: message, Data: r.RemoteAddr})
	_, _ = fmt.Fprint(w, string(res))
}

// joinScheme 拼接http
func joinScheme(scheme, url string) string {
	if url[:len(scheme)] == scheme {
		return url
	}
	return strings.Join([]string{scheme, url}, "")
}

// httpRequest http请求
func httpRequest(url string) (err error) {
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	tt.Log(string(bytes))
	return nil
}

// TestWebServer 启动 web server
func TestWebServer(t *testing.T) {
	tt = t
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/", index)
	ipv6Server := NewIPv6Server(address, PORT, types.TCP, serveMux)
	err := ipv6Server.StartWebServer()
	if err != nil {
		t.Error(err)
		return
	}
	select {}
}

// TestAccessWebServer 访问 web server
func TestAccessWebServer(t *testing.T) {
	tt = t
	ch := &sync.WaitGroup{}
	ch.Add(len(address))
	for _, v := range address {
		go func(ip string) {
			defer ch.Done()
			if err := httpRequest(joinScheme(HTTP, ip)); err != nil {
				t.Error(err)
			}
		}(net.JoinHostPort(v, PORT))
	}
	ch.Wait()
}

// TestServe 测试IPv4、IPv6地址监听
func TestServe(t *testing.T) {
	tt = t

	handler := http.NewServeMux()
	handler.HandleFunc("/", index)

	// create Listen
	listeners := make([]net.Listener, 0)
	for _, ip := range address {
		listen, err := net.Listen(types.TCP, net.JoinHostPort(ip, PORT))
		if err != nil {
			t.Error(err)
			return
		}
		listeners = append(listeners, listen)
	}

	errors := make(chan error, 1)
	defer close(errors)

	// start Server
	for _, v := range listeners {
		go func(listen net.Listener) {
			errors <- Serve(listen, handler)
		}(v)
	}

	t.Log("listen", strings.Join(address, ":"+PORT+","), ", start server...")
	t.Error(<-errors)
}

// TestAccessServer 访问 TestServe
func TestAccessServer(t *testing.T) {
	TestAccessWebServer(t)
}

// getTLSFiles 获取 私钥，证书
func getTLSFiles() (certFile, keyFile string) {
	max := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, _ := rand.Int(rand.Reader, max)
	subject := pkix.Name{
		Country:            []string{"CN"},
		Province:           []string{"Shenzhen"},
		Organization:       []string{"Devops"},
		OrganizationalUnit: []string{"certDevops"},
		CommonName:         "127.0.0.1",
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject:      subject,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}

	// private key
	key, _ := rsa.GenerateKey(rand.Reader, 2048)

	// cert
	cert, _ := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)

	// path
	dir, _ := ioutil.TempDir(os.TempDir(), "ca")

	// cert file
	certfile, _ := ioutil.TempFile(dir, "server.pem")
	_ = pem.Encode(certfile, &pem.Block{Type: "CERTIFICATE", Bytes: cert})

	// key file
	keyfile, _ := ioutil.TempFile(dir, "server.key")
	_ = pem.Encode(keyfile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})

	return certfile.Name(), keyfile.Name()
}

// TestServeTLS 启动 server(tls)
func TestServeTLS(t *testing.T) {
	tt = t

	handler := http.NewServeMux()
	handler.HandleFunc("/", index)

	certFile, keyFile := getTLSFiles()
	t.Logf("certFile:%s,keyFile:%s", certFile, keyFile)
	defer func() {
		_ = os.Remove(keyFile)
		_ = os.Remove(certFile)
	}()

	// create Listen
	listeners := make([]net.Listener, 0)
	for _, ip := range address {
		listen, err := net.Listen(types.TCP, net.JoinHostPort(ip, PORT))
		if err != nil {
			t.Error(err)
			return
		}
		listeners = append(listeners, listen)
	}

	errors := make(chan error, 1)
	defer close(errors)

	// start Server
	for _, v := range listeners {
		go func(listen net.Listener) {
			errors <- ServeTLS(listen, handler, certFile, keyFile)
		}(v)
	}

	t.Log("listen", strings.Join(address, ":"+PORT+","), ", start server(tls)...")
	t.Error(<-errors)
}

// httpsRequest https 请求
func httpsRequest(url string) error {
	// 创建一个请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		tt.Log(err)
		return err
	}

	tls13Transport := &http.Transport{
		MaxIdleConns: 10,
		TLSClientConfig: &tls.Config{
			MaxVersion:         tls.VersionTLS13,
			InsecureSkipVerify: true, // 不校验服务端证书，直接信任
		},
	}

	client := &http.Client{Transport: tls13Transport}

	response, err := client.Do(req)
	if err != nil {
		tt.Log(err)
		return err
	}

	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		tt.Log(err)
		return err
	}
	defer response.Body.Close()

	tt.Log(string(bytes))
	return nil
}

// TestAccessServeTLS 访问 TestServeTLS
func TestAccessServeTLS(t *testing.T) {
	tt = t
	ch := &sync.WaitGroup{}
	ch.Add(len(address))
	for _, v := range address {
		go func(ip string) {
			defer ch.Done()
			if err := httpsRequest(joinScheme(HTTPS, ip)); err != nil {
				t.Error(err)
			}
		}(net.JoinHostPort(v, PORT))
	}

	ch.Wait()
}

// TestListenAndServe 启动 ListenAndServe
func TestListenAndServe(t *testing.T) {
	tt = t

	handler := http.NewServeMux()
	handler.HandleFunc("/", index)

	errors := make(chan error, 1)
	defer close(errors)

	// start Server
	for _, v := range address {
		go func(ip string) {
			errors <- ListenAndServe(net.JoinHostPort(ip, PORT), handler)
		}(v)
	}

	t.Log("listen", strings.Join(address, ":"+PORT+","), ", start server...")

	select {
	case err := <-errors:
		t.Error(err)
	}
}

// TestAccessListenAndServe 访问 TestListenAndServe
func TestAccessListenAndServe(t *testing.T) {
	TestAccessWebServer(t)
}

// TestListenAndServeTLS 启动 ListenAndServe(TLS)
func TestListenAndServeTLS(t *testing.T) {

	tt = t

	handler := http.NewServeMux()
	handler.HandleFunc("/", index)

	certFile, keyFile := getTLSFiles()
	t.Logf("certFile:%s,keyFile:%s", certFile, keyFile)
	defer func() {
		_ = os.Remove(keyFile)
		_ = os.Remove(certFile)
	}()

	errors := make(chan error, 1)
	defer close(errors)

	// start Server
	for _, v := range address {
		go func(ip string) {
			errors <- ListenAndServeTLS(net.JoinHostPort(ip, PORT), certFile, keyFile, handler)
		}(v)
	}

	t.Log("listen", strings.Join(address, ":"+PORT+","), ", start server(tls)...")
	t.Error(<-errors)
}

// TestAccessListenAndServeTLS 访问 TestListenAndServeTLS
func TestAccessListenAndServeTLS(t *testing.T) {
	TestAccessServeTLS(t)
}

func TestIPv6Server_Serve(t *testing.T) {
	tt = t

	ipv6Server := &IPv6Server{
		Server: &http.Server{},
	}

	http.HandleFunc("/", index)

	// create Listen
	listeners := make([]net.Listener, 0)
	for _, ip := range address {
		listen, err := net.Listen(types.TCP, net.JoinHostPort(ip, PORT))
		if err != nil {
			t.Error(err)
			return
		}
		listeners = append(listeners, listen)
	}

	errors := make(chan error, 1)
	defer close(errors)

	// start Server
	for _, v := range listeners {
		go func(listen net.Listener) {
			errors <- ipv6Server.Serve(listen)
		}(v)
	}

	t.Log("listen", strings.Join(address, ":"+PORT+","), ", start server...")
	select {
	case err := <-errors:
		t.Error(err)
	}
}

func TestIPv6Server_ServeTLS(t *testing.T) {
	tt = t

	http.HandleFunc("/", index)

	ipv6Server := &IPv6Server{
		Server: &http.Server{},
	}

	certFile, keyFile := getTLSFiles()
	t.Logf("certFile:%s,keyFile:%s", certFile, keyFile)
	defer func() {
		_ = os.Remove(keyFile)
		_ = os.Remove(certFile)
	}()

	// create Listen
	listeners := make([]net.Listener, 0)
	for _, ip := range address {
		listen, err := net.Listen(types.TCP, net.JoinHostPort(ip, PORT))
		if err != nil {
			t.Error(err)
			return
		}
		listeners = append(listeners, listen)
	}

	errors := make(chan error, 1)
	defer close(errors)

	// start Server
	for _, v := range listeners {
		go func(listen net.Listener) {
			errors <- ipv6Server.ServeTLS(listen, certFile, keyFile)
		}(v)
	}

	t.Log("listen", strings.Join(address, ":"+PORT+","), ", start server...")
	select {
	case err := <-errors:
		t.Error(err)
	}
}

func TestIPv6Server_ListenAndServeTLS(t *testing.T) {
	tt = t

	certFile, keyFile := getTLSFiles()
	t.Logf("certFile:%s,keyFile:%s", certFile, keyFile)
	defer func() {
		_ = os.Remove(keyFile)
		_ = os.Remove(certFile)
	}()

	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/", index)
	ipv6Server := NewIPv6Server(address, PORT, types.TCP, serveMux)
	err := ipv6Server.ListenAndServeTLS(certFile, keyFile)
	if err != nil {
		t.Error(err)
		return
	}
	select {}
}
