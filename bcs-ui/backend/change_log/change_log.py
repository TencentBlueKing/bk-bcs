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
import time
from pathlib import Path, PosixPath
from typing import List

from django.conf import settings


class ChangeLog:
    """查询变更版本

    版本日志文件名格式为 `版本_时间.md`，例如: v1.0.0_2021-11-01.md

    :param path: 版本日志文件路径
    """

    def __init__(self, path: str = settings.CHANGE_LOG_PATH, language: str = settings.LANGUAGE_CODE):
        self.path = path
        self.language = language

    def list(self) -> List:
        p = Path(self.path)
        # 如果不是目录，则返回空
        if not p.is_dir():
            return []
        # 根据语言获取对应的目录
        log_path = self._get_path_by_language(p)
        # 解析文件，并按版本逆序
        log_list = []
        for file_path in log_path.iterdir():
            # 必须是以md结尾的文件
            if not (file_path.is_file() and file_path.suffix == ".md"):
                continue
            # 通过文件名，获取版本及对应的日期
            version, suffix = file_path.stem.split("_")
            date = self._get_date(file_path, suffix)
            # 获取文件内容
            content = file_path.read_text()
            log_list.append({"version": version, "date": date, "content": content})

        # 以时间逆序
        log_list.sort(key=lambda x: x["version"], reverse=True)
        return log_list

    def _get_path_by_language(self, p: PosixPath) -> PosixPath:
        # 仅支持中文和英文
        name = "zh_CN" if self.language == settings.LANGUAGE_CODE else "en"
        return p.joinpath(name)

    def _get_date(self, file_path: PosixPath, date: str = "") -> str:
        """获取日期，如果日期为空，获取文件的最后修改日期"""
        if date:
            return date
        timestamp = file_path.stat().st_mtime
        return time.strftime('%Y-%m-%d', time.localtime(timestamp))
