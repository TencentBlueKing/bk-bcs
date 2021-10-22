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

package model

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
)

func init() {
	blog.InitLogs(
		conf.LogConfig{
			LogDir:       "",
			LogMaxSize:   500,
			LogMaxNum:    10,
			ToStdErr:     true,
			AlsoToStdErr: true,
			Verbosity:    5,
		},
	)
}

func createTmpDir(prefix string) (string, error) {
	return ioutil.TempDir(os.TempDir(), prefix)
}

func cleanTmpDir(dir string, t *testing.T) {
	err := os.RemoveAll(dir)
	if err != nil {
		t.Fatalf("failed to clean tmp dir %s, err %s", dir, err.Error())
	}
}

func createFile(dir string, filename string, content []byte) error {
	f, err := os.Create(filepath.Join(dir, filename))
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(content)
	if err != nil {
		return err
	}
	return nil
}

func overwriteFile(dir string, filename string, content []byte) error {
	f, err := os.OpenFile(filepath.Join(dir, filename), os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(content)
	if err != nil {
		return err
	}
	return nil
}

func appendFile(dir string, filename string, content []byte) error {
	f, err := os.OpenFile(filepath.Join(dir, filename), os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(content)
	if err != nil {
		return err
	}
	return nil
}

func renameFile(dir string, oldname string, newname string) error {
	return os.Rename(filepath.Join(dir, oldname), filepath.Join(dir, newname))
}

func deleteFile(dir string, filename string) error {
	err := os.Remove(filepath.Join(dir, filename))
	if err != nil {
		return fmt.Errorf("delete file %s failed, err %s", filepath.Join(dir, filename), err.Error())
	}
	return nil
}

func TestDirWatchAddUpdateDelete(t *testing.T) {
	dir, err := createTmpDir("k8s-clb-test")
	if err != nil {
		t.Fatalf("failed to create tmp dir with prefix %s, err %s", "k8s-clb-test", err.Error())
	}
	fmt.Printf("created dir %s\n", dir)
	defer cleanTmpDir(dir, t)

	watcher := NewFileWatcher(time.Minute * 1)

	ch, err := watcher.WatchDir(dir)
	if err != nil {
		t.Fatalf("failed to watch dir %s, err %s", dir, err.Error())
	}

	time.Sleep(time.Second * 3)

	err = createFile(dir, "tmp1", []byte("test content"))
	if err != nil {
		t.Fatalf("failed to create file, err %s", err.Error())
	}

	event := <-ch
	if event.Type != FileEventCreate || event.Filename != filepath.Join(dir, "tmp1") {
		t.Fatalf("expect %d-%s, but get %d-%s", FileEventCreate, filepath.Join(dir, "tmp1"), event.Type, event.Filename)
	}

	err = overwriteFile(dir, "tmp1", []byte("updated content"))
	if err != nil {
		t.Fatalf("failed to overwriteFile %s", filepath.Join(dir, "tmp1"))
	}

	event = <-ch
	if event.Type != FileEventUpdate || event.Filename != filepath.Join(dir, "tmp1") {
		t.Fatalf("expect %d-%s, but get %d-%s", FileEventUpdate, filepath.Join(dir, "tmp1"), event.Type, event.Filename)
	}

	err = renameFile(dir, "tmp1", "tmp2")
	if err != nil {
		t.Fatalf("failed to renameFile from %s to %s", filepath.Join(dir, "tmp1"), filepath.Join(dir, "tmp2"))
	}

	event = <-ch
	if event.Type != FileEventDelete || event.Filename != filepath.Join(dir, "tmp1") {
		t.Fatalf("expect %d-%s, but get %d-%s", FileEventDelete, filepath.Join(dir, "tmp1"), event.Type, event.Filename)
	}

	event = <-ch
	if event.Type != FileEventCreate || event.Filename != filepath.Join(dir, "tmp2") {
		t.Fatalf("expect %d-%s, but get %d-%s", FileEventDelete, filepath.Join(dir, "tmp2"), event.Type, event.Filename)
	}

	err = deleteFile(dir, "tmp2")
	if err != nil {
		t.Fatalf("failed to delete, err %s", err.Error())
	}

	event = <-ch
	if event.Type != FileEventDelete || event.Filename != filepath.Join(dir, "tmp2") {
		t.Fatalf("expect %d-%s, but get %d-%s", FileEventDelete, filepath.Join(dir, "tmp2"), event.Type, event.Filename)
	}
}

func TestFileWatchAddUpdateDelete(t *testing.T) {
	dir, err := createTmpDir("k8s-clb-test")
	if err != nil {
		t.Fatalf("failed to create tmp dir with prefix %s, err %s", "k8s-clb-test", err.Error())
	}
	fmt.Printf("created dir %s\n", dir)
	defer cleanTmpDir(dir, t)

	watcher := NewFileWatcher(time.Minute * 1)

	err = createFile(dir, "tmp1", []byte("test content"))
	if err != nil {
		t.Fatalf("failed to create file, err %s", err.Error())
	}

	time.Sleep(time.Second * 3)

	ch, err := watcher.WatchFile(filepath.Join(dir, "tmp1"))
	if err != nil {
		t.Fatalf("failed to watch dir %s, err %s", dir, err.Error())
	}

	err = overwriteFile(dir, "tmp1", []byte("updated content"))
	if err != nil {
		t.Fatalf("failed to overwriteFile %s", filepath.Join(dir, "tmp1"))
	}

	event := <-ch
	if event.Type != FileEventUpdate || event.Filename != filepath.Join(dir, "tmp1") {
		t.Fatalf("expect %d-%s, but get %d-%s", FileEventUpdate, filepath.Join(dir, "tmp1"), event.Type, event.Filename)
	}

	err = deleteFile(dir, "tmp1")
	if err != nil {
		t.Fatalf("failed to delete, err %s", err.Error())
	}

	event = <-ch
	if event.Type != FileEventDelete || event.Filename != filepath.Join(dir, "tmp1") {
		t.Fatalf("expect %d-%s, but get %d-%s", FileEventDelete, filepath.Join(dir, "tmp1"), event.Type, event.Filename)
	}
}
