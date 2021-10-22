# -*- coding: utf-8 -*-
#
# Tencent is pleased to support the open source community by making 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community Edition) available.
# Copyright (C) 2017-2019 THL A29 Limited, a Tencent company. All rights reserved.
# Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://opensource.org/licenses/MIT
#
# Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
# an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
# specific language governing permissions and limitations under the License.
#
from typing import Dict


class FakeNamespace:
    def __init__(self, *args, **kwargs):
        pass

    def register(self, *args, **kwargs) -> Dict:
        return {"code": 0}


class FakeCluster:
    def __init__(self, *args, **kwargs):
        pass

    def had_perm(self, *args, **kwargs) -> bool:
        return True

    def can_create(self, *args, **kwargs) -> bool:
        return True

    def can_edit(self, *args, **kwargs) -> bool:
        return True

    def can_delete(self, *args, **kwargs) -> bool:
        return True

    def can_view(self, *args, **kwargs) -> bool:
        return True

    def can_use(self, *args, **kwargs) -> bool:
        return True
