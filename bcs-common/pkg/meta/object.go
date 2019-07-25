package meta

import (
	"fmt"
	"strings"
	"time"
)

//ObjectNewFn return target data object pointer
type ObjectNewFn func() Object

//ObjectListNewFn creat Object List
type ObjectListNewFn func(raw []byte) ([]Object, error)

//Object offer common interface to access all data objects
type Object interface {
	GetName() string
	SetName(name string)
	GetNamespace() string
	SetNamespace(ns string)
	GetCreationTimestamp() time.Time
	SetCreationTimestamp(timestamp time.Time)
	GetLabels() map[string]string
	SetLabels(lables map[string]string)
	GetAnnotations() map[string]string
	SetAnnotations(annotation map[string]string)
	GetClusterName() string
	SetClusterName(clusterName string)
}

//List list for objects
type List interface {
	GetItems() []Object
	SetItems([]Object)
}

const (
	NamespaceIndex string = "namespace"
	//only taskgroup has the application index
	ApplicationIndex string = "application"
)

var errNotObject = fmt.Errorf("object does not implement the Object interfaces")

// NamespaceIndexFunc is a default index function that indexes based on an object's namespace
func NamespaceIndexFunc(obj interface{}) ([]string, error) {
	switch t := obj.(type) {
	case Object:
		return []string{t.GetNamespace()}, nil

	default:
		return nil, errNotObject
	}
}

// ApplicationIndexFunc is a default index function that indexes based on an object's application name
// only taskgroup has this index
func ApplicationIndexFunc(obj interface{}) ([]string, error) {
	switch t := obj.(type) {
	case Object:
		//name: mars-test-lb-0 and mars-test-lb is appname
		//application is appname.namespace
		index := strings.LastIndex(t.GetName(), "-")
		if index == -1 {
			return nil, fmt.Errorf("Taskgroup(%s:%s) is invalid", t.GetNamespace(), t.GetName())
		}

		return []string{fmt.Sprintf("%s.%s", t.GetName()[:index], t.GetNamespace())}, nil

	default:
		return nil, errNotObject
	}
}
