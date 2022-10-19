/*
Copyright 2015 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package editor

import (
	"bytes"
	"io/ioutil"
	"k8s.io/kubectl/pkg/cmd/util/editor"
	"os"
	"reflect"
	"sigs.k8s.io/yaml"
	"strings"
	"testing"
)

func TestArgs(t *testing.T) {
	if e, a := []string{"/bin/bash", "-c \"test\""}, (Editor{Args: []string{"/bin/bash", "-c"}, Shell: true}).args("test"); !reflect.DeepEqual(e, a) {
		t.Errorf("unexpected args: %v", a)
	}
	if e, a := []string{"/bin/bash", "-c", "test"}, (Editor{Args: []string{"/bin/bash", "-c"}, Shell: false}).args("test"); !reflect.DeepEqual(e, a) {
		t.Errorf("unexpected args: %v", a)
	}
	if e, a := []string{"/bin/bash", "-i -c \"test\""}, (Editor{Args: []string{"/bin/bash", "-i -c"}, Shell: true}).args("test"); !reflect.DeepEqual(e, a) {
		t.Errorf("unexpected args: %v", a)
	}
	if e, a := []string{"/test", "test"}, (Editor{Args: []string{"/test"}}).args("test"); !reflect.DeepEqual(e, a) {
		t.Errorf("unexpected args: %v", a)
	}
}

func TestEditor(t *testing.T) {
	//go test -v editor_test.go -test.run TestEditor
	edit := editor.NewDefaultEditor([]string{})
	testStr := "test something\n"
	y, err := yaml.JSONToYAML([]byte(`{
    "code": 0,
    "message": "success",
    "data": {
        "createTime": "2022-10-17T15:46:48+08:00",
        "updateTime": "2022-10-17T16:26:56+08:00",
        "creator": "admin",
        "updater": "admin",
        "managers": "admin,",
        "projectID": "7da12ea6af35464a8be39961a21e95d9",
        "name": "测试2-1",
        "projectCode": "7da12ea6af35464a8be39961a21e95d9",
        "useBKRes": false,
        "description": "string",
        "isOffline": false,
        "kind": "k8s",
        "businessID": "10000",
        "isSecret": true,
        "projectType": 0,
        "deployType": 0,
        "BGID": "string",
        "BGName": "string",
        "deptID": "string",
        "deptName": "string",
        "centerID": "string",
        "centerName": "string"
    },
    "requestID": "66c0455fa9ce4f8b88eceeeb0eabaefd",
    "webAnnotations": {
        "perms": null
    }
}`))

	contents, path, err := edit.LaunchTempFile("", "someprefix", bytes.NewBufferString(string(y)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("no temp file: %s", path)
	}
	defer os.Remove(path)
	if disk, err := ioutil.ReadFile(path); err != nil || !bytes.Equal(contents, disk) {
		t.Errorf("unexpected file on disk: %v %s", err, string(disk))
	}
	if !bytes.Equal(contents, []byte(testStr)) {
		t.Errorf("unexpected contents: %s", string(contents))
	}
	if !strings.Contains(path, "someprefix") {
		t.Errorf("path not expected: %s", path)
	}
}
