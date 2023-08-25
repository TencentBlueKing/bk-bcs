package controllers

import (
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	v1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-netservice-controller/api/v1"
)

// IPFilter filter for BCSNetIP event
type IPFilter struct {
	filterName string
}

// NewIPFilter create bcsNetIP filter
func NewIPFilter() *IPFilter {
	return &IPFilter{
		filterName: "bcsNetIP",
	}
}

var _ handler.EventHandler = &IPFilter{}

// Create is called in response to a create event
func (f *IPFilter) Create(event event.CreateEvent, q workqueue.RateLimitingInterface) {}

// Update is called in response to an update event
func (f *IPFilter) Update(event event.UpdateEvent, q workqueue.RateLimitingInterface) {
	_, ok := event.ObjectOld.(*v1.BCSNetIP)
	if !ok {
		blog.Errorf("update object is not BCSNetIP, event %+v", event)
		return
	}
	ip, ok := event.ObjectNew.(*v1.BCSNetIP)
	if !ok {
		blog.Errorf("update object is not BCSNetIP, event %+v", event)
		return
	}
	poolName, ok := ip.Labels["pool"]
	if !ok {
		blog.Errorf("can not find pool name by labels for IP [%s]", ip.Name)
		return
	}

	q.Add(reconcile.Request{NamespacedName: types.NamespacedName{Name: poolName}})
}

// Delete is called in response to a delete event
func (f *IPFilter) Delete(event event.DeleteEvent, q workqueue.RateLimitingInterface) {
	ip, ok := event.Object.(*v1.BCSNetIP)
	if !ok {
		blog.Errorf("delete object is not BCSNetIP, event %+v", event)
		return
	}

	poolName, ok := ip.Labels["pool"]
	if !ok {
		blog.Errorf("can not find pool name by labels for IP [%s]", ip.Name)
		return
	}

	q.Add(reconcile.Request{NamespacedName: types.NamespacedName{Name: poolName}})
}

// Generic is called in response to an event of an unknown type or a synthetic event triggered as a cron or
// external trigger request
func (f *IPFilter) Generic(event event.GenericEvent, q workqueue.RateLimitingInterface) {}
