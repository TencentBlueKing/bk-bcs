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
import pytest
from requests_mock import ANY

from backend.components import bk_repo

PROJECT_EXIST_CODE = bk_repo.BkRepoClient.PROJECT_EXIST_CODE
REPO_EXIST_CODE = bk_repo.BkRepoClient.REPO_EXIST_CODE

fake_access_token = "access_token"
fake_username = "admin"
fake_pwd = "pwd"
fake_project_code = "project_code"
fake_project_name = "project_name"
fake_project_description = "this is a test"
fake_chart_name = "chart_name"
fake_chart_version = "0.0.1"
fake_error_code = 1


class TestBkRepoClient:
    @pytest.mark.parametrize(
        "raw_resp,expected_resp",
        [
            ({"code": 0, "data": {"foo": "bar"}}, {"code": 0, "data": {"foo": "bar"}}),
            (
                {"code": PROJECT_EXIST_CODE, "message": "project [xxx] existed"},
                {"code": PROJECT_EXIST_CODE, "message": "project [xxx] existed"},
            ),
        ],
    )
    def test_create_project_ok(self, raw_resp, expected_resp, requests_mock):
        requests_mock.post(ANY, json=raw_resp)

        client = bk_repo.BkRepoClient(username=fake_username, access_token=fake_access_token)
        resp_data = client.create_project(fake_project_code, fake_project_name, fake_project_description)
        assert resp_data == expected_resp
        assert requests_mock.called

    def test_create_project_error(self, requests_mock):
        requests_mock.post(ANY, json={"code": fake_error_code, "message": "error message"})

        client = bk_repo.BkRepoClient(username=fake_username, access_token=fake_access_token)
        with pytest.raises(bk_repo.BkRepoCreateProjectError):
            client.create_project(fake_project_code, fake_project_name, fake_project_description)

    @pytest.mark.parametrize(
        "raw_resp,expected_resp",
        [
            ({"code": 0, "data": {"foo": "bar"}}, {"code": 0, "data": {"foo": "bar"}}),
            (
                {"code": REPO_EXIST_CODE, "message": "project [xxx] existed"},
                {"code": REPO_EXIST_CODE, "data": {"foo": "bar"}},
            ),
        ],
    )
    def test_create_repo_ok(self, raw_resp, expected_resp, requests_mock):
        requests_mock.post(ANY, json=raw_resp)

        client = bk_repo.BkRepoClient(username=fake_username, access_token=fake_access_token)
        resp_data = client.create_repo(fake_project_code)
        assert resp_data == raw_resp
        assert requests_mock.request_history[0].method == "POST"

    def test_create_repo_error(self, requests_mock):
        requests_mock.post(ANY, json={"code": fake_error_code, "messahe": "error message"})

        client = bk_repo.BkRepoClient(username=fake_username, access_token=fake_access_token)
        with pytest.raises(bk_repo.BkRepoCreateRepoError):
            client.create_repo(fake_project_code)

    def test_set_auth(self, requests_mock):
        requests_mock.post(ANY, json={"result": True, "data": {"foo": "bar"}})

        client = bk_repo.BkRepoClient(username=fake_username, access_token=fake_access_token)
        resp_data = client.set_auth(fake_project_code, fake_username, fake_pwd)
        assert resp_data == {"foo": "bar"}
        assert requests_mock.request_history[0].method == "POST"

    def test_list_charts(self, requests_mock):
        requests_mock.get(ANY, json={"foo": "bar"})

        client = bk_repo.BkRepoClient(username=fake_username, password=fake_pwd)
        resp_data = client.list_charts(fake_project_code, fake_project_code)
        assert resp_data == {"foo": "bar"}
        assert requests_mock.called

    def test_get_chart_versions(self, requests_mock):
        requests_mock.get(ANY, json={"foo": "bar"})

        client = bk_repo.BkRepoClient(username=fake_username, password=fake_pwd)
        resp_data = client.get_chart_versions(fake_project_code, fake_project_code, fake_chart_name)
        assert resp_data == {"foo": "bar"}
        assert requests_mock.called
        assert requests_mock.request_history[0].method == "GET"

    def test_get_chart_version_detail(self, requests_mock):
        requests_mock.get(ANY, json={"foo": "bar"})

        client = bk_repo.BkRepoClient(username=fake_username, password=fake_pwd)
        resp_data = client.get_chart_version_detail(
            fake_project_code, fake_project_code, fake_chart_name, fake_chart_version
        )
        assert resp_data == {"foo": "bar"}
        assert requests_mock.called
        assert requests_mock.request_history[0].method == "GET"

    @pytest.mark.parametrize(
        "raw_resp,expected_resp",
        [
            ({"deleted": True}, {"deleted": True}),
            (
                {"error": "remove /test.tgz failed: no such file or directory"},
                {"error": "remove /test.tgz failed: no such file or directory"},
            ),
        ],
    )
    def test_delete_chart_version_ok(self, raw_resp, expected_resp, requests_mock):
        requests_mock.delete(ANY, json=raw_resp)

        client = bk_repo.BkRepoClient(username=fake_username, password=fake_pwd)
        resp = client.delete_chart_version(fake_project_code, fake_project_code, fake_chart_name, fake_chart_version)
        assert resp == expected_resp
        assert requests_mock.called
        assert requests_mock.request_history[0].method == "DELETE"

    def test_delete_chart_version_error(self, requests_mock):
        requests_mock.delete(ANY, json={"error": "error message"})

        client = bk_repo.BkRepoClient(username=fake_username, password=fake_pwd)
        with pytest.raises(bk_repo.BkRepoDeleteVersionError):
            client.delete_chart_version(fake_project_code, fake_project_code, fake_chart_name, fake_chart_version)

    def test_list_images(self, requests_mock):
        requests_mock.get(
            ANY,
            json={
                "code": 0,
                "data": {
                    "totalRecords": 10,
                    "records": [
                        {
                            "name": "busybox",
                            "lastModifiedBy": fake_username,
                            "lastModifiedDate": "2021-10-29T15:48:55.121",
                            "downloadCount": 0,
                            "logoUrl": "",
                            "description": "",
                        }
                    ],
                },
            },
        )

        client = bk_repo.BkRepoClient(username=fake_username, password=fake_pwd)
        resp_data = client.list_images(fake_project_code, fake_project_code, bk_repo.PageData())
        assert requests_mock.called
        assert resp_data["totalRecords"] == 10
        assert resp_data["records"][0]["name"] == "busybox"

    def test_list_image_tags(self, requests_mock):
        requests_mock.get(
            ANY,
            json={
                "code": 0,
                "data": {
                    "totalRecords": 10,
                    "records": [
                        {
                            "tag": "latest",
                            "stageTag": "",
                            "size": 527,
                            "lastModifiedBy": fake_username,
                            "lastModifiedDate": "2021-10-29T15:48:55.121",
                            "downloadCount": 0,
                        }
                    ],
                },
            },
        )

        client = bk_repo.BkRepoClient(username=fake_username, password=fake_pwd)
        resp_data = client.list_image_tags(fake_project_code, fake_project_code, "test", bk_repo.PageData())
        assert requests_mock.called
        assert resp_data["totalRecords"] == 10
        assert resp_data["records"][0]["tag"] == "latest"
