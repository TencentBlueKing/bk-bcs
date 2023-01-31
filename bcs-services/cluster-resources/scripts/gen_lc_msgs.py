# -*- coding: utf-8 -*-
"""
Tencent is pleased to support the open source community by making Blueking Container Service available.
Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at

    http://opensource.org/licenses/MIT

Unless required by applicable law or agreed to in writing, software distributed under,
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
"""

# 遍历扫描代码中的国际化数据，提取 msgID，合并已有的 lc_msgs.yaml，排序后输出
import os
import re
from enum import Enum
from pathlib import Path
from typing import Dict

import yaml
from attr import dataclass

# Golang 源码中国际化数据正则匹配
GO_SRC_I18N_REGEX = re.compile(r".*?i18n.GetMsg.*?\"(.*)\"[,)]")
# 模板文件中国际化数据正则匹配
TMPL_I18N_REGEX = re.compile(r".*?i18n.*?\"(.*)\" \.lang")


class ExtLangEnum(str, Enum):
    EN = "en"  # 英文
    # RU = "ru"  # 俄语
    # JA = "ja"  # 日语


@dataclass
class LcMsg:
    zh: str
    en: str = ""
    # ru: str = ""
    # ja: str = ""


class LcMsgsGenerator:

    ls_msgs: Dict[str, LcMsg]

    def __init__(self, base_dir: Path):
        self.base_dir = base_dir
        self.ls_msgs = {}
        self.lc_msgs_filepath = base_dir / "i18n" / "locale" / "lc_msgs.yaml"

    def execute(self):
        self.scan_and_collect()
        self.load_and_merge()
        self.write_to_file()

    def scan_and_collect(self):
        """扫描并采集数据"""
        for root, _, files in os.walk(self.base_dir):
            for file in files:
                if not (file.endswith(".go") or file.endswith(".yaml") or file.endswith(".tpl")):
                    continue
                self.regex_match(root + "/" + file)

    def regex_match(self, filepath: str):
        """通过正则匹配的方式，遍历指定文件，采集国际化数据"""
        regex = GO_SRC_I18N_REGEX if filepath.endswith(".go") else TMPL_I18N_REGEX
        with open(filepath) as fr:
            file_contents = fr.readlines()

        for line in file_contents:
            if not line:
                continue
            match = regex.match(line)
            if match:
                msg_id = match.group(1)
                # msg_id 即中文数据
                self.ls_msgs[msg_id] = LcMsg(zh=msg_id)

    def load_and_merge(self):
        """加载并与现有数据合并"""
        with open(self.lc_msgs_filepath) as fr:
            old_lc_msgs = yaml.load(fr.read(), yaml.SafeLoader)
        for msg in old_lc_msgs:
            msg_id = msg["msgID"]
            # 可能存在不再使用的国际化数据，这里忽略掉
            if msg_id not in self.ls_msgs:
                continue

            ext_lang_msgs = {lang: msg.get(lang, "") for lang in ExtLangEnum}
            self.ls_msgs[msg_id] = LcMsg(zh=msg_id, **ext_lang_msgs)

    def write_to_file(self):
        """覆盖输出到文件"""
        with open(self.lc_msgs_filepath, "w") as fw:
            for msg_id, lc_msg in self.ls_msgs.items():
                fw.write("- msgID: " + self.quote_when_necessary(msg_id) + "\n")
                for lang in ExtLangEnum:
                    c = getattr(lc_msg, lang) or "<TODO>"
                    fw.write(f"  {lang}: " + self.quote_when_necessary(c) + "\n")

    @staticmethod
    def quote_when_necessary(s: str) -> str:
        if " " in s or ":" in s:
            return '"{}"'.format(s)
        return s


if __name__ == "__main__":
    base_dir = Path(__file__).resolve().parents[1] / "pkg"
    LcMsgsGenerator(base_dir).execute()
