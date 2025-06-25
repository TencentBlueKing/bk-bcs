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

package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/pkg/recorder"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/pkg/utils"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/pkg/utils/tail"
)

// recorderQuery query the record events
type recorderQuery struct {
	Limit     int64
	RequestID string
	Registry  string
	Repo      string
	Digest    string
	ShowID    bool
	Follow    bool
}

// parseRecorderQuery parse the recorde query object
func (s *CustomRegistry) parseRecorderQuery(r *http.Request) *recorderQuery {
	var limit int64
	limitStr := r.URL.Query().Get("limit")
	if limitStr == "" {
		limit = 200
	} else {
		v, err := strconv.Atoi(limitStr)
		if err != nil {
			limit = 300
		} else {
			limit = int64(v)
		}
	}
	return &recorderQuery{
		Limit:     limit,
		RequestID: r.URL.Query().Get("requestID"),
		Registry:  r.URL.Query().Get("registry"),
		Repo:      r.URL.Query().Get("repo"),
		Digest:    r.URL.Query().Get("digest"),
		ShowID:    r.URL.Query().Get("showID") == "true",
		Follow:    r.URL.Query().Get("follow") == "true",
	}
}

// filterEvent filter event by query
func (s *CustomRegistry) filterEvent(query *recorderQuery, line string) *recorder.Event {
	event := new(recorder.Event)
	if line == "" {
		return nil
	}
	if err := json.Unmarshal([]byte(line), event); err != nil {
		return nil
	}
	if query.RequestID != "" && !strings.Contains(event.RequestID, query.RequestID) {
		return nil
	}
	if query.Registry != "" && !strings.Contains(event.Registry, query.Registry) {
		return nil
	}
	if query.Repo != "" && !strings.Contains(event.Repo, query.Repo) {
		return nil
	}
	if query.Digest != "" && !strings.Contains(event.Digest, query.Digest) {
		return nil
	}
	return event
}

// printRecorderHead print analysis head
func (s *CustomRegistry) printRecorderHead(query *recorderQuery, tw *tablewriter.Table) {
	if query.ShowID {
		tw.SetHeader(func() []string {
			return []string{
				"REQ-ID", "TIME", "REGISTRY", "REPO", "TYPE", "MESSAGE",
			}
		}())
		return
	} else {
		tw.SetHeader(func() []string {
			return []string{
				"TIME", "REGISTRY", "REPO", "TYPE", "MESSAGE",
			}
		}())
	}
}

// printRecorderRow print analysis row
func (s *CustomRegistry) printRecorderRow(query *recorderQuery, tw *tablewriter.Table, event *recorder.Event) {
	if query.ShowID {
		tw.Append([]string{
			event.RequestID, event.CreatedAt.Format("01-02/15:04:05"), event.Registry,
			event.Repo, string(event.EventType), event.Message,
		})
	} else {
		tw.Append([]string{
			event.CreatedAt.Format("01-02/15:04:05"), event.Registry,
			event.Repo, string(event.EventType), event.Message,
		})
	}
}

// Recorder returns analysis data
func (s *CustomRegistry) Recorder(rw http.ResponseWriter, r *http.Request) (interface{}, error) {
	query := s.parseRecorderQuery(r)
	tw := utils.DefaultTableWriter(rw)
	if !query.Follow {
		lines, err := tail.OnceTailLines(s.op.EventFile, int(query.Limit))
		if err != nil {
			return nil, errors.Wrap(err, "reader event file failed")
		}
		result := make([]*recorder.Event, 0, len(lines))
		for i := range lines {
			event := s.filterEvent(query, lines[i])
			if event != nil {
				result = append(result, event)
			}
		}
		s.printRecorderHead(query, tw)
		for _, event := range result {
			s.printRecorderRow(query, tw, event)
		}
		tw.Render()
		return nil, nil
	}

	ctx := r.Context()
	ch, err := tail.FollowTailLines(ctx, s.op.EventFile, int(query.Limit))
	if err != nil {
		return nil, err
	}
	s.printRecorderHead(query, tw)
	tw.Render()
	for {
		select {
		case <-ctx.Done():
			return nil, nil
		case line, ok := <-ch:
			if !ok {
				return nil, nil
			}
			event := s.filterEvent(query, line)
			if event == nil {
				continue
			}

			// s.printRecorderRow(query, tw, event)
			fmt.Fprintf(rw, fmt.Sprintf("%s\r%s\r%s\r%s\r%s\r\n",
				event.CreatedAt.Format("01-02/15:04:05"), event.Registry,
				event.Repo, string(event.EventType), event.Message))
		}
	}
}

// TorrentStatus return the torrent status
func (s *CustomRegistry) TorrentStatus(rw http.ResponseWriter, r *http.Request) (interface{}, error) {
	cl := s.torrentHandler.GetClient()
	cl.WriteStatus(rw)
	return nil, nil
}
