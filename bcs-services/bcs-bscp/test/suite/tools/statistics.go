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

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"runtime"
	"strings"

	"github.com/tidwall/gjson"
)

// statistics according to the json file exported by the goconvey test,
// statistical analysis needs to show the integration test-related information.
func statistics(fPath string) ([]*StatisticalResults, error) {
	results, err := getResultFromFile(fPath)
	if err != nil {
		return nil, err
	}

	srs := make([]*StatisticalResults, 0)
	for _, one := range results {
		srs = append(srs, statisticsResult(one))
	}

	// reading files into memory consumes more memory, so after calculating a test file,
	// you need to manually perform GC
	runtime.GC()
	return srs, nil
}

// statisticsResult statistics goconvey test result.
func statisticsResult(one *Result) *StatisticalResults {
	sr := new(StatisticalResults)

	failed := 0
	failedMap := make(map[string]*FailedInfo, 0)
	for _, one := range one.Assertions {
		if one.Line != 0 {
			failed++
			line := getSplitLine(one.File, one.Line)

			if _, ok := failedMap[line]; !ok {
				failedMap[line] = &FailedInfo{
					Line:    getSplitLine(one.File, one.Line),
					Message: compressStr(one.Failure, one.Error),
					Total:   1,
				}

				continue
			}

			failedMap[line].Total++
		}
	}

	failedInfos := make([]*FailedInfo, 0)
	for _, v := range failedMap {
		failedInfos = append(failedInfos, v)
	}

	sr.Title = one.Title
	sr.Total = len(one.Assertions)
	sr.Failed = failed
	sr.Succeed = sr.Total - failed
	sr.FailedInfos = failedInfos

	return sr
}

// getResultFromFile get goconvey test result from file.
func getResultFromFile(path string) ([]*Result, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	content := string(bytes)

	splitContents := strings.Split(content, "<\n>")
	results := make([]*Result, 0)
	for _, one := range splitContents {
		subResults, err := getResultFromContent(one)
		if err != nil {
			log.Printf("get result from content(%s) failed, err: %v\n", one, err)
			return nil, err
		}
		results = append(results, subResults...)
	}

	return results, nil
}

// getResultFromContent get goconvey test result from string content, and to parse
func getResultFromContent(content string) ([]*Result, error) {
	first := strings.Index(content, "{")
	last := strings.LastIndex(content, "}")
	content = content[first : last+1]

	split := strings.Split(content, "},{")

	results := make([]*Result, 0)
	for _, one := range split {
		if one[0] != '{' {
			one = "{" + one
		}

		if one[len(one)-1] != '}' {
			one += "}"
		}

		r := new(Result)
		err := r.Unmarshal([]byte(one))
		if err != nil {
			return nil, err
		}
		results = append(results, r)
	}

	return results, nil
}

// Result peer convey test result info.
type Result struct {
	Title      string      `json:"Title"`
	File       string      `json:"File"`
	Line       int         `json:"Line"`
	Depth      int         `json:"Depth"`
	Assertions []Assertion `json:"Assertions"`
	Output     string      `json:"Output"`
}

// Unmarshal a json raw to this Result
func (r *Result) Unmarshal(raw []byte) error {
	if err := json.Unmarshal(raw, &r); err != nil {
		return err
	}

	var parsed = gjson.GetManyBytes(raw, "Assertions")
	assertions := parsed[0]

	ass := make([]Assertion, 0)
	if err := json.Unmarshal([]byte(assertions.Raw), &ass); err != nil {
		return err
	}
	r.Assertions = ass

	return nil
}

// Assertion goconvey so test assertion info.
type Assertion struct {
	File       string `json:"File"`
	Line       int    `json:"Line"`
	Expected   string `json:"Expected"`
	Actual     string `json:"Actual"`
	Failure    string `json:"Failure"`
	Error      string `json:"Error"`
	StackTrace string `json:"StackTrace"`
	Skipped    bool   `json:"Skipped"`
}

// StatisticalResults statistical results for a single Result.
type StatisticalResults struct {
	Title       string        `json:"title"`
	Total       int           `json:"total"`
	Failed      int           `json:"failed"`
	Succeed     int           `json:"succeed"`
	FailedInfos []*FailedInfo `json:"failed_infos"`
}

// FailedInfo failed info.
type FailedInfo struct {
	Line    string `json:"line"`
	Message string `json:"message"`
	Total   int    `json:"total"`
}

// getSplitLine get the truncated file lines shown to the user.
func getSplitLine(file string, line int) string {
	length := len(file)
	if length <= 30 {
		return fmt.Sprintf("%s:%d", file, line)
	}

	return fmt.Sprintf("...%s:%d", file[length-30:], line)
}

// compressStr compress characters because the goconvey error message contains invalid spaces and needs to be deleted.
func compressStr(str, err string) string {
	if str == "" && err == "" {
		return ""
	}

	if str == "" && err != "" {
		return "" + err
	}

	reg := regexp.MustCompile("\\s+") // nolint
	return reg.ReplaceAllString(str, " ") + err
}
