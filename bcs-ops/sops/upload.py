#!/usr/bin/env python3
#######################################
# Tencent is pleased to support the open source community by making Blueking Container Service available.
# Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
# Licensed under the MIT License (the "License"); you may not use this file except
# in compliance with the License. You may obtain a copy of the License at
# http://opensource.org/licenses/MIT
# Unless required by applicable law or agreed to in writing, software distributed under
# the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
# either express or implied. See the License for the specific language governing permissions and
# limitations under the License.
#######################################

"""
upload.py is a script to upload bcs scripts package to bkrepo
and upload relate sops file to common flow.
Notice!!!!!!
BKREPO URL is diverse in in different blueking environments.
So `SCRIPT_URL_PLACEHOLDER` is used as a matching pattern in sops template flow.
"""

import argparse
import base64
import requests
import ujson
import hashlib
import logging
import os
import re

LOG_LEVEL = os.environ.get("LOG_LEVEL", "INFO")

REPO_FILE = os.environ.get("REPO_FILE", "bcs-ops.tar.gz")
SOPS_FILE = os.environ.get("SOPS_FILE", "bcs_bk_sops_common.dat")
SOPS_PAT = os.environ.get("SOPS_PAT", r"SCRIPT_URL_PLACEHOLDER")

BKAPI_HOST = os.environ["BKAPI_HOST"]
APP_CODE = os.environ["APP_CODE"]
APP_SECRET = os.environ["APP_SECRET"]

REPO_HOST = os.environ["REPO_HOST"]
REPO_PROJECT = os.environ["REPO_PROJECT"]
REPO_BUCKET = os.environ["REPO_BUCKET"]
REPO_PATH = os.environ["REPO_PATH"]
REPO_USER = os.environ["REPO_USER"]
REPO_PASSWD = os.environ["REPO_PASSWD"]


def bkrepo_upload(file: str, bkrepo_url: str, override: bool = True) -> bool:
    with open(file, "rb") as f:
        content = f.read()
        content_hash = hashlib.md5(content).hexdigest()

        headers = {
            "X-BKREPO-OVERWRITE": str(override).lower(),
            "X-BKREPO-MD5": content_hash,
        }
        response = requests.put(
            bkrepo_url, auth=(REPO_USER, REPO_PASSWD), headers=headers, data=content
        )
        if response.status_code == 200:
            logging.info(f"Upload {file} to {bkrepo_url} succeeded.")
            return True
        else:
            logging.error(
                f"http_code: {response.status_code}, Upload {file} to {bkrepo_url} failed: {response.text}"
            )
            return False


class SOPS_UPLOAD_API:
    def __init__(
        self, file: str, paas_host: str, app_code: str, app_secret: str
    ) -> None:
        self.url = f"http://{paas_host}/api/c/compapi/v2/sops/import_common_template/"
        self.salt = r"821a11587ea434eb85c2f5327a90ae54"
        self.app_code = app_code
        self.app_secret = app_secret
        self.file = file
        with open(self.file, "rb") as f:
            self.data = f.read()
            self._b64dec_unmarshal()

    def replace_data(self, sub_pat: str, sub_str: str) -> None:
        if sub_pat == "":
            logging.warn(f"missing sub_pat, skip replace")
            self._b64en_salt_marshal()
            return
        logging.debug(f"sub_pat: {sub_pat}, sub_str: {sub_str}")
        if isinstance(self.data, dict):
            self.data = ujson.dumps(self.data, sort_keys=True)
        if re.search(sub_pat, self.data):
            self.data = re.sub(sub_pat, sub_str, self.data)
        else:
            logging.warn(f"data not found: {sub_pat}")
        self.data = ujson.loads(self.data)
        self._b64en_salt_marshal()

    def _b64dec_unmarshal(self) -> None:
        self.data = ujson.loads(base64.b64decode(self.data).decode("utf-8"))

    def _b64en_salt_marshal(self) -> None:
        if isinstance(self.data, str):
            self.data = ujson.loads(self.data)
        template_data_string = (
            ujson.dumps(self.data["template_data"], sort_keys=True) + self.salt
        ).encode("utf-8")
        digest = hashlib.md5(template_data_string).hexdigest()
        self.data["digest"] = digest
        self.data = base64.b64encode(
            ujson.dumps(self.data, sort_keys=True).encode("utf-8")
        ).decode("utf-8")

    def upload(self, override: bool = True) -> bool:
        # 构建请求数据
        data = {
            "bk_app_code": self.app_code,
            "bk_app_secret": self.app_secret,
            "bk_username": "admin",
            "template_data": self.data,
            "override": override,
        }
        headers = {"Content-Type": "application/json", "cache-control": "no-cache"}
        # 发送 POST 请求
        response = requests.post(self.url, headers=headers, data=ujson.dumps(data))
        if response.status_code == 200:
            response_obj = ujson.loads(response.content)
            if response_obj["result"] == True:
                logging.info(f"Upload succeeded: {self.file}")
                return True
        logging.error(f"Upload failed: {response.text}")
        return False


def main():
    # set log-level
    logging.basicConfig(level=LOG_LEVEL)

    # bkrepo_url
    bkrepo_url = f"http://{REPO_HOST}/generic/{REPO_PROJECT}/{REPO_BUCKET}/{REPO_PATH}/{os.path.basename(REPO_FILE)}"
    logging.debug(f"bkrepo_url: {bkrepo_url}")
    # bcs sops file

    parser = argparse.ArgumentParser(description="Upload bcs scripts package to bkrepo")
    parser.add_argument(
        "upload",
        choices=["bkrepo", "sops"],
        type=str,
        default="bkrepo",
        help="upload to [sops] or [bkrepo]",
    )
    args = parser.parse_args()
    if args.upload:
        if args.upload == "bkrepo":
            bkrepo_upload(REPO_FILE, bkrepo_url, override=True)
        elif args.upload == "sops":
            s = SOPS_UPLOAD_API(SOPS_FILE, BKAPI_HOST, APP_CODE, APP_SECRET)
            s.replace_data(SOPS_PAT, bkrepo_url)
            s.upload()
        else:
            raise ValueError(
                "Invalid choice for upload, please choose either 'sops' or 'bkrepo'"
            )


if __name__ == "__main__":
    main()
