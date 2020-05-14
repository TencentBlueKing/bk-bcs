/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package testdata

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	pb "bk-bscp/internal/protocol/accessserver"
)

func randName(prefix string) string {
	return fmt.Sprintf("%s-%s-%d", prefix, time.Now().Format("2006-01-02-15:04:05"), time.Now().Nanosecond())
}

// CreateBusinessTestData returns test data for create business case.
func CreateBusinessTestData() (string, error) {
	req := &pb.CreateBusinessReq{
		Seq:     0,
		Name:    randName("e2e-bu"),
		Depid:   "e2e",
		Dbid:    "db01",
		Dbname:  "testdb",
		Creator: "e2e",
		Memo:    "e2e testing",
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}

// UpdateBusinessTestData returns test data for update business case.
func UpdateBusinessTestData(bid string) (string, error) {
	req := &pb.UpdateBusinessReq{
		Seq:      0,
		Bid:      bid,
		Name:     randName("e2e-bu"),
		Depid:    "e2e update",
		State:    0,
		Operator: "e2e",
		Memo:     "e2e testing update",
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}

// CreateAppTestData returns test data for create application case.
func CreateAppTestData(bid string) (string, error) {
	req := &pb.CreateAppReq{
		Seq:        0,
		Bid:        bid,
		Name:       randName("e2e-app"),
		DeployType: 0,
		Creator:    "e2e",
		Memo:       "e2e testing",
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}

// UpdateAppTestData returns test data for update application case.
func UpdateAppTestData(bid, appid string) (string, error) {
	req := &pb.UpdateAppReq{
		Seq:        0,
		Bid:        bid,
		Appid:      appid,
		Name:       randName("e2e-app"),
		DeployType: 0,
		Memo:       "e2e testing update",
		State:      0,
		Operator:   "e2e",
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}

// CreateClusterTestData returns test data for create cluster case.
func CreateClusterTestData(bid, appid string) (string, error) {
	req := &pb.CreateClusterReq{
		Seq:        0,
		Bid:        bid,
		Name:       randName("e2e-cluster"),
		Appid:      appid,
		RClusterid: "rclusterid",
		Creator:    "e2e",
		Memo:       "e2e testing",
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}

// CreateClusterTestDataWithLabel returns test data for create cluster case.
func CreateClusterTestDataWithLabel(bid, appid string) (string, error) {
	req := &pb.CreateClusterReq{
		Seq:        0,
		Bid:        bid,
		Name:       randName("e2e-cluster"),
		Appid:      appid,
		RClusterid: "rclusterid",
		Creator:    "e2e",
		Memo:       "e2e testing",
		Labels:     "{\"environment\":\"test\"}",
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}

// UpdateClusterTestData returns test data for update cluster case.
func UpdateClusterTestData(bid, clusterid string) (string, error) {
	req := &pb.UpdateClusterReq{
		Seq:        0,
		Bid:        bid,
		Clusterid:  clusterid,
		Name:       randName("e2e-cluster"),
		RClusterid: "rclusterid update",
		Memo:       "e2e testing update",
		State:      0,
		Operator:   "e2e",
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}

// CreateZoneTestData returns test data for create zone case.
func CreateZoneTestData(bid, appid, clusterid string) (string, error) {
	req := &pb.CreateZoneReq{
		Seq:       0,
		Bid:       bid,
		Appid:     appid,
		Clusterid: clusterid,
		Name:      randName("e2e-zone"),
		Creator:   "e2e",
		Memo:      "e2e testing",
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}

// UpdateZoneTestData returns test data for update zone case.
func UpdateZoneTestData(bid, zoneid string) (string, error) {
	req := &pb.UpdateZoneReq{
		Seq:      0,
		Bid:      bid,
		Zoneid:   zoneid,
		Name:     randName("e2e-zone"),
		Memo:     "e2e testing update",
		State:    0,
		Operator: "e2e",
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}

// CreateConfigSetTestData returns test data for create config set case.
func CreateConfigSetTestData(bid, appid string) (string, error) {
	req := &pb.CreateConfigSetReq{
		Seq:     0,
		Bid:     bid,
		Appid:   appid,
		Name:    randName("e2e-configset"),
		Creator: "e2e",
		Memo:    "e2e testing",
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}

// UpdateConfigSetTestData returns test data for update config set case.
func UpdateConfigSetTestData(bid, cfgsetid string) (string, error) {
	req := &pb.UpdateConfigSetReq{
		Seq:      0,
		Bid:      bid,
		Cfgsetid: cfgsetid,
		Name:     randName("e2e-configset"),
		Memo:     "e2e testing update",
		State:    0,
		Operator: "e2e",
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}

// LockConfigSetTestData returns test data for lock config set case.
func LockConfigSetTestData(bid, cfgsetid string) (string, error) {
	req := &pb.LockConfigSetReq{
		Seq:      0,
		Bid:      bid,
		Cfgsetid: cfgsetid,
		Operator: randName("e2e-operator"),
		Memo:     "e2e testing",
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}

// UnlockConfigSetTestData returns test data for unlock config set case.
func UnlockConfigSetTestData(bid, cfgsetid, operator string) (string, error) {
	if len(operator) == 0 {
		operator = randName("e2e-operator")
	}

	req := &pb.UnlockConfigSetReq{
		Seq:      0,
		Bid:      bid,
		Cfgsetid: cfgsetid,
		Operator: operator,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}

// CreateCommitTestData returns test data for create commit case.
func CreateCommitTestData(bid, appid, cfgsetid string) (string, error) {
	req := &pb.CreateCommitReq{
		Seq:      0,
		Bid:      bid,
		Appid:    appid,
		Cfgsetid: cfgsetid,
		Op:       0,
		Operator: randName("e2e-operator"),
		Configs:  []byte("e2e-configs"),
		Changes:  "e2e changes",
		Memo:     "e2e testing",
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}

// CreateCommitWithTplTestData returns test data for create commit with template case.
func CreateCommitWithTplTestData(bid, appid, cfgsetid, clusterName, zoneName string) (string, error) {
	req := &pb.CreateCommitReq{
		Seq:          0,
		Bid:          bid,
		Appid:        appid,
		Cfgsetid:     cfgsetid,
		Op:           0,
		Operator:     randName("e2e-operator"),
		Templateid:   randName("e2e-tplid"),
		Template:     "{{ .k }}",
		TemplateRule: fmt.Sprintf("[{\"type\": 0, \"name\": \"%s\", \"vars\": {\"k\": \"v1\"}}, {\"type\": 1, \"name\": \"%s\", \"vars\": {\"k\": \"v2\"}}]", clusterName, zoneName),
		Changes:      "e2e changes",
		Memo:         "e2e testing",
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}

// UpdateCommitTestData returns test data for update commit case.
func UpdateCommitTestData(bid, commitid string) (string, error) {
	req := &pb.UpdateCommitReq{
		Seq:      0,
		Bid:      bid,
		Commitid: commitid,
		Configs:  []byte("e2e configs update"),
		Changes:  "e2e changes update",
		Memo:     "e2e testing update",
		Operator: "e2e",
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}

// ConfirmCommitTestData returns test data for confirm commit case.
func ConfirmCommitTestData(bid, commitid string) (string, error) {
	req := &pb.ConfirmCommitReq{
		Seq:      0,
		Bid:      bid,
		Commitid: commitid,
		Operator: "e2e",
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}

// CancelCommitTestData returns test data for cancel commit case.
func CancelCommitTestData(bid, commitid string) (string, error) {
	req := &pb.CancelCommitReq{
		Seq:      0,
		Bid:      bid,
		Commitid: commitid,
		Operator: "e2e",
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}

// CreateReleaseTestData returns test data for create release case.
func CreateReleaseTestData(bid, commitid, strategyid string) (string, error) {
	req := &pb.CreateReleaseReq{
		Seq:        0,
		Bid:        bid,
		Name:       randName("e2e-release"),
		Commitid:   commitid,
		Strategyid: strategyid,
		Memo:       "e2e testing",
		Creator:    "e2e",
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}

// UpdateReleaseTestData returns test data for update release case.
func UpdateReleaseTestData(bid, releaseid string) (string, error) {
	req := &pb.UpdateReleaseReq{
		Seq:       0,
		Bid:       bid,
		Releaseid: releaseid,
		Name:      randName("e2e-release"),
		Memo:      "e2e testing update",
		Operator:  "e2e",
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}

// PublishReleaseTestData returns test data for publish release case.
func PublishReleaseTestData(bid, releaseid string) (string, error) {
	req := &pb.PublishReleaseReq{
		Seq:       0,
		Bid:       bid,
		Releaseid: releaseid,
		Operator:  "e2e",
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}

// CancelReleaseTestData returns test data for cancel release case.
func CancelReleaseTestData(bid, releaseid string) (string, error) {
	req := &pb.CancelReleaseReq{
		Seq:       0,
		Bid:       bid,
		Releaseid: releaseid,
		Operator:  "e2e",
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}

// CreateStrategyTestData returns test data for create strategy case.
func CreateStrategyTestData(bid, appid string) (string, error) {
	req := &pb.CreateStrategyReq{
		Seq:        0,
		Bid:        bid,
		Appid:      appid,
		Name:       randName("e2e-strategy"),
		Clusterids: []string{},
		Zoneids:    []string{},
		Dcs:        []string{},
		IPs:        []string{},
		Labels:     map[string]string{},
		Memo:       "e2e testing",
		Creator:    "e2e",
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}

// CreateVarTestData returns test data for create vars
func CreateVarTestData(varType int32, bid, cluster, clusterLabels, zone string) (string, error) {
	req := &pb.CreateVariableReq{
		Seq:           0,
		Bid:           bid,
		Cluster:       cluster,
		ClusterLabels: clusterLabels,
		Zone:          zone,
		Type:          varType,
		Key:           randName("e2e-var"),
		Value:         randName("e2e-var-value"),
		Memo:          "e2e test vars",
		Creator:       "e2e",
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}

// UpdateVarTestData returns update data for vars
func UpdateVarTestData(varType int32, bid, vid string) (string, error) {
	req := &pb.UpdateVariableReq{
		Seq:      0,
		Bid:      bid,
		Vid:      vid,
		Type:     varType,
		Key:      randName("e2e-var"),
		Value:    randName("e2e-var-value"),
		Memo:     "e2e test vars",
		Operator: "e2e",
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}

// CreateConfigTemplateSetTestData return create data for config template set
func CreateConfigTemplateSetTestData(bid, fpath string) (string, error) {
	req := &pb.CreateConfigTemplateSetReq{
		Seq:     0,
		Bid:     bid,
		Fpath:   fpath,
		Name:    randName("e2e-tpl-set"),
		Memo:    "e2e test create template set",
		Creator: "e2e",
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}

// UpdateConfigTemplateSetTestData return update data for config template set
func UpdateConfigTemplateSetTestData(bid, setid string) (string, error) {
	req := &pb.UpdateConfigTemplateSetReq{
		Seq:      0,
		Bid:      bid,
		Setid:    setid,
		Name:     randName("e2e-tpl-set"),
		Memo:     "e2e test update template set",
		Operator: "e2e",
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}

// CreateConfigTemplateTestData returns create data for config template
func CreateConfigTemplateTestData(bid, setid string) (string, error) {
	req := &pb.CreateConfigTemplateReq{
		Seq:          0,
		Bid:          bid,
		Setid:        setid,
		Name:         randName("config") + ".txt",
		Memo:         "e2e test create config template",
		User:         "root",
		Group:        "root",
		Permission:   644,
		FileEncoding: "utf8",
		EngineType:   0,
		Creator:      "e2e",
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}

// UpdateConfigTemplateTestData returns update data for config template
func UpdateConfigTemplateTestData(bid, tid string) (string, error) {
	req := &pb.UpdateConfigTemplateReq{
		Seq:          0,
		Bid:          bid,
		Templateid:   tid,
		Name:         randName("config") + ".txt",
		Memo:         "e2e test update config template",
		User:         "rootupdate",
		Group:        "rootupdate",
		Permission:   645,
		FileEncoding: "gbk",
		Operator:     "e2e",
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}

// CreateTemplateVersionTestData returns create data for config template version
func CreateTemplateVersionTestData(bid, templateid string) (string, error) {
	str := `
	{{ .TITLE }}
君不見黃河之水天上來，奔流到海不復回！
君不見高堂明鏡悲白髮，朝如青絲暮成雪。
人生得意須盡歡，莫使金樽空對月。
天生我材必有用，千金散盡還復來。
烹羊宰牛且為樂，會須一飲三百杯。
岑夫子，丹丘生，將進酒，杯莫停。
與君歌一曲，請君為我傾耳聽。
鐘鼓饌玉何足貴，但願長醉不復醒。
古來聖賢皆寂寞，唯有飲者留其名。
陳王昔時宴平樂，斗酒十千恣讙謔。
主人何為言少錢，徑須沽取對君酌。
五花馬，千金裘，
呼兒將出換美酒，與爾同銷萬古愁。

作者: {{ .AUTHOR }}
朝代: {{ .DEST }}

暗号: {{ .InnerIP }}
地点: {{ .Placement }}
`
	req := &pb.CreateTemplateVersionReq{
		Seq:         0,
		Bid:         bid,
		Templateid:  templateid,
		VersionName: randName("version"),
		Memo:        "e2e template version",
		Creator:     "e2e",
		Content:     str,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}

// UpdateTemplateVersionTestData returns update data for config template version
func UpdateTemplateVersionTestData(bid, versionid string) (string, error) {
	str := `
	{{ .TITLE }}
君不見黃河之水天上來，奔流到海不復回！
君不見高堂明鏡悲白髮，朝如青絲暮成雪。
人生得意須盡歡，莫使金樽空對月。
天生我材必有用，千金散盡還復來。
烹羊宰牛且為樂，會須一飲三百杯。
岑夫子，丹丘生，將進酒，杯莫停。
與君歌一曲，請君為我傾耳聽。
鐘鼓饌玉何足貴，但願長醉不復醒。
古來聖賢皆寂寞，唯有飲者留其名。
陳王昔時宴平樂，斗酒十千恣讙謔。
主人何為言少錢，徑須沽取對君酌。
五花馬，千金裘，
呼兒將出換美酒，與爾同銷萬古愁。

诗人: {{ .AUTHOR }}
朝代: {{ .DEST }}

暗号: {{ .InnerIP }}
地点: {{ .Placement }}
`
	req := &pb.UpdateTemplateVersionReq{
		Seq:         0,
		Bid:         bid,
		Versionid:   versionid,
		VersionName: randName("version"),
		Memo:        "e2e update template version",
		Operator:    "e2e",
		Content:     str,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}

// CreateTemplateBindingTestData returns create data for template binding
func CreateTemplateBindingTestData(bid, templateid, appid, versionid, bindingParam string) (string, error) {
	req := &pb.CreateConfigTemplateBindingReq{
		Seq:           0,
		Bid:           bid,
		Templateid:    templateid,
		Appid:         appid,
		Versionid:     versionid,
		BindingParams: bindingParam,
		Creator:       "e2e",
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}

// UpdateTemplateBindingTestData returns update data for template binding
func UpdateTemplateBindingTestData(bid, templateid, appid, versionid, bindingParam string) (string, error) {
	req := &pb.SyncConfigTemplateBindingReq{
		Seq:           0,
		Bid:           bid,
		Templateid:    templateid,
		Appid:         appid,
		Versionid:     versionid,
		BindingParams: bindingParam,
		Operator:      "e2e",
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}

// CreateCertainVarTestData returns create data for certain var
func CreateCertainVarTestData(bid string, varType int, cluster, clusterLabels, zone, key, value string) (string, error) {
	req := &pb.CreateVariableReq{
		Seq:           0,
		Bid:           bid,
		Cluster:       cluster,
		ClusterLabels: clusterLabels,
		Zone:          zone,
		Type:          int32(varType),
		Key:           key,
		Value:         value,
		Memo:          "e2e create var",
		Creator:       "e2e",
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}

// CreatePreviewCommitTestData returns create data for preview commit
func CreatePreviewCommitTestData(bid, commitid string) (string, error) {
	req := &pb.PreviewCommitReq{
		Seq:      0,
		Bid:      bid,
		Commitid: commitid,
		Operator: "e2e",
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}

// CreateConfirmCommitWithTemplateTestData returns confirm commit data for render
func CreateConfirmCommitWithTemplateTestData(bid string, commitid string) (string, error) {
	req := &pb.ConfirmCommitReq{
		Seq:      0,
		Bid:      bid,
		Commitid: commitid,
		Operator: "e2e",
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("test data empty")
	}
	return string(data), nil
}
