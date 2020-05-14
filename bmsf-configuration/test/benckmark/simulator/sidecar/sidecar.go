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

	// it would fucking spin here?
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

	fmt.Println("\n压测完成, 结果如下:")
	fmt.Printf("	并发数:%d  请求量:%d  耗时:%d秒  QPS:%d  成功数量:%d  失败数量:%d  平均响应时间:%d毫秒  最小响应时间:%d毫秒  最大响应时间:%d毫秒\n",
		concurrence, reqCount, (stopTime-startTime)/1000, reqCount*1000/(stopTime-startTime), succCount, failCount, rtimeCount/reqCount, minRtime, maxRtime)
}

// genTargetReleaseCmd generates a query target release flow benchmark command.
func genTargetReleaseCmd() *cobra.Command {
	r := &pb.PullReleaseReq{}

	f := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	f.StringVar(&r.Bid, "bid", "", "Business id.")
	f.StringVar(&r.Appid, "appid", "", "Application id.")
	f.StringVar(&r.Clusterid, "clusterid", "", "Sidecar cluster id.")
	f.StringVar(&r.Zoneid, "zoneid", "", "Sidecar zone id.")
	f.StringVar(&r.Dc, "dc", "", "sidecar datacenter tag.")
	f.StringVar(&r.IP, "ip", "", "sidecar ip.")
	f.StringVar(&r.Labels, "labels", "", "sidecar labels.")
	f.StringVar(&r.Cfgsetid, "cfgsetid", "", "Config set id.")
	f.StringVar(&r.Releaseid, "releaseid", "", "Release id.")

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
				go func() {
					// single client for coroutine.
					conn, client := makeClient()
					defer conn.Close()

					// core func call.
					call := func() {
						ctx, cancel := context.WithTimeout(context.Background(), timeout)
						defer cancel()

						r.Seq = common.Sequence()
						rtime := time.Now()

						resp, err := client.PullRelease(ctx, r)
						if err != nil {
							Stat(false, rtime)
							log.Printf("can't pull release, %+v", err)
							return
						}

						if resp.ErrCode != pbcommon.ErrCode_E_OK {
							Stat(false, rtime)
							log.Printf("can't pull release, %s", resp.ErrMsg)
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
				}()
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
	cmd.Flags().AddGoFlag(f.Lookup("bid"))
	cmd.MarkFlagRequired("bid")
	cmd.Flags().AddGoFlag(f.Lookup("appid"))
	cmd.MarkFlagRequired("appid")
	cmd.Flags().AddGoFlag(f.Lookup("clusterid"))
	cmd.MarkFlagRequired("clusterid")
	cmd.Flags().AddGoFlag(f.Lookup("zoneid"))
	cmd.MarkFlagRequired("zoneid")
	cmd.Flags().AddGoFlag(f.Lookup("dc"))
	cmd.MarkFlagRequired("dc")
	cmd.Flags().AddGoFlag(f.Lookup("ip"))
	cmd.MarkFlagRequired("ip")
	cmd.Flags().AddGoFlag(f.Lookup("labels"))
	cmd.Flags().AddGoFlag(f.Lookup("cfgsetid"))
	cmd.MarkFlagRequired("cfgsetid")
	cmd.Flags().AddGoFlag(f.Lookup("releaseid"))
	cmd.MarkFlagRequired("releaseid")

	return cmd
}

// genNewestReleaseCmd generates a query newest release flow benchmark command.
func genNewestReleaseCmd() *cobra.Command {
	r := &pb.PullReleaseReq{}

	f := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	f.StringVar(&r.Bid, "bid", "", "Business id.")
	f.StringVar(&r.Appid, "appid", "", "Application id.")
	f.StringVar(&r.Clusterid, "clusterid", "", "Sidecar cluster id.")
	f.StringVar(&r.Zoneid, "zoneid", "", "Sidecar zone id.")
	f.StringVar(&r.Dc, "dc", "", "sidecar datacenter tag.")
	f.StringVar(&r.IP, "ip", "", "sidecar ip.")
	f.StringVar(&r.Labels, "labels", "", "sidecar labels.")
	f.StringVar(&r.Cfgsetid, "cfgsetid", "", "Config set id.")
	f.StringVar(&r.LocalReleaseid, "lreleaseid", "", "Local release id.")

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
				go func() {
					// single client for coroutine.
					conn, client := makeClient()
					defer conn.Close()

					// core func call.
					call := func() {
						ctx, cancel := context.WithTimeout(context.Background(), timeout)
						defer cancel()

						r.Seq = common.Sequence()
						rtime := time.Now()

						resp, err := client.PullRelease(ctx, r)
						if err != nil {
							Stat(false, rtime)
							log.Printf("can't pull newest release, %+v", err)
							return
						}

						if resp.ErrCode != pbcommon.ErrCode_E_OK {
							Stat(false, rtime)
							log.Printf("can't pull newest release, %s", resp.ErrMsg)
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
				}()
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
	cmd.Flags().AddGoFlag(f.Lookup("bid"))
	cmd.MarkFlagRequired("bid")
	cmd.Flags().AddGoFlag(f.Lookup("appid"))
	cmd.MarkFlagRequired("appid")
	cmd.Flags().AddGoFlag(f.Lookup("clusterid"))
	cmd.MarkFlagRequired("clusterid")
	cmd.Flags().AddGoFlag(f.Lookup("zoneid"))
	cmd.MarkFlagRequired("zoneid")
	cmd.Flags().AddGoFlag(f.Lookup("dc"))
	cmd.MarkFlagRequired("dc")
	cmd.Flags().AddGoFlag(f.Lookup("ip"))
	cmd.MarkFlagRequired("ip")
	cmd.Flags().AddGoFlag(f.Lookup("labels"))
	cmd.Flags().AddGoFlag(f.Lookup("cfgsetid"))
	cmd.MarkFlagRequired("cfgsetid")
	cmd.Flags().AddGoFlag(f.Lookup("lreleaseid"))

	return cmd
}

// genReleaseConfigsCmd generates a query target release configs
// content flow benchmark command.
func genReleaseConfigsCmd() *cobra.Command {
	r := &pb.PullReleaseConfigsReq{}

	f := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	f.StringVar(&r.Bid, "bid", "", "Business id.")
	f.StringVar(&r.Appid, "appid", "", "Application id.")
	f.StringVar(&r.Clusterid, "clusterid", "", "Sidecar cluster id.")
	f.StringVar(&r.Zoneid, "zoneid", "", "Sidecar zone id.")
	f.StringVar(&r.Cfgsetid, "cfgsetid", "", "Config set id.")
	f.StringVar(&r.Releaseid, "releaseid", "", "Release id.")
	f.StringVar(&r.Cid, "cid", "", "Release configs cid.")

	cmd := &cobra.Command{
		Use:   "ReleaseConfigs",
		Short: "Target Release Configs Content Flow Benchmark command.",

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
				go func() {
					// single client for coroutine.
					conn, client := makeClient()
					defer conn.Close()

					// core func call.
					call := func() {
						ctx, cancel := context.WithTimeout(context.Background(), timeout)
						defer cancel()

						r.Seq = common.Sequence()
						rtime := time.Now()

						resp, err := client.PullReleaseConfigs(ctx, r)
						if err != nil {
							Stat(false, rtime)
							log.Printf("can't pull release configs, %+v", err)
							return
						}

						if resp.ErrCode != pbcommon.ErrCode_E_OK {
							Stat(false, rtime)
							log.Printf("can't pull release configs, %s", resp.ErrMsg)
							return
						}

						Stat(true, rtime)
						if debug {
							log.Printf("Command: pull release configs response, %+v", resp)
						}
					}

					// benchmark.
					for j := 0; j < peerCount; j++ {
						call()
						wg.Done()
					}
				}()
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
	cmd.Flags().AddGoFlag(f.Lookup("bid"))
	cmd.MarkFlagRequired("bid")
	cmd.Flags().AddGoFlag(f.Lookup("appid"))
	cmd.MarkFlagRequired("appid")
	cmd.Flags().AddGoFlag(f.Lookup("clusterid"))
	cmd.MarkFlagRequired("clusterid")
	cmd.Flags().AddGoFlag(f.Lookup("zoneid"))
	cmd.MarkFlagRequired("zoneid")
	cmd.Flags().AddGoFlag(f.Lookup("cfgsetid"))
	cmd.MarkFlagRequired("cfgsetid")
	cmd.Flags().AddGoFlag(f.Lookup("releaseid"))
	cmd.MarkFlagRequired("releaseid")
	cmd.Flags().AddGoFlag(f.Lookup("cid"))
	cmd.MarkFlagRequired("cid")

	return cmd
}

// genTargetReleaseAndConfigsCmd generates a query target release
// and configs content flow benchmark command.
func genTargetReleaseAndConfigsCmd() *cobra.Command {
	rRelease := &pb.PullReleaseReq{}
	rConfigs := &pb.PullReleaseConfigsReq{}

	f := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	f.StringVar(&rRelease.Bid, "bid", "", "Business id.")
	f.StringVar(&rRelease.Appid, "appid", "", "Application id.")
	f.StringVar(&rRelease.Clusterid, "clusterid", "", "Sidecar cluster id.")
	f.StringVar(&rRelease.Zoneid, "zoneid", "", "Sidecar zone id.")
	f.StringVar(&rRelease.Dc, "dc", "", "sidecar datacenter tag.")
	f.StringVar(&rRelease.IP, "ip", "", "sidecar ip.")
	f.StringVar(&rRelease.Labels, "labels", "", "sidecar labels.")
	f.StringVar(&rRelease.Cfgsetid, "cfgsetid", "", "Config set id.")
	f.StringVar(&rRelease.Releaseid, "releaseid", "", "Release id.")
	f.StringVar(&rConfigs.Cid, "cid", "", "Release cid.")

	cmd := &cobra.Command{
		Use:   "TargetReleaseAndConfigs",
		Short: "Target Release And Configs Content Flow Benchmark command.",

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
				go func() {
					conn, client := makeClient()
					defer conn.Close()

					// core func call.
					callRelease := func() {
						ctx, cancel := context.WithTimeout(context.Background(), timeout)
						defer cancel()

						rRelease.Seq = common.Sequence()
						rtime := time.Now()

						resp, err := client.PullRelease(ctx, rRelease)
						if err != nil {
							Stat(false, rtime)
							log.Printf("can't pull release, %+v", err)
							return
						}

						if resp.ErrCode != pbcommon.ErrCode_E_OK {
							Stat(false, rtime)
							log.Printf("can't pull release, %s", resp.ErrMsg)
							return
						}

						Stat(true, rtime)
						if debug {
							log.Printf("Command: pull release response, %+v", resp)
						}
					}

					// core func call.
					callConfigs := func() {
						ctx, cancel := context.WithTimeout(context.Background(), timeout)
						defer cancel()

						rConfigs.Seq = common.Sequence()
						rConfigs.Bid = rRelease.Bid
						rConfigs.Appid = rRelease.Appid
						rConfigs.Clusterid = rRelease.Clusterid
						rConfigs.Zoneid = rRelease.Zoneid
						rConfigs.Cfgsetid = rRelease.Cfgsetid
						rConfigs.Releaseid = rRelease.Releaseid

						rtime := time.Now()

						resp, err := client.PullReleaseConfigs(ctx, rConfigs)
						if err != nil {
							Stat(false, rtime)
							log.Printf("can't pull release configs, %+v", err)
							return
						}

						if resp.ErrCode != pbcommon.ErrCode_E_OK {
							Stat(false, rtime)
							log.Printf("can't pull release configs, %s", resp.ErrMsg)
							return
						}

						Stat(true, rtime)
						if debug {
							log.Printf("Command: pull release configs response, %+v", resp)
						}
					}

					// benchmark.
					for j := 0; j < peerCount; j++ {
						callRelease()
						callConfigs()
						wg.Done()
					}
				}()
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
	cmd.Flags().AddGoFlag(f.Lookup("bid"))
	cmd.MarkFlagRequired("bid")
	cmd.Flags().AddGoFlag(f.Lookup("appid"))
	cmd.MarkFlagRequired("appid")
	cmd.Flags().AddGoFlag(f.Lookup("clusterid"))
	cmd.MarkFlagRequired("clusterid")
	cmd.Flags().AddGoFlag(f.Lookup("zoneid"))
	cmd.MarkFlagRequired("zoneid")
	cmd.Flags().AddGoFlag(f.Lookup("dc"))
	cmd.MarkFlagRequired("dc")
	cmd.Flags().AddGoFlag(f.Lookup("ip"))
	cmd.MarkFlagRequired("ip")
	cmd.Flags().AddGoFlag(f.Lookup("labels"))
	cmd.Flags().AddGoFlag(f.Lookup("cfgsetid"))
	cmd.MarkFlagRequired("cfgsetid")
	cmd.Flags().AddGoFlag(f.Lookup("releaseid"))
	cmd.MarkFlagRequired("releaseid")
	cmd.Flags().AddGoFlag(f.Lookup("cid"))
	cmd.MarkFlagRequired("cid")

	return cmd
}

// genSubCmds returns sub commands.
func genSubCmds() []*cobra.Command {
	cmds := []*cobra.Command{}
	cmds = append(cmds, genTargetReleaseCmd())
	cmds = append(cmds, genNewestReleaseCmd())
	cmds = append(cmds, genReleaseConfigsCmd())
	cmds = append(cmds, genTargetReleaseAndConfigsCmd())
	return cmds
}

// bscp simulator sidecar.
func main() {
	// root command.
	rootCmd := &cobra.Command{Use: "bscp-simulator-sidecar"}

	// sub commands.
	subCmds := genSubCmds()

	// add sub commands.
	rootCmd.AddCommand(subCmds...)

	// run root command.
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
