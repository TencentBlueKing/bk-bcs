# -*- coding: utf-8 -*-
"""
Tencent is pleased to support the open source community by making 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community
Edition) available.
Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://opensource.org/licenses/MIT

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
specific language governing permissions and limitations under the License.

数据平台 标准日志、非标准日志、Metric接入相关方法
"""
import logging
import time
from abc import ABCMeta
from enum import Enum

from django.conf import settings

from backend.components.utils import http_post

from .constants import DATA_API_V3_PREFIX, DATA_TOKEN

logger = logging.getLogger(__name__)


def get_data_id_by_name(raw_data_name):
    return False


class DataType(Enum):
    # 标准日志
    SLOG = "slog"
    # 非标准日志
    CLOG = "clog"
    # Metric
    METRIC = "metric"


# 标准日志的清洗规则
SLOG_CLEAN_FIELDS = [
    {"field_name": "log", "field_alias": "log", "field_type": "string", "is_dimension": False, "field_index": 1},
    {"field_name": "stream", "field_alias": "stream", "field_type": "string", "is_dimension": False, "field_index": 2},
    {
        "field_name": "logfile",
        "field_alias": "logfile",
        "field_type": "string",
        "is_dimension": False,
        "field_index": 3,
    },
    {
        "field_name": "gseindex",
        "field_alias": "gseindex",
        "field_type": "string",
        "is_dimension": False,
        "field_index": 4,
    },
    {
        "field_name": "bcs_appid",
        "field_alias": "bcs_appid",
        "field_type": "string",
        "is_dimension": False,
        "field_index": 5,
    },
    {
        "field_name": "bcs_cluster",
        "field_alias": "bcs_cluster",
        "field_type": "string",
        "is_dimension": False,
        "field_index": 6,
    },
    {
        "field_name": "container_id",
        "field_alias": "container_id",
        "field_type": "string",
        "is_dimension": False,
        "field_index": 7,
    },
    {
        "field_name": "bcs_namespace",
        "field_alias": "bcs_namespace",
        "field_type": "string",
        "is_dimension": False,
        "field_index": 8,
    },
    {
        "field_name": "timestamp_orig",
        "field_alias": "timestamp_orig",
        "field_type": "string",
        "is_dimension": False,
        "field_index": 9,
    },
    {
        "field_name": "bcs_custom_labels",
        "field_alias": "bcs_custom_labels",
        "field_type": "string",
        "is_dimension": False,
        "field_index": 10,
    },
    {
        "field_name": "event_time",
        "field_alias": "event_time",
        "field_type": "string",
        "is_dimension": False,
        "field_index": 11,
    },
]

SLOG_JSON_CONFIG = """{"extract": {"args": [], "type": "fun", "method": "from_json", "next": {"type": "branch",
"name": "", "next": [{"subtype": "assign_obj", "type": "assign", "assign": [{"type": "string", "assign_to":
"logfile","key": "logfile"}, {"type": "string", "assign_to": "gseindex", "key": "gseindex"}], "next": null},
{"subtype":"access_obj", "type": "access", "key": "container", "next": {"type": "branch", "name": "", "next":
[{"subtype":"assign_obj", "type": "assign", "assign": [{"type": "string", "assign_to": "container_id", "key": "id"}],
"next":null}, {"subtype": "access_obj", "type": "access", "key": "labels", "next": {"type": "branch", "name": "",
"next": [{"subtype": "assign_obj", "type": "assign", "assign": [{"type": "string", "assign_to": "bcs_appid", "key":
"io.tencent.bcs.app.appid"}, {"type": "string", "assign_to": "bcs_cluster", "key": "io.tencent.bcs.cluster"}, {"type":
"string", "assign_to": "bcs_namespace", "key": "io.tencent.bcs.namespace"}], "next": null}, {"subtype": "access_obj",
"next": {"next": {"subtype": "assign_json", "next": null, "type": "assign", "assign": [{"type": "string", "assign_to":
"bcs_custom_labels", "key": "__all_keys__"}]}, "args": [], "type": "fun", "method": "from_json"}, "type": "access",
"key": "io.tencent.bcs.custom.labels"}]}}]}}, {"subtype": "access_obj", "type": "access", "key": "log", "next":
{"args":[], "type": "fun", "method": "iterate", "next": {"type": "branch", "name": "", "next": [{"subtype":
"assign_obj","type": "assign", "assign": [{"type": "string", "assign_to": "stream", "key": "stream"}, {"type":
"string", "assign_to":"log", "key": "log"}, {"type": "string", "assign_to": "timestamp_orig", "key":
"timestamp_orig"}], "next": null},{"subtype": "access_obj", "type": "access", "key": "time", "next": {"args": ["+"],
"type": "fun", "method": "split","next": {"index": "0", "subtype": "access_pos", "type": "access", "next": {"args":
["."], "type": "fun", "method":"split", "next": {"index": "0", "subtype": "access_pos", "type": "access", "next":
{"args": ["T", "", ":", "", "-", ""],"type": "fun", "method": "replace", "next": {"subtype": "assign_pos", "type":
"assign", "assign": [{"index": "0","assign_to": "event_time", "type": "string"}], "next": null}}}}}}}]}}}]}}, "conf":
{"timestamp_len": 0, "encoding":"UTF8", "time_format": "yyyyMMddHHmmss", "timezone": 0, "output_field_name":
"timestamp", "time_field_name":"event_time"}}"""

SLOG_STORAGE_FIELDS = [
    {
        "field_name": "log",
        "field_alias": "log",
        "field_type": "string",
        "physical_field": "log",
        "is_dimension": False,
        "field_index": 1,
        "is_index": False,
        "is_key": False,
        "is_value": False,
        "is_analyzed": True,
        "is_doc_values": False,
        "is_json": False,
    },
    {
        "field_name": "bcs_custom_labels",
        "field_alias": "bcs_custom_labels",
        "field_type": "string",
        "physical_field": "bcs_custom_labels",
        "is_dimension": False,
        "field_index": 2,
        "is_index": False,
        "is_key": False,
        "is_value": False,
        "is_analyzed": False,
        "is_doc_values": False,
        "is_json": True,
    },
]
# 非标准日志清洗规则
CLOG_CLEAN_FIELDS = SLOG_CLEAN_FIELDS
CLOG_JSON_CONFIG = """{"extract":{"args":[],"type":"fun","method":"from_json","next":{"type":"branch","name":"","next":
[{"subtype":"assign_obj","type":"assign","assign":[{"assign_to":"logfile","type":"string","key":"logfile"},{"assign_to":
"gseindex","type":"string","key":"gseindex"}]},{"subtype":"access_obj","type":"access","key":"container","next":{"type":
"branch","name":"","next":[{"subtype":"assign_obj","type":"assign","assign":[{"assign_to":"container_id","type":"string"
,"key":"id"}]},{"subtype":"access_obj","type":"access","key":"labels","next":{"type":"branch","name":"","next":
[{"subtype":"assign_obj","type":"assign","assign":[{"assign_to":"bcs_appid","type":"string","key":
"io.tencent.bcs.app.appid"},{"assign_to":"bcs_cluster","type":"string","key":"io.tencent.bcs.cluster"},{"assign_to":
"bcs_namespace","type":"string","key":"io.tencent.bcs.namespace"}]},{"subtype":"access_obj","type":"access","key":
"io.tencent.bcs.custom.labels","next":{"args":[],"type":"fun","method":"from_json","next":{"subtype":"assign_json",
"type":"assign","assign":[{"assign_to":"bcs_custom_labels","type":"string","key":"__all_keys__"}]}}}]}}]}},{"subtype":
"access_obj","type":"access","key":"log","next":{"args":[],"type":"fun","method":"iterate","next":{"subtype":
"assign_pos","type":"assign","assign":[{"assign_to":"log","type":"string","index":"0"}]}}},{"subtype":"access_obj",
"type":"access","key":"timestamp","next":{"args":["+"],"type":"fun","method":"split","next":{"subtype":"access_pos",
"next":{"args":["."],"type":"fun","method":"split","next":{"subtype":"access_pos","next":{"args":["T","",":","","-",""],
"type":"fun","method":"replace","next":{"subtype":"assign_pos","type":"assign","assign":[{"assign_to":"event_time",
"type":"string","index":"0"}]}},"type":"access","index":"0"}},"type":"access","index":"0"}}}]}},"conf":{"timestamp_len":
0,"encoding":"UTF8","time_format":"yyyyMMddHHmmss","timezone":0,"output_field_name":"timestamp","time_field_name":
"event_time"}}"""

CLOG_STORAGE_FIELDS = SLOG_STORAGE_FIELDS

# Metric 清洗规则
METRIC_CLEAN_FIELDS = [
    {"field_name": "labels", "field_alias": "labels", "field_type": "string", "is_dimension": False, "field_index": 1},
    {
        "field_name": "metric_name",
        "field_alias": "metric_name",
        "field_type": "string",
        "is_dimension": False,
        "field_index": 2,
    },
    {
        "field_name": "metric_value",
        "field_alias": "metric_value",
        "field_type": "double",
        "is_dimension": False,
        "field_index": 3,
    },
]
METRIC_JSON_CONFIG = '{"extract": {"next": {"next": [{"subtype": "access_obj", "next": {"next": {"index": "0", "next": {"next": {"subtype": "assign_pos", "next": null, "type": "assign", "assign": [{"index": "0", "assign_to": "time", "type": "string"}], "label": null}, "args": [":", "", "T", "", "-", ""], "type": "fun", "method": "replace", "label": null}, "type": "access", "subtype": "access_pos", "label": null}, "args": ["."], "type": "fun", "method": "split", "label": null}, "type": "access", "key": "@timestamp", "label": null}, {"subtype": "access_obj", "next": {"subtype": "access_obj", "next": {"subtype": "access_obj", "next": {"next": {"next": [{"subtype": "assign_obj", "next": null, "type": "assign", "assign": [{"type": "string", "assign_to": "metric_name", "key": "key"}, {"type": "double", "assign_to": "metric_value", "key": "value"}], "label": null}, {"subtype": "assign_json", "next": null, "type": "assign", "assign": [{"type": "string", "assign_to": "labels", "key": "labels"}], "label": null}], "type": "branch", "name": "", "label": null}, "args": [], "type": "fun", "method": "iterate", "label": null}, "type": "access", "key": "metrics", "label": null}, "type": "access", "key": "collector", "label": null}, "type": "access", "key": "prometheus", "label": null}], "type": "branch", "name": "", "label": null}, "args": [], "type": "fun", "method": "from_json", "label": null}, "conf": {"timestamp_len": 0, "encoding": "UTF8", "time_format": "yyyyMMddHHmmss", "timezone": 0, "output_field_name": "timestamp", "time_field_name": "time"}}'  # noqa
METRIC_STORAGE_CONFIG = '{"analyzed_fields": [], "doc_values_fields": [], "json_fields": ["labels"]}'  # noqa

PLAN_DESCRIPTION = {DataType.SLOG.value: "标准日志采集", DataType.CLOG.value: "非标准日志采集", DataType.METRIC.value: "BCS metric"}


def compose_auth_params(username):
    return {
        "bk_app_code": settings.APP_ID,
        "bk_app_secret": settings.APP_TOKEN,
        "bk_username": username,
        "bkdata_authentication_method": "user",
    }


def generate_table_name(data_type):
    return f"{data_type[:4]}_{int(time.time())}"


def deploy_plan(username, cc_biz_id, data_name, data_type):
    """
    提交接入部署计划,获取dataid
    data_type: slog/clog/metric
    """
    description = PLAN_DESCRIPTION.get(data_type, "")
    params = {
        "data_scenario": "custom",
        "bk_biz_id": cc_biz_id,
        "description": description,
        "data_token": DATA_TOKEN,
        "access_raw_data": {
            "raw_data_alias": data_name,
            "raw_data_name": data_name,
            "maintainer": username,
            "description": description,
            "data_category": "sys_performance",
            "data_source": "business_server",
            "data_encoding": "UTF-8",
            "tags": ["SV"],
            "sensitivity": "private",
        },
    }
    params.update(compose_auth_params(username))
    try:
        resp = http_post(f"{DATA_API_V3_PREFIX}/access/deploy_plan/", json=params)
        if resp.get("result"):
            return True, resp.get("data", {}).get("raw_data_id")

        logger.error(resp.get("message"))
        return True, 0
    except Exception as e:
        logger.exception(e)
        return True, 0


def create_data_bus(data_type):
    if data_type == DataType.SLOG.value:
        return STDLogDataBus
    else:
        return NSTDLogDataBus


class DataBus(metaclass=ABCMeta):
    def __init__(self, project_data):
        self.project_data = project_data

    def setup_clean(self, username, table_name):
        """
        创建清洗配置
        """
        params = {
            "raw_data_id": self.raw_data_id,
            "json_config": self.json_config,
            "pe_config": "",
            "bk_biz_id": self.project_data.cc_biz_id,
            "result_table_name": table_name,
            "result_table_name_alias": table_name,
            "fields": self.clean_fields,
            "clean_config_name": f"clean_{self.data_type}_config",
            "description": f"clean_{self.data_type}",
        }
        params.update(compose_auth_params(username))
        resp = http_post(f"{DATA_API_V3_PREFIX}/databus/cleans/", json=params)
        if resp.get("result"):
            return True, resp.get("data", {}).get("id")
        return False, resp.get("message")

    def storage_data(self, username, table_name):
        """
        入库
        """
        params = {
            "raw_data_id": self.raw_data_id,
            "data_type": "clean",
            "result_table_name": table_name,
            "result_table_name_alias": table_name,
            "storage_type": "es",
            "storage_cluster": "bkee-es",
            "expires": "7d",
            "fields": self.storage_fields,
        }
        params.update(compose_auth_params(username))
        resp = http_post(f"{DATA_API_V3_PREFIX}/databus/data_storages/", json=params)
        if resp.get("result"):
            return True, resp.get("data")
        return False, resp.get("message")

    def clean_and_storage_data(self, username):

        if self.storage_success:
            return True, "success"

        if not self.table_name:
            table_name = generate_table_name(self.data_type)
            result, message = self.setup_clean(username, table_name)
            if not result:
                return False, message
            self._update_table_name(table_name)

        result, message = self.storage_data(username, self.table_name)
        if not result:
            self._update_storage_success(False)
            return False, message
        self._update_storage_success(True)

        return True, "success"

    def _update_table_name(self, table_name):
        self.table_name = table_name

    def _update_storage_success(self, is_success):
        self.storage_success = is_success


class STDLogDataBus(DataBus):
    def __init__(self, project_data):
        self.raw_data_id = project_data.standard_data_id
        self.table_name = project_data.standard_table_name
        self.storage_success = project_data.standard_storage_success

        self.json_config = SLOG_JSON_CONFIG
        self.clean_fields = SLOG_CLEAN_FIELDS
        self.storage_fields = SLOG_STORAGE_FIELDS
        self.data_type = DataType.SLOG.value
        super().__init__(project_data)

    def _update_table_name(self, table_name):
        super()._update_table_name(table_name)
        self.project_data.standard_table_name = table_name
        self.project_data.save(update_fields=["standard_table_name"])

    def _update_storage_success(self, is_success):
        super()._update_storage_success(is_success)
        self.project_data.standard_storage_success = is_success
        self.project_data.save(update_fields=["standard_storage_success"])


class NSTDLogDataBus(DataBus):
    def __init__(self, project_data):
        self.raw_data_id = project_data.non_standard_data_id
        self.table_name = project_data.non_standard_table_name
        self.storage_success = project_data.non_standard_storage_success

        self.json_config = CLOG_JSON_CONFIG
        self.clean_fields = CLOG_CLEAN_FIELDS
        self.storage_fields = CLOG_STORAGE_FIELDS
        self.data_type = DataType.CLOG.value
        super().__init__(project_data)

    def _update_table_name(self, table_name):
        super()._update_table_name(table_name)
        self.project_data.non_standard_table_name = table_name
        self.project_data.save(update_fields=["non_standard_table_name"])

    def _update_storage_success(self, is_success):
        super()._update_storage_success(is_success)
        self.project_data.non_standard_storage_success = is_success
        self.project_data.save(update_fields=["non_standard_storage_success"])


try:
    from .databus_ext import *  # noqa
except ImportError as e:
    logger.debug("Load extension failed: %s", e)
