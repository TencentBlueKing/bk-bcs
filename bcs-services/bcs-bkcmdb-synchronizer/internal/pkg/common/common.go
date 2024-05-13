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

// Package common define common methods
package common

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"reflect"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/encrypt"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/option"
)

// InArray checks if a value is in an array
func InArray(val interface{}, array interface{}) (exists bool, index int) {
	exists = false
	index = -1

	switch reflect.TypeOf(array).Kind() { // nolint
	case reflect.Slice:
		s := reflect.ValueOf(array)

		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(val, s.Index(i).Interface()) {
				index = i
				exists = true
				return
			}
		}
	}

	return exists, index
}

// DecryptCMOption decrypts the CostManagerOption
func DecryptCMOption(bkcmdbSynchronizerOption *option.BkcmdbSynchronizerOption) error {
	needToDecrypt := []*string{
		&bkcmdbSynchronizerOption.Client.ClientCrtPwd,
		&bkcmdbSynchronizerOption.Bcsapi.BearerToken,
		&bkcmdbSynchronizerOption.Bcsapi.ProjectToken,
		&bkcmdbSynchronizerOption.RabbitMQ.Password,
		&bkcmdbSynchronizerOption.CMDB.AppCode,
		&bkcmdbSynchronizerOption.CMDB.AppSecret,
		&bkcmdbSynchronizerOption.CMDB.BKUserName,
	}

	for _, value := range needToDecrypt {
		if err := decrypt(value); err != nil {
			return err
		}
	}
	return nil
}

func decrypt(value *string) error {
	// decrypt
	decrypted, err := encrypt.DesDecryptFromBase([]byte(*value))
	if err != nil {
		return err
	}
	*value = string(decrypted)
	return nil
}

// HashCode returns the hash code of the string
func HashCode(str string) string {
	h := fnv.New32a()
	h.Write([]byte(str))
	return fmt.Sprint(h.Sum32())
}

// Recoverer is a function that recovers from panics and calls the given function f.
// It will continue to recover from panics up to the specified maxPanics times.
func Recoverer(maxPanics int, f func()) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			if maxPanics == 0 {
				fmt.Printf("Too many panics, exiting")
			} else {
				time.Sleep(1 * time.Second)
				go Recoverer(maxPanics-1, f)
			}
		}
	}()
	f()
}

// InterfaceToStruct is a function that converts an interface to a struct.
// It first marshals the input interface to JSON data, and then unmarshals it to the output struct.
func InterfaceToStruct(in interface{}, out interface{}) error {
	data, err := json.Marshal(in)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, out)
}

// FirstLower first lowercase
func FirstLower(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToLower(s[:1]) + s[1:]
}
