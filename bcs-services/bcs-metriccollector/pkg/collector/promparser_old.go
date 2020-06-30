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
 *
 */

package collector

import (
	"bufio"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"io"
	"strconv"
	"strings"
)

type CollectorWrapper struct {
	Collector CollectorMetrics `json:"collector"`
}

type CollectorMetrics struct {
	Metrics []*PromMetric `json:"metrics"`
}

type PromMetric struct {
	Name   string            `json:"key"`
	Labels map[string]string `json:"labels,omitempty"`
	Value  float64           `json:"value"`
}

func ParsePromTextToOldVersion(in io.Reader, cLabels map[string]string) (*CollectorWrapper, error) {
	metrics := make([]*PromMetric, 0)
	buf := bufio.NewReader(in)
outter:
	for {
		wholeline := make([]byte, 0)
		for {
			line, isPrefix, err := buf.ReadLine()
			if err != nil {
				if err == io.EOF {
					break outter
				} else {
					return nil, err
				}
			}
			wholeline = append(wholeline, line...)
			if isPrefix {
				continue
			} else {
				break
			}
		}

		text, needSkip := skipBlankTab(string(wholeline))
		if needSkip {
			continue
		}
		name, value, labels, valid := parse(text)
		if valid {
			// add user defined labels for now.
			for k, v := range cLabels {
				labels[k] = v
			}
			metrics = append(metrics, &PromMetric{
				Name:   name,
				Value:  value,
				Labels: labels,
			})
		} else {
			blog.Infof("parse prometric metric, find an invalid metric term: %s", wholeline)
		}
	}
	return &CollectorWrapper{Collector: CollectorMetrics{Metrics: metrics}}, nil
}

func skipBlankTab(line string) (string, bool) {
	base := 0
	for index, b := range line {
		if b == ' ' || b == '\t' {
			base = index
		} else {
			if b == '\r' || b == '\n' {
				if index+1 <= len(line)-1 {
					return line[index+1:], false
				}

				return "", true
			} else if b == '#' {
				return "", true
			}

			if base == 0 {
				return line, false
			} else if base+1 <= len(line)-1 {
				return line[base+1:], false
			}

			return "", true
		}
	}
	return "", true
}

func parse(line string) (metric_name string, metric_value float64, metric_label map[string]string, valid bool) {
	metric_label = make(map[string]string)
	var label_start, label_end, value int
	for index := range line {
		switch line[index] {
		case '{':
			label_start = index
		case '}':
			label_end = index
		case ' ':
			value = index
		}
	}

	if label_start == label_end && label_start != 0 {
		return metric_name, metric_value, metric_label, false
	}

	if label_start != label_end && label_start == 0 {
		return metric_name, metric_value, metric_label, false
	}

	if label_start > label_end {
		return metric_name, metric_value, metric_label, false
	}

	if value <= label_end {
		return metric_name, metric_value, metric_label, false
	}

	metric_value, err := strconv.ParseFloat(strings.TrimSpace(line[value+1:]), 10)
	if err != nil {
		blog.Errorf("format metric value[%s] to float64 failed, err: %v", line[value+1:], err)
		// fmt.Printf("format metric value[%s] to float64 failed, err: %v\n", line[value+1:], err)
		return metric_name, metric_value, nil, false
	}

	if label_start == label_end && label_start == 0 {
		metric_name = line[:value]
	} else {
		metric_name = line[:label_start]
		label := line[label_start+1 : label_end]
		labelArrary := strings.Split(label, ",")
		for _, ele := range labelArrary {
			kv := strings.Split(ele, "=")
			if len(kv) == 2 {
				value := strings.Trim(kv[1], " ")
				value = strings.Trim(value, "\"")
				metric_label[strings.Trim(kv[0], " ")] = value
			}
		}
	}
	return metric_name, metric_value, metric_label, true
}
