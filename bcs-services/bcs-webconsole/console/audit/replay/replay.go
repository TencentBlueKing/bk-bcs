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

package replay

import (
	"os"
	"time"

	player "github.com/xakep666/asciinema-player/v3"
)

var (
	maxWait time.Duration = 2 * time.Second
	speed   float64       = 1
)

// Replay 回放文件
func Replay(f string) error {
	file, err := os.Open(f)
	if err != nil {
		return err
	}
	source, err := player.NewStreamFrameSource(file)
	if err != nil {
		return err
	}
	term, err := player.NewOSTerminal()
	if err != nil {
		return err
	}

	p, err := player.NewPlayer(source, term, player.WithSpeed(speed), player.WithMaxWait(maxWait), player.WithIgnoreSizeCheck())
	if err != nil {
		return err
	}
	err = p.Start()
	if err != nil {
		return err
	}

	defer func() {
		file.Close()
		term.Close()
	}()
	return nil
}
