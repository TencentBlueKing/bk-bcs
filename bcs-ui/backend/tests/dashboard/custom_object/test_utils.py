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
from unittest.mock import patch

import pytest

from backend.dashboard.custom_object.utils import to_table_format

no_additional_printer_columns = {
    "crd_dict": {
        "apiVersion": "apiextensions.k8s.io/v1",
        "kind": "CustomResourceDefinition",
        "metadata": {
            "annotations": {"api-approved.kubernetes.io": "unapproved, experimental-only"},
            "creationTimestamp": "2021-01-14T04:15:27Z",
            "generation": 1,
            "name": "foos.samplecontroller.k8s.io",
            "resourceVersion": "858834",
            "uid": "68ee935e-6b93-4839-8a42-4952ad01a318",
        },
        "spec": {
            "conversion": {"strategy": "None"},
            "group": "samplecontroller.k8s.io",
            "names": {"kind": "Foo", "listKind": "FooList", "plural": "foos", "singular": "foo"},
            "preserveUnknownFields": True,
            "scope": "Namespaced",
            "versions": [{"name": "v1alpha1", "served": True, "storage": True}],
        },
    },
    "cobj_list": [
        {
            "apiVersion": "samplecontroller.k8s.io/v1alpha1",
            "kind": "Foo",
            "metadata": {
                "creationTimestamp": "2021-01-14T04:20:31Z",
                "generation": 2,
                "name": "example-foo",
                "namespace": "default",
                "resourceVersion": "859054",
                "selfLink": "/apis/samplecontroller.k8s.io/v1alpha1/namespaces/default/foos/example-foo",
                "uid": "856f4f07-5cd7-4e82-8669-106cc3976a71",
            },
            "spec": {"deploymentName": "example-foo", "replicas": 1},
            "status": {"availableReplicas": 0},
        }
    ],
    "table_format": {
        "th_list": ["name", "cluster_id", "namespace", "AGE"],
        "td_list": [{"name": "example-foo", "cluster_id": "", "namespace": "default", "AGE": "2h1m"}],
    },
}

with_additional_printer_columns = {
    "crd_dict": {
        'apiVersion': 'apiextensions.k8s.io/v1beta1',
        'kind': 'CustomResourceDefinition',
        'metadata': {
            'labels': {'release': 'gamedeployment'},
            'name': 'gamedeployments.tkex.tencent.com',
        },
        'spec': {
            'additionalPrinterColumns': [
                {
                    'JSONPath': '.spec.replicas',
                    'description': 'The desired number of pods.',
                    'name': 'DESIRED',
                    'type': 'integer',
                },
                {
                    'JSONPath': '.status.updatedReplicas',
                    'description': 'The number of pods updated.',
                    'name': 'UPDATED',
                    'type': 'integer',
                },
                {
                    'JSONPath': '.metadata.creationTimestamp',
                    "description": "creationTimestamp",
                    'name': 'AGE',
                    'type': 'date',
                },
            ],
            'group': 'tkex.tencent.com',
            'names': {
                'kind': 'GameDeployment',
                'listKind': 'GameDeploymentList',
                'plural': 'gamedeployments',
                'singular': 'gamedeployment',
            },
            'scope': 'Namespaced',
            'subresources': {
                'scale': {
                    'labelSelectorPath': '.status.labelSelector',
                    'specReplicasPath': '.spec.replicas',
                    'statusReplicasPath': '.status.replicas',
                },
                'status': {},
            },
            'version': 'v1alpha1',
        },
    },
    "cobj_list": [
        {
            "apiVersion": "tkex.tencent.com/v1alpha1",
            "kind": "GameDeployment",
            "metadata": {
                "creationTimestamp": "2020-12-07T08:53:39Z",
                "generation": 7,
                "name": "test-gamedeployment",
                "namespace": "default",
            },
            "spec": {
                "minReadySeconds": 10,
                "preDeleteUpdateStrategy": {"hook": {"templateName": "test"}},
                "replicas": 2,
                "selector": {"matchLabels": {"app": "test"}},
            },
            "status": {
                "availableReplicas": 2,
                "readyReplicas": 2,
                "replicas": 2,
                "updateRevision": "test-gamedeployment-675f4ffc75",
                "updatedReadyReplicas": 2,
                "updatedReplicas": 2,
            },
        },
    ],
    "table_format": {
        'th_list': ["name", "cluster_id", "namespace", "DESIRED", "UPDATED", "AGE"],
        'td_list': [
            {
                "name": "test-gamedeployment",
                "cluster_id": "",
                "namespace": "default",
                "DESIRED": 2,
                "UPDATED": 2,
                "AGE": "2h1m",
            }
        ],
    },
}


@pytest.mark.parametrize(
    "crd_dict, cobj_list, table_format",
    [
        ({}, [], {'td_list': [], 'th_list': []}),
        (
            no_additional_printer_columns["crd_dict"],
            no_additional_printer_columns["cobj_list"],
            no_additional_printer_columns["table_format"],
        ),
        (
            with_additional_printer_columns["crd_dict"],
            with_additional_printer_columns["cobj_list"],
            with_additional_printer_columns["table_format"],
        ),
    ],
)
@patch("backend.dashboard.custom_object.utils.calculate_age", return_value="2h1m")
def test_to_table_format(mock_calculate_age, crd_dict, cobj_list, table_format):
    assert to_table_format(crd_dict, cobj_list) == table_format
