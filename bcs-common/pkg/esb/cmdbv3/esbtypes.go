/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmdbv3

import (
	"time"
)

// ESBBaseResp base resp from esb
type ESBBaseResp struct {
	Code      int64  `json:"code"`
	Result    bool   `json:"result"`
	Message   string `json:"message"`
	ReqeustID string `json:"request_id,omitempty"`
}

// ESBBusinessInfo business info struct
type ESBBusinessInfo struct {
	BkAlarmRvcMan     string      `json:"bk_alarm_rvc_man"`
	BkAppAbbr         string      `json:"bk_app_abbr"`
	BkAppDevBak       string      `json:"bk_app_dev_bak"`
	BkAppDevteam      string      `json:"bk_app_devteam"`
	BkAppDirector     string      `json:"bk_app_director"`
	BkAppForumURL     string      `json:"bk_app_forum_url"`
	BkAppGameTypeid   string      `json:"bk_app_game_typeid"`
	BkAppImptLevel    string      `json:"bk_app_impt_level"`
	BkAppOperManual   string      `json:"bk_app_oper_manual"`
	BkAppSummary      string      `json:"bk_app_summary"`
	BkAppType         string      `json:"bk_app_type"`
	BkAppURL          string      `json:"bk_app_url"`
	BkAppUserManual   string      `json:"bk_app_user_manual"`
	BkArcDoc          string      `json:"bk_arc_doc"`
	BkBipAppName      string      `json:"bk_bip_app_name"`
	BkBipID           string      `json:"bk_bip_id"`
	BkBizDeveloper    string      `json:"bk_biz_developer"`
	BkBizID           int64       `json:"bk_biz_id"`
	BkBizMaintainer   string      `json:"bk_biz_maintainer"`
	BkBizName         string      `json:"bk_biz_name"`
	BkBizProductor    string      `json:"bk_biz_productor"`
	BkBizTester       string      `json:"bk_biz_tester"`
	BkDbaBak          string      `json:"bk_dba_bak"`
	BkDeptNameID      string      `json:"bk_dept_name_id"`
	BkIdipID          string      `json:"bk_idip_id"`
	BkIsBip           string      `json:"bk_is_bip"`
	BkMblQqAppid      string      `json:"bk_mbl_qq_appid"`
	BkOperGrpName     string      `json:"bk_oper_grp_name"`
	BkOperGrpNameID   int64       `json:"bk_oper_grp_name_id"`
	BkOperPlan        string      `json:"bk_oper_plan"`
	BkOperateDeptID   int64       `json:"bk_operate_dept_id"`
	BkOperateDeptName string      `json:"bk_operate_dept_name"`
	BkPmpCmMan        string      `json:"bk_pmp_cm_man"`
	BkPmpCmreqman     string      `json:"bk_pmp_cmreqman"`
	BkPmpComPlot      string      `json:"bk_pmp_com_plot"`
	BkPmpDbaMajor     string      `json:"bk_pmp_dba_major"`
	BkPmpGroupUser    string      `json:"bk_pmp_group_user"`
	BkPmpIdipMan      string      `json:"bk_pmp_idip_man"`
	BkPmpLogo         string      `json:"bk_pmp_logo"`
	BkPmpOpeExpert    string      `json:"bk_pmp_ope_expert"`
	BkPmpOpePm        string      `json:"bk_pmp_ope_pm"`
	BkPmpOperDevMan   string      `json:"bk_pmp_oper_dev_man"`
	BkPmpOssMan       string      `json:"bk_pmp_oss_man"`
	BkPmpPotlMan      string      `json:"bk_pmp_potl_man"`
	BkPmpQa           string      `json:"bk_pmp_qa"`
	BkPmpQc           string      `json:"bk_pmp_qc"`
	BkPmpSa           string      `json:"bk_pmp_sa"`
	BkPmpSafeMan      string      `json:"bk_pmp_safe_man"`
	BkPmpSensCol      string      `json:"bk_pmp_sens_col"`
	BkPmpSvcPm        string      `json:"bk_pmp_svc_pm"`
	BkPmpTestTm       string      `json:"bk_pmp_test_tm"`
	BkProductName     string      `json:"bk_product_name"`
	BkSourceID        string      `json:"bk_source_id"`
	BkSupplierAccount string      `json:"bk_supplier_account"`
	BkTclsID          string      `json:"bk_tcls_id"`
	BkTcmID           string      `json:"bk_tcm_id"`
	BkTestResource    string      `json:"bk_test_resource"`
	BkTlogMan         string      `json:"bk_tlog_man"`
	BkVaskeyID        string      `json:"bk_vaskey_id"`
	BkVipdlID         string      `json:"bk_vipdl_id"`
	BkVisitorAppid    string      `json:"bk_visitor_appid"`
	BkWechatAppid     string      `json:"bk_wechat_appid"`
	Bs2NameID         int64       `json:"bs2_name_id"`
	BusinessDeptID    int64       `json:"business_dept_id"`
	BusinessDeptName  string      `json:"business_dept_name"`
	Default           int64       `json:"default"`
	Language          string      `json:"language"`
	LastTime          interface{} `json:"last_time"`
	LifeCycle         string      `json:"life_cycle"`
	Operator          string      `json:"operator"`
	TimeZone          string      `json:"time_zone"`
}

// ESBBusinessInfoList business info list for search business result
type ESBBusinessInfoList struct {
	Count int64             `json:"count"`
	Info  []ESBBusinessInfo `json:"info"`
}

// ESBSearchBusinessResult result for search business
type ESBSearchBusinessResult struct {
	ESBBaseResp `json:",inline"`
	Data        *ESBBusinessInfoList `json:"data"`
}

// ESBTopoInst topo instance
type ESBTopoInst struct {
	HostCount  int64         `json:"host_count"`
	Default    int64         `json:"default"`
	BkObjName  string        `json:"bk_obj_name"`
	BkObjID    string        `json:"bk_obj_id"`
	Child      []ESBTopoInst `json:"child"`
	BkInstID   int64         `json:"bk_inst_id"`
	BkInstName string        `json:"bk_inst_name"`
}

// ESBSearchBizInstTopoResult result for search business inst topo
type ESBSearchBizInstTopoResult struct {
	ESBBaseResp `json:",inline"`
	Data        *ESBTopoInst `json:"data"`
}

// ESBTransferHostModuleResult result for transfer host module
type ESBTransferHostModuleResult struct {
	ESBBaseResp `json:",inline"`
}

// ESBListHostsWitoutBizRequest request for list biz hosts
type ESBListHostsWitoutBizRequest struct {
	Page               *BasePage              `json:"page"`
	BkBizID            int64                  `json:"bk_biz_id"`
	HostPropertyFilter map[string]interface{} `json:"host_property_filter,omitempty"`
}

// ESBHostBsInfo bk bs info for host
type ESBHostBsInfo struct {
	Bs3Name   string `json:"bs3_name"`
	Bs3NameID int64  `json:"bs3_name_id"`
	Bs1NameID int64  `json:"bs1_name_id"`
	Bs2NameID int64  `json:"bs2_name_id"`
	Bs2Name   string `json:"bs2_name"`
	Bs1Name   string `json:"bs1_name"`
}

// ESBHostInfo host info
type ESBHostInfo struct {
	BkIspName            string          `json:"bk_isp_name"`
	Domain               string          `json:"domain"`
	OuterNetworkSegment  string          `json:"outer_network_segment"`
	NetStructID          int64           `json:"net_struct_id"`
	BkOsName             string          `json:"bk_os_name"`
	SrvOutBandManageType string          `json:"srv_out_band_manage_type"`
	BkSvcIDArr           string          `json:"bk_svc_id_arr"`
	BkHostID             int64           `json:"bk_host_id"`
	BkOsVersion          string          `json:"bk_os_version"`
	BkPositionName       string          `json:"bk_position_name"`
	BkHostName           string          `json:"bk_host_name"`
	BkStrVersion         string          `json:"bk_str_version"`
	IdcUnitID            int64           `json:"idc_unit_id"`
	BkLogicZoneID        string          `json:"bk_logic_zone_id"`
	BkProduct            string          `json:"bk_product"`
	BkIdcAreaID          int64           `json:"bk_idc_area_id"`
	SubZoneID            string          `json:"sub_zone_id"`
	BkIPOperName         string          `json:"bk_ip_oper_name"`
	BkInnerNetIdc        string          `json:"bk_inner_net_idc"`
	BkComment            string          `json:"bk_comment"`
	BkBsInfo             []ESBHostBsInfo `json:"bk_bs_info"`
	Dbrole               string          `json:"dbrole"`
	SrvStatus            string          `json:"srv_status"`
	BkBakOperator        string          `json:"bk_bak_operator"`
	BkLogicZone          string          `json:"bk_logic_zone"`
	IdcID                int64           `json:"idc_id"`
	BkIsVirtual          string          `json:"bk_is_virtual"`
	Operator             string          `json:"operator"`
	BkSvrDeviceClsName   string          `json:"bk_svr_device_cls_name"`
	IdcCityID            int64           `json:"idc_city_id"`
	NetStructName        string          `json:"net_struct_name"`
	SvrInputTime         time.Time       `json:"svr_input_time"`
	ClassifyLevelName    string          `json:"classify_level_name"`
	BkHostInnerip        string          `json:"bk_host_innerip"`
	RackID               string          `json:"rack_id"`
	IFix                 string          `json:"接入iFix"`
	RaidID               string          `json:"raid_id"`
	HardMemo             string          `json:"hard_memo"`
	OuterSwitchPort      string          `json:"outer_switch_port"`
	SvrFirstTime         time.Time       `json:"svr_first_time"`
	LogicDomain          string          `json:"logic_domain"`
	BkAssetID            string          `json:"bk_asset_id"`
	IdcName              string          `json:"idc_name"`
	SvrTypeName          string          `json:"svr_type_name"`
	BkCloudID            int64           `json:"bk_cloud_id"`
	SvrID                int64           `json:"svr_id"`
	BkSLA                string          `json:"bk_sla"`
	BkOuterEquipID       string          `json:"bk_outer_equip_id"`
	RaidName             string          `json:"raid_name"`
	BkSvrTypeID          int64           `json:"bk_svr_type_id"`
	InnerSwitchPort      string          `json:"inner_switch_port"`
	BkHostOuterip        string          `json:"bk_host_outerip"`
	SvrDeviceTypeID      int64           `json:"svr_device_type_id"`
	GroupName            string          `json:"group_name"`
	BkSn                 string          `json:"bk_sn"`
	SvrDeviceID          string          `json:"svr_device_id"`
	LogicDomainID        string          `json:"logic_domain_id"`
	InnerNetworkSegment  string          `json:"inner_network_segment"`
	SvrDeviceClass       string          `json:"svr_device_class"`
	IdcCityName          string          `json:"idc_city_name"`
	IdcUnitName          string          `json:"idc_unit_name"`
	BkInnerSwitchIP      string          `json:"bk_inner_switch_ip"`
	SubZone              string          `json:"sub_zone"`
	Rack                 string          `json:"rack"`
	NetDeviceID          string          `json:"net_device_id"`
	SrvImportantLevel    string          `json:"srv_important_level"`
	DeptName             string          `json:"dept_name"`
	BkInnerEquipID       string          `json:"bk_inner_equip_id"`
	BkServiceArr         string          `json:"bk_service_arr"`
	BkZoneName           string          `json:"bk_zone_name"`
	BkOuterSwitchIP      string          `json:"bk_outer_switch_ip"`
	BkSupplierAccount    string          `json:"bk_supplier_account"`
	SvrDeviceTypeName    string          `json:"svr_device_type_name"`
	ModuleName           string          `json:"module_name"`
	BkManageType         string          `json:"bk_manage_type"`
	IsSpecial            string          `json:"is_special"`
	BkIdcArea            string          `json:"bk_idc_area"`
}

// ESBHostListInfo host list info
type ESBHostListInfo struct {
	Count int64         `json:"count"`
	Info  []ESBHostInfo `json:"info"`
}

// ESBListHostsWitoutBizResult result for list biz hosts
type ESBListHostsWitoutBizResult struct {
	ESBBaseResp `json:",inline"`
	Data        *ESBHostListInfo `json:"data"`
}

// ESBBizLocation struct for biz location
type ESBBizLocation struct {
	BkBizID    int64  `json:"bk_biz_id"`
	BkLocation string `json:"bk_location"`
}

// ESBGetBizLocationResult result for get biz location
type ESBGetBizLocationResult struct {
	ESBBaseResp `json:",inline"`
	Data        []ESBBizLocation `json:"data"`
}
