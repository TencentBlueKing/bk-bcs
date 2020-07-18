/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package blogr

import (
	"fmt"
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/go-logr/logr"
)

// Blogger logger for go-logr used by k8s controller, wrapping blog
type Blogger struct {
	v         string
	name      string
	keyValues map[string]interface{}
}

var _ logr.Logger = &Blogger{}

// Info implements go-logr
func (l *Blogger) Info(msg string, kvs ...interface{}) {
	str := fmt.Sprintf("[%s]\t%s\t", l.name, msg)
	for k, v := range l.keyValues {
		str = str + fmt.Sprintf("%s: %+v ", k, v)
	}
	for i := 0; i < len(kvs); i += 2 {
		str = str + fmt.Sprintf("%s: %+v ", kvs[i], kvs[i+1])
	}
	blog.Infof(str)
}

// Enabled implements go-logr
func (_ *Blogger) Enabled() bool {
	return true
}

// Error implements go-logr
func (l *Blogger) Error(err error, msg string, kvs ...interface{}) {
	str := fmt.Sprintf("[%s]\t%s\t", l.name, msg)
	str = str + fmt.Sprintf("err: %+v ", err.Error())
	for k, v := range l.keyValues {
		str = str + fmt.Sprintf("%s: %+v ", k, v)
	}
	for i := 0; i < len(kvs); i += 2 {
		str = str + fmt.Sprintf("%s: %+v ", kvs[i], kvs[i+1])
	}
	blog.Errorf(str)
}

// V implements go-logr
func (l *Blogger) V(level int) logr.InfoLogger {
	return &Blogger{
		v:         strconv.Itoa(level),
		name:      l.name,
		keyValues: l.keyValues,
	}
}

// WithName implements go-logr
func (l *Blogger) WithName(name string) logr.Logger {
	return &Blogger{
		v:         l.v,
		name:      l.name + "." + name,
		keyValues: l.keyValues,
	}
}

// WithValues implements go-logr
func (l *Blogger) WithValues(kvs ...interface{}) logr.Logger {
	newMap := make(map[string]interface{}, len(l.keyValues)+len(kvs)/2)
	for k, v := range l.keyValues {
		newMap[k] = v
	}
	for i := 0; i < len(kvs); i += 2 {
		newMap[kvs[i].(string)] = kvs[i+1]
	}
	return &Blogger{
		v:         l.v,
		name:      l.name,
		keyValues: newMap,
	}
}

// NewBlogger create blogger
func NewBlogger() logr.Logger {
	return &Blogger{}
}
