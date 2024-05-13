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

package run

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
)

var reporter *Reporter

// Reporter can count Metrics data and generate reports
type Reporter struct {
	Data []metricsData
}

// Archive archive bench data.
func (a *Reporter) Archive(t string, m Metrics) {
	reporter.Data = append(reporter.Data, metricsData{
		Title: t,
		Metrics: Metrics{
			SustainSeconds:    m.SustainSeconds,
			Concurrent:        m.Concurrent,
			TotalRequest:      m.TotalRequest,
			SucceedRequest:    m.SucceedRequest,
			FailedRequest:     m.FailedRequest,
			OnTheFlyRequest:   m.OnTheFlyRequest,
			QPS:               floatFormat(m.QPS),
			MaxDuration:       floatFormat(m.MaxDuration),
			MinDuration:       floatFormat(m.MinDuration),
			MedianDuration:    floatFormat(m.MedianDuration),
			AverageDuration:   floatFormat(m.AverageDuration),
			Percent85Duration: floatFormat(m.Percent85Duration),
			Percent95Duration: floatFormat(m.Percent95Duration),
		},
	})
}

// GenReport gen bench report.
func (a *Reporter) GenReport(path string) error {
	outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("open report file failed, err: %v", err)
	}
	defer outFile.Close()

	write := bufio.NewWriter(outFile)
	tp.render(write, reporter.Data)
	write.Flush() // nolint error not checked

	return nil
}

// metricsData
type metricsData struct {
	Title   string
	Metrics Metrics
}

func init() {
	reporter = &Reporter{
		Data: make([]metricsData, 0),
	}
}

// Archive metrics.
func Archive(t string, m Metrics) {
	reporter.Archive(t, m)
}

// GenReport gen html metrics report.
func GenReport(path string) error {
	return reporter.GenReport(path)
}

// floatFormat format float save two decimal places.
func floatFormat(value float64) float64 {
	value, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", value), 64)
	return value
}
