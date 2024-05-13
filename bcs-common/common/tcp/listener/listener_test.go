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

package listener

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/http/ipv6server"
	"go-micro.dev/v4"
	"go-micro.dev/v4/client"
	"go-micro.dev/v4/debug/handler"
	proto "go-micro.dev/v4/debug/proto"
	"go-micro.dev/v4/registry"
	"go-micro.dev/v4/server"
	"go-micro.dev/v4/transport"
	"go-micro.dev/v4/util/test"
	"google.golang.org/grpc"
	"google.golang.org/grpc/examples/helloworld/helloworld"
)

var tt *testing.T

const (
	IPv4 = "127.0.0.1"
	IPv6 = "::1"
	Port = "8080"
)

// Result Result
type Result struct {
	Result  bool        `json:"result"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// index 主页
func index(w http.ResponseWriter, r *http.Request) {
	message := r.Host
	tt.Log(r.Host, r.URL)
	res, _ := json.Marshal(&Result{Result: true, Message: message, Data: r.RemoteAddr})
	_, _ = fmt.Fprint(w, string(res))
}

// TestAddListener 测试AddListener
func TestAddListener(t *testing.T) {
	dualStackListener := NewDualStackListener()
	defer dualStackListener.Close()
	err := dualStackListener.AddListener(IPv4, Port)
	if err != nil {
		t.Error(err)
		return
	}
	err = dualStackListener.AddListener(IPv6, Port)
	if err != nil {
		t.Error(err)
		return
	}
}

// TestAcceptToHttpServer 测试http server使用
func TestAcceptToHttpServer(t *testing.T) {
	// 创建双栈
	dualStackListener := NewDualStackListener()
	err := dualStackListener.AddListener(IPv4, Port)
	if err != nil {
		t.Error(err)
		return
	}
	err = dualStackListener.AddListener(IPv6, Port)
	if err != nil {
		t.Error(err)
		return
	}
	// 设置index tt
	tt = t

	// 设置handle
	http.HandleFunc("/", index)

	//创建http server
	err = http.Serve(dualStackListener, nil)
	if err != nil {
		t.Error(err)
		return
	}
}

// TestAcceptToIPv6Server1 测试IPv6 server使用
func TestAcceptToIPv6Server1(t *testing.T) {
	// 创建双栈
	dualStackListener := NewDualStackListener()
	err := dualStackListener.AddListener(IPv4, Port)
	if err != nil {
		t.Error(err)
		return
	}
	err = dualStackListener.AddListener(IPv6, Port)
	if err != nil {
		t.Error(err)
		return
	}
	// 设置index tt
	tt = t

	// 设置handle
	http.HandleFunc("/", index)

	//创建IPv6 server
	err = ipv6server.Serve(dualStackListener, nil)
	if err != nil {
		t.Error(err)
		return
	}
}

// TestAcceptToIPv6Server2 测试IPv6 server使用
func TestAcceptToIPv6Server2(t *testing.T) {
	// 创建双栈
	dualStackListener := NewDualStackListener()
	err := dualStackListener.AddListener(IPv4, Port)
	if err != nil {
		t.Error(err)
		return
	}
	err = dualStackListener.AddListener(IPv6, Port)
	if err != nil {
		t.Error(err)
		return
	}
	// 设置index tt
	tt = t

	// 设置handle
	http.HandleFunc("/", index)

	//创建IPv6 server
	srv := ipv6server.NewParameterlessIPv6Server()
	err = srv.Serve(dualStackListener)
	if err != nil {
		t.Error(err)
		return
	}
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

// TestAcceptToIPv6ServerTLS1
func TestAcceptToIPv6ServerTLS1(t *testing.T) {
	tt = t

	//handler := http.NewServeMux()
	//handler.HandleFunc("/", index)

	certFile, keyFile := getTLSFiles()
	t.Logf("certFile:%s,keyFile:%s", certFile, keyFile)
	defer func() {
		_ = os.Remove(keyFile)
		_ = os.Remove(certFile)
	}()

	// 创建双栈
	dualStackListener := NewDualStackListener()
	err := dualStackListener.AddListener(IPv4, Port)
	if err != nil {
		t.Error(err)
		return
	}
	err = dualStackListener.AddListener(IPv6, Port)
	if err != nil {
		t.Error(err)
		return
	}
	// 设置index tt
	tt = t

	// 设置handle
	http.HandleFunc("/", index)

	//创建IPv6 server
	srv := ipv6server.NewParameterlessIPv6Server()
	err = srv.ServeTLS(dualStackListener, certFile, keyFile)
	if err != nil {
		t.Error(err)
		return
	}
}

// TestAcceptToIPv6ServerTLS2
func TestAcceptToIPv6ServerTLS2(t *testing.T) {
	tt = t

	//handler := http.NewServeMux()
	//handler.HandleFunc("/", index)

	certFile, keyFile := getTLSFiles()
	t.Logf("certFile:%s,keyFile:%s", certFile, keyFile)
	defer func() {
		_ = os.Remove(keyFile)
		_ = os.Remove(certFile)
	}()

	// 创建双栈
	dualStackListener := NewDualStackListener()
	err := dualStackListener.AddListener(IPv4, Port)
	if err != nil {
		t.Error(err)
		return
	}
	err = dualStackListener.AddListener(IPv6, Port)
	if err != nil {
		t.Error(err)
		return
	}
	// 设置index tt
	tt = t

	// 设置handle
	http.HandleFunc("/", index)

	//创建IPv6 server
	err = ipv6server.ServeTLS(dualStackListener, nil, certFile, keyFile)
	if err != nil {
		t.Error(err)
		return
	}
}

// TestClose 测试Close
func TestClose(t *testing.T) {
	// 创建双栈
	dualStackListener := NewDualStackListener()
	err := dualStackListener.AddListener(IPv4, Port)
	if err != nil {
		t.Error(err)
		return
	}
	err = dualStackListener.AddListener(IPv6, Port)
	if err != nil {
		t.Error(err)
		return
	}
	// 设置index tt
	tt = t

	// 设置handle
	http.HandleFunc("/", index)

	//创建http server
	go func() {
		err = http.Serve(dualStackListener, nil)
		if err != nil {
			t.Error(err)
			return
		}
	}()

	// 5s
	time.Sleep(5 * time.Second)
	// 尝试结束
	err = dualStackListener.Close()
	if err != nil {
		t.Error(err)
		return
	}
}

// TestAddr 测试Addr
func TestAddr(t *testing.T) {
	// 创建双栈
	dualStackListener := NewDualStackListener()
	err := dualStackListener.AddListener(IPv4, Port)
	if err != nil {
		t.Error(err)
		return
	}
	err = dualStackListener.AddListener(IPv6, Port)
	if err != nil {
		t.Error(err)
		return
	}
	// 设置index tt
	tt = t
	// Addr()
	t.Log(dualStackListener.Addr())
}

type GreeterServer struct {
	helloworld.UnimplementedGreeterServer
}

func (gs *GreeterServer) SayHello(ctx context.Context, request *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	value := request.Name
	tt.Log(value)
	return &helloworld.HelloReply{Message: "hello " + value}, nil
}

func TestGrpcServer(t *testing.T) {
	tt = t
	dualStackListener := NewDualStackListener()
	err := dualStackListener.AddListener(IPv4, Port)
	if err != nil {
		t.Error(err)
		return
	}
	err = dualStackListener.AddListener(IPv6, Port)
	if err != nil {
		t.Error(err)
		return
	}

	srv := grpc.NewServer()

	helloworld.RegisterGreeterServer(srv, &GreeterServer{})

	// 注入 dualStackListener
	err = srv.Serve(dualStackListener)
	if err != nil {
		t.Error(err)
		return
	}

}

func gRPCRequest(ip string) {
	conn, err := grpc.Dial(ip, grpc.WithInsecure())
	if err != nil {
		tt.Error(err)
		return
	}
	defer conn.Close()

	gClient := helloworld.NewGreeterClient(conn)

	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()

	reply, err := gClient.SayHello(ctx, &helloworld.HelloRequest{Name: "Arvin"})
	if err != nil {
		tt.Error(err)
		return
	}
	tt.Log(reply)
}

func TestGrpcClientIPv4(t *testing.T) {
	tt = t
	gRPCRequest(net.JoinHostPort(IPv4, Port))
}

func TestGrpcClientIPv6(t *testing.T) {
	tt = t
	gRPCRequest(net.JoinHostPort(IPv6, Port))
}

func testShutdown(wg *sync.WaitGroup, cancel func()) {
	// add 1
	wg.Add(1)
	// shutdown the service
	cancel()
	// wait for stop
	wg.Wait()
}

func testCustomListenService(ctx context.Context, customListener net.Listener, wg *sync.WaitGroup, name string) micro.Service {
	// add self
	wg.Add(1)

	r := registry.NewMemoryRegistry(registry.Services(test.Data))

	//httpTransport := transport.NewHTTPTransport(transport.NetListener2(customListener))

	//broker := broker.NewBroker(broker.SetNetListen(customListener))

	// create service
	srv := micro.NewService(
		micro.Name(name),
		micro.Context(ctx),
		micro.Registry(r),

		//micro.Transport(httpTransport),

		//micro.Broker(broker),

		micro.Address(net.JoinHostPort(IPv4, Port)),
		// injection customListener
		micro.AddListenOption(server.ListenOption(transport.NetListener(customListener))),

		micro.AfterStart(func() error {
			wg.Done()
			return nil
		}),
		micro.AfterStop(func() error {
			wg.Done()
			return nil
		}),
	)

	micro.RegisterHandler(srv.Server(), handler.NewHandler(srv.Client()))

	return srv
}

func testRequest(ctx context.Context, c client.Client, name string) error {
	// test call debug
	req := c.NewRequest(
		name,
		"Debug.Health",
		new(proto.HealthRequest),
	)

	rsp := new(proto.HealthResponse)

	err := c.Call(context.TODO(), req, rsp)
	if err != nil {
		return err
	}

	if rsp.Status != "ok" {
		return errors.New("service response: " + rsp.Status)
	}

	return nil
}

func benchmarkCustomListenService(b *testing.B, n int, name string) {
	// stop the timer
	b.StopTimer()

	// waitgroup for server start
	var wg sync.WaitGroup

	// cancellation context
	ctx, cancel := context.WithCancel(context.Background())

	// 双栈监听
	dualStackListener := NewDualStackListener()
	dualStackListener.addSubListener(net.JoinHostPort(IPv4, Port))
	dualStackListener.addSubListener(net.JoinHostPort(IPv6, Port))
	// create test server
	service := testCustomListenService(ctx, dualStackListener, &wg, name)

	//// create custom listen
	//customListen, err := net.Listen("tcp", net.JoinHostPort(IPv4, Port))
	//if err != nil {
	//	b.Fatal(err)
	//}
	////create test server
	//service := testCustomListenService(ctx, customListen, &wg, name)

	// start the server
	go func() {
		if err := service.Run(); err != nil {
			b.Fatal(err)
		}
	}()

	// wait for service to start
	wg.Wait()

	// make a test call to warm the cache
	for i := 0; i < 10; i++ {
		if err := testRequest(ctx, service.Client(), name); err != nil {
			b.Fatal(err)
		}
	}

	// start the timer
	b.StartTimer()

	// number of iterations
	for i := 0; i < b.N; i++ {
		// for concurrency
		for j := 0; j < n; j++ {
			wg.Add(1)

			go func() {
				err := testRequest(ctx, service.Client(), name)
				wg.Done()
				if err != nil {
					b.Fatal(err)
				}
			}()
		}

		// wait for test completion
		wg.Wait()
	}

	// stop the timer
	b.StopTimer()

	// shutdown service
	testShutdown(&wg, cancel)
}

func BenchmarkCustomListenService1(b *testing.B) {
	benchmarkCustomListenService(b, 1, "test.service.1")
}
