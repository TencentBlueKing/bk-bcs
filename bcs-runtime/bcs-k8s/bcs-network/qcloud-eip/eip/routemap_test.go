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

package eip

import (
	"testing"
)

func TestAddRecord(t *testing.T) {
	rm := newRouteMap("./tmp.json")
	rm.addRecord("127.0.0.1", 101)
	rm.addRecord("0.0.0.0", 100)
	if rm.Data["127.0.0.1"] != 101 {
		t.Errorf("127.0.0.1's should be %d but get %d", 101, rm.Data["127.0.0.1"])
	}
	if rm.Data["0.0.0.0"] != 100 {
		t.Errorf("0.0.0.0's should be %d but get %d", 100, rm.Data["0.0.0.0"])
	}
}

func TestLoadAndSave(t *testing.T) {
	rm := newRouteMap("./tmp.json")
	rm.addRecord("127.0.0.1", 101)
	rm.addRecord("0.0.0.0", 100)
	err := rm.saveToFile()
	if err != nil {
		t.Errorf("save to file error %s", err.Error())
	}

	newRm := newRouteMap("./tmp.json")
	err = newRm.loadFromFile()
	if err != nil {
		t.Errorf("load from file err %s", err.Error())
	}
	if newRm.Data["127.0.0.1"] != 101 {
		t.Errorf("127.0.0.1's should be %d but get %d", 101, newRm.Data["127.0.0.1"])
	}
	if newRm.Data["0.0.0.0"] != 100 {
		t.Errorf("0.0.0.0's should be %d but get %d", 100, newRm.Data["0.0.0.0"])
	}

	err = newRm.removeFile()
	if err != nil {
		t.Errorf("remove file err %s", err.Error())
	}

}
