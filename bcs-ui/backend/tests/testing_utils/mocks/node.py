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
from backend.utils import FancyDict


class StubNodeClient:
    def __init__(self, *args, **kwargs):
        pass

    def list(self, *args, **kwargs):
        return self.make_node_data()

    @staticmethod
    def make_node_data():
        return [
            {
                "apiVersion": "v1",
                "kind": "Node",
                "metadata": {
                    "annotations": {
                        "node.alpha.kubernetes.io/ttl": "0",
                        "volumes.kubernetes.io/controller-managed-attach-detach": "true",
                    },
                    "creationTimestamp": "2020-09-16T05:24:43Z",
                    "labels": {
                        "kubernetes.io/arch": "amd64",
                        "kubernetes.io/os": "linux",
                    },
                    "name": "bcs-test-node",
                    "resourceVersion": "95278508",
                    "selfLink": "/api/v1/nodes/bcs-test-node",
                    "uid": "ed65e985-f7dc-11ea-a432-525400ed6cb7",
                },
                "spec": {"podCIDR": "127.0.0.1/26", "providerID": "bcs"},
                "status": {
                    "addresses": [
                        {"address": "127.0.0.1", "type": "InternalIP"},
                        {"address": "bcs-test-node", "type": "Hostname"},
                    ],
                    "conditions": [
                        {
                            "lastHeartbeatTime": "2021-07-07T04:13:48Z",
                            "lastTransitionTime": "2020-09-16T05:24:53Z",
                            "message": "kubelet is posting ready status",
                            "reason": "KubeletReady",
                            "status": "True",
                            "type": "Ready",
                        },
                    ],
                    "images": [
                        {
                            "names": ["bcs.demo.com/bkpaas/test@sha256"],
                            "sizeBytes": 2140479994,
                        },
                    ],
                },
            }
        ]
