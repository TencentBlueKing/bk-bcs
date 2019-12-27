// Copyright 2019 HAProxy Technologies
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package runtime

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/go-openapi/strfmt"

	"github.com/haproxytech/models"
	"github.com/mitchellh/mapstructure"
)

//GetStats fetches HAProxy stats from runtime API
func (s *SingleRuntime) GetStats() *models.NativeStatsCollection {
	rAPI := ""
	if s.worker != 0 {
		rAPI = fmt.Sprintf("%s@%v", s.socketPath, s.worker)
	} else {
		rAPI = s.socketPath
	}
	result := &models.NativeStatsCollection{RuntimeAPI: rAPI}
	rawdata, err := s.ExecuteRaw("show stat")
	if err != nil {
		result.Error = err.Error()
		return result
	}
	lines := strings.Split(rawdata[2:], "\n")
	stats := []*models.NativeStat{}
	keys := strings.Split(lines[0], ",")
	//data := []map[string]string{}
	for i := 1; i < len(lines); i++ {
		data := map[string]string{}
		line := strings.Split(lines[i], ",")
		if len(line) < len(keys) {
			continue
		}
		for index, key := range keys {
			if len(line[index]) > 0 {
				data[key] = line[index]
			}
		}
		oneLineData := &models.NativeStat{}
		tString := strings.ToLower(line[1])
		if tString == "backend" || tString == "frontend" {
			oneLineData.Name = line[0]
			oneLineData.Type = tString
		} else {
			oneLineData.Name = tString
			oneLineData.Type = "server"
			oneLineData.BackendName = line[0]
		}

		var st models.NativeStatStats
		decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			Result:           &st,
			WeaklyTypedInput: true,
			TagName:          "json",
		})
		if err != nil {
			continue
		}

		err = decoder.Decode(data)
		if err != nil {
			continue
		}
		oneLineData.Stats = &st

		stats = append(stats, oneLineData)
	}
	result.Stats = stats
	return result
}

//GetInfo fetches HAProxy info from runtime API
func (s *SingleRuntime) GetInfo() (models.ProcessInfoHaproxy, error) {
	dataStr, err := s.ExecuteRaw("show info typed")
	data := models.ProcessInfoHaproxy{}
	if err != nil {
		fmt.Println(err.Error())
		return data, err
	}
	return parseInfo(dataStr)
}

func parseInfo(info string) (models.ProcessInfoHaproxy, error) {
	data := models.ProcessInfoHaproxy{}

	for _, line := range strings.Split(info, "\n") {
		fields := strings.Split(line, ":")
		fID := strings.TrimSpace(strings.Split(fields[0], ".")[0])
		switch fID {
		case "1":
			data.Version = fields[3]
		case "2":
			d := strfmt.Date{}
			err := d.Scan(strings.Replace(fields[3], "/", "-", -1))
			if err == nil {
				data.ReleaseDate = d
			}
		case "4":
			nbproc, err := strconv.ParseInt(fields[3], 10, 64)
			if err == nil {
				data.Processes = &nbproc
			}
		case "6":
			pid, err := strconv.ParseInt(fields[3], 10, 64)
			if err == nil {
				data.Pid = &pid
			}
		case "8":
			uptime, err := strconv.ParseInt(fields[3], 10, 64)
			if err == nil {
				data.Uptime = &uptime
			}
		}
	}

	return data, nil
}
