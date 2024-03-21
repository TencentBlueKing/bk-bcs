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

package controller

import (
	"context"
	"fmt"
	"net/http"
	"net/http/pprof"
	"runtime/debug"
	"sync"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/options"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/pkg/apis"
	deschedulev1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/pkg/apis/tkex/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/pkg/controller/cachemanager"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/pkg/controller/migrator"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/pkg/controller/sched/extender"
)

type ControllerManager struct {
	op *options.DeSchedulerOption

	mgr          manager.Manager
	cacheManager cachemanager.CacheInterface
	httpExtender extender.HttpExtenderInterface
	migrator     migrator.DescheduleMigratorInterface

	client client.Client
	scheme *runtime.Scheme

	internalWg     *sync.WaitGroup
	internalCtx    context.Context
	internalCancel context.CancelFunc
}

// NewControllerManager create the instance of ControllerManager
func NewControllerManager() *ControllerManager {
	op := options.GlobalConfigHandler().GetOptions()
	return &ControllerManager{
		op:           op,
		migrator:     migrator.GlobalMigratorManager(),
		httpExtender: extender.NewHTTPExtender(),
		cacheManager: cachemanager.NewCacheManager(),
	}
}

// Init all the controllers with config options.
func (m *ControllerManager) Init() error {
	// init cache informer for kubernetes
	if err := m.cacheManager.Init(); err != nil {
		return errors.Wrapf(err, "init cache manager failed")
	}

	// init controller runtime manager
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(deschedulev1alpha1.AddToScheme(scheme))
	var err error
	m.mgr, err = ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Host:               m.op.Address,
		Port:               int(m.op.HttpPort),
		CertDir:            m.op.WebhookCertDir,
		Scheme:             scheme,
		LeaderElection:     true,
		LeaderElectionID:   apis.ElectionID,
		MetricsBindAddress: fmt.Sprintf("%s:%d", m.op.Address, m.op.MetricPort),
	})
	if err != nil {
		return errors.Wrapf(err, "unable to create manager")
	}

	m.client = m.mgr.GetClient()
	m.scheme = m.mgr.GetScheme()

	// init http extender for scheduler
	if err := m.httpExtender.Init(); err != nil {
		return errors.Wrapf(err, "init http extender failed")
	}

	// setup with manager
	if err := ctrl.NewControllerManagedBy(m.mgr).
		For(&deschedulev1alpha1.DeschedulePolicy{}).WithEventFilter(m.predicate()).
		Complete(m); err != nil {
		return errors.Wrapf(err, "init deschedule controller failed")
	}
	if err = m.initWebhook(); err != nil {
		return errors.Wrapf(err, "init webhook server failed")
	}
	blog.Infof("init deschedule controller success.")
	return nil
}

func (m *ControllerManager) initWebhook() error {
	annotator := &deschedulev1alpha1.DeschedulePolicyAnnotator{}
	if err := annotator.SetupWebhookWithManager(m.mgr); err != nil {
		return errors.Wrapf(err, "init deschedule webhook failed")
	}
	annotator.RegisterValidateCreate(m.RegisterValidateCreate)
	annotator.RegisterValidateUpdate(m.RegisterValidateUpdate)
	hookServer := m.mgr.GetWebhookServer()
	hookServer.CertName = m.op.WebhookCertName
	hookServer.KeyName = m.op.WebhookKeyName
	if hookServer.CertDir == "" || hookServer.CertName == "" || hookServer.KeyName == "" {
		return errors.Errorf("config webhook certDir/certName/keyName cannot be empty.")
	}
	//m.initPProf(hookServer.WebhookMux)
	blog.Infof("init deschedule webhook success, address: %s:%d.", hookServer.Host, hookServer.Port)
	return nil
}

func (m *ControllerManager) initPProf(mux *http.ServeMux) {
	if !m.op.Debug {
		blog.Infof("pprof is disabled")
		return
	}
	blog.Infof("pprof is enabled")
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
}

// runHTPExtender start the http extender which is the extender of k8s scheduler. It will
// handle the schedule with Filter/Prioritize/ProcessPreemption
func (m *ControllerManager) runHTPExtender(errChan chan error) {
	var err error
	defer func() {
		m.internalWg.Done()
		if r := recover(); r != nil {
			err = fmt.Errorf("http_extender panic, err: %v, stack:\n%s", r, string(debug.Stack()))
		}
		if err != nil {
			blog.Errorf("http_extender exited: %s", err.Error())
			errChan <- err
		} else {
			blog.Infof("http_extender is stopped.")
		}
	}()

	blog.Infof("http_extender is started.")
	if err = m.httpExtender.Run(m.internalCtx); err != nil {
		err = errors.Wrapf(err, "http_extender exit with error")
	}
}

// runCacheManager will start some informers for k8s. The resource of k8s will
// cache at local.
func (m *ControllerManager) runCacheManager(errChan chan error) {
	var err error
	defer func() {
		m.internalWg.Done()
		if r := recover(); r != nil {
			err = errors.Errorf("cache_manager panic, err: %v, stack:\n%s", r, string(debug.Stack()))
		}
		if err != nil {
			blog.Errorf("cache_manager exited: %s", err.Error())
			errChan <- err
		} else {
			blog.Infof("cache_manager is stopped.")
		}
	}()

	blog.Infof("cache_manager is started.")
	if err = m.cacheManager.Start(m.internalCtx); err != nil {
		err = errors.Wrapf(err, "cache_manager exit with error")
	}
}

// runControllerManager will start the manager of controller-runtime, it will handle
// the event of DeschedulePolicy.
func (m *ControllerManager) runControllerManager(errChan chan error) {
	var err error
	defer func() {
		m.internalWg.Done()
		if r := recover(); r != nil {
			err = errors.Errorf("controller_manager panic, err: %v, stack:\n%s", r, string(debug.Stack()))
		}
		if err != nil {
			blog.Errorf("controller_manager exited: %s", err.Error())
			errChan <- err
		} else {
			blog.Infof("controller_manager is stopped.")
		}
	}()

	blog.Infof("controller_manager is started.")
	if err = m.mgr.Start(m.internalCtx); err != nil {
		err = errors.Wrapf(err, "controller_manager exit with error")
	}
}

// run will run goroutines and add wait group for them
func (m *ControllerManager) run(fs ...func(errChan chan error)) chan error {
	m.internalWg = &sync.WaitGroup{}
	m.internalWg.Add(len(fs))
	errChan := make(chan error, len(fs))
	for i := range fs {
		go fs[i](errChan)
	}
	return errChan
}

func (m *ControllerManager) Run(ctx context.Context) error {
	m.internalCtx, m.internalCancel = context.WithCancel(ctx)
	errChan := m.run(m.runControllerManager, m.runCacheManager, m.runHTPExtender)
	defer func() {
		m.internalCancel()
		m.internalWg.Wait()
	}()
	for {
		select {
		case err := <-errChan:
			return errors.Wrapf(err, "server exit with error")
		case <-ctx.Done():
			return nil
		}
	}
}
