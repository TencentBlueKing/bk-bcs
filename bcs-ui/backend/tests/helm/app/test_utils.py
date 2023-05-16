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
from ruamel.yaml.comments import CommentedMap

from backend.helm.app.utils import remove_updater_creator_from_manifest, ruamel_yaml_dump, yaml_dump

FAKE_MANIFEST_YAML = """
apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: test12-redis\n  labels:\n    app: bk-redis\n    chart: bk-redis-0.1.29\n    release: test12\n    heritage: Helm\n    io.tencent.paas.source_type: helm\n    io.tencent.paas.projectid: xxx\n    io.tencent.bcs.clusterid: BCS-K8S-00000\n    io.tencent.bcs.namespace: test-tes123\n    io.tencent.bcs.controller.type: Deployment\n    io.tencent.bcs.controller.name: test12-redis\n  annotations:\n    io.tencent.paas.version: 0.1.29\n    io.tencent.bcs.clusterid: BCS-K8S-00000\nspec:\n  replicas: 1\n  selector:\n    matchLabels:\n      app: bk-redis\n      release: test12\n  template:\n    metadata:\n      labels:\n        app: bk-redis\n        release: test12\n        app-name: test-db\n        io.tencent.paas.source_type: helm\n        io.tencent.paas.projectid: xxx\n        io.tencent.bcs.clusterid: BCS-K8S-00000\n        io.tencent.bcs.namespace: test-tes123\n        io.tencent.bcs.controller.type: Deployment\n        io.tencent.bcs.controller.name: test12-redis\n    spec:\n      containers:\n      - name: bk-redis\n        image: /paas/test/test:latest\n        imagePullPolicy: IfNotPresent\n        env:\n        - name: test\n          value: test\n        - name: test\n          value: test123\n        - name: test\n          value: ieod\n        - name: test\n          value: test\n        - name: test\n          value: \"80\"\n        - name: test\n          value: \"true\"\n        - name: test\n          value: test\n        - name: io_tencent_bcs_namespace\n          value: test-tes123\n        - name: io_tencent_bcs_custom_labels\n          value: '{}'\n        command:\n        - bash -c\n        ports:\n        - name: http\n          containerPort: 80\n          protocol: TCP\n        livenessProbe:\n          httpGet:\n            path: /\n            port: http\n        readinessProbe:\n          httpGet:\n            path: /\n            port: http\n        resources: {}\n      imagePullSecrets:\n      - name: paas.image.registry.test-tes123\n---\napiVersion: batch/v1\nkind: Job\nmetadata:\n  name: test12-db-migrate\n  labels:\n    io.tencent.paas.source_type: helm\n    io.tencent.paas.projectid: xxx\n    io.tencent.bcs.clusterid: BCS-K8S-00000\n    io.tencent.bcs.namespace: test-tes123\n    io.tencent.bcs.controller.type: Job\n    io.tencent.bcs.controller.name: test12-db-migrate\n  annotations:\n    io.tencent.paas.version: 0.1.29\n    io.tencent.bcs.clusterid: BCS-K8S-00000\nspec:\n  backoffLimit: 0\n  template:\n    metadata:\n      name: test12\n      labels:\n        app.kubernetes.io/managed-by: Helm\n        app.kubernetes.io/instance: test12\n        helm.sh/chart: bk-redis-0.1.29\n        io.tencent.paas.source_type: helm\n        io.tencent.paas.projectid: xxx\n        io.tencent.bcs.clusterid: BCS-K8S-00000\n        io.tencent.bcs.namespace: test-tes123\n        io.tencent.bcs.controller.type: Job\n        io.tencent.bcs.controller.name: test12-db-migrate\n    spec:\n      restartPolicy: Never\n      containers:\n      - name: pre-install-job\n        image: /paas/test/test:latest\n        command:\n        - /bin/bash\n        - -c\n        args:\n        - python manage.py migrate\n        env:\n        - name: test\n          value: test\n        - name: test\n          value: test\n        - name: test\n          value: test\n        - name: test\n          value: \"80\"\n        - name: test\n          value: \"true\"\n        - name: test\n          value: test\n        - name: io_tencent_bcs_namespace\n          value: test-tes123\n        - name: io_tencent_bcs_custom_labels\n          value: '{}'\n        imagePullPolicy: Always\n      imagePullSecrets:\n      - name: paas.image.registry.test-tes123\n
"""  # noqa

EXPECTED_MANIFEST_YAML = """
apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: test12-redis\n  labels:\n    app: bk-redis\n    chart: bk-redis-0.1.29\n    release: test12\n    heritage: Helm\n    io.tencent.paas.source_type: helm\n    io.tencent.paas.projectid: xxx\n    io.tencent.bcs.clusterid: BCS-K8S-00000\n    io.tencent.bcs.namespace: test-tes123\n    io.tencent.bcs.controller.type: Deployment\n    io.tencent.bcs.controller.name: test12-redis\n  annotations:\n    io.tencent.paas.version: 0.1.29\n    io.tencent.bcs.clusterid: BCS-K8S-00000\nspec:\n  replicas: 1\n  selector:\n    matchLabels:\n      app: bk-redis\n      release: test12\n  template:\n    metadata:\n      labels:\n        app: bk-redis\n        release: test12\n        app-name: test-db\n        io.tencent.paas.source_type: helm\n        io.tencent.paas.projectid: xxx\n        io.tencent.bcs.clusterid: BCS-K8S-00000\n        io.tencent.bcs.namespace: test-tes123\n        io.tencent.bcs.controller.type: Deployment\n        io.tencent.bcs.controller.name: test12-redis\n    spec:\n      containers:\n      - name: bk-redis\n        image: /paas/test/test:latest\n        imagePullPolicy: IfNotPresent\n        env:\n        - name: test\n          value: test\n        - name: test\n          value: test123\n        - name: test\n          value: ieod\n        - name: test\n          value: test\n        - name: test\n          value: \"80\"\n        - name: test\n          value: \"true\"\n        - name: test\n          value: test\n        - name: io_tencent_bcs_namespace\n          value: test-tes123\n        - name: io_tencent_bcs_custom_labels\n          value: '{}'\n        command:\n        - bash -c\n        ports:\n        - name: http\n          containerPort: 80\n          protocol: TCP\n        livenessProbe:\n          httpGet:\n            path: /\n            port: http\n        readinessProbe:\n          httpGet:\n            path: /\n            port: http\n        resources: {}\n      imagePullSecrets:\n      - name: paas.image.registry.test-tes123\n---\napiVersion: batch/v1\nkind: Job\nmetadata:\n  name: test12-db-migrate\n  labels:\n    io.tencent.paas.source_type: helm\n    io.tencent.paas.projectid: xxx\n    io.tencent.bcs.clusterid: BCS-K8S-00000\n    io.tencent.bcs.namespace: test-tes123\n    io.tencent.bcs.controller.type: Job\n    io.tencent.bcs.controller.name: test12-db-migrate\n  annotations:\n    io.tencent.paas.version: 0.1.29\n    io.tencent.bcs.clusterid: BCS-K8S-00000\nspec:\n  backoffLimit: 0\n  template:\n    metadata:\n      name: test12\n      labels:\n        app.kubernetes.io/managed-by: Helm\n        app.kubernetes.io/instance: test12\n        helm.sh/chart: bk-redis-0.1.29\n        io.tencent.paas.source_type: helm\n        io.tencent.paas.projectid: xxx\n        io.tencent.bcs.clusterid: BCS-K8S-00000\n        io.tencent.bcs.namespace: test-tes123\n        io.tencent.bcs.controller.type: Job\n        io.tencent.bcs.controller.name: test12-db-migrate\n    spec:\n      restartPolicy: Never\n      containers:\n      - name: pre-install-job\n        image: /paas/test/test:latest\n        command:\n        - /bin/bash\n        - -c\n        args:\n        - python manage.py migrate\n        env:\n        - name: test\n          value: test\n        - name: test\n          value: test\n        - name: test\n          value: test\n        - name: test\n          value: \"80\"\n        - name: test\n          value: \"true\"\n        - name: test\n          value: test\n        - name: io_tencent_bcs_namespace\n          value: test-tes123\n        - name: io_tencent_bcs_custom_labels\n          value: '{}'\n        imagePullPolicy: Always\n      imagePullSecrets:\n      - name: paas.image.registry.test-tes123\n
"""  # noqa


def test_remove_updater_creator():
    updater_creator = ["io.tencent.paas.updator: admin", "io.tencent.paas.creator: admin"]
    mf = remove_updater_creator_from_manifest(FAKE_MANIFEST_YAML)
    for key in updater_creator:
        assert key not in mf
    assert mf == EXPECTED_MANIFEST_YAML


@pytest.mark.parametrize(
    "json_src,expected_yaml",
    [
        (CommentedMap([('a', '1')]), "a: '1'\n"),
        (CommentedMap([('a', 1)]), "a: 1\n"),
        (CommentedMap([('a', 'Y')]), "a: Y\n"),
        (CommentedMap([('a', CommentedMap([('a1', '1')]))]), "a:\n  a1: '1'\n"),
        (CommentedMap([('a', CommentedMap([('a1', 1)]))]), "a:\n  a1: 1\n"),
    ],
)
def test_ruamel_yaml_dump(json_src, expected_yaml):
    assert ruamel_yaml_dump(json_src) == expected_yaml


@pytest.mark.parametrize(
    "json_src,expected_yaml",
    [
        ({"a": "1"}, '"a": "1"\n'),
        ({"a": 1}, '"a": 1\n'),
        ({"a": "Y"}, '"a": "Y"\n'),
        ({"a": "foo\nbar"}, '"a": |-\n  foo\n  bar\n'),
        ({"a": {"a1": "1"}}, '"a":\n  "a1": "1"\n'),
        ({"a": {"a1": 1}}, '"a":\n  "a1": 1\n'),
    ],
)
def test_yaml_dump(json_src, expected_yaml):
    assert yaml_dump(json_src) == expected_yaml
