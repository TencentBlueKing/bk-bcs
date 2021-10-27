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
from backend.resources.constants import ConditionStatus, PodConditionType, PodPhase

# PodStatus Failed
FailedStatusPodConfig = {
    'status': {
        'phase': PodPhase.PodFailed.value,
        'conditions': [
            {
                'type': PodConditionType.PodInitialized.value,
                'status': ConditionStatus.ConditionTrue.value,
            }
        ],
    }
}

# PodStatus Succeeded
SucceededStatusPodConfig = {
    'status': {
        'phase': PodPhase.PodSucceeded.value,
        'conditions': [
            {
                'type': PodConditionType.PodInitialized.value,
                'status': ConditionStatus.ConditionTrue.value,
            }
        ],
    }
}

# PodStatus Running
RunningStatusPodConfig = {
    'status': {
        'phase': PodPhase.PodRunning.value,
        'conditions': [
            {
                'type': PodConditionType.PodInitialized.value,
                'status': ConditionStatus.ConditionTrue.value,
            },
            {
                'type': PodConditionType.PodReady.value,
                'status': ConditionStatus.ConditionTrue.value,
            },
        ],
    }
}

# PodStatus Pending
PendingStatusPodConfig = {
    'status': {
        'phase': PodPhase.PodPending.value,
        'conditions': [
            {
                'type': PodConditionType.PodInitialized.value,
                'status': ConditionStatus.ConditionFalse.value,
            }
        ],
    }
}

# PodStatus Terminating
TerminatingStatusPodConfig = {
    'metadata': {
        'deletionTimestamp': '2021-01-01T10:00:00Z',
    },
    'status': {
        'phase': PodPhase.PodRunning.value,
    },
}

# PodStatus Unknown
UnknownStatusPodConfig = {
    'metadata': {
        'deletionTimestamp': '2021-01-01T10:00:00Z',
    },
    'status': {
        'phase': PodPhase.PodRunning.value,
        'reason': 'NodeLost',
    },
}

# PodStatus Completed
CompletedStatusPodConfig = {
    'status': {
        'phase': PodPhase.PodSucceeded.value,
        'containerStatuses': [
            {
                'state': {
                    'terminated': {
                        'reason': 'Completed',
                    }
                }
            }
        ],
    }
}

# PodStatus CreateContainerError
CreateContainerErrorStatusPodConfig = {
    'status': {
        'phase': PodPhase.PodPending.value,
        'containerStatuses': [
            {
                'state': {
                    'waiting': {
                        'message': 'Error response from daemon: No command specified',
                        'reason': 'CreateContainerError',
                    }
                }
            }
        ],
    }
}
