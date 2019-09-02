package bcs

// import (
// 	"bcs/bmsf-mesh/bmsf-mesos-adaptor/pkg/cache"
// 	bcstypes "bcs/bmsf-mesh/bmsf-mesos-adaptor/pkg/datatype/bcs/common/types"
// 	schetypes "bcs/bmsf-mesh/bmsf-mesos-adaptor/pkg/datatype/bcs/scheduler/types"
// 	"bcs/bmsf-mesh/bmsf-mesos-adaptor/pkg/meta"
// 	"bcs/bmsf-mesh/bmsf-mesos-adaptor/pkg/storage/zookeeper"
// 	"bcs/bmsf-mesh/bmsf-mesos-adaptor/pkg/watch"
// 	"bcs/control-common/blog"
// 	"fmt"
// 	"path"
// 	"time"

// 	"golang.org/x/net/context"
// )

// //DiscoveryApp adaptor for bcs-scheduler application to object data
// type DiscoveryApp struct {
// 	schetypes.Application `json:",inline"`
// }

// //GetName for ObjectMeta
// func (om *DiscoveryApp) GetName() string {
// 	return om.ObjectMeta.GetName()
// }

// //SetName set object name
// func (om *DiscoveryApp) SetName(name string) {
// 	om.ObjectMeta.Name = name
// }

// //GetNamespace for ObjectMeta
// func (om *DiscoveryApp) GetNamespace() string {
// 	return om.ObjectMeta.NameSpace
// }

// //SetNamespace set object namespace
// func (om *DiscoveryApp) SetNamespace(ns string) {
// 	om.ObjectMeta.NameSpace = ns
// }

// //GetCreationTimestamp get create timestamp
// func (om *DiscoveryApp) GetCreationTimestamp() time.Time {
// 	return om.ObjectMeta.CreationTimestamp
// }

// //SetCreationTimestamp set creat timestamp
// func (om *DiscoveryApp) SetCreationTimestamp(timestamp time.Time) {
// 	om.ObjectMeta.CreationTimestamp = timestamp
// }

// //GetLabels for ObjectMeta
// func (om *DiscoveryApp) GetLabels() map[string]string {
// 	return om.ObjectMeta.Labels
// }

// //SetLabels set objec labels
// func (om *DiscoveryApp) SetLabels(labels map[string]string) {
// 	om.ObjectMeta.Labels = labels
// }

// //GetAnnotations for ObjectMeta
// func (om *DiscoveryApp) GetAnnotations() map[string]string {
// 	return om.ObjectMeta.Annotations
// }

// //SetAnnotations get annotation name
// func (om *DiscoveryApp) SetAnnotations(annotation map[string]string) {
// 	om.ObjectMeta.Annotations = annotation
// }

// //GetClusterName get cluster name
// func (om *DiscoveryApp) GetClusterName() string {
// 	return om.ObjectMeta.ClusterName
// }

// //SetClusterName set cluster name
// func (om *DiscoveryApp) SetClusterName(clusterName string) {
// 	om.ObjectMeta.ClusterName = clusterName
// }

// func newAppCache(hosts []string, handler cache.EventInterface) (*AppStore, *controller, error) {
// 	c := cache.CreateCache(AppObjectKeyFn)
// 	as := &AppStore{
// 		Cache: *c,
// 	}
// 	//create namespace client for zookeeper
// 	zkConfig := &zookeeper.ZkConfig{
// 		Hosts:         hosts,
// 		PrefixPath:    "/blueking/application",
// 		Name:          "application",
// 		Codec:         &meta.JsonCodec{},
// 		ObjectNewFunc: AppObjectNewFn,
// 	}
// 	nsclient, err := zookeeper.NewStorage(zkConfig)
// 	if err != nil {
// 		blog.Errorf("bcs mesos discovery create application namespace client failed, %s", err)
// 		return nil, nil, err
// 	}
// 	//create listwatcher
// 	listwatcher := &cache.ListWatch{
// 		ListFn: func() ([]meta.Object, error) {
// 			return nsclient.List(context.Background(), "", nil)
// 		},
// 		WatchFn: func() (watch.Interface, error) {
// 			return nsclient.Watch(context.Background(), "", "", nil)
// 		},
// 	}
// 	//create reflector
// 	reflector := cache.NewReflector(fmt.Sprintf("Reflector-%s", zkConfig.Name), as, listwatcher, time.Second*600, handler)
// 	ctl := &controller{
// 		eventStorage: nsclient,
// 		reflector:    reflector,
// 	}
// 	return as, ctl, nil
// }

// //AppObjectKeyFn create key for ServiceObject
// func AppObjectKeyFn(obj interface{}) (string, error) {
// 	app, ok := obj.(*DiscoveryApp)
// 	if !ok {
// 		return "", fmt.Errorf("error object type")
// 	}
// 	return path.Join(app.GetNamespace(), app.GetName()), nil
// }

// //AppObjectNewFn create new Service Object
// func AppObjectNewFn() meta.Object {
// 	return new(DiscoveryApp)
// }

// //AppStore wrapper for Store, offer convenient access for bcs-scheduler application
// type AppStore struct {
// 	cache.Cache //drive from cache for data storage
// }

// //GetTaskGroupIDByService get all taskgroup id list with service
// func (as *AppStore) GetTaskGroupIDByService(svc *bcstypes.BcsService) ([]string, error) {
// 	//get all applications from local cache
// 	apps := as.List()
// 	if len(apps) == 0 {
// 		return nil, nil
// 	}
// 	var targets []string
// 	for _, obj := range apps {
// 		app, ok := obj.(*DiscoveryApp)
// 		if !ok {
// 			blog.Errorf("AppStore got error object data type, object: %v", obj)
// 			continue
// 		}
// 		//check namespace
// 		if svc.GetNamespace() != app.GetNamespace() {
// 			continue
// 		}
// 		if ok := serviceSelector(svc.Spec.Selector, app.GetLabels()); !ok {
// 			continue
// 		}
// 		for _, info := range app.Pods {
// 			targets = append(targets, info.Name)
// 		}
// 	}
// 	return targets, nil
// }
