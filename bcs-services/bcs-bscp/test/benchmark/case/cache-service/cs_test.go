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

package cs_test

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"sync"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pbcs "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/cache-service"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/test/benchmark/run"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/test/util"
)

const (
	stressBizID     = 11
	stressAppID     = 501
	stressReleaseID = 501
)

var (
	// logCfg is log config
	logCfg util.LogConfig
	// conSize cache service conn pool size.
	conSize int
	// debug if debug is true, bench only request one, and print response result.
	debug bool
	// csPool cache service conn pool.
	csPool *pool
	// outputPath statistics result html file that by bench test result, save file path
	outputPath string
)

type pool struct {
	lock sync.Mutex
	idx  int
	conn []pbcs.CacheClient
}

// Pick get cache service client.
func (p *pool) Pick() pbcs.CacheClient {
	p.lock.Lock()
	defer func() {
		p.lock.Unlock()
	}()

	if p.idx == conSize-1 {
		p.idx = 0
		return p.conn[p.idx]
	}

	p.idx++
	return p.conn[p.idx]
}

func init() {
	var host string
	csPool = new(pool)

	flag.StringVar(&host, "host", "127.0.0.1:9514", "cache service grpc address")
	flag.IntVar(&conSize, "pool-size", 10, "cache service grpc client conn pool size")
	flag.IntVar(&run.Concurrent, "concurrent", 100, "concurrent request during the load test.")
	flag.Float64Var(&run.SustainSeconds, "sustain-seconds", 10, "the load test sustain time in seconds ")
	flag.Int64Var(&run.TotalRequest, "total-request", 0, "the load test total request,it has higher priority than "+
		"SustainSeconds")
	flag.BoolVar(&debug, "debug", false, "debug model only request one, and print response result")
	flag.StringVar(&outputPath, "output-path", "./bench.html", "statistics result html "+
		"file that by bench test result, save file path")
	flag.UintVar(&logCfg.Verbosity, "log-verbosity", 0, "log verbosity")
	testing.Init()
	flag.Parse()

	util.SetLogger(logCfg)

	// build cache service conn pool.
	csPool.conn = make([]pbcs.CacheClient, conSize)
	opts := make([]grpc.DialOption, 0)
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithWriteBufferSize(16*1024*1024),
		grpc.WithReadBufferSize(32*1024*1024), grpc.WithInitialConnWindowSize(32*1024*1024))
	for i := 0; i < conSize; i++ {
		conn, err := grpc.Dial(host, opts...)
		if err != nil {
			log.Println(err)
			return
		}

		csPool.conn[i] = pbcs.NewCacheClient(conn)
	}
}

// TestReport perform routine stress tests and generate stress test reports.
func TestReport(t *testing.T) {
	TestBenchAppMeta(t)
	// NOTE: strategy related test depends on group, add group test first
	//TestBenchAppCPS(t)
	TestBenchAppCRIMeta(t)
	TestBenchReleasedCI(t)
	TestGetAppMeta(t)
	TestGetReleasedCI(t)
	TestGetAppInstanceRelease(t)
	// NOTE: strategy related test depends on group, add group test first
	//TestGetAppReleasedStrategy(t)

	if err := run.GenReport(outputPath); err != nil {
		fmt.Println(err)
		return
	}
}

// TestBenchAppMeta test cache service query db's app meta func.
func TestBenchAppMeta(t *testing.T) {
	req := &pbcs.BenchAppMetaReq{
		BizId:  stressBizID,
		AppIds: []uint32{stressAppID},
	}

	if debug {
		resp, err := csPool.Pick().BenchAppMeta(context.Background(), req)
		if err != nil {
			log.Println(err)
			return
		}

		fmt.Printf("BenchAppMeta Resp: %+v", resp)
		return
	}

	m := run.FireLoadTest(func() error {
		resp, err := csPool.Pick().BenchAppMeta(context.Background(), req)
		if err != nil {
			return err
		}

		if len(resp.Meta) == 0 {
			return errors.New("BenchAppMeta return data not right")
		}

		return nil
	})

	run.Archive("TestBenchAppMeta", m)
	fmt.Printf("\nTestBenchAppMeta: \n" + m.Format())
}

// TestBenchAppCPS test cache service query db's app current published strategy func.
func TestBenchAppCPS(t *testing.T) {
	req := &pbcs.BenchAppCPSReq{
		BizId: stressBizID,
		AppId: stressAppID,
	}

	if debug {
		resp, err := csPool.Pick().BenchAppCPS(context.Background(), req)
		if err != nil {
			log.Println(err)
			return
		}

		fmt.Printf("BenchAppCPS Resp: %+v\n", resp)
		return
	}

	m := run.FireLoadTest(func() error {
		resp, err := csPool.Pick().BenchAppCPS(context.Background(), req)
		if err != nil {
			return err
		}

		if len(resp.Meta) == 0 {
			return errors.New("BenchAppCPS return data not right")
		}

		return nil
	})

	run.Archive("TestBenchAppCPS", m)
	fmt.Printf("\nTestBenchAppCPS: \n" + m.Format())
}

// TestBenchAppCRIMeta test cache service query db's app current released instance func.
func TestBenchAppCRIMeta(t *testing.T) {
	req := &pbcs.BenchAppCRIMetaReq{
		BizId: stressBizID,
		AppId: stressAppID,
	}

	if debug {
		resp, err := csPool.Pick().BenchAppCRIMeta(context.Background(), req)
		if err != nil {
			log.Println(err)
			return
		}

		fmt.Printf("BenchAppCRIMeta Resp: %+v\n", resp)
		return
	}

	m := run.FireLoadTest(func() error {
		resp, err := csPool.Pick().BenchAppCRIMeta(context.Background(), req)
		if err != nil {
			return err
		}

		if len(resp.Meta) == 0 {
			return errors.New("BenchAppCRIMeta return data not right")
		}

		return nil
	})

	run.Archive("TestBenchAppCRIMeta", m)
	fmt.Printf("\nTestBenchAppCRIMeta: \n" + m.Format())
}

// TestBenchReleasedCI test cache service query db's released config item func.
func TestBenchReleasedCI(t *testing.T) {
	req := &pbcs.BenchReleasedCIReq{
		BizId:     stressBizID,
		ReleaseId: stressReleaseID,
	}

	if debug {
		resp, err := csPool.Pick().BenchReleasedCI(context.Background(), req)
		if err != nil {
			log.Println(err)
			return
		}

		fmt.Printf("BenchReleasedCI Resp: %+v\n", resp)
		return
	}

	m := run.FireLoadTest(func() error {
		resp, err := csPool.Pick().BenchReleasedCI(context.Background(), req)
		if err != nil {
			return err
		}

		if len(resp.Meta) != 5 {
			return fmt.Errorf("BenchReleasedCI return data not right")
		}

		return nil
	})

	run.Archive("TestBenchReleasedCI", m)
	fmt.Printf("\nTestBenchReleasedCI: \n" + m.Format())
}

// TestBenchAverage four db query interfaces perform pressure test at the same time.
func TestBenchAverage(t *testing.T) {
	wg := sync.WaitGroup{}
	defer wg.Wait()

	wg.Add(1)
	go func() {
		req := &pbcs.BenchAppMetaReq{
			BizId:  stressBizID,
			AppIds: []uint32{stressAppID},
		}

		m := run.FireLoadTest(func() error {
			resp, err := csPool.Pick().BenchAppMeta(context.Background(), req)
			if err != nil {
				return err
			}

			if len(resp.Meta) == 0 {
				return errors.New("BenchAppMeta return data not right")
			}

			return nil
		})

		run.Archive("TestBenchAverage-BenchAppMeta", m)
		fmt.Printf("\nTestBenchAverage-BenchAppMeta: \n" + m.Format())
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		req := &pbcs.BenchAppCPSReq{
			BizId: stressBizID,
			AppId: stressAppID,
		}

		m := run.FireLoadTest(func() error {
			resp, err := csPool.Pick().BenchAppCPS(context.Background(), req)
			if err != nil {
				return err
			}

			if len(resp.Meta) == 0 {
				return errors.New("BenchAppCPS return data not right")
			}

			return nil
		})

		run.Archive("TestBenchAverage-BenchAppCPS", m)
		fmt.Printf("\nTestBenchAverage-BenchAppCPS: \n" + m.Format())
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		req := &pbcs.BenchAppCRIMetaReq{
			BizId: stressBizID,
			AppId: stressAppID,
		}

		m := run.FireLoadTest(func() error {
			resp, err := csPool.Pick().BenchAppCRIMeta(context.Background(), req)
			if err != nil {
				return err
			}

			if len(resp.Meta) == 0 {
				return errors.New("BenchAppCRIMeta return data not right")
			}

			return nil
		})

		run.Archive("TestBenchAverage-BenchAppCRIMeta", m)
		fmt.Printf("\nTestBenchAverage-BenchAppCRIMeta: \n" + m.Format())
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		req := &pbcs.BenchReleasedCIReq{
			BizId:     stressBizID,
			ReleaseId: stressReleaseID,
		}

		m := run.FireLoadTest(func() error {
			resp, err := csPool.Pick().BenchReleasedCI(context.Background(), req)
			if err != nil {
				return err
			}

			if len(resp.Meta) != 5 {
				return fmt.Errorf("BenchReleasedCI return data not right")
			}

			return nil
		})

		run.Archive("TestBenchAverage-BenchReleasedCI", m)
		fmt.Printf("\nTestBenchAverage-BenchReleasedCI: \n" + m.Format())
		wg.Done()
	}()
}

// TestGetAppMeta test cache service GetAppMeta interface.
func TestGetAppMeta(t *testing.T) {
	req := &pbcs.GetAppMetaReq{
		BizId: stressBizID,
		AppId: stressAppID,
	}

	if debug {
		resp, err := csPool.Pick().GetAppMeta(context.Background(), req)
		if err != nil {
			log.Println(err)
			return
		}

		fmt.Printf("GetAppMeta Resp: %+v\n", resp)
		return
	}

	m := run.FireLoadTest(func() error {
		resp, err := csPool.Pick().GetAppMeta(context.Background(), req)
		if err != nil {
			return err
		}

		if len(resp.JsonRaw) == 0 {
			return fmt.Errorf("GetAppMeta return data not right")
		}

		return nil
	})

	run.Archive("TestGetAppMeta", m)
	fmt.Printf("\nTestGetAppMeta: \n" + m.Format())
}

// TestGetReleasedCI test cache service GetReleasedCI interface.
func TestGetReleasedCI(t *testing.T) {
	req := &pbcs.GetReleasedCIReq{
		BizId:     stressBizID,
		ReleaseId: stressReleaseID,
	}

	if debug {
		resp, err := csPool.Pick().GetReleasedCI(context.Background(), req)
		if err != nil {
			log.Println(err)
			return
		}

		fmt.Printf("GetReleasedCI Resp: %+v\n", resp)
		return
	}

	m := run.FireLoadTest(func() error {
		resp, err := csPool.Pick().GetReleasedCI(context.Background(), req)
		if err != nil {
			return err
		}

		if len(resp.JsonRaw) == 0 {
			return errors.New("GetReleasedCI return data not right")
		}

		return nil
	})

	run.Archive("TestGetReleasedCI", m)
	fmt.Printf("\nTestGetReleasedCI: \n" + m.Format())
}

// TestGetAppInstanceRelease test cache service GetAppInstanceRelease interface.
func TestGetAppInstanceRelease(t *testing.T) {
	req := &pbcs.GetAppInstanceReleaseReq{
		BizId: stressBizID,
		AppId: stressAppID,
		Uid:   "961b6dd3ede3cb8ecbaacbd68de040cd78eb2ed5889130cceb4c49268ea4d506",
	}

	if debug {
		resp, err := csPool.Pick().GetAppInstanceRelease(context.Background(), req)
		if err != nil {
			log.Println(err)
			return
		}

		fmt.Printf("GetAppInstanceRelease Resp: %+v\n", resp)
		return
	}

	m := run.FireLoadTest(func() error {
		resp, err := csPool.Pick().GetAppInstanceRelease(context.Background(), req)
		if err != nil {
			return err
		}

		if resp.ReleaseId == 0 {
			return errors.New("GetAppInstanceRelease return data not right")
		}

		return nil
	})

	run.Archive("TestGetAppInstanceRelease", m)
	fmt.Printf("\nTestGetAppInstanceRelease: \n" + m.Format())
}

// TestGetAppReleasedStrategy test cache service GetAppReleasedStrategy interface.
func TestGetAppReleasedStrategy(t *testing.T) {

	cpsReq := &pbcs.GetAppCpsIDReq{
		BizId: stressBizID,
		AppId: stressAppID,
	}
	cpsResp, err := csPool.Pick().GetAppCpsID(context.Background(), cpsReq)
	if err != nil {
		log.Println(err)
		return
	}

	req := &pbcs.GetAppReleasedStrategyReq{
		BizId: stressBizID,
		AppId: stressAppID,
		CpsId: cpsResp.CpsId,
	}

	if debug {
		resp, err := csPool.Pick().GetAppReleasedStrategy(context.Background(), req)
		if err != nil {
			log.Println(err)
			return
		}

		fmt.Printf("GetAppReleasedStrategy Resp: %+v\n", resp)
		return
	}

	m := run.FireLoadTest(func() error {
		resp, err := csPool.Pick().GetAppReleasedStrategy(context.Background(), req)
		if err != nil {
			return err
		}

		if len(resp.JsonRaw) == 0 {
			return errors.New("CheckAppHasReleasedInstance return data not right")
		}

		return nil
	})

	run.Archive("TestGetAppReleasedStrategy", m)
	fmt.Printf("\nTestGetAppReleasedStrategy: \n" + m.Format())
}
