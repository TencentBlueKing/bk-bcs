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

package job

import (
	"context"
	"testing"
)

func NewClient() *Client {
	cli, _ := NewJobClient(Options{
		AppCode:    "bcs-xxx",
		AppSecret:  "xxx",
		BKUserName: "xxx",
		Server:     "xxx",
		Debug:      true,
	})

	return cli
}

var content = "IyEvYmluL2Jhc2gKCmFueW5vd3RpbWU9ImRhdGUgKyclWS0lbS0lZCAlSDolTTolUyciCk5PVz0i\nZWNobyBbXGAkYW55bm93dGltZVxgXVtQSUQ6JCRdIgoKIyMjIyMg5Y+v5Zyo6ISa5pys5byA5aeL\n6L+Q6KGM5pe26LCD55So77yM5omT5Y2w5b2T5pe255qE5pe26Ze05oiz5Y+KUElE44CCCmZ1bmN0\naW9uIGpvYl9zdGFydAp7CiAgICBlY2hvICJgZXZhbCAkTk9XYCBqb2Jfc3RhcnQiCn0KCiMjIyMj\nIOWPr+WcqOiEmuacrOaJp+ihjOaIkOWKn+eahOmAu+i+keWIhuaUr+WkhOiwg+eUqO+8jOaJk+WN\nsOW9k+aXtueahOaXtumXtOaIs+WPilBJROOAgiAKZnVuY3Rpb24gam9iX3N1Y2Nlc3MKewogICAg\nTVNHPSIkKiIKICAgIGVjaG8gImBldmFsICROT1dgIGpvYl9zdWNjZXNzOlskTVNHXSIKICAgIGV4\naXQgMAp9CgojIyMjIyDlj6/lnKjohJrmnKzmiafooYzlpLHotKXnmoTpgLvovpHliIbmlK/lpITo\nsIPnlKjvvIzmiZPljbDlvZPml7bnmoTml7bpl7TmiLPlj4pQSUTjgIIKZnVuY3Rpb24gam9iX2Zh\naWwKewogICAgTVNHPSIkKiIKICAgIGVjaG8gImBldmFsICROT1dgIGpvYl9mYWlsOlskTVNHXSIK\nICAgIGV4aXQgMQp9Cgpqb2Jfc3RhcnQKCiMjIyMjIyDkvZzkuJrlubPlj7DkuK3miafooYzohJrm\nnKzmiJDlip/lkozlpLHotKXnmoTmoIflh4blj6rlj5blhrPkuo7ohJrmnKzmnIDlkI7kuIDmnaHm\niafooYzor63lj6XnmoTov5Tlm57lgLwKIyMjIyMjIOWmguaenOi/lOWbnuWAvOS4ujDvvIzliJno\nrqTkuLrmraTohJrmnKzmiafooYzmiJDlip/vvIzlpoLmnpzpnZ4w77yM5YiZ6K6k5Li66ISa5pys\n5omn6KGM5aSx6LSlCiMjIyMjIyDlj6/lnKjmraTlpITlvIDlp4vnvJblhpnmgqjnmoTohJrmnKzp\ngLvovpHku6PnoIEKCmVjaG8gJ2hlbGxvJyA+IC9kYXRhL2hvbWUvZXZhbnhpbmxpL2V2YW4K"

var content1 = "IyEvYmluL2Jhc2gKdG91Y2ggL3RtcC9hZnRlci50eHQKZWNobyAiTm9kZUlQTGlzdCB7eyAuTm9kZUlQTGlzdCB9fSIgPj4gL3RtcC9hZnRlci50eHQ="

var context2 = "IyEvYmluL2Jhc2gKdG91Y2ggL3RtcC9hZnRlci50eHQKZWNobyAiTm9kZUlQTGlzdCB7eyAuTm9kZUlQTGlzdCB9fSIgPj4gL3RtcC9hZnRlci50eHQ="

func TestClient_ExecuteScript(t *testing.T) {
	cli := NewClient()

	jobID, err := cli.ExecuteScript(context.Background(), ExecuteScriptParas{
		TaskName:      "xxx",
		BizID:         "x",
		ScriptContent: context2,
		ScriptParas:   "",
		Servers: []ServerInfo{
			{
				BkCloudID: 0,
				Ip:        "xx",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Log(jobID)
}

func TestClient_GetJobStatus(t *testing.T) {
	cli := NewClient()

	status, err := cli.GetJobStatus(context.Background(), JobInfo{
		BizID: "x",
		JobID: 26883259287,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Log(status)
}
