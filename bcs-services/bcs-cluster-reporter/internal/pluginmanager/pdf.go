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

// Package pluginmanager xxx
package pluginmanager

import (
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/util"
	"github.com/jung-kurt/gofpdf"
	"k8s.io/klog/v2"
)

// GetBizReport xxx
func GetBizReport(bizID string, pluginStr string) (gofpdf.Pdf, error) {
	pdf := gofpdf.New("P", "mm", "A3", "")

	return pdf, nil
}

// GetClusterReport xxx
func GetClusterReport(clusterID string, option CheckOption) (gofpdf.Pdf, error) {
	pdf := gofpdf.New("P", "mm", "A3", "")
	clusterConfigs := Pm.GetConfig().ClusterConfigs

	clusterConfig := clusterConfigs[clusterID]
	if clusterConfig == nil {
		err := fmt.Errorf("%s not found in clusters", clusterID)
		klog.Errorf(err.Error())
		return nil, err
	}
	result := Pm.GetClusterResult(clusterID, option)

	pdf.AddPage()
	pdf.AddUTF8Font("tencent", "", "TencentSans-W3.ttf")
	pdf.AddUTF8Font("tencent", "B", "TencentSans-W7.ttf")
	pdf.SetFont("tencent", "", 12)
	pdf.SetXY(0, 10)

	// 打印集群信息
	infoTable := ConvertInfoItemToPDFTable(InfoItem{ItemName: ClusterID, Result: clusterConfig.ClusterID}, "集群信息")
	WriteClusterInfo(clusterConfig, infoTable)
	util.WritePDFTable(pdf, *infoTable, true)

	// 打印节点信息
	typeNumMap := make(map[string]int)
	regionNumMap := make(map[string]int)
	zoneNumMap := make(map[string]int)
	for _, nodeInfo := range clusterConfig.NodeInfo {
		if _, ok := typeNumMap[nodeInfo.Type]; !ok {
			typeNumMap[nodeInfo.Type] = 1
		} else {
			typeNumMap[nodeInfo.Type] = typeNumMap[nodeInfo.Type] + 1
		}

		if _, ok := regionNumMap[nodeInfo.Region]; !ok {
			regionNumMap[nodeInfo.Region] = 1
		} else {
			regionNumMap[nodeInfo.Region] = regionNumMap[nodeInfo.Region] + 1
		}

		if _, ok := zoneNumMap[nodeInfo.Zone]; !ok {
			zoneNumMap[nodeInfo.Zone] = 1
		} else {
			zoneNumMap[nodeInfo.Zone] = zoneNumMap[nodeInfo.Zone] + 1
		}
	}

	nodeInfoTable := NewInfoItemToPDFTable("节点信息")
	for key, value := range typeNumMap {
		AddInfoItemToPDFTable(InfoItem{ItemName: "机型:" + key, Result: value}, nodeInfoTable)
	}
	for key, value := range regionNumMap {
		AddInfoItemToPDFTable(InfoItem{ItemName: "地域:" + key, Result: value}, nodeInfoTable)
	}
	for key, value := range zoneNumMap {
		AddInfoItemToPDFTable(InfoItem{ItemName: "可用区:" + key, Result: value}, nodeInfoTable)
	}
	util.WritePDFTable(pdf, *nodeInfoTable, true)

	// 打印各类检查项
	for pluginName, checkItemList := range result {
		checkItemTable := NewCheckItemPDFTable(pluginName)

		for _, checkItem := range checkItemList.Items {
			//if checkItem.Normal {
			//	continue
			//}
			AddCheckItemToPDFTable(checkItem, checkItemTable)
		}

		if checkItemTable.Line == 0 {
			continue
		}
		util.WritePDFTable(pdf, *checkItemTable, true)
	}
	return pdf, nil
}

// WriteClusterInfo xxx
func WriteClusterInfo(clusterConfig *ClusterConfig, infoTable *util.PDFTable) {
	if clusterConfig.BCSCluster.ClusterID != "" {
		AddInfoItemToPDFTable(InfoItem{ItemName: "ClusterName", Result: clusterConfig.BCSCluster.ClusterName}, infoTable)
		AddInfoItemToPDFTable(InfoItem{ItemName: "Creator", Result: clusterConfig.BCSCluster.Creator}, infoTable)
		AddInfoItemToPDFTable(InfoItem{ItemName: "Managetype", Result: clusterConfig.BCSCluster.ManageType}, infoTable)
		AddInfoItemToPDFTable(InfoItem{ItemName: "CreateTime", Result: clusterConfig.BCSCluster.CreateTime}, infoTable)
		AddInfoItemToPDFTable(InfoItem{ItemName: "Systemid", Result: clusterConfig.BCSCluster.SystemID}, infoTable)
		AddInfoItemToPDFTable(InfoItem{ItemName: "Vpc", Result: clusterConfig.BCSCluster.VpcID}, infoTable)
	}
	AddInfoItemToPDFTable(InfoItem{ItemName: "ClusterType", Result: clusterConfig.ClusterType}, infoTable)
	AddInfoItemToPDFTable(InfoItem{ItemName: "BusinessID", Result: clusterConfig.BusinessID}, infoTable)
	AddInfoItemToPDFTable(InfoItem{ItemName: "Master", Result: clusterConfig.Master}, infoTable)
	AddInfoItemToPDFTable(InfoItem{ItemName: "ServiceCidr", Result: clusterConfig.ServiceCidr}, infoTable)
	AddInfoItemToPDFTable(InfoItem{ItemName: "ServiceMaxNum", Result: clusterConfig.ServiceMaxNum}, infoTable)
	AddInfoItemToPDFTable(InfoItem{ItemName: "ServiceNum", Result: clusterConfig.ServiceNum}, infoTable)
	AddInfoItemToPDFTable(InfoItem{ItemName: "Cidr", Result: clusterConfig.Cidr}, infoTable)
	AddInfoItemToPDFTable(InfoItem{ItemName: "NodeNum", Result: clusterConfig.NodeNum}, infoTable)
}

// NewInfoItemToPDFTable xxx
func NewInfoItemToPDFTable(title string) *util.PDFTable {
	keys := make([]util.Column, 0, 0)

	result := &util.PDFTable{
		Header: append(keys, util.Column{Content: StringMap[CheckItemName]}, util.Column{Content: StringMap[CheckItemResult]}),
		Title:  util.Column{Content: title},
		Data:   [][]util.Column{},
	}

	return result
}

// ConvertInfoItemToPDFTable xxx
func ConvertInfoItemToPDFTable(item InfoItem, title string) *util.PDFTable {
	keys := make([]util.Column, 0, 0)

	result := &util.PDFTable{
		Header: append(keys, util.Column{Content: StringMap[CheckItemName]}, util.Column{Content: StringMap[CheckItemResult]}),
		Title:  util.Column{Content: item.ItemName},
		Data:   [][]util.Column{},
	}

	if title != "" {
		result.Title.Content = title
	}

	AddInfoItemToPDFTable(item, result)
	return result
}

// AddInfoItemToPDFTable xxx
func AddInfoItemToPDFTable(item InfoItem, table *util.PDFTable) {
	values := make([]util.Column, 0, 0)
	values = append(values, util.Column{Content: item.ItemName}, util.Column{Content: fmt.Sprintf("%v", item.Result)})

	table.Data = append(table.Data, values)
}

// NewCheckItemPDFTable xxx
func NewCheckItemPDFTable(title string) *util.PDFTable {
	keys := make([]util.Column, 0, 0)

	result := &util.PDFTable{
		Header: append(keys, util.Column{Content: StringMap[CheckItemName]}, util.Column{Content: StringMap[CheckItemTarget]},
			util.Column{Content: StringMap[CheckItemLevel]}, util.Column{Content: StringMap[CheckItemResult]}, util.Column{Content: StringMap[CheckItemDetail]}),
		Title: util.Column{Content: title},
		Data:  [][]util.Column{},
	}

	return result
}

// AddCheckItemToPDFTable xxx
func AddCheckItemToPDFTable(item CheckItem, table *util.PDFTable) {
	values := make([]util.Column, 0, 0)
	if !item.Normal {
		values = append(values, util.Column{Content: item.ItemName}, util.Column{Content: item.ItemTarget}, util.Column{Content: item.Level},
			util.Column{Content: item.Status, Color: util.Color{Red: 238, Green: 56, Blue: 43}}, util.Column{Content: item.Detail})
	} else {
		values = append(values, util.Column{Content: item.ItemName}, util.Column{Content: item.ItemTarget}, util.Column{Content: item.Level},
			util.Column{Content: item.Status}, util.Column{Content: item.Detail})
	}
	table.Line++

	table.Data = append(table.Data, values)
}

// SolutionTable xxx
type SolutionTable struct {
	ItemName   util.Column
	ItemType   util.Column
	ItemTarget util.Column
	Level      util.Column
	Result     util.Column
	Advise     util.Column
}
