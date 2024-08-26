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

// Package tasks ...
package tasks

import (
	"errors"
	"strings"
	"time"

	"github.com/RichardKnop/machinery/v2/log"
)

// Add ...
func Add(args ...int64) (int64, error) {
	sum := int64(0)
	for _, arg := range args {
		sum += arg
	}
	return sum, nil
}

// Multiply ...
func Multiply(args ...int64) (int64, error) {
	sum := int64(1)
	for _, arg := range args {
		sum *= arg
	}
	return sum, nil
}

// SumInts ...
func SumInts(numbers []int64) (int64, error) {
	var sum int64
	for _, num := range numbers {
		sum += num
	}
	return sum, nil
}

// SumFloats ...
func SumFloats(numbers []float64) (float64, error) {
	var sum float64
	for _, num := range numbers {
		sum += num
	}
	return sum, nil
}

// Concat ...
func Concat(strs []string) (string, error) {
	var res string
	for _, s := range strs {
		res += s
	}
	return res, nil
}

// Split ...
func Split(str string) ([]string, error) {
	return strings.Split(str, ""), nil
}

// PanicTask ...
func PanicTask() (string, error) {
	panic(errors.New("oops"))
}

// LongRunningTask ...
func LongRunningTask() error {
	log.INFO.Print("Long running task started")
	for i := 0; i < 10; i++ {
		log.INFO.Print(10 - i)
		time.Sleep(1 * time.Second)
	}
	log.INFO.Print("Long running task finished")
	return nil
}
