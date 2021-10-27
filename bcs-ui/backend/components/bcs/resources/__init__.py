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

from .configmap import ConfigMap
from .daemonset import DaemonSet
from .deployment import Deployment
from .endpoints import Endpoints
from .event import Event
from .ingress import Ingress
from .job import Job
from .namespace import Namespace
from .node import Node
from .pod import Pod
from .pv import PersistentVolume
from .pvc import PersistentVolumeClaim
from .replicaset import ReplicaSet
from .secret import Secret
from .service import Service
from .statefulset import StatefulSet
from .storageclass import StorageClass
