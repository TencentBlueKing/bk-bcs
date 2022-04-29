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
import json
from typing import Dict

import yaml
from django.conf import settings
from django.utils.crypto import get_random_string
from django.utils.translation import ugettext_lazy as _

from backend.dashboard.examples.constants import (
    DEMO_RESOURCE_MANIFEST_DIR,
    EXAMPLE_CONFIG_DIR,
    EXAMPLE_LANG_MAP,
    RANDOM_SUFFIX_LENGTH,
    RESOURCE_REFERENCES_DIR,
    SUFFIX_ALLOWED_CHARS,
)
from backend.dashboard.exceptions import LanguageForExampleUnsupported


def load_resource_template(lang: str, kind: str) -> Dict:
    """获取指定 资源类型模版 信息"""
    with open(f'{EXAMPLE_CONFIG_DIR}/{lang}/{kind}.json') as fr:
        return json.loads(fr.read())


def load_resource_references(lang: str, kind: str) -> str:
    """获取指定 资源类型参考资料"""
    with open(f'{RESOURCE_REFERENCES_DIR}/{lang}/{kind}.md') as fr:
        return fr.read()


def load_demo_manifest(file_path: str) -> Dict:
    """指定资源类型的 Demo 配置信息"""
    with open(f'{DEMO_RESOURCE_MANIFEST_DIR}/{file_path}.yaml') as fr:
        manifest = yaml.load(fr.read(), yaml.SafeLoader)

    # 避免名称重复，每次默认添加随机后缀
    random_suffix = get_random_string(length=RANDOM_SUFFIX_LENGTH, allowed_chars=SUFFIX_ALLOWED_CHARS)
    manifest['metadata']['name'] = f"{manifest['metadata']['name']}-{random_suffix}"
    return manifest


def get_example_lang_from_cookie(request) -> str:
    """从 Cookie 获取示例文件/配置语言版本"""
    bk_lang = request.COOKIES.get(settings.LANGUAGE_COOKIE_NAME, 'zh-CN').lower()
    if bk_lang not in EXAMPLE_LANG_MAP:
        raise LanguageForExampleUnsupported(_('语言版本 {} 不受支持').format(bk_lang))
    return EXAMPLE_LANG_MAP[bk_lang]
