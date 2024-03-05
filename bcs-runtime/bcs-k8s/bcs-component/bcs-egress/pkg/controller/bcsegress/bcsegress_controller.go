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

// Package bcsegress xxx
package bcsegress

import (
	"context"
	"fmt"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	bkbcsv1alpha1 "bcs-egress/pkg/apis/bkbcs/v1alpha1"
)

const (
	labelReference           = "bcsegress"
	defaultReconcileInterval = time.Second * 5
)

// Add creates a new BCSEgress Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, option *EgressOption) error {
	recon := NewBCSEgressReconciler(mgr, option)
	if recon == nil {
		return fmt.Errorf("init BCSEgressController err")
	}
	// ready to init Reconciler
	return recon.Init(mgr)
}

// NewBCSEgressReconciler returns a new reconcile.Reconciler
// *config: configuration file for Reconciler
func NewBCSEgressReconciler(mgr manager.Manager, option *EgressOption) *ReconcileBCSEgress {
	p, err := NewNginx(option)
	if err != nil {
		klog.Errorf("init BCSEgressReconciler failed, %s", err.Error())
		return nil
	}
	egress := &ReconcileBCSEgress{
		client:   mgr.GetClient(),
		scheme:   mgr.GetScheme(),
		option:   option,
		identity: fmt.Sprintf("%s.%s", option.Name, option.Namespace),
		proxy:    p,
	}
	return egress
}

// blank assignment to verify that ReconcileBCSEgress implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileBCSEgress{}
var _ predicate.Predicate = &ReconcileBCSEgress{}

// ReconcileBCSEgress reconciles a BCSEgress object
type ReconcileBCSEgress struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
	option *EgressOption
	// identity comes from namespace/name in ENV
	identity string
	// proxy is interface for network flow control
	proxy Proxy
	// LastError for BCSEgress error recording
	// * key is reconcile.Request.String()
	// lastError map[string]error
	// local cache for diffing nginx rules
	// httpsList map[string]string
	// tcpPortList use for resolving tcp port conflict
	// its principle is first come first serves. egress will
	// drop all later conflict ports.
	// *key is port, value is namespace/name of BCSEgress
	// tcpPortList map[string]string
	stop     context.Context
	stopFunc context.CancelFunc
}

// Reconcile reads that state of the cluster for a BCSEgress object and makes changes based on the state read
// and what is in the BCSEgress.Spec
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
// nolint 
func (r *ReconcileBCSEgress) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	klog.Infof(">BCSEgress %s reconcile", request.String())
	// Fetch the BCSEgress instance
	instance := &bkbcsv1alpha1.BCSEgress{}
	controlReference := map[string]string{
		labelReference: request.String(),
	}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// BCSEgressController has multiple instance, it's hard to use finalizer because of s
			// Return and don't requeue
			klog.Infof("BCSEgress %s is actually deleted, try to clean all rules reference to %s", request.String(),
				request.String())
			if err = r.cleanRulesByLabel(controlReference); err != nil {
				klog.Errorf(">sync all reference rules failed when %s Egress deleted, %s. try next reconcile", request.String(),
					err.Error())
				return reconcile.Result{RequeueAfter: time.Second * 3}, err
			}
			klog.Infof(">BCSEgress %s all rules SYNCED trigger by deletion event", request.String())
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		klog.Errorf("reading BCSEgress %s failed, requeue...", request.String())
		return reconcile.Result{}, err
	}
	// port conflict verification and filter, if port conflicts,
	// do nothing until client fix them all.
	tcps, https, err := r.fromBCSEgressToList(instance)
	if err != nil {
		klog.Errorf("BCSEgress %s port definition conflicts: %s, just update Egress status.", request.String(), err.Error())
		instance.Status.Reason = err.Error()
		instance.Status.State = bkbcsv1alpha1.EgressStateError
		instance.Status.SyncedAt = metav1.Now()
		if err = r.client.Update(context.TODO(), instance); err != nil {
			klog.Errorf("update BCSEgress %s port conflict status failed, %s, #now drop reconcile until client fix#",
				request.String(), err.Error())
			// should evaluate necessary of reconcile
		} else {
			klog.Warningf(">update BCSEgress %s port conflict status successfully, #wait client fix#", request.String())
		}
		return reconcile.Result{}, nil
	}
	// all tcp rules & http rules from BCSEgress are verified,
	// we get all affective rules from proxy' cache, then find
	// difference between them for clean, add or update
	cacheTCPS, err := r.proxy.ListTCPRulesByLabel(controlReference)
	if err != nil {
		klog.Errorf("List all TCP by Label %+v failed, %s, try next reconcile after 5 seconds", controlReference, err.Error())
		return reconcile.Result{RequeueAfter: defaultReconcileInterval}, err
	}
	cacheHTTPS, err := r.proxy.ListHTTPRulesByLabel(controlReference)
	if err != nil {
		klog.Errorf("list all HTTP by Label %+v failed, %s, try next reconcile after 5 seconds", controlReference,
			err.Error())
		return reconcile.Result{RequeueAfter: defaultReconcileInterval}, err
	}
	httpChanged, err := r.reconcileHTTPRules(https, cacheHTTPS)
	if err != nil {
		klog.Errorf("Reconcile %s HTTP %d rules failed( %d in caches), %s", request.String(), len(https), len(cacheHTTPS),
			err.Error())
		return reconcile.Result{RequeueAfter: defaultReconcileInterval}, err
	}
	tcpChanged, err := r.reconcileTCPRules(tcps, cacheTCPS)
	if err != nil {
		klog.Errorf("Reconcile %s TCP %d rules failed( %d in caches), %s", request.String(), len(tcps), len(cacheTCPS),
			err.Error())
		return reconcile.Result{RequeueAfter: defaultReconcileInterval}, err
	}
	// nothing changed & there is no relative error for BCSEgress
	// just update synchronization state
	if !tcpChanged && !httpChanged && r.proxy.LastError(request.String()) == nil {
		klog.Infof("Reconcile %s proxy rules, but nothing changed & proxy works corectly~ wait for next Reconciler",
			request.String())
		// all rules done successfully, try to update BCSEgress status
		instance.Status.Reason = "all rules SYNCED"
		instance.Status.State = bkbcsv1alpha1.EgressStateSynced
		instance.Status.SyncedAt = metav1.Now()
		if err = r.client.Update(context.TODO(), instance); err != nil {
			klog.Errorf(
				"update BCSEgress %s last Reconcile status[%s] failed, %s. "+
					"push to ReconcileQueue for updating again in 5 seconds", request.String(), instance.Status.State,
				err.Error())
			return reconcile.Result{RequeueAfter: time.Second * 5}, nil
		}
		return reconcile.Result{}, nil
	}
	if err = r.proxy.Reload(request.String()); err != nil {
		klog.Errorf("EgressController reload %s proxy rules failed, %s. try to reload in 15 seconds", request.String(),
			err.Error())
		// try to update BCSEgress status
		instance.Status.Reason = fmt.Sprintf("EgressController reload internal failed: %s", err.Error())
		instance.Status.State = bkbcsv1alpha1.EgressStateError
		instance.Status.SyncedAt = metav1.Now()
		if err = r.client.Update(context.TODO(), instance); err != nil {
			klog.Errorf("update BCSEgress %s reload status failed, %s", request.String(), err.Error())
		} else {
			klog.Warningf("update BCSEgress %s reload failed status done, try to reload in 15 seconds", request.String())
		}
		return reconcile.Result{RequeueAfter: time.Second * 15}, err
	}
	// all rules done successfully, try to update BCSEgress status
	instance.Status.Reason = "all rules SYNCED"
	instance.Status.State = bkbcsv1alpha1.EgressStateSynced
	instance.Status.SyncedAt = metav1.Now()
	if err = r.client.Update(context.TODO(), instance); err != nil {
		klog.Errorf(
			"update BCSEgress %s last Reconcile status[%s] after proxy reload failed, %s. "+
				"try to Update again in next reconciler", request.String(), instance.Status.State, err.Error())
		return reconcile.Result{RequeueAfter: time.Second * 5}, nil
	}
	klog.Warningf(">update BCSEgress %s Reconcile status done, Status [%s]", request.String(), instance.Status.State)
	return reconcile.Result{}, nil
}

// Init all BCSEgressController  instance running requirement
func (r *ReconcileBCSEgress) Init(mgr manager.Manager) error {
	// Create a new controller
	c, err := controller.New("bcsegress-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		klog.Errorf("BCSEgress controller %s creation failed, %s", r.identity, err.Error())
		return err
	}

	// Watch for changes to primary resource BCSEgress
	// done(DeveloperJim): predicate for controller information filter
	err = c.Watch(
		&source.Kind{Type: &bkbcsv1alpha1.BCSEgress{}},
		&handler.EnqueueRequestForObject{},
		r,
	)
	if err != nil {
		klog.Errorf("BCSEgressController %s init watch BCSEgress failed with self identity failed, %s", r.identity,
			err.Error())
		return err
	}
	return nil
}

// Run ready to start egress backgroup worker
func (r *ReconcileBCSEgress) Run() {
	// start ticker for all egress synchronization
	tick := time.NewTicker(time.Second * 180)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			// time to check sync flag
			klog.Infoln("Force synchronization in tick...")
			r.synchronizationAllRules() // nolint
			klog.Infoln("nothing need to synchronize in tick...")
		case <-r.stop.Done():
			// ready to exit
			klog.Infof("BCSEgress %s is asked to exit...", r.identity)
			return
		}
	}
}

// Stop stop egress controller
func (r *ReconcileBCSEgress) Stop() {
	r.stopFunc()
}

// synchronizationAllRules xxx
// ! Most importance part, synchronize all BCSEgress rule to network policy.
// ! Every synchronization maybe affect business container, it's never too much
// ! to emphasize importance of synchronization. Sometimes synchronization will
// ! interrupt long connection that maintained by proxy without expectation
// nolint
func (r *ReconcileBCSEgress) synchronizationAllRules() error {
	// all datas changed Egress were in update channel
	klog.V(5).Infof("BCSEgress ready to sync all rules")
	return nil
}

// cleanRulesByLabel when BCSEgress is deleted, No details retrieve from reconciler
// so clean all rules reference to specified label
func (r *ReconcileBCSEgress) cleanRulesByLabel(cleanLabel map[string]string) error {
	// first try to get all Rules by Label
	tcps, err := r.proxy.ListTCPRulesByLabel(cleanLabel)
	if err != nil {
		klog.Errorf("EgressController get specified %+v TCP rules failed when try to clean egress rules, %s", cleanLabel,
			err.Error())
		return err
	}
	https, err := r.proxy.ListHTTPRulesByLabel(cleanLabel)
	if err != nil {
		klog.Errorf("EgressController get %+v HTTP rules failed when try to clean egress rules, %s", cleanLabel, err.Error())
		return err
	}
	// clean these delete Rules
	for index, tcprule := range tcps {
		if err := r.proxy.DeleteTCPRule(tcprule.Key()); err != nil {
			klog.V(5).Infof("clean tcp Rule %s under %+v failed, %s", tcprule.Key(), cleanLabel, err.Error())
			return err
		}
		klog.Infof("[index %d]clean tcp rule %s in proxy cache successfully", index, tcprule.Key())
	}
	for index, httprule := range https {
		if err := r.proxy.DeleteHTTPRule(httprule.Key()); err != nil {
			klog.V(5).Infof("clean http Rule %s under %+v failed, %s", httprule.Key(), cleanLabel, err.Error())
			return err
		}
		klog.Infof("[index %d]clean http rule %s in proxy cache successfully", index, httprule.Key())
	}
	klog.V(5).Infof("EgressController clean %s all relative egress rules successfully, try to reload...",
		cleanLabel[labelReference])
	// try to reload
	if err := r.proxy.Reload(cleanLabel[labelReference]); err != nil {
		klog.Errorf("EgressController Reload proxy with %s failed, %s", cleanLabel[labelReference], err.Error())
		return err
	}
	klog.Infof("EgressController clean %s operation all success. cheer~", cleanLabel[labelReference])
	return nil
}

// fromBCSEgressToList convert egress proxy rule to local cache
// in conversion, we have to verify:
// * first, find all conflict ports within this egress
// * second, find all port conflicted with other egresses(already exist ones)
// all return Configs are verified
func (r *ReconcileBCSEgress) fromBCSEgressToList(egress *bkbcsv1alpha1.BCSEgress) ([]*TCPConfig, []*HTTPConfig, error) {
	var httpList []*HTTPConfig
	var tcpList []*TCPConfig
	egressIndexer := fmt.Sprintf("%s/%s", egress.Namespace, egress.Name)
	nameMap := make(map[string]string)
	for _, httprule := range egress.Spec.HTTPS {
		// http name can not conflict in one BCSEgress definition
		if _, ok := nameMap[httprule.Name]; ok {
			return nil, nil, fmt.Errorf("http name %s conflicts", httprule.Name)
		}
		nameMap[httprule.Name] = httprule.Name
		// http rule must be unique in global scope
		httpConfig := SimpleHTTPConfig(httprule.Name, httprule.Host, httprule.DestPort)
		httpConfig.Label[labelReference] = egressIndexer
		destConfig, err := r.proxy.GetHTTPRule(httpConfig.Key())
		if err != nil {
			klog.Errorf("EgressController get HTTPRule [%s] error when formating BCSEgress %s: %s", httpConfig.Key(),
				egressIndexer, err.Error())
			return nil, nil, fmt.Errorf("EgressController internal error: %s", err.Error())
		}
		if destConfig == nil {
			// new http rule in BCSEgress, adopt
			httpList = append(httpList, httpConfig)
			continue
		}
		if !httpConfig.LabelFilter(destConfig.Label) {
			klog.Errorf("BCSEgress %s http rule %s conflicts with egress %s, rule formating error, drop BCSEgress %s",
				egressIndexer,
				httpConfig.Key(),
				destConfig.Label[labelReference],
				egressIndexer,
			)
			return nil, nil, fmt.Errorf("http rule %s conflict with %s", httpConfig.Key(), destConfig.Label[labelReference])
		}
		// http rule is valid, push to available list
		httpList = append(httpList, httpConfig)
	}
	// tcp rule validation
	nameMap = make(map[string]string)
	portMap := make(map[uint]uint)
	for _, tcprule := range egress.Spec.TCPS {
		// check name unique in definition
		if _, ok := nameMap[tcprule.Name]; ok {
			return nil, nil, fmt.Errorf("tcp name %s conflicts", tcprule.Name)
		}
		nameMap[tcprule.Name] = tcprule.Name
		// check port unique in definition
		if _, ok := portMap[tcprule.SourcePort]; ok {
			return nil, nil, fmt.Errorf("tcp proxy source port %d conflicts", tcprule.SourcePort)
		}
		portMap[tcprule.SourcePort] = tcprule.SourcePort
		tcpConfig := &TCPConfig{
			Name:            tcprule.Name,
			ProxyPort:       tcprule.SourcePort,
			Domain:          tcprule.Domain,
			DestinationPort: tcprule.DestPort,
			Algorithm:       tcprule.Algorithm,
			Label: map[string]string{
				labelReference: egressIndexer,
			},
		}
		if len(tcprule.IPs) != 0 {
			tcpConfig.IPs = strings.Split(tcprule.IPs, ",")
			tcpConfig.HasBackend = true
			tcpConfig.SortIPs()
		}
		// check proxy port is only maintained by this BCSEgress
		destConfig, err := r.proxy.GetTCPRuleByPort(tcprule.SourcePort)
		if err != nil {
			klog.Errorf("EgressController get TCPRule [%s] error when formating BCSEgress %s: %s", tcpConfig.Key(),
				egressIndexer, err.Error())
			return nil, nil, fmt.Errorf("EgressController internal error: %s", err.Error())
		}
		if destConfig == nil {
			// new rule in BCSEgress, adopt
			tcpList = append(tcpList, tcpConfig)
			continue
		}
		if !tcpConfig.LabelFilter(destConfig.Label) {
			klog.Errorf("BCSEgress %s tcp rule %s conflicts with egress %s, rule formating error, drop BCSEgress %s",
				egressIndexer,
				tcpConfig.Key(),
				destConfig.Label[labelReference],
				egressIndexer,
			)
			return nil, nil, fmt.Errorf("tcp rule %s conflicts with %s", tcpConfig.Key(), destConfig.Label[labelReference])
		}
		// all check is passed, tcp rule is valid
		tcpList = append(tcpList, tcpConfig)
	}
	return tcpList, httpList, nil
}

// reconcileHTTPRules try reconcile difference between these two HTTPConfig slices
func (r *ReconcileBCSEgress) reconcileHTTPRules(https, cacheHTTPS []*HTTPConfig) (bool, error) {
	isChanged := false
	if len(https) == 0 && len(cacheHTTPS) == 0 {
		klog.V(3).Infof("empty BCSEgress http rules & proxy http cache rules.")
		return isChanged, nil
	}
	if len(https) != len(cacheHTTPS) {
		isChanged = true
	}
	// egressRule use for Add/Update
	egressRules := make(map[string]*HTTPConfig)
	// cacheRules use for Delete
	cacheRules := make(map[string]*HTTPConfig)
	for _, newhttprule := range https {
		egressRules[newhttprule.Key()] = newhttprule
	}
	for _, cachehttprule := range cacheHTTPS {
		cacheRules[cachehttprule.Key()] = cachehttprule
	}
	// try to clean no changed rules in these two rulesMap
	for key, newrule := range egressRules {
		if oldrule, ok := cacheRules[key]; ok {
			delete(cacheRules, key)
			if !newrule.IsChanged(oldrule) {
				// nothing changed, clean in these two map
				delete(egressRules, key)
			} else {
				isChanged = true
			}
		} else {
			// we don't find rules in cache, this rule must add into cache & reload then
			isChanged = true
		}
	}
	// all left in egressRules are need to Update/Add
	for k, v := range egressRules {
		if err := r.proxy.UpdateHTTPRule(v); err != nil {
			klog.Errorf("EgressController Update http rule %s in reconcile failed, %s. details: %+v", k, err.Error(), v)
			return isChanged, err
		}
		klog.V(5).Infof("EgressController update http rule %s in cache successfully", k)
	}
	// all left in cacheRules are need to Delete
	for k, v := range cacheRules {
		if err := r.proxy.DeleteHTTPRule(k); err != nil {
			klog.Errorf("EgressController delete http rule %s in reconcil failed, %s, details: %+v", k, err.Error(), v)
			return isChanged, err
		}
		klog.V(5).Infof("EgressController delete http rule %s in cache successfully", k)
	}
	return isChanged, nil
}

// reconcileTCPRules try reconcile difference between these two TCPConfig slices
func (r *ReconcileBCSEgress) reconcileTCPRules(tcps, cacheTCPs []*TCPConfig) (bool, error) {
	isChanged := false
	if len(tcps) == 0 && len(cacheTCPs) == 0 {
		klog.V(3).Infof("empty BCSEgress http rules & proxy http cache rules.")
		return isChanged, nil
	}
	if len(tcps) != len(cacheTCPs) {
		isChanged = true
	}
	// egressRule use for Add/Update
	egressRules := make(map[string]*TCPConfig)
	// cacheRules use for Delete
	cacheRules := make(map[string]*TCPConfig)
	for _, newtcprule := range tcps {
		egressRules[newtcprule.Key()] = newtcprule
	}
	for _, cachetcprule := range cacheTCPs {
		cacheRules[cachetcprule.Key()] = cachetcprule
	}
	// try to clean no changed rules in these two rulesMap
	for key, newrule := range egressRules {
		if oldrule, ok := cacheRules[key]; ok {
			delete(cacheRules, key)
			if !newrule.IsChanged(oldrule) {
				// nothing changed, clean in these two map
				delete(egressRules, key)
			} else {
				isChanged = true
			}
		} else {
			// we don't find rules in cache, this rule must add into cache & reload then
			isChanged = true
		}
	}
	// all left in egressRules are need to Update/Add
	for k, v := range egressRules {
		if err := r.proxy.UpdateTCPRule(v); err != nil {
			klog.Errorf("EgressController Update tcp rule %s in reconcile failed, %s. details: %+v", k, err.Error(), v)
			return isChanged, err
		}
		klog.V(5).Infof("EgressController update tcp rule %s in cache successfully", k)
	}
	// all left in cacheRules are need to Delete
	for k, v := range cacheRules {
		if err := r.proxy.DeleteTCPRule(k); err != nil {
			klog.Errorf("EgressController delete tcp rule %s in reconcil failed, %s, details: %+v", k, err.Error(), v)
			return isChanged, err
		}
		klog.V(5).Infof("EgressController delete tcp rule %s in cache successfully", k)
	}
	return isChanged, nil
}

// Create returns true if the Create event should be processed
func (r *ReconcileBCSEgress) Create(e event.CreateEvent) bool {
	egress := e.Object.(*bkbcsv1alpha1.BCSEgress)
	if egress.Spec.Controller.Namespace != r.option.Namespace {
		return false
	}
	if egress.Spec.Controller.Name != r.option.Name {
		return false
	}
	return true
}

// Delete returns true if the Delete event should be processed
func (r *ReconcileBCSEgress) Delete(e event.DeleteEvent) bool {
	egress := e.Object.(*bkbcsv1alpha1.BCSEgress)
	if egress.Spec.Controller.Namespace != r.option.Namespace {
		return false
	}
	if egress.Spec.Controller.Name != r.option.Name {
		return false
	}
	return true
}

// Update returns true if the Update event should be processed
func (r *ReconcileBCSEgress) Update(e event.UpdateEvent) bool {
	egress := e.ObjectNew.(*bkbcsv1alpha1.BCSEgress)
	if egress.Spec.Controller.Namespace != r.option.Namespace {
		return false
	}
	if egress.Spec.Controller.Name != r.option.Name {
		return false
	}
	return true
}

// Generic returns true if the Generic event should be processed
func (r *ReconcileBCSEgress) Generic(e event.GenericEvent) bool {
	egress := e.Object.(*bkbcsv1alpha1.BCSEgress)
	if egress.Spec.Controller.Namespace != r.option.Namespace {
		return false
	}
	if egress.Spec.Controller.Name != r.option.Name {
		return false
	}
	return true
}
