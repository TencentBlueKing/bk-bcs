/*
Copyright (C) 2019 The BlueKing Authors. All rights reserved.

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
of the Software, and to permit persons to whom the Software is furnished to do
so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package cache

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
