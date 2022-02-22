/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package timex

import (
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/hako/durafmt"
)

// TimeLayout 时间格式
const TimeLayout = "2006-01-02 15:04:05"

// CalcDuration 计算 起始 至 终止时间 间时间间隔（带单位），例：
// 1. start: '2021-04-01 12:35:30' end: '2021-04-03 14:00:00' => '2d1h'
// 2. start: '2021-04-01 12:35:30' end: '2021-04-01 12:59:59' => '24m29s'
func CalcDuration(start string, end string) string {
	startTime, _ := dateparse.ParseStrict(start)
	endTime := time.Now()
	if end != "" {
		endTime, _ = dateparse.ParseStrict(end)
	}
	timeDuration := endTime.Sub(startTime)
	// 将时间间隔解析成字符串，最大单位为 d（days），保留两位有效信息
	duration := durafmt.Parse(timeDuration).LimitToUnit("days").LimitFirstN(2).InternationalString()
	// 去除解析结果的空格
	duration = strings.Join(strings.Split(duration, " "), "")
	return duration
}

// CalcAge 计算存在时间
func CalcAge(createTime string) string {
	return CalcDuration(createTime, "")
}

// NormalizeDatetime 标准化时间格式
func NormalizeDatetime(datetime string) (string, error) {
	t, err := dateparse.ParseStrict(datetime)
	if err != nil {
		return "", err
	}
	return t.Format(TimeLayout), nil
}

// Current 获取当前时间
func Current() string {
	return time.Now().Format(TimeLayout)
}
