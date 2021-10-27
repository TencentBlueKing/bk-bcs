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
import copy
import json
import logging

import dpath

logger = logging.getLogger(__name__)


def merge_to_path(obj, glob, value, separator="/", afilter=None):
    """
    Given a path glob, set all existing elements in the document
    to the given value. Returns the number of elements changed.
    """
    changed = 0
    globlist = dpath.util.__safe_path__(glob, separator)
    for path in dpath.util._inner_search(obj, globlist, separator):
        changed += 1
        old_value = dpath.path.get(obj, path)
        old_value = copy.deepcopy(old_value)
        logger.debug("merge_to_path origin data: %s, merge data: %s", json.dumps(old_value), json.dumps(value))
        dpath.util.merge(dst=old_value, src=value, separator=separator, afilter=afilter)
        logger.debug("merge_to_path final data: %s", json.dumps(old_value))

        dpath.path.set(obj, path, old_value, create_missing=False, afilter=afilter)
    return changed
