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
from backend.utils.errcodes import ErrorCode
from backend.utils.error_codes import error_codes

from .. import constants, utils


class Driver:
    @classmethod
    def get_deployment_rs_name_list(cls, client, ns_name, inst_name, extra=None):
        """获取deployment关联的rs名称
        主要是用作查询关联的pod信息
        """
        extra = extra or {constants.REFERENCE_RESOURCE_LABEL: inst_name}
        extra = utils.base64_encode_params(extra)
        rs_resp = client.get_rs({'extra': extra, 'namespace': ns_name, 'field': 'resourceName'})
        if rs_resp.get('code') != ErrorCode.NoError:
            raise error_codes.APIError(rs_resp.get('message'))
        data = rs_resp.get('data') or []
        return [info['resourceName'] for info in data if info.get('resourceName')]

    @classmethod
    def get_unit_info_by_name(cls, client, ns_name, pod_name, field):
        params = {'namespace': ns_name, 'name': pod_name}
        # 组装field
        field = field or constants.STORAGE_FIELD_LIST
        pod_resp = client.get_pod(extra=None, field=','.join(field), params=params)
        if pod_resp.get('code') != ErrorCode.NoError:
            raise error_codes.APIError(pod_resp.get('message'))

        return pod_resp.get('data')

    @classmethod
    def get_events(cls, client, params):
        return client.get_events(params)
