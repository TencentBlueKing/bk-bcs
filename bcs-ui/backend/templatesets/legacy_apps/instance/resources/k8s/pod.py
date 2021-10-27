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
from backend.templatesets.legacy_apps.instance.resources import BCSResource, utils


class Pod(BCSResource):
    def _strategy_params_to_int(self, roll_update_strategy, resource_name, metadata_name, is_preview, is_validate):
        strategy_params = ['maxUnavailable', 'maxSurge']
        for p in strategy_params:
            roll_update_strategy[p] = utils.handle_number_var(
                roll_update_strategy[p], f'{resource_name}[{metadata_name}]{p}', is_preview, is_validate
            )

    def set_base_spec(self, spec, resource_name, metadata_name, is_preview, is_validate):
        # 处理minReadySeconds
        min_readys = 'minReadySeconds'
        if min_readys in spec:
            spec[min_readys] = utils.handle_number_var(
                spec[min_readys], f'{resource_name}[{metadata_name}]{min_readys}', is_preview, is_validate
            )

    def set_strategy(self, strategy, resource_name, metadata_name, is_preview, is_validate):
        if strategy.get('type') == 'Recreate':
            if 'rollingUpdate' in strategy:
                del strategy['rollingUpdate']
            return

        self._strategy_params_to_int(strategy['rollingUpdate'], resource_name, metadata_name, is_preview, is_validate)
