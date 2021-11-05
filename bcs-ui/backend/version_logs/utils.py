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
import os
import time
from typing import List, Dict

from django.conf import settings


class VersionLogs:
    """查询版本列表

    版本日志文件名格式为 `版本_时间.md`，例如: v1.0.0_2021-11-01.md

    :param path: 版本日志文件路径
    """

    def __init__(self, path: str = settings.VERSION_LOG_PATH, language: str = settings.LANGUAGE_CODE):
        self.path = path
        self.language = language

    def get_version_list(self) -> List[Dict]:
        # 判断为目录
        if not self._is_dir(self.path):
            return []
        # 根据语言获取对应的目录
        version_log_path = self._get_path_by_language()
        if not self._is_dir(version_log_path):
            return []
        # 解析文件
        version_log_list = []
        for filename in os.listdir(version_log_path):
            # 必须以md文件
            if not filename.endswith(".md"):
                continue
            # 获取文件内容
            file_path = os.path.join(version_log_path, filename)
            with open(file_path) as f:
                content = f.read()
            full_name = os.path.splitext(filename)[0]
            # 通过文件名，获取版本及对应的日期
            version, _, date = full_name.partition("_")
            date = self._get_date(file_path, date)
            version_log_list.append({"version": version, "date": date, "content": content})

        # 以时间逆序
        version_log_list.sort(key=lambda x: x["version"], reverse=True)
        return version_log_list

    def _is_dir(self, path: str) -> bool:
        return os.path.isdir(path)

    def _get_path_by_language(self) -> str:
        # 仅支持中文和英文
        name = "zh_CN" if self.language == settings.LANGUAGE_CODE else "en"
        return os.path.join(self.path, name)

    def _get_date(self, file_path: str, date: str = "") -> str:
        """获取日期，如果日期为空，获取文件的最后修改日期"""
        if date:
            return date
        timestamp = os.stat(file_path).st_mtime
        return time.strftime('%Y-%m-%d', time.localtime(timestamp))
