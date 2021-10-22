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

from backend.resources.utils.format import ResourceDefaultFormatter


@pytest.fixture
def pod_resource():
    return {
        'apiVersion': 'v1',
        'kind': 'Pod',
        'metadata': {
            'creationTimestamp': '2021-01-07T12:18:40Z',
            'name': 'demo-pod',
            'namespace': 'demo-pod-ns',
            'resourceVersion': '1828',
            'selfLink': '/api/v1/namespaces/vfuqk1or/pods/vfuqk1or',
            'uid': '7a0173ef-50e2-11eb-8821-0242ac150002',
            'labels': {'io.tencent.bcs.clusterid': 'demo-cluster'},
        },
        'spec': {},
    }


class TestResourceDefaultFormatter:
    @pytest.mark.parametrize(
        'labels,cluster_id',
        [
            (None, ''),
            ({'io.tencent.bcs.clusterid': 'demo-cluster'}, 'demo-cluster'),
        ],
    )
    def test_format_diff_cluster_id(self, labels, cluster_id, pod_resource):
        pod_resource['metadata']['labels'] = labels
        formatted_data = ResourceDefaultFormatter().format_dict(pod_resource)

        assert formatted_data['clusterId'] == cluster_id
        assert formatted_data['resourceType'] == 'Pod'
        assert formatted_data['resourceName'] == 'demo-pod'
        assert formatted_data['namespace'] == 'demo-pod-ns'
        assert formatted_data['data']['metadata']['annotations'] == {}
