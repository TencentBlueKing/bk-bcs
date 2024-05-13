/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package dbm

import (
	"fmt"
	"log"
	"os"
	"regexp"
)

const (
	bkAppCodePattern   = `"bk_app_code":[ ]*"[^"]*"`
	bkAppSecretPattern = `"bk_app_secret":[ ]*"[^"]*"`
	bkUsernamePattern  = `"bk_username":[ ]*"[^"]*"`
	hiddenAppCode      = `"bk_app_code":"***"`
	hiddenAppSecret    = `"bk_app_secret":"***"`
	hiddenUsername     = `"bk_username":"***"`
)

var (
	regexAppCode, _   = regexp.Compile(bkAppCodePattern)
	regexAppSecret, _ = regexp.Compile(bkAppSecretPattern)
	regexUsername, _  = regexp.Compile(bkUsernamePattern)
)

func hideAppCode(input string) string {
	return regexAppCode.ReplaceAllString(input, hiddenAppCode)
}

func hideAppSecret(input string) string {
	return regexAppSecret.ReplaceAllString(input, hiddenAppSecret)
}

func hideUsername(input string) string {
	return regexUsername.ReplaceAllString(input, hiddenUsername)
}

func hideAuth(input string) string {
	return hideUsername(hideAppSecret(hideAppCode(input)))
}

func newLogger() *deIdentLogger {
	l := log.New(os.Stderr, "[gorequest]", log.LstdFlags)
	return &deIdentLogger{
		l: l,
	}
}

// 打印日志时，将信息脱敏
type deIdentLogger struct {
	l *log.Logger
}

// SetPrefix 实现接口 https://github.com/parnurzeal/gorequest/blob/develop/logger.go
func (dl *deIdentLogger) SetPrefix(prefix string) {
	dl.l.SetPrefix(prefix)
}

// Printf 实现接口 https://github.com/parnurzeal/gorequest/blob/develop/logger.go
func (dl *deIdentLogger) Printf(format string, v ...interface{}) {
	result := fmt.Sprintf(format, v...)
	dl.l.Printf(hideAuth(result))
}

// Println 实现接口 https://github.com/parnurzeal/gorequest/blob/develop/logger.go
func (dl *deIdentLogger) Println(v ...interface{}) {
	result := fmt.Sprint(v...)
	dl.l.Println(hideAuth(result))
}
