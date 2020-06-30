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

package exec

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	v1 "github.com/Tencent/bk-bcs/bcs-services/bcs-client/pkg/storage/v1"
	//"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/utils"
	v4 "github.com/Tencent/bk-bcs/bcs-services/bcs-client/pkg/scheduler/v4"
	"github.com/docker/docker/pkg/signal"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	gosignal "os/signal"
	"runtime"
	"time"

	"context"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/exec/streams"
	//v1 "github.com/Tencent/bk-bcs/bcs-services/bcs-client/pkg/storage/v1"
	"fmt"
	//"github.com/docker/docker/api/types"
	"github.com/moby/term"
	"github.com/urfave/cli"
)

// Streams is an interface which exposes the standard input and output streams
type Streams interface {
	In() *streams.In
	Out() *streams.Out
	Err() io.Writer
}

type execOptions struct {
	tty         bool
	interactive bool
	clusterId   string
	container   string
	command     []string
}

type ExecCli struct {
	scheduler v4.Scheduler
	ClusterId string
	ExecId    string
	HostIp    string
	in        *streams.In
	out       *streams.Out
	err       io.Writer
}

// Out returns the writer used for stdout
func (cli *ExecCli) Out() *streams.Out {
	return cli.out
}

// In returns the reader used for stdin
func (cli *ExecCli) In() *streams.In {
	return cli.in
}

// Err returns the writer used for stderr
func (cli *ExecCli) Err() io.Writer {
	return cli.err
}

//NewExecCommand sub command exec registration
func NewExecCommand() cli.Command {
	return cli.Command{
		Name:  "exec",
		Usage: "exec [OPTIONS] CONTAINER COMMAND [ARG...]",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "clusterid",
				Usage: "Cluster ID",
			},
			cli.BoolFlag{
				Name:  "tty, t",
				Usage: "Allocate a pseudo-TTY",
			},
			cli.BoolFlag{
				Name:  "interactive, i",
				Usage: "Keep STDIN open even if not attached",
			},
			cli.StringFlag{
				Name:  "container, c",
				Usage: "Container name. If omitted, the first container in the taskgroup will be chosen",
			},
			cli.StringFlag{
				Name:  "namespace, ns",
				Usage: "Namespace",
				Value: "",
			},
		},
		Action: func(c *cli.Context) error {
			return exec(utils.NewClientContext(c))
		},
	}
}

func exec(c *utils.ClientContext) error {
	if err := c.MustSpecified(utils.OptionClusterID, utils.OptionNamespace); err != nil {
		return err
	}

	args := c.Args()
	if len(args) < 2 {
		return fmt.Errorf("require at least 2 args, first one is pod name, others are commands")
	}

	// inspect the taskgroup to get the hostIp of the taskgroup
	storage := v1.NewBcsStorage(utils.GetClientOption())
	single, err := storage.InspectTaskGroup(c.ClusterID(), c.Namespace(), c.Args()[0])
	if err != nil {
		return fmt.Errorf("failed to inspect taskgroup: %v", err)
	}
	if single.Data.Status != types.Pod_Running {
		return fmt.Errorf("can't exec into a pod whose status is not running")
	}

	var containerId, hostIp string
	hostIp = single.Data.HostIP
	containerName := c.String("container")
	if len(containerName) == 0 {
		if len(single.Data.ContainerStatuses) > 0 {
			usageString := fmt.Sprintf("Use the first container in pod, container name is %s.", single.Data.ContainerStatuses[0].Name)
			containerId = single.Data.ContainerStatuses[0].ContainerID
			fmt.Println(usageString)
		}
	} else {
		for _, c := range single.Data.ContainerStatuses {
			if c.Name == containerName {
				containerId = c.ContainerID
				break
			}
		}
		if containerId == "" {
			return fmt.Errorf("container name invalid, please check your container name")
		}
	}

	// call the consoleproxy create_exec api
	scheduler := v4.NewBcsScheduler(utils.GetClientOption())
	execId, err := scheduler.CreateContainerExec(c.ClusterID(), containerId, hostIp, c.Args()[1:])
	if err != nil {
		return fmt.Errorf("failed to create container exec: %v", err)
	}

	stdin, stdout, stderr := term.StdStreams()
	cli := &ExecCli{
		in:        streams.NewIn(stdin),
		out:       streams.NewOut(stdout),
		err:       stderr,
		scheduler: scheduler,
		ClusterId: c.ClusterID(),
		ExecId:    execId,
		HostIp:    hostIp,
	}

	// call the consoleproxy start_exec api
	ctx := context.Background()
	resp, err := scheduler.StartContainerExec(ctx, c.ClusterID(), execId, containerId, hostIp)
	if err != nil {
		return err
	}
	defer resp.Close()

	errCh := make(chan error, 1)

	go func() {
		defer close(errCh)
		errCh <- func() error {
			streamer := hijackedIOStreamer{
				streams:      cli,
				inputStream:  cli.In(),
				outputStream: cli.Out(),
				errorStream:  cli.Out(),
				resp:         resp,
				tty:          c.Bool("tty"),
			}

			return streamer.stream(ctx)
		}()
	}()

	// resize the terminal
	if c.Bool("tty") && cli.In().IsTerminal() {
		if err := MonitorTtySize(ctx, *cli, true); err != nil {
			fmt.Fprintln(cli.Err(), "Error monitoring TTY size:", err)
		}
	}

	if err := <-errCh; err != nil {
		return err
	}

	return nil
}

// MonitorTtySize updates the container tty size when the terminal tty changes size
func MonitorTtySize(ctx context.Context, cli ExecCli, isExec bool) error {
	initTtySize(ctx, cli, isExec, resizeTty)
	if runtime.GOOS == "windows" {
		go func() {
			prevH, prevW := cli.Out().GetTtySize()
			for {
				time.Sleep(time.Millisecond * 250)
				h, w := cli.Out().GetTtySize()

				if prevW != w || prevH != h {
					resizeTty(ctx, cli, isExec)
				}
				prevH = h
				prevW = w
			}
		}()
	} else {
		sigchan := make(chan os.Signal, 1)
		gosignal.Notify(sigchan, signal.SIGWINCH)
		go func() {
			for range sigchan {
				resizeTty(ctx, cli, isExec)
			}
		}()
	}
	return nil
}

// initTtySize is to init the tty's size to the same as the window, if there is an error, it will retry 5 times.
func initTtySize(ctx context.Context, cli ExecCli, isExec bool, resizeTtyFunc func(ctx context.Context, cli ExecCli, isExec bool) error) {
	rttyFunc := resizeTtyFunc
	if rttyFunc == nil {
		rttyFunc = resizeTty
	}
	if err := rttyFunc(ctx, cli, isExec); err != nil {
		go func() {
			var err error
			for retry := 0; retry < 5; retry++ {
				time.Sleep(10 * time.Millisecond)
				if err = rttyFunc(ctx, cli, isExec); err == nil {
					break
				}
			}
			if err != nil {
				fmt.Fprintln(cli.Err(), "failed to resize tty, using default size")
			}
		}()
	}
}

// resizeTty is to resize the tty with cli out's tty size
func resizeTty(ctx context.Context, cli ExecCli, isExec bool) error {
	height, width := cli.Out().GetTtySize()
	return resizeTtyTo(ctx, cli, height, width, isExec)
}

// resizeTtyTo resizes tty to specific height and width
func resizeTtyTo(ctx context.Context, cli ExecCli, height, width uint, isExec bool) error {
	if height == 0 && width == 0 {
		return nil
	}

	err := cli.scheduler.ResizeContainerExec(cli.ClusterId, cli.ExecId, cli.HostIp, int(height), int(width))

	if err != nil {
		logrus.Debugf("Error resize: %s\r", err)
	}
	return err
}
