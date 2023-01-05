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

# 自动替换 lc_msgs 中的 <TO-DO> 项，依赖 google 翻译
from enum import Enum
from pathlib import Path
from typing import Dict

import requests
import yaml
from attr import dataclass


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


class LcMsgsTranslater:

    ls_msgs: Dict[str, LcMsg]

    def __init__(self, base_dir: Path):
        self.base_dir = base_dir
        self.ls_msgs = {}
        self.lc_msgs_filepath = base_dir / "i18n" / "locale" / "lc_msgs.yaml"
        self.translate_api = "http://translate.google.com/translate_a/single?client=gtx&dt=t&sl={sl}&tl={tl}&q={q}"

    def execute(self):
        self.load_from_file()
        self.translate()
        self.write_to_file()

    def load_from_file(self):
        """加载已有的数据"""
        with open(self.lc_msgs_filepath) as fr:
            exists_lc_msgs = yaml.load(fr.read(), yaml.SafeLoader)

        for msg in exists_lc_msgs:
            msg_id = msg["msgID"]
            ext_lang_msgs = {lang: msg.get(lang, "") for lang in ExtLangEnum}
            self.ls_msgs[msg_id] = LcMsg(zh=msg_id, **ext_lang_msgs)

    def translate(self):
        """调用 API 获取翻译"""
        for msg_id, lc_msg in self.ls_msgs.items():
            for lang in ExtLangEnum:
                if getattr(lc_msg, lang) != "<TODO>":
                    continue

                ret = requests.get(self.translate_api.format(sl="zh", tl=lang, q=msg_id)).json()[0][0][0]
                print(f"msg id: {msg_id}, translate result: {ret}")
                setattr(lc_msg, lang, ret)

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
    LcMsgsTranslater(base_dir).execute()
