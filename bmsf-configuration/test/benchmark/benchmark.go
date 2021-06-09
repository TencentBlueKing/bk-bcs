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

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/connserver"
	"bk-bscp/internal/types"
	"bk-bscp/pkg/common"
)

var (
	// target is the endpoint of target connserver.
	target string

	// gRPC client timeout.
	timeout time.Duration

	// concurrence num.
	concurrence int

	// peer concurrence request num.
	peerCount int

	// request count.
	reqCount uint64

	// success count.
	succCount uint64

	// failed count.
	failCount uint64

	// request time count.
	rtimeCount uint64

	// max rtime.
	maxRtime uint64

	// min rtime.
	minRtime uint64

	// debug prints logs.
	debug bool

	// benchmark start time.
	startTime uint64

	// benchmark stop time.
	stopTime uint64
)

func init() {
	flag.StringVar(&target, "target", "127.0.0.1:9516", "Endpoint of target connserver.")
	flag.DurationVar(&timeout, "timeout", 3*time.Second, "Timeout of gRPC client.")
	flag.IntVar(&concurrence, "concurrence", 1, "Concurrence num.")
	flag.IntVar(&peerCount, "peercount", 1, "Peer concurrence request num.")
	flag.BoolVar(&debug, "debug", false, "Debug mode, print response.")
}

// make connserver gRPC client.
func makeClient() (*grpc.ClientConn, pb.ConnectionClient) {
	grpcOpts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithInsecure(),
		grpc.WithTimeout(timeout),
	}

	c, err := grpc.Dial(target, grpcOpts...)
	if err != nil {
		log.Fatal("can't connect, ", err)
	}
	return c, pb.NewConnectionClient(c)
}

// Stat stats the benchmark data records.
func Stat(isSuccess bool, rtime time.Time) {
	cost := uint64(common.ToMSTimestamp(time.Now()) - common.ToMSTimestamp(rtime))

	if atomic.LoadUint64(&reqCount) == 0 {
		startTime = uint64(common.ToMSTimestamp(time.Now()))
	}

	atomic.AddUint64(&rtimeCount, cost)
	atomic.AddUint64(&reqCount, 1)

	if isSuccess {
		atomic.AddUint64(&succCount, 1)
	} else {
		atomic.AddUint64(&failCount, 1)
	}

	for {
		max := atomic.LoadUint64(&maxRtime)

		if cost > max {
			if atomic.CompareAndSwapUint64(&maxRtime, max, cost) {
				break
			}
		} else {
			break
		}
	}

	for {
		min := atomic.LoadUint64(&minRtime)

		if cost < min || min == 0 {
			if atomic.CompareAndSwapUint64(&minRtime, min, cost) {
				break
			}
		} else {
			break
		}
	}
}

// Done prints benchmark result in a stupid mod.
func Done() {
	stopTime = uint64(common.ToMSTimestamp(time.Now()))

	totalCost := stopTime - startTime
	if totalCost == 0 {
		totalCost = 1
	}
	if reqCount == 0 {
		reqCount = 1
	}
	fmt.Println("\n压测完成, 结果如下:")

	fmt.Printf("	并发数:%d  请求量:%d  耗时:%d秒  QPS:%d  成功数量:%d  失败数量:%d  "+
		"平均响应时间:%d毫秒  最小响应时间:%d毫秒  最大响应时间:%d毫秒\n",
		concurrence, reqCount, totalCost/1000, reqCount*1000/totalCost, succCount, failCount,
		rtimeCount/reqCount, minRtime, maxRtime)
}

// genQueryMetadataCmd generates a query metadata flow benchmark command.
func genQueryMetadataCmd() *cobra.Command {
	// #lizard forgives
	r := pb.QueryAppMetadataReq{Seq: "benchmark-" + common.Sequence()}

	f := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	f.StringVar(&r.BizId, "biz_id", "", "Business id.")
	f.StringVar(&r.AppId, "app_id", "", "Application id.")

	cmd := &cobra.Command{
		Use:   "QueryAppMetadata",
		Short: "Query Target App Metadata Flow Benchmark command.",

		// benchmark done.
		PostRun: func(cmd *cobra.Command, args []string) {
			Done()
		},

		// benchmark.
		Run: func(cmd *cobra.Command, args []string) {
			wg := sync.WaitGroup{}
			wg.Add(concurrence * peerCount)

			// concurrences in single coroutine.
			for i := 0; i < concurrence; i++ {

				// new coroutine.
				go func(r pb.QueryAppMetadataReq) {
					// single client for coroutine.
					conn, client := makeClient()
					defer conn.Close()

					// core func call.
					call := func() {
						ctx, cancel := context.WithTimeout(context.Background(), timeout)
						defer cancel()

						r.Seq = common.Sequence()
						rtime := time.Now()

						resp, err := client.QueryAppMetadata(ctx, &r)
						if err != nil {
							Stat(false, rtime)
							log.Printf("can't query metadata, %+v", err)
							return
						}

						if resp.Code != pbcommon.ErrCode_E_OK {
							Stat(false, rtime)
							log.Printf("can't query metadata, %s", resp.Message)
							return
						}

						Stat(true, rtime)
						if debug {
							log.Printf("Command: query metadata response, %+v", resp)
						}
					}

					// benchmark.
					for j := 0; j < peerCount; j++ {
						call()
						wg.Done()
					}
				}(r)
			}

			// done.
			wg.Wait()
		},
	}

	// flags.
	cmd.Flags().AddGoFlag(flag.CommandLine.Lookup("target"))
	cmd.Flags().AddGoFlag(flag.CommandLine.Lookup("timeout"))
	cmd.Flags().AddGoFlag(flag.CommandLine.Lookup("concurrence"))
	cmd.Flags().AddGoFlag(flag.CommandLine.Lookup("peercount"))
	cmd.Flags().AddGoFlag(flag.CommandLine.Lookup("debug"))
	cmd.Flags().AddGoFlag(f.Lookup("biz_id"))
	cmd.MarkFlagRequired("biz_id")
	cmd.Flags().AddGoFlag(f.Lookup("app_id"))
	cmd.MarkFlagRequired("app_id")

	return cmd
}

// genConfigListCmd generates a query config list flow benchmark command.
func genConfigListCmd() *cobra.Command {
	// #lizard forgives
	r := pb.PullConfigListReq{Seq: "benchmark-" + common.Sequence()}

	var returnTotal bool
	var start int
	var limit int

	f := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	f.StringVar(&r.BizId, "biz_id", "", "Business id.")
	f.StringVar(&r.AppId, "app_id", "", "Application id.")
	f.BoolVar(&returnTotal, "return_total", false, "Query page return total flag.")
	f.IntVar(&start, "start", 0, "Query page start.")
	f.IntVar(&limit, "limit", 100, "Query page limit.")

	r.Page = &pbcommon.Page{ReturnTotal: returnTotal, Start: int32(start), Limit: int32(limit)}

	cmd := &cobra.Command{
		Use:   "PullConfigList",
		Short: "Pull Config List of Target App Flow Benchmark command.",

		// benchmark done.
		PostRun: func(cmd *cobra.Command, args []string) {
			Done()
		},

		// benchmark.
		Run: func(cmd *cobra.Command, args []string) {
			wg := sync.WaitGroup{}
			wg.Add(concurrence * peerCount)

			// concurrences in single coroutine.
			for i := 0; i < concurrence; i++ {

				// new coroutine.
				go func(r pb.PullConfigListReq) {
					// single client for coroutine.
					conn, client := makeClient()
					defer conn.Close()

					// core func call.
					call := func() {
						ctx, cancel := context.WithTimeout(context.Background(), timeout)
						defer cancel()

						r.Seq = common.Sequence()
						rtime := time.Now()

						resp, err := client.PullConfigList(ctx, &r)
						if err != nil {
							Stat(false, rtime)
							log.Printf("can't pull config list, %+v", err)
							return
						}

						if resp.Code != pbcommon.ErrCode_E_OK {
							Stat(false, rtime)
							log.Printf("can't pull config list, %s", resp.Message)
							return
						}

						Stat(true, rtime)
						if debug {
							log.Printf("Command: pull config list response, %+v", resp)
						}
					}

					// benchmark.
					for j := 0; j < peerCount; j++ {
						call()
						wg.Done()
					}
				}(r)
			}

			// done.
			wg.Wait()
		},
	}

	// flags.
	cmd.Flags().AddGoFlag(flag.CommandLine.Lookup("target"))
	cmd.Flags().AddGoFlag(flag.CommandLine.Lookup("timeout"))
	cmd.Flags().AddGoFlag(flag.CommandLine.Lookup("concurrence"))
	cmd.Flags().AddGoFlag(flag.CommandLine.Lookup("peercount"))
	cmd.Flags().AddGoFlag(flag.CommandLine.Lookup("debug"))
	cmd.Flags().AddGoFlag(f.Lookup("biz_id"))
	cmd.MarkFlagRequired("biz_id")
	cmd.Flags().AddGoFlag(f.Lookup("app_id"))
	cmd.MarkFlagRequired("app_id")
	cmd.Flags().AddGoFlag(f.Lookup("return_total"))
	cmd.Flags().AddGoFlag(f.Lookup("start"))
	cmd.Flags().AddGoFlag(f.Lookup("limit"))

	return cmd
}

// genReportCmd generates a report flow benchmark command.
func genReportCmd() *cobra.Command {
	// #lizard forgives
	r := pb.ReportReq{Seq: "benchmark-" + common.Sequence()}

	var cfgID string
	var releaseID string

	f := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	f.StringVar(&r.BizId, "biz_id", "", "Business id.")
	f.StringVar(&r.AppId, "app_id", "", "Application id.")
	f.StringVar(&r.CloudId, "cloud_id", "", "Sidecar tag.")
	f.StringVar(&r.Ip, "ip", "", "Sidecar ip.")
	f.StringVar(&r.Path, "path", "", "Config effect cache path.")
	f.StringVar(&r.Labels, "labels", "", "Sidecar labels.")
	f.StringVar(&cfgID, "cfg_id", "", "Config id.")
	f.StringVar(&releaseID, "release_id", "", "Release id.")

	r.Infos = []*pbcommon.ReportInfo{&pbcommon.ReportInfo{
		CfgId:      cfgID,
		ReleaseId:  releaseID,
		EffectCode: types.EffectCodeSuccess,
		EffectMsg:  types.EffectMsgSuccess,
		EffectTime: time.Now().Format("2006-01-02 15:04:05"),
	}}

	cmd := &cobra.Command{
		Use:   "Report",
		Short: "Report Flow Benchmark command.",

		// benchmark done.
		PostRun: func(cmd *cobra.Command, args []string) {
			Done()
		},

		// benchmark.
		Run: func(cmd *cobra.Command, args []string) {
			wg := sync.WaitGroup{}
			wg.Add(concurrence * peerCount)

			// concurrences in single coroutine.
			for i := 0; i < concurrence; i++ {

				// new coroutine.
				go func(r pb.ReportReq) {
					// single client for coroutine.
					conn, client := makeClient()
					defer conn.Close()

					// core func call.
					call := func() {
						ctx, cancel := context.WithTimeout(context.Background(), timeout)
						defer cancel()

						r.Seq = common.Sequence()
						rtime := time.Now()

						resp, err := client.Report(ctx, &r)
						if err != nil {
							Stat(false, rtime)
							log.Printf("can't report, %+v", err)
							return
						}

						if resp.Code != pbcommon.ErrCode_E_OK {
							Stat(false, rtime)
							log.Printf("can't report, %s", resp.Message)
							return
						}

						Stat(true, rtime)
						if debug {
							log.Printf("Command: report response, %+v", resp)
						}
					}

					// benchmark.
					for j := 0; j < peerCount; j++ {
						call()
						wg.Done()
					}
				}(r)
			}

			// done.
			wg.Wait()
		},
	}

	// flags.
	cmd.Flags().AddGoFlag(flag.CommandLine.Lookup("target"))
	cmd.Flags().AddGoFlag(flag.CommandLine.Lookup("timeout"))
	cmd.Flags().AddGoFlag(flag.CommandLine.Lookup("concurrence"))
	cmd.Flags().AddGoFlag(flag.CommandLine.Lookup("peercount"))
	cmd.Flags().AddGoFlag(flag.CommandLine.Lookup("debug"))
	cmd.Flags().AddGoFlag(f.Lookup("biz_id"))
	cmd.MarkFlagRequired("biz_id")
	cmd.Flags().AddGoFlag(f.Lookup("app_id"))
	cmd.MarkFlagRequired("app_id")
	cmd.Flags().AddGoFlag(f.Lookup("cloud_id"))
	cmd.MarkFlagRequired("cloud_id")
	cmd.Flags().AddGoFlag(f.Lookup("ip"))
	cmd.MarkFlagRequired("ip")
	cmd.Flags().AddGoFlag(f.Lookup("path"))
	cmd.MarkFlagRequired("path")
	cmd.Flags().AddGoFlag(f.Lookup("labels"))
	cmd.Flags().AddGoFlag(f.Lookup("cfg_id"))
	cmd.MarkFlagRequired("cfg_id")
	cmd.Flags().AddGoFlag(f.Lookup("release_id"))
	cmd.MarkFlagRequired("release_id")

	return cmd
}

// genTargetReleaseCmd generates a query target release flow benchmark command.
func genTargetReleaseCmd() *cobra.Command {
	// #lizard forgives
	r := pb.PullReleaseReq{Seq: "benchmark-" + common.Sequence()}

	f := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	f.StringVar(&r.BizId, "biz_id", "", "Business id.")
	f.StringVar(&r.AppId, "app_id", "", "Application id.")
	f.StringVar(&r.CloudId, "cloud_id", "", "Sidecar tag.")
	f.StringVar(&r.Ip, "ip", "", "Sidecar ip.")
	f.StringVar(&r.Path, "path", "", "Config effect cache path.")
	f.StringVar(&r.Labels, "labels", "", "Sidecar labels.")
	f.StringVar(&r.CfgId, "cfg_id", "", "Config id.")
	f.StringVar(&r.ReleaseId, "release_id", "", "Target release id.")

	cmd := &cobra.Command{
		Use:   "TargetRelease",
		Short: "Target Release Flow Benchmark command.",

		// benchmark done.
		PostRun: func(cmd *cobra.Command, args []string) {
			Done()
		},

		// benchmark.
		Run: func(cmd *cobra.Command, args []string) {
			wg := sync.WaitGroup{}
			wg.Add(concurrence * peerCount)

			// concurrences in single coroutine.
			for i := 0; i < concurrence; i++ {

				// new coroutine.
				go func(r pb.PullReleaseReq) {
					// single client for coroutine.
					conn, client := makeClient()
					defer conn.Close()

					// core func call.
					call := func() {
						ctx, cancel := context.WithTimeout(context.Background(), timeout)
						defer cancel()

						r.Seq = common.Sequence()
						rtime := time.Now()

						resp, err := client.PullRelease(ctx, &r)
						if err != nil {
							Stat(false, rtime)
							log.Printf("can't pull release, %+v", err)
							return
						}

						if resp.Code != pbcommon.ErrCode_E_OK {
							Stat(false, rtime)
							log.Printf("can't pull release, %s", resp.Message)
							return
						}

						Stat(true, rtime)
						if debug {
							log.Printf("Command: pull release response, %+v", resp)
						}
					}

					// benchmark.
					for j := 0; j < peerCount; j++ {
						call()
						wg.Done()
					}
				}(r)
			}

			// done.
			wg.Wait()
		},
	}

	// flags.
	cmd.Flags().AddGoFlag(flag.CommandLine.Lookup("target"))
	cmd.Flags().AddGoFlag(flag.CommandLine.Lookup("timeout"))
	cmd.Flags().AddGoFlag(flag.CommandLine.Lookup("concurrence"))
	cmd.Flags().AddGoFlag(flag.CommandLine.Lookup("peercount"))
	cmd.Flags().AddGoFlag(flag.CommandLine.Lookup("debug"))
	cmd.Flags().AddGoFlag(f.Lookup("biz_id"))
	cmd.MarkFlagRequired("biz_id")
	cmd.Flags().AddGoFlag(f.Lookup("app_id"))
	cmd.MarkFlagRequired("app_id")
	cmd.Flags().AddGoFlag(f.Lookup("cloud_id"))
	cmd.MarkFlagRequired("cloud_id")
	cmd.Flags().AddGoFlag(f.Lookup("ip"))
	cmd.MarkFlagRequired("ip")
	cmd.Flags().AddGoFlag(f.Lookup("path"))
	cmd.MarkFlagRequired("path")
	cmd.Flags().AddGoFlag(f.Lookup("labels"))
	cmd.Flags().AddGoFlag(f.Lookup("cfg_id"))
	cmd.MarkFlagRequired("cfg_id")
	cmd.Flags().AddGoFlag(f.Lookup("release_id"))
	cmd.MarkFlagRequired("release_id")

	return cmd
}

// genNewestReleaseCmd generates a query newest release flow benchmark command.
func genNewestReleaseCmd() *cobra.Command {
	// #lizard forgives
	r := pb.PullReleaseReq{Seq: "benchmark-" + common.Sequence()}

	f := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	f.StringVar(&r.BizId, "biz_id", "", "Business id.")
	f.StringVar(&r.AppId, "app_id", "", "Application id.")
	f.StringVar(&r.CloudId, "cloud_id", "", "Sidecar tag.")
	f.StringVar(&r.Ip, "ip", "", "Sidecar ip.")
	f.StringVar(&r.Path, "path", "", "Config effect cache path.")
	f.StringVar(&r.Labels, "labels", "", "Sidecar labels.")
	f.StringVar(&r.CfgId, "cfg_id", "", "Config id.")
	f.StringVar(&r.LocalReleaseId, "local_release_id", "", "Local release id.")

	cmd := &cobra.Command{
		Use:   "NewestRelease",
		Short: "Newest Release Flow Benchmark command.",

		// benchmark done.
		PostRun: func(cmd *cobra.Command, args []string) {
			Done()
		},

		// benchmark.
		Run: func(cmd *cobra.Command, args []string) {
			wg := sync.WaitGroup{}
			wg.Add(concurrence * peerCount)

			// concurrences in single coroutine.
			for i := 0; i < concurrence; i++ {

				// new coroutine.
				go func(r pb.PullReleaseReq) {
					// single client for coroutine.
					conn, client := makeClient()
					defer conn.Close()

					// core func call.
					call := func() {
						ctx, cancel := context.WithTimeout(context.Background(), timeout)
						defer cancel()

						r.Seq = common.Sequence()
						rtime := time.Now()

						resp, err := client.PullRelease(ctx, &r)
						if err != nil {
							Stat(false, rtime)
							log.Printf("can't pull newest release, %+v", err)
							return
						}

						if resp.Code != pbcommon.ErrCode_E_OK {
							Stat(false, rtime)
							log.Printf("can't pull newest release, %s", resp.Message)
							return
						}

						Stat(true, rtime)
						if debug {
							log.Printf("Command: pull newest release response, %+v", resp)
						}
					}

					// benchmark.
					for j := 0; j < peerCount; j++ {
						call()
						wg.Done()
					}
				}(r)
			}

			// done.
			wg.Wait()
		},
	}

	// flags.
	cmd.Flags().AddGoFlag(flag.CommandLine.Lookup("target"))
	cmd.Flags().AddGoFlag(flag.CommandLine.Lookup("timeout"))
	cmd.Flags().AddGoFlag(flag.CommandLine.Lookup("concurrence"))
	cmd.Flags().AddGoFlag(flag.CommandLine.Lookup("peercount"))
	cmd.Flags().AddGoFlag(flag.CommandLine.Lookup("debug"))
	cmd.Flags().AddGoFlag(f.Lookup("biz_id"))
	cmd.MarkFlagRequired("biz_id")
	cmd.Flags().AddGoFlag(f.Lookup("app_id"))
	cmd.MarkFlagRequired("app_id")
	cmd.Flags().AddGoFlag(f.Lookup("cloud_id"))
	cmd.MarkFlagRequired("cloud_id")
	cmd.Flags().AddGoFlag(f.Lookup("ip"))
	cmd.MarkFlagRequired("ip")
	cmd.Flags().AddGoFlag(f.Lookup("path"))
	cmd.MarkFlagRequired("path")
	cmd.Flags().AddGoFlag(f.Lookup("labels"))
	cmd.Flags().AddGoFlag(f.Lookup("cfg_id"))
	cmd.MarkFlagRequired("cfg_id")
	cmd.Flags().AddGoFlag(f.Lookup("local_release_id"))
	cmd.MarkFlagRequired("local_release_id")

	return cmd
}

// genSubCmds returns sub commands.
func genSubCmds() []*cobra.Command {
	cmds := []*cobra.Command{}
	cmds = append(cmds, genQueryMetadataCmd())
	cmds = append(cmds, genConfigListCmd())
	cmds = append(cmds, genReportCmd())
	cmds = append(cmds, genTargetReleaseCmd())
	cmds = append(cmds, genNewestReleaseCmd())
	return cmds
}

// bscp benchmark tool.
func main() {
	// root command.
	rootCmd := &cobra.Command{Use: "bk-bscp-benchmark-tool"}

	// sub commands.
	subCmds := genSubCmds()

	// add sub commands.
	rootCmd.AddCommand(subCmds...)

	// run root command.
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
