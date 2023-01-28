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
"""
from typing import Dict, List

from backend.resources.utils.common import calculate_age
from backend.utils.basic import getitems

creation_timestamp_path = ".metadata.creationTimestamp"


def parse_column_data(co_item: Dict, columns: List[Dict], **kwargs: str) -> Dict:
    column_data = {}
    for col in columns:
        col_name = col["col_name"]
        if "json_path" not in col:
            column_data[col_name] = kwargs.get(col_name, "")
            continue

        json_path = col["json_path"]
        value = getitems(co_item, json_path)
        if json_path == creation_timestamp_path:
            column_data[col_name] = calculate_age(value)
        else:
            column_data[col_name] = value
    return column_data


def parse_columns(crd_dict: Dict) -> List[Dict]:
    """
    解析出crd中col名以及其值在spec中的位置(json_path)

    note: 解析时, additionalPrinterColumns 字段在不同的 api_version 中, 位置不同
    """
    columns = [
        {"col_name": "name", "json_path": ".metadata.name"},
        {"col_name": "cluster_id"},
        {"col_name": "namespace", "json_path": ".metadata.namespace"},
    ]

    # 默认按 apiextensions.k8s.io/v1beta1 版本读取
    additional_printer_columns = getitems(crd_dict, "spec.additionalPrinterColumns", [])

    # 未取到 additionalPrinterColumns 有效值时, 再按照 apiextensions.k8s.io/v1 版本读取
    if not additional_printer_columns:
        versions = getitems(crd_dict, "spec.versions", [{'additionalPrinterColumns': []}])
        additional_printer_columns = versions[0].get('additionalPrinterColumns', [])

    for add_col in additional_printer_columns:
        json_path = add_col.get('JSONPath') or add_col.get('jsonPath')

        # 忽略创建时间戳, 后面再添加
        if json_path == creation_timestamp_path:
            continue

        columns.append({"col_name": add_col["name"], "json_path": json_path})

    columns.append({"col_name": "AGE", "json_path": creation_timestamp_path})
    return columns


def to_table_format(crd_dict: Dict, cobj_list: List, **kwargs: str) -> Dict:
    """
    :return: 返回给前端约定的表格结构，th_list是表头内容，td_list是对应的表格内容
    """
    columns = parse_columns(crd_dict)
    column_data_list = [parse_column_data(co_item, columns, **kwargs) for co_item in cobj_list]
    if column_data_list:
        return {"th_list": [col["col_name"] for col in columns], "td_list": column_data_list}
    return {"th_list": [], "td_list": []}
