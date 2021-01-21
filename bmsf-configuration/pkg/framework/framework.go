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

package framework

import (
	"flag"
	"os"
	"runtime/debug"
	"time"

	"bk-bscp/pkg/common"
)

var (
	// real service instance.
	instance Service

	// system pprof flags.
	httpprof, cpuprofile, memprofile string

	// config filepath flag.
	configfile string

	// cpuprofile file.
	cpuprofileOut *os.File
)

// Setting is settings for server.
type Setting struct {
	Configfile string
}

// Service is service abstraction.
type Service interface {
	// Init init the service instance.
	Init(setting Setting)

	// Run runs the service instance.
	Run()

	// Stop stops the service instance.
	Stop()
}

func init() {
	flag.StringVar(&httpprof, "httpprof", "", "Setup a pprof httpserver by the endpoint.")
	flag.StringVar(&cpuprofile, "cpuprofile", "", "Dump cpu profile info to the file (eg:pprof/cpuprofile.pprof).")
	flag.StringVar(&memprofile, "memprofile", "", "Dump memory profile to the file (eg:pprof/memprofile.pprof).")
	flag.StringVar(&configfile, "configfile", "./etc/server.yaml", "The config path of server.")
}

// beforeRun does something before run services.
func beforeRun() {
	// handle os signals.
	common.HandleSignals(func() {
		cleanup()

		// TODO alert
	})

	// setup pprof services.
	if httpprof != "" {
		common.SetupHTTPPprof(httpprof)
	}

	if cpuprofile != "" {
		common.SetupCPUPprof(cpuprofile, &cpuprofileOut)
	}
}

// cleanup handles cleaning up environments.
func cleanup() {
	// stop service.
	if instance != nil {
		instance.Stop()
	}

	// collect cpu pprof data.
	if cpuprofile != "" {
		common.CollectCPUPprofData(cpuprofileOut)
	}

	// collect memory pprof data.
	if memprofile != "" {
		common.CollectMemPprofData(memprofile)
	}

	// wait for the async cleanup.
	time.Sleep(time.Second)
}

// realRun runs real services.
func realRun(service Service) {
	setting := Setting{
		Configfile: configfile,
	}
	instance = service

	// run service.
	instance.Init(setting)
	instance.Run()
}

// Run runs services.
func Run(service Service) {
	// clean up the runtime and OS environments when shutdown.
	defer cleanup()

	// recover service from panicking.
	defer func() {
		if err := recover(); err != nil {
			debug.PrintStack()
		}
	}()

	// do something before run the real services,
	// register the OS signals and setup the pprof services.
	beforeRun()

	// run services in main gCoroutine.
	realRun(service)
}
