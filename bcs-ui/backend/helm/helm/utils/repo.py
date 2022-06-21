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

this is the functions for fetch content from repo
"""

import base64
import hashlib
import io
import logging
import tarfile

import requests
import yaml

from backend.components.utils import http_get
from backend.helm.helm.utils.repo_bk import get_charts_info, make_requests_auth
from backend.utils.cache import rd_client

logger = logging.getLogger(__name__)


def prepareRepoCharts(url, name, auths):
    """
    NOTE: currently not support git
    """
    charts_info, charts_info_hash = _prepareHelmRepoPath(url, name, auths)
    return charts_info, charts_info_hash


def _prepareHelmRepoPath(url, name, auths):
    ok, charts_info, charts_info_hash = get_charts_info(url, auths)
    if not ok:
        logger.error("get charts info from url fail! %s", url)
        return None, None

    return charts_info, charts_info_hash


def download_icon_data(url, auths):
    """
    download icon
    """
    try:
        if not auths:
            resp = requests.get(url)
        else:
            for auth in auths:
                resp = requests.get(url, auth=make_requests_auth(auth))
                if resp.status_code != 401:
                    break
    except Exception as e:
        logger.warning("Download icon fail: [url=%s, error=%s]", url, e)
        return False, None

    if resp.status_code != 200:
        logger.error("Download icon fail: [url=%s, status=%s]", url, resp.status_code)
        return False, None

    content_type = resp.headers.get("Content-Type")
    b64_content = base64.b64encode(resp.content).decode()

    data = "data:{content_type};base64,{content}".format(content_type=content_type, content=b64_content)
    return True, data


SUPPORT_FILES = ["questions.yml", "questions.yaml"]


def is_binary_string(bytes):
    textchars = bytearray({7, 8, 9, 10, 12, 13, 27} | set(range(0x20, 0x100)) - {0x7F})
    return bool(bytes.translate(None, textchars))


def download_template_data(chart_name, url, auths):
    # https://kubernetes-charts-incubator.storage.googleapis.com/kafka-0.4.6.tgz
    if not url:
        return False, None, None

    if not auths:
        resp = requests.get(url, stream=True, verify=False)
    else:
        for auth in auths:
            resp = requests.get(url, stream=True, auth=make_requests_auth(auth), verify=False)
            if resp.status_code != 401:
                break

    if resp.status_code != 200:
        # just retry once
        resp = requests.get(url, stream=True)
        if resp.status_code != 200:
            logger.error("Download template data fail: [url=%s]", url)
            return False, None, None

    tar = tarfile.open(mode="r:*", fileobj=io.BytesIO(resp.content))

    support_file_list = {"{chart}/{file}".format(chart=chart_name, file=f): 1 for f in SUPPORT_FILES}

    files = {}
    questions = {}

    # for file_path in tar.getnames():
    tar.getnames()
    for member in tar.members:
        if member.isdir():
            continue
        file_path = member.path
        file_content = tar.extractfile(file_path).read()

        if is_binary_string(file_content[:1024]):
            logger.warning(
                "file %s seems to be a binary file, skipped it. content: %s", file_path, file_content[:1024]
            )
            continue

        if file_path in support_file_list:
            questions = file_content.decode()
        try:
            files[file_path] = file_content.decode()
        except Exception as e:
            logger.exception(
                "download_template_data failed %s, file_path=%s, file_content: %s", e, file_path, file_content
            )
            return False, None, None

    if not questions:
        return True, files, questions

    questions = yaml.load(questions)
    return True, files, questions


class InProcessSign(object):
    def __init__(self, repo_id):
        self.repo_id = repo_id
        self.key = "bcs_k8s:helm:repo:{repo_id}".format(repo_id=repo_id)

    def exists(self):
        value = rd_client.get(self.key)
        if value:
            return True
        return False

    def create(self, expires=300):
        rd_client.setex(self.key, 1, expires)

    def update(self):
        self.create(expires=300)

    def delete(self):
        rd_client.delete(self.key)


if __name__ == '__main__':
    # git
    # url = "https://git.rancher.io/charts"
    # name = "test"
    # prepareRepoPath(url, name)

    # helm
    url = "https://kubernetes-charts-incubator.storage.googleapis.com/"
    name = "test"
