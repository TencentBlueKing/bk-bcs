/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package feedserver

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"testing"
	"time"

	"bscp.io/cmd/feed-server/bll/types"
	"bscp.io/pkg/criteria/constant"
	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/criteria/uuid"
	"bscp.io/test/benchmark/run"
	"bscp.io/test/client/feed"

	_ "github.com/go-sql-driver/mysql" // import mysql drive, used to create conn.
	"github.com/jmoiron/sqlx"
)

var (
	// cli feed server client.
	cli *feed.Client
	// db config.
	dbCfg DBConfig
	// debug if debug is true, bench only request one, and print response result.
	debug bool
	// gwOpt api gateway bench request need reqs.
	gwOpt ApiGatewayOpt
	// outputPath statistics result html file that by bench test result, save file path
	outputPath string
	// goroutineNum the load test goroutine num at same time, only used to TestScene11
	goroutineNum int
	// appRandomNum randomly select the total number of stress testing applications, only used to TestScene12-14
	appRandomNum int
)

// DBConfig db config.
type DBConfig struct {
	IP       string
	Port     int64
	User     string
	Password string
	DB       string
}

// ApiGatewayOpt api gateway bench request need reqs.
type ApiGatewayOpt struct {
	AppCode     string
	AppSecret   string
	Ticket      string
	AccessToken string
	Jwt         string
}

func init() {
	var host string

	flag.StringVar(&host, "host", "http://127.0.0.1:8080", "feed server http address")
	flag.IntVar(&run.Concurrent, "concurrent", 1000, "concurrent request during the load test.")
	flag.Float64Var(&run.SustainSeconds, "sustain-seconds", 10, "the load test sustain time in seconds ")
	flag.Int64Var(&run.TotalRequest, "total-request", 0, "the load test total request,it has higher priority than "+
		"SustainSeconds")
	flag.IntVar(&goroutineNum, "goroutine-num", 1, "the load test goroutine num at same time, only used to TestScene11")
	flag.IntVar(&appRandomNum, "app-num", 1000, "randomly select the total number of stress testing applications, "+
		"only used to TestScene12-14")

	// mysql related flag req.
	flag.StringVar(&dbCfg.IP, "mysql-ip", "127.0.0.1", "mysql ip address")
	flag.Int64Var(&dbCfg.Port, "mysql-port", 3306, "mysql port")
	flag.StringVar(&dbCfg.User, "mysql-user", "root", "mysql login user")
	flag.StringVar(&dbCfg.Password, "mysql-passwd", "admin", "mysql login password")
	flag.StringVar(&dbCfg.DB, "mysql-db", "bk_bscp_admin", "mysql database")

	// api gateway bench related flag req.
	flag.StringVar(&gwOpt.AppCode, "app-code", "bscp", "request api gateway's app code")
	flag.StringVar(&gwOpt.AppSecret, "app-secret", "xxxxxx", "request api gateway's app secret")
	flag.StringVar(&gwOpt.Ticket, "ticket", "xxxxxx", "request api gateway's ticket")
	flag.StringVar(&gwOpt.AccessToken, "access-token", "xxxxxx", "request api gateway's access token")
	flag.StringVar(&gwOpt.Jwt, "api-jwt", "xxxxxx", "api gateway generate jwt")

	flag.BoolVar(&debug, "debug", false, "debug model only request one, and print response result, default false")
	flag.StringVar(&outputPath, "output-path", "./bench.html", "statistics result html "+
		"file that by bench test result, save file path")
	testing.Init()
	flag.Parse()

	// build feed server client.
	var err error
	cli, err = feed.NewFeedClient(host, nil)
	if err != nil {
		log.Printf("new feed server client failed, err: %v", err)
		return
	}
}

// TestReport bench scene 1-10.
func TestReport(t *testing.T) {
	TestScene12(t)

	if err := run.GenReport(outputPath); err != nil {
		fmt.Println(err)
		return
	}
}

// TestScene1 ?????????2????????????????????????????????????
func TestScene1(t *testing.T) {
	req := &types.ListFileAppLatestReleaseMetaReq{
		BizId: 2001,
		AppId: 100002,
		Uid:   "4fc82b26aecb47d2868c4efbe3581732a3e7cbcc6c2efb32062c08170a05eeb8",
	}

	if debug {
		resp, err := cli.ListAppFileLatestRelease(context.Background(), header(), req)
		if err != nil {
			log.Println(err)
			return
		}

		fmt.Printf("TestScene1 Resp: %+v", resp)
		return
	}

	m := run.FireLoadTest(func() error {
		resp, err := cli.ListAppFileLatestRelease(context.Background(), header(), req)
		if err != nil {
			return err
		}

		if resp.Code != errf.OK {
			return errf.New(resp.Code, resp.Message)
		}

		return nil
	})

	run.Archive("TestScene1", m)
	fmt.Printf("TestScene1: \n" + m.Format())
}

// TestScene2 ?????????3?????????????????????????????? label ??????Normal??????
func TestScene2(t *testing.T) {
	req := &types.ListFileAppLatestReleaseMetaReq{
		BizId: 2001,
		AppId: 100003,
		Uid:   "4fc82b26aecb47d2868c4efbe3581732a3e7cbcc6c2efb32062c08170a05eeb8",
		Labels: map[string]string{
			"biz": "2001",
		},
	}

	if debug {
		resp, err := cli.ListAppFileLatestRelease(context.Background(), header(), req)
		if err != nil {
			log.Println(err)
			return
		}

		fmt.Printf("TestScene2 Resp: %+v", resp)
		return
	}

	m := run.FireLoadTest(func() error {
		resp, err := cli.ListAppFileLatestRelease(context.Background(), header(), req)
		if err != nil {
			return err
		}

		if resp.Code != errf.OK {
			return errf.New(resp.Code, resp.Message)
		}

		return nil
	})

	run.Archive("TestScene2", m)
	fmt.Printf("TestScene2: \n" + m.Format())
}

// TestScene3 ?????????4????????????????????????????????????
func TestScene3(t *testing.T) {
	req := &types.ListFileAppLatestReleaseMetaReq{
		BizId: 2001,
		AppId: 100004,
		Uid:   "961b6dd3ede3cb8ecbaacbd68de040cd78eb2ed5889130cceb4c49268ea4d506",
	}

	if debug {
		resp, err := cli.ListAppFileLatestRelease(context.Background(), header(), req)
		if err != nil {
			log.Println(err)
			return
		}

		fmt.Printf("TestScene3 Resp: %+v", resp)
		return
	}

	m := run.FireLoadTest(func() error {
		resp, err := cli.ListAppFileLatestRelease(context.Background(), header(), req)
		if err != nil {
			return err
		}

		if resp.Code != errf.OK {
			return errf.New(resp.Code, resp.Message)
		}

		return nil
	})

	run.Archive("TestScene3", m)
	fmt.Printf("TestScene3: \n" + m.Format())
}

// TestScene4 ?????????5????????????????????????????????????
func TestScene4(t *testing.T) {
	req := &types.ListFileAppLatestReleaseMetaReq{
		BizId: 2001,
		AppId: 100005,
		Uid:   "4fc82b26aecb47d2868c4efbe3581732a3e7cbcc6c2efb32062c08170a05eeb8",
	}

	if debug {
		resp, err := cli.ListAppFileLatestRelease(context.Background(), header(), req)
		if err != nil {
			log.Println(err)
			return
		}

		fmt.Printf("TestScene4 Resp: %+v", resp)
		return
	}

	m := run.FireLoadTest(func() error {
		resp, err := cli.ListAppFileLatestRelease(context.Background(), header(), req)
		if err != nil {
			return err
		}

		if resp.Code != errf.OK {
			return errf.New(resp.Code, resp.Message)
		}

		return nil
	})

	run.Archive("TestScene4", m)
	fmt.Printf("TestScene4: \n" + m.Format())
}

// TestScene5 ?????????5?????????????????????????????? label ??????Normal??????
func TestScene5(t *testing.T) {
	req := &types.ListFileAppLatestReleaseMetaReq{
		BizId: 2001,
		AppId: 100005,
		Uid:   "4fc82b26aecb47d2868c4efbe3581732a3e7cbcc6c2efb32062c08170a05eeb8",
		Labels: map[string]string{
			"biz": "2001",
		},
	}

	if debug {
		resp, err := cli.ListAppFileLatestRelease(context.Background(), header(), req)
		if err != nil {
			log.Println(err)
			return
		}

		fmt.Printf("TestScene5 Resp: %+v", resp)
		return
	}

	m := run.FireLoadTest(func() error {
		resp, err := cli.ListAppFileLatestRelease(context.Background(), header(), req)
		if err != nil {
			return err
		}

		if resp.Code != errf.OK {
			return errf.New(resp.Code, resp.Message)
		}

		return nil
	})

	run.Archive("TestScene5", m)
	fmt.Printf("TestScene5: \n" + m.Format())
}

// TestScene6 ?????????5????????????????????????????????????
func TestScene6(t *testing.T) {
	req := &types.ListFileAppLatestReleaseMetaReq{
		BizId: 2001,
		AppId: 100005,
		Uid:   "961b6dd3ede3cb8ecbaacbd68de040cd78eb2ed5889130cceb4c49268ea4d506",
	}

	if debug {
		resp, err := cli.ListAppFileLatestRelease(context.Background(), header(), req)
		if err != nil {
			log.Println(err)
			return
		}

		fmt.Printf("TestScene6 Resp: %+v", resp)
		return
	}

	m := run.FireLoadTest(func() error {
		resp, err := cli.ListAppFileLatestRelease(context.Background(), header(), req)
		if err != nil {
			return err
		}

		if resp.Code != errf.OK {
			return errf.New(resp.Code, resp.Message)
		}

		return nil
	})

	run.Archive("TestScene6", m)
	fmt.Printf("TestScene6: \n" + m.Format())
}

// TestScene7 ?????????5?????????????????????????????????
func TestScene7(t *testing.T) {
	req := &types.ListFileAppLatestReleaseMetaReq{
		BizId: 2001,
		AppId: 100005,
		Uid:   "4fc82b26aecb47d2868c4efbe3581732a3e7cbcc6c2efb32062c08170a05eeb8",
		Labels: map[string]string{
			"sub": "true",
		},
	}

	if debug {
		resp, err := cli.ListAppFileLatestRelease(context.Background(), header(), req)
		if err != nil {
			log.Println(err)
			return
		}

		fmt.Printf("TestScene7 Resp: %+v", resp)
		return
	}

	m := run.FireLoadTest(func() error {
		resp, err := cli.ListAppFileLatestRelease(context.Background(), header(), req)
		if err != nil {
			return err
		}

		if resp.Code != errf.OK {
			return errf.New(resp.Code, resp.Message)
		}

		return nil
	})

	run.Archive("TestScene7", m)
	fmt.Printf("TestScene7: \n" + m.Format())
}

// TestScene8 ?????????5?????????????????????????????? label ??????Normal??????
func TestScene8(t *testing.T) {
	req := &types.ListFileAppLatestReleaseMetaReq{
		BizId: 2001,
		AppId: 100005,
		Uid:   "4fc82b26aecb47d2868c4efbe3581732a3e7cbcc6c2efb32062c08170a05eeb8",
		Labels: map[string]string{
			"biz":    "2002",
			"set":    "4",
			"module": "3",
			"game":   "stress1",
		},
	}

	if debug {
		resp, err := cli.ListAppFileLatestRelease(context.Background(), header(), req)
		if err != nil {
			log.Println(err)
			return
		}

		fmt.Printf("TestScene8 Resp: %+v", resp)
		return
	}

	m := run.FireLoadTest(func() error {
		resp, err := cli.ListAppFileLatestRelease(context.Background(), header(), req)
		if err != nil {
			return err
		}

		if resp.Code != errf.OK {
			return errf.New(resp.Code, resp.Message)
		}

		return nil
	})

	run.Archive("TestScene8", m)
	fmt.Printf("TestScene8: \n" + m.Format())
}

// TestScene9 ?????????6????????????????????????????????????
func TestScene9(t *testing.T) {
	req := &types.ListFileAppLatestReleaseMetaReq{
		BizId:     2001,
		AppId:     100006,
		Uid:       "4fc82b26aecb47d2868c4efbe3581732a3e7cbcc6c2efb32062c08170a05eeb8",
		Namespace: "namespace",
	}

	if debug {
		resp, err := cli.ListAppFileLatestRelease(context.Background(), header(), req)
		if err != nil {
			log.Println(err)
			return
		}

		fmt.Printf("TestScene9 Resp: %+v", resp)
		return
	}

	m := run.FireLoadTest(func() error {
		resp, err := cli.ListAppFileLatestRelease(context.Background(), header(), req)
		if err != nil {
			return err
		}

		if resp.Code != errf.OK {
			return errf.New(resp.Code, resp.Message)
		}

		return nil
	})

	run.Archive("TestScene9", m)
	fmt.Printf("TestScene9: \n" + m.Format())
}

// TestScene10 ?????????6????????????????????????Namespace??????
func TestScene10(t *testing.T) {
	req := &types.ListFileAppLatestReleaseMetaReq{
		BizId:     2001,
		AppId:     100006,
		Uid:       "4fc82b26aecb47d2868c4efbe3581732a3e7cbcc6c2efb32062c08170a05eeb8",
		Namespace: "namespace0",
	}

	if debug {
		resp, err := cli.ListAppFileLatestRelease(context.Background(), header(), req)
		if err != nil {
			log.Println(err)
			return
		}

		fmt.Printf("TestScene10 Resp: %+v", resp)
		return
	}

	m := run.FireLoadTest(func() error {
		resp, err := cli.ListAppFileLatestRelease(context.Background(), header(), req)
		if err != nil {
			return err
		}

		if resp.Code != errf.OK {
			return errf.New(resp.Code, resp.Message)
		}

		return nil
	})

	run.Archive("TestScene10", m)
	fmt.Printf("TestScene10: \n" + m.Format())
}

// TestScene11 ?????????6????????????????????????????????????????????????????????????????????????
func TestScene11(t *testing.T) {
	req := &types.ListFileAppLatestReleaseMetaReq{
		BizId:     2001,
		AppId:     100006,
		Uid:       "4fc82b26aecb47d2868c4efbe3581732a3e7cbcc6c2efb32062c08170a05eeb8",
		Namespace: "namespace",
	}

	if debug {
		resp, err := cli.ListAppFileLatestRelease(context.Background(), header(), req)
		if err != nil {
			log.Println(err)
			return
		}

		fmt.Printf("TestScene11 Resp: %+v", resp)
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(goroutineNum)
	for i := 0; i < goroutineNum; i++ {
		go func() {
			m := run.FireLoadTest(func() error {
				resp, err := cli.ListAppFileLatestRelease(context.Background(), header(), req)
				if err != nil {
					return err
				}

				if resp.Code != errf.OK {
					return errf.New(resp.Code, resp.Message)
				}

				return nil
			})

			fmt.Printf("TestScene11: \n" + m.Format())
			wg.Done()
		}()
	}
	wg.Wait()
}

// TestScene12 ?????????????????????1000?????????????????????Namespace??????
func TestScene12(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	list, err := getQueryReleaseMeta()
	if err != nil {
		log.Println(err)
		return
	}

	if debug {
		// random debug an application.
		meta := list[rand.Intn(appRandomNum)]

		req := &types.ListFileAppLatestReleaseMetaReq{
			BizId:     meta.BizID,
			AppId:     meta.AppID,
			Uid:       uuid.UUID(),
			Namespace: meta.Namespace,
		}

		resp, err := cli.ListAppFileLatestRelease(context.Background(), header(), req)
		if err != nil {
			log.Println(err)
			return
		}

		fmt.Printf("TestScene12 Resp: %+v", resp)
		return
	}

	m := run.FireLoadTest(func() error {
		meta := list[rand.Intn(appRandomNum)]
		req := &types.ListFileAppLatestReleaseMetaReq{
			BizId:     meta.BizID,
			AppId:     meta.AppID,
			Uid:       uuid.UUID(),
			Namespace: meta.Namespace,
		}

		resp, err := cli.ListAppFileLatestRelease(context.Background(), header(), req)
		if err != nil {
			return err
		}

		if resp.Code != errf.OK {
			return errf.New(resp.Code, resp.Message)
		}

		return nil
	})

	run.Archive(fmt.Sprintf("%d App Random Pull", appRandomNum), m)
	fmt.Printf("TestScene12: \n" + m.Format())
}

// TestScene13 ?????????????????????1000?????????????????????Namespace?????????????????????????????? FeedServer ??? ApiGateway???
func TestScene13(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	list, err := getQueryReleaseMeta()
	if err != nil {
		log.Println(err)
		return
	}

	// build request header.
	authHeader := fmt.Sprintf(`{"bk_ticket": "%s", "bk_app_code": "%s", "bk_app_secret": "%s", "access_token": "%s"}`,
		gwOpt.Ticket, gwOpt.AppCode, gwOpt.AppSecret, gwOpt.AccessToken)
	h := http.Header{}
	h.Set("x-bkapi-authorization", authHeader)

	if debug {
		// random debug an application.
		meta := list[rand.Intn(appRandomNum)]

		req := &types.ListFileAppLatestReleaseMetaReq{
			BizId:     meta.BizID,
			AppId:     meta.AppID,
			Uid:       uuid.UUID(),
			Namespace: meta.Namespace,
		}

		resp, err := cli.ListAppFileLatestRelease(context.Background(), h, req)
		if err != nil {
			log.Println(err)
			return
		}

		fmt.Printf("TestScene13 Resp: %+v", resp)
		return
	}

	m := run.FireLoadTest(func() error {
		meta := list[rand.Intn(appRandomNum)]
		req := &types.ListFileAppLatestReleaseMetaReq{
			BizId:     meta.BizID,
			AppId:     meta.AppID,
			Uid:       uuid.UUID(),
			Namespace: meta.Namespace,
		}

		resp, err := cli.ListAppFileLatestRelease(context.Background(), h, req)
		if err != nil {
			return err
		}

		if resp.Code != errf.OK {
			return errf.New(resp.Code, resp.Message)
		}

		return nil
	})

	run.Archive("TestScene13", m)
	fmt.Printf("TestScene13: \n" + m.Format())
}

// header http request need header.
func header() http.Header {
	header := http.Header{}
	header.Set(constant.UserKey, "stress")
	header.Set(constant.RidKey, uuid.UUID())
	header.Set(constant.AppCodeKey, "test")
	header.Set(constant.BKGWJWTTokenKey, gwOpt.Jwt)
	return header
}

// QueryReleaseMeta used to query app latest release info.
type QueryReleaseMeta struct {
	BizID     uint32 `db:"biz_id"`
	AppID     uint32 `db:"app_id"`
	Namespace string `db:"namespace"`
}

// getQueryReleaseMeta get query app latest release info meta.
func getQueryReleaseMeta() ([]*QueryReleaseMeta, error) {
	dsn := fmt.Sprintf("%s:%s@(%s:%d)/%s?charset=utf8&parseTime=True&loc=UTC",
		dbCfg.User, dbCfg.Password, dbCfg.IP, dbCfg.Port, dbCfg.DB)
	db := sqlx.MustConnect("mysql", dsn)

	list := make([]*QueryReleaseMeta, 0)
	if err := db.Select(&list, fmt.Sprintf("SELECT biz_id, app_id, namespace FROM strategy ORDER BY id LIMIT %d",
		appRandomNum)); err != nil {
		return nil, err
	}

	return list, nil
}
