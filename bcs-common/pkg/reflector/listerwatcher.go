package reflector

import (
	"bk-bcs/bcs-common/pkg/meta"
	"bk-bcs/bcs-common/pkg/watch"
	"fmt"
)

// ListerWatcher is interface perform list all objects and start a watch
type ListerWatcher interface {
	// List should return a list type object
	List() ([]meta.Object, error)
	// Watch should begin a watch from remote storage
	Watch() (watch.Interface, error)
}

//ListFunc define list function for ListerWatcher
type ListFunc func() ([]meta.Object, error)

//WatchFunc define watch function for ListerWatcher
type WatchFunc func() (watch.Interface, error)

//ListWatch implements ListerWatcher interface
//protect ListFunc or WatchFunc is nil
type ListWatch struct {
	ListFn  ListFunc
	WatchFn WatchFunc
}

// List should return a list type object
func (lw *ListWatch) List() ([]meta.Object, error) {
	if lw.ListFn == nil {
		return nil, nil
	}
	return lw.ListFn()
}

// Watch should begin a watch from remote storage
func (lw *ListWatch) Watch() (watch.Interface, error) {
	if lw.WatchFn == nil {
		return nil, fmt.Errorf("Lost watcher function")
	}
	return lw.WatchFn()
}
