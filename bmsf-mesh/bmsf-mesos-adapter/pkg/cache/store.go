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
	"fmt"
)

//Store memory storage interface
type Store interface {
	Add(obj interface{}) error
	Update(obj interface{}) error
	Delete(obj interface{}) error
	List() []interface{}
	ListKeys() []string
	Get(obj interface{}) (item interface{}, exists bool, err error)
	GetByKey(key string) (item interface{}, exists bool, err error)
	//Num will return data counts in Store
	Num() int
	//Clear will drop all data in Store
	Clear()
	// Replace will delete the contents of the store, using instead the
	// given list. Store takes ownership of the list, you should not reference
	// it after calling this function.
	Replace([]interface{}) error
	//Reset clean data first and then setting data
	Reset([]interface{}) error
}

//ObjectKeyFunc define make object to a uniq key
type ObjectKeyFunc func(obj interface{}) (string, error)

//KeyError wrapper error return from ObjectKeyFunc
type KeyError struct {
	Obj interface{}
	Err error
}

//Error return string info for KeyError
func (k KeyError) Error() string {
	return fmt.Sprintf("create key for object %+v failed: %v", k.Obj, k.Err)
}

//DataNoExist return when No data in Store
type DataNoExist struct {
	Obj interface{}
}

//Error return string info
func (k DataNoExist) Error() string {
	return fmt.Sprintf("no data object %+v in Store", k.Obj)
}
