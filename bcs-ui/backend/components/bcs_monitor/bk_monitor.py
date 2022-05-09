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

蓝鲸监控接口封装
"""
import logging
import time

from django.conf import settings

from backend.components.utils import http_post

logger = logging.getLogger(__name__)

BK_MONITOR_QUERY_HOST = settings.BK_MONITOR_QUERY_HOST

# 磁盘统计 忽略的设备类型, 数据来源蓝鲸监控主机查询规则
IGNORE_DEVICE_TYPE = "iso9660|tmpfs|udf"
# 磁盘统计 允许的挂载目录
DISK_MOUNTPOINT = "/data"


def query_range(query, start, end, step, project_id=None, milliseconds=True):
    """范围请求API"""
    url = f'{BK_MONITOR_QUERY_HOST}/query/ts/promql'
    data = {"promql": query, "start": str(int(start)), "end": str(int(end)), "step": f"{step}s"}
    logger.info("prometheus query_range: %s", data)
    bkmonitor_resp = http_post(url, json=data, timeout=120, raise_exception=False)
    prom_resp = bkmonitor_resp2prom_range(bkmonitor_resp, milliseconds)
    return prom_resp


def query(_query, timestamp=None, project_id=None, milliseconds=True):
    """查询API"""
    end = time.time()
    # 蓝鲸监控没有实时数据接口, 这里的方案是向前追溯5分钟, 取最新的一个点
    start = end - 300
    url = f'{BK_MONITOR_QUERY_HOST}/query/ts/promql'
    data = {"promql": _query, "start": str(int(start)), "end": str(int(end)), "step": "60s"}
    logger.info("prometheus query: %s", data)
    bkmonitor_resp = http_post(url, json=data, timeout=120, raise_exception=False)
    logger.info("prometheus query_range: %s", bkmonitor_resp)
    prom_resp = bkmonitor_resp2prom(bkmonitor_resp, milliseconds)
    return prom_resp


def series2prom(resp_series, milliseconds=True):
    """蓝鲸监控数据返回转换为prom返回"""
    result = []
    series_list = resp_series.get('series') or []
    for series in series_list:
        metric = dict(zip(series['group_keys'], series['group_values']))
        values = []

        # 蓝鲸监控返回的values可能会变化, 通过 columes 字段顺序判断
        if series["columns"][0] == "_value":
            value_index = 0
            timestamp_index = 1
        else:
            value_index = 1
            timestamp_index = 0

        for value in series["values"]:

            # 是否使用毫秒单位
            if milliseconds is True:
                timestamp = value[timestamp_index]  # 蓝鲸监控固定单位
            else:
                timestamp = int(value[timestamp_index] / 1000)  # 部分 prom 使用了秒做单位

            values.append((timestamp, str(value[value_index])))

        result.append({'metric': metric, 'values': values})
    return result


def bkmonitor_resp2prom_range(response, milliseconds=True):
    """蓝鲸监控数据返回转换为prom返回 matrix 格式"""
    data = {'resultType': 'matrix', 'result': []}
    result = series2prom(response, milliseconds)
    data['result'] = result
    prom_resp = {'data': data}
    return prom_resp


def bkmonitor_resp2prom(response, milliseconds=True):
    """蓝鲸监控数据返回转换为prom返回 vector 格式"""
    data = {'resultType': 'vector', 'result': []}
    result = series2prom(response, milliseconds)
    for i in result:
        i['value'] = i['values'][-1]
        i.pop('values')
    data['result'] = result
    prom_resp = {'data': data}
    return prom_resp


def get_first_value(prom_resp, fill_zero=True):
    """获取返回的第一个值"""
    data = prom_resp.get("data") or {}
    result = data.get("result") or []
    if not result:
        if fill_zero:
            # 返回0字符串, 和promtheus保存一致
            return "0"
        return None

    value = result[0]["value"]
    if not value:
        if fill_zero:
            return "0"
        return None

    return value[1]


def get_targets(project_id, cluster_id, dedup=True):
    """获取集群的targets"""
    resp = []
    return resp


def get_cluster_cpu_usage(cluster_id, node_ip_list, bk_biz_id=None):
    """获取集群nodeCPU使用率"""
    node_ip_list = "|".join(node_ip_list)

    cpu_used_prom_query = f"""
        sum(bkmonitor:system:cpu_detail:usage{{bk_biz_id="{bk_biz_id}", ip=~"{node_ip_list}"}}) / 100
    """  # noqa

    cpu_count_prom_query = f"""
        count(bkmonitor:system:cpu_detail:usage{{bk_biz_id="{bk_biz_id}", ip=~"{node_ip_list}"}})
    """  # noqa

    data = {"used": get_first_value(query(cpu_used_prom_query)), "total": get_first_value(query(cpu_count_prom_query))}
    return data


def get_cluster_cpu_usage_range(cluster_id, node_ip_list, bk_biz_id=None):
    """获取集群nodeCPU使用率"""
    end = time.time()
    start = end - 3600
    step = 60

    node_ip_list = "|".join(node_ip_list)
    prom_query = f"""
        sum(bkmonitor:system:cpu_detail:usage{{bk_biz_id="{bk_biz_id}", ip=~"{node_ip_list}"}}) /
        count(bkmonitor:system:cpu_detail:usage{{bk_biz_id="{bk_biz_id}", ip=~"{node_ip_list}"}})"""  # noqa

    resp = query_range(prom_query, start, end, step)
    return resp.get("data") or {}


def get_cluster_memory_usage(cluster_id, node_ip_list, bk_biz_id=None):
    """获取集群nodeCPU使用率"""

    node_ip_list = "|".join(node_ip_list)

    memory_total_prom_query = f"""
        sum(bkmonitor:system:mem:total{{bk_biz_id="{bk_biz_id}", ip=~"{node_ip_list}"}})
    """

    memory_used_prom_query = f"""
        sum(bkmonitor:system:mem:used{{bk_biz_id="{bk_biz_id}", ip=~"{node_ip_list}"}})
    """  # noqa

    data = {
        "used_bytes": get_first_value(query(memory_used_prom_query)),
        "total_bytes": get_first_value(query(memory_total_prom_query)),
    }
    return data


def get_cluster_memory_usage_range(cluster_id, node_ip_list, bk_biz_id=None):
    """获取集群nodeCPU使用率"""
    end = time.time()
    start = end - 3600
    step = 60

    node_ip_list = "|".join(node_ip_list)
    prom_query = f"""
        (sum(bkmonitor:system:mem:used{{bk_biz_id="{bk_biz_id}", ip=~"{node_ip_list}"}}) /
        sum(bkmonitor:system:mem:total{{bk_biz_id="{bk_biz_id}", ip=~"{node_ip_list}"}})) *
        100
    """  # noqa

    resp = query_range(prom_query, start, end, step)
    return resp.get("data") or {}


def get_cluster_disk_usage(cluster_id, node_ip_list, bk_biz_id=None):
    """获取集群nodeCPU使用率"""
    node_ip_list = "|".join(node_ip_list)

    disk_total_prom_query = f"""
        sum(bkmonitor:system:disk:total{{bk_biz_id="{bk_biz_id}", device_type!~"{ IGNORE_DEVICE_TYPE }", ip=~"{node_ip_list}"}})
    """  # noqa

    disk_used_prom_query = f"""
        sum(bkmonitor:system:disk:used{{bk_biz_id="{bk_biz_id}", device_type!~"{ IGNORE_DEVICE_TYPE }", ip=~"{node_ip_list}"}})
    """  # noqa

    data = {
        "used_bytes": get_first_value(query(disk_used_prom_query)),
        "total_bytes": get_first_value(query(disk_total_prom_query)),
    }
    return data


def get_cluster_disk_usage_range(cluster_id, node_ip_list, bk_biz_id=None):
    """获取k8s集群磁盘使用率"""
    end = time.time()
    start = end - 3600
    step = 60

    node_ip_list = "|".join(node_ip_list)

    prom_query = f"""
        sum(bkmonitor:system:disk:used{{bk_biz_id="{bk_biz_id}", device_type!~"{ IGNORE_DEVICE_TYPE }", ip=~"{node_ip_list}"}}) /
        sum(bkmonitor:system:disk:total{{bk_biz_id="{bk_biz_id}", device_type!~"{ IGNORE_DEVICE_TYPE }", ip=~"{node_ip_list}"}})
    """  # noqa

    resp = query_range(prom_query, start, end, step)
    return resp.get("data") or {}


def get_node_info(cluster_id, ip, bk_biz_id=None):
    info_resp = {"resultType": "vector", "result": []}

    node_info_query = f"""
        cadvisor_version_info{{bk_biz_id="{bk_biz_id}", bcs_cluster_id="{cluster_id}", bk_instance=~"{ip}:.*"}}
    """
    resp_data = query(node_info_query).get('data') or []
    if resp_data and resp_data.get('result'):
        result = resp_data['result'][0]
        # 字段和 prom 对齐
        result['metric']['sysname'] = result['metric']['osVersion']
        result['metric']['release'] = result['metric']['kernelVersion']
        info_resp['result'].append(result)

    node_core_count_query = f"""
        count(bkmonitor:system:cpu_detail:usage{{bk_biz_id="{bk_biz_id}", ip="{ip}"}})
    """
    resp_data = query(node_core_count_query).get('data') or []
    if resp_data and resp_data.get('result'):
        result = resp_data['result'][0]
        result['metric']['metric_name'] = "cpu_count"
        info_resp['result'].append(result)

    node_memory_query = f"""
        sum(bkmonitor:system:mem:total{{bk_biz_id="{bk_biz_id}", ip="{ip}"}})
    """
    resp_data = query(node_memory_query).get('data') or []
    if resp_data and resp_data.get('result'):
        result = resp_data['result'][0]
        result['metric']['metric_name'] = "memory"
        info_resp['result'].append(result)

    node_disk_size_query = f"""
        sum(bkmonitor:system:disk:total{{bk_biz_id="{bk_biz_id}", device_type!~"{ IGNORE_DEVICE_TYPE }", ip="{ip}"}})
    """
    resp_data = query(node_disk_size_query).get('data') or []
    if resp_data and resp_data.get('result'):
        result = resp_data['result'][0]
        result['metric']['metric_name'] = "disk"
        info_resp['result'].append(result)

    return info_resp


def get_container_pod_count(cluster_id, ip, bk_biz_id=None):
    """获取K8S节点容器/Pod数量"""
    count_resp = {"resultType": "vector", "result": []}

    # 注意 k8s 1.19 版本以前的 metrics 是 kubelet_running_container_count
    container_count_query = f"""
        max by( bk_instance ) (kubelet_running_containers{{bk_biz_id="{bk_biz_id}", bcs_cluster_id="{cluster_id}", container_state="running", bk_instance=~"{ip}:.*"}})
    """  # noqa
    resp_data = query(container_count_query).get('data') or []
    if resp_data and resp_data.get('result'):
        result = resp_data['result'][0]
        result['metric']['metric_name'] = "container_count"
        count_resp['result'].append(result)

    # 注意 k8s 1.19 版本以前的 metrics 是 kubelet_running_pod_count
    pod_count_query = f"""
        max by( bk_instance ) (kubelet_running_pods{{bk_biz_id="{bk_biz_id}", bcs_cluster_id="{cluster_id}", bk_instance=~"{ip}:.*"}})
    """  # noqa
    resp_data = query(pod_count_query).get('data') or []
    if resp_data and resp_data.get('result'):
        result = resp_data['result'][0]
        result['metric']['metric_name'] = "pod_count"
        count_resp['result'].append(result)

    return count_resp


def get_node_cpu_usage(cluster_id, ip, bk_biz_id=None):
    """获取CPU总使用率"""
    bk_biz_id = "2"
    prom_query = f"""
        sum(bkmonitor:system:cpu_detail:usage{{bk_biz_id="{bk_biz_id}", ip="{ip}"}}) /
        count(bkmonitor:system:cpu_detail:usage{{bk_biz_id="{bk_biz_id}", ip="{ip}"}})"""  # noqa

    resp = query(prom_query)
    value = get_first_value(resp)
    return value


def get_node_cpu_usage_range(cluster_id, ip, start, end, bk_biz_id=None):
    """获取CPU总使用率
    start, end单位为毫秒，和数据平台保持一致
    """
    step = (end - start) // 60

    prom_query = f"""
        sum(bkmonitor:system:cpu_detail:usage{{bk_biz_id="{bk_biz_id}", ip="{ip}"}}) /
        count(bkmonitor:system:cpu_detail:usage{{bk_biz_id="{bk_biz_id}", ip="{ip}"}})"""  # noqa

    resp = query_range(prom_query, start, end, step)
    return resp.get("data") or {}


def get_node_memory_usage(cluster_id, ip, bk_biz_id=None):
    """获取节点内存使用率"""
    prom_query = f"""
        (sum(bkmonitor:system:mem:used{{bk_biz_id="{bk_biz_id}", ip="{ip}"}}) /
        sum(bkmonitor:system:mem:total{{bk_biz_id="{bk_biz_id}", ip="{ip}"}})) *
        100
    """  # noqa

    resp = query(prom_query)
    value = get_first_value(resp)
    return value


def get_node_memory_usage_range(cluster_id, ip, start, end, bk_biz_id=None):
    """获取CPU总使用率
    start, end单位为毫秒，和数据平台保持一致
    """
    step = (end - start) // 60
    prom_query = f"""
        (sum(bkmonitor:system:mem:used{{bk_biz_id="{bk_biz_id}", ip="{ip}"}}) /
        sum(bkmonitor:system:mem:total{{bk_biz_id="{bk_biz_id}", ip="{ip}"}})) *
        100
    """  # noqa

    resp = query_range(prom_query, start, end, step)
    return resp.get("data") or {}


def get_node_disk_usage(cluster_id, ip, bk_biz_id=None):
    prom_query = f"""
        (sum(bkmonitor:system:disk:used{{bk_biz_id="{bk_biz_id}", device_type!~"{ IGNORE_DEVICE_TYPE }", ip="{ip}"}}) /
        sum(bkmonitor:system:disk:total{{bk_biz_id="{bk_biz_id}", device_type!~"{ IGNORE_DEVICE_TYPE }", ip="{ip}"}})) *
        100
    """  # noqa

    value = get_first_value(query(prom_query))
    return value


def get_node_network_receive(cluster_id, ip, start, end, bk_biz_id=None):
    """获取网络数据
    start, end单位为毫秒，和数据平台保持一致
    数据单位KB/s
    """
    step = (end - start) // 60
    prom_query = f"""
        max(bkmonitor:system:net:speed_recv{{bk_biz_id="{bk_biz_id}", ip="{ ip }"}})
    """  # noqa
    resp = query_range(prom_query, start, end, step)
    return resp.get("data") or {}


def get_node_network_transmit(cluster_id, ip, start, end, bk_biz_id=None):
    step = (end - start) // 60
    prom_query = f"""
        max(bkmonitor:system:net:speed_sent{{bk_biz_id="{bk_biz_id}", ip="{ ip }"}})
        """  # noqa
    resp = query_range(prom_query, start, end, step)
    return resp.get("data") or {}


def get_node_diskio_usage(cluster_id, ip, bk_biz_id=None):
    """获取当前磁盘IO"""
    prom_query = f"""
        max(bkmonitor:system:io:util{{bk_biz_id="{bk_biz_id}", ip="{ip}"}}) * 100
    """  # noqa

    value = get_first_value(query(prom_query))
    return value


def get_node_diskio_usage_range(cluster_id, ip, start, end, bk_biz_id=None):
    """获取磁盘IO数据
    start, end单位为毫秒，和数据平台保持一致
    数据单位KB/s
    """
    step = (end - start) // 60
    prom_query = f"""
        max(bkmonitor:system:io:util{{bk_biz_id="{bk_biz_id}", ip="{ip}"}}) * 100
    """  # noqa

    resp = query_range(prom_query, start, end, step)
    return resp.get("data") or {}


def get_pod_cpu_usage_range(cluster_id, namespace, pod_name_list, start, end, bk_biz_id=None):
    """获取CPU总使用率
    start, end单位为毫秒，和数据平台保持一致
    """
    step = (end - start) // 60
    pod_name_list = "|".join(pod_name_list)

    porm_query = f"""
        sum by (pod_name) (rate(container_cpu_usage_seconds_total{{bk_biz_id="{bk_biz_id}", bcs_cluster_id="{cluster_id}", namespace=~"{ namespace }",
        pod_name=~"{ pod_name_list }", container_name!="", container_name!="POD"}}[2m])) * 100
        """  # noqa
    resp = query_range(porm_query, start, end, step)

    return resp.get("data") or {}


def get_pod_memory_usage_range(cluster_id, namespace, pod_name_list, start, end, bk_biz_id=None):
    """获取CPU总使用率
    start, end单位为毫秒，和数据平台保持一致
    """
    step = (end - start) // 60
    pod_name_list = "|".join(pod_name_list)

    porm_query = f"""
        sum by (pod_name) (container_memory_working_set_bytes{{bk_biz_id="{bk_biz_id}", bcs_cluster_id="{cluster_id}", namespace=~"{ namespace }", pod_name=~"{ pod_name_list }",
        container_name!="", container_name!="POD"}})
        """  # noqa
    resp = query_range(porm_query, start, end, step)

    return resp.get("data") or {}


def get_pod_network_receive(cluster_id, namespace, pod_name_list, start, end, bk_biz_id=None):
    """获取网络数据
    start, end单位为毫秒，和数据平台保持一致
    """
    step = (end - start) // 60
    pod_name_list = "|".join(pod_name_list)

    prom_query = f"""
        sum by(pod_name) (rate(container_network_receive_bytes_total{{bk_biz_id="{bk_biz_id}", bcs_cluster_id="{cluster_id}", namespace=~"{ namespace }", pod_name=~"{ pod_name_list }"}}[2m]))
        """  # noqa

    resp = query_range(prom_query, start, end, step)
    return resp.get("data") or {}


def get_pod_network_transmit(cluster_id, namespace, pod_name_list, start, end, bk_biz_id=None):
    step = (end - start) // 60
    pod_name_list = "|".join(pod_name_list)

    prom_query = f"""
        sum by(pod_name) (rate(container_network_transmit_bytes_total{{bk_biz_id="{bk_biz_id}", bcs_cluster_id="{cluster_id}",  namespace=~"{ namespace }", pod_name=~"{ pod_name_list }"}}[2m]))
        """  # noqa

    resp = query_range(prom_query, start, end, step)
    return resp.get("data") or {}


def get_container_cpu_usage_range(cluster_id, namespace, pod_name, container_name, start, end, bk_biz_id=None):
    """获取CPU总使用率
    start, end单位为毫秒，和数据平台保持一致
    """
    step = (end - start) // 60

    prom_query = f"""
        sum by(container_name) (rate(container_cpu_usage_seconds_total{{bk_biz_id="{bk_biz_id}", bcs_cluster_id="{cluster_id}", namespace=~"{ namespace }", pod_name=~"{pod_name}",
        container_name=~"{ container_name }", container_name!="", container_name!="POD", BcsNetworkContainer!="true"}}[2m])) * 100
        """  # noqa

    resp = query_range(prom_query, start, end, step, milliseconds=False)
    return resp.get("data") or {}


def get_container_cpu_limit(cluster_id, namespace, pod_name, container_name, bk_biz_id=None):
    """获取CPU总使用率
    start, end单位为毫秒，和数据平台保持一致
    """

    prom_query = f"""
        max by(container_name) (container_spec_cpu_quota{{bk_biz_id="{bk_biz_id}", bcs_cluster_id="{cluster_id}", namespace=~"{ namespace }", pod_name=~"{pod_name}",
        container_name=~"{ container_name }", container_name!="", container_name!="POD", BcsNetworkContainer!="true"}})
        """  # noqa

    resp = query(prom_query, milliseconds=False)
    return resp.get("data") or {}


def get_container_memory_usage_range(cluster_id, namespace, pod_name, container_name, start, end, bk_biz_id=None):
    """获取CPU总使用率
    start, end单位为毫秒，和数据平台保持一致
    """
    step = (end - start) // 60

    prom_query = f"""
        sum by(container_name) (container_memory_working_set_bytes{{bk_biz_id="{bk_biz_id}", bcs_cluster_id="{cluster_id}", namespace=~"{ namespace }",pod_name=~"{pod_name}",
        container_name=~"{ container_name }", container_name!="", container_name!="POD", BcsNetworkContainer!="true"}})
        """  # noqa

    resp = query_range(prom_query, start, end, step, milliseconds=False)
    return resp.get("data") or {}


def get_container_memory_limit(cluster_id, namespace, pod_name, container_name, bk_biz_id=None):
    """获取CPU总使用率
    start, end单位为毫秒，和数据平台保持一致
    """

    prom_query = f"""
        max by(container_name) (container_spec_memory_limit_bytes{{bk_biz_id="{bk_biz_id}", bcs_cluster_id="{cluster_id}", namespace=~"{ namespace }", pod_name=~"{pod_name}",
        container_name=~"{ container_name }", container_name!="", container_name!="POD", BcsNetworkContainer!="true"}}) > 0
        """  # noqa

    resp = query(prom_query, milliseconds=False)
    return resp.get("data") or {}


def get_container_disk_read(cluster_id, namespace, pod_name, container_name, start, end, bk_biz_id=None):
    step = (end - start) // 60

    prom_query = f"""
        sum by(container_name) (container_fs_reads_bytes_total{{bk_biz_id="{bk_biz_id}", bcs_cluster_id="{cluster_id}", namespace=~"{ namespace }", pod_name=~"{pod_name}",
        container_name=~"{ container_name }", container_name!="", container_name!="POD", BcsNetworkContainer!="true"}})
        """  # noqa

    resp = query_range(prom_query, start, end, step)
    return resp.get("data") or {}


def get_container_disk_write(cluster_id, namespace, pod_name, container_name, start, end, bk_biz_id=None):
    step = (end - start) // 60

    prom_query = f"""
        sum by(container_name) (container_fs_writes_bytes_total{{bk_biz_id="{bk_biz_id}", bcs_cluster_id="{cluster_id}", namespace=~"{ namespace }", pod_name=~"{pod_name}",
        container_name=~"{ container_name }", container_name!="", container_name!="POD", BcsNetworkContainer!="true"}})
        """  # noqa

    resp = query_range(prom_query, start, end, step)
    return resp.get("data") or {}


def mesos_agent_memory_usage(cluster_id, ip, bk_biz_id=None):
    """mesos内存使用率"""
    data = {"total": "0", "remain": "0"}
    return data


def mesos_agent_cpu_usage(cluster_id, ip, bk_biz_id=None):
    """mesosCPU使用率"""
    data = {"total": "0", "remain": "0"}
    return data


def mesos_agent_ip_remain_count(cluster_id, ip, bk_biz_id=None):
    """mesos 剩余IP数量"""
    value = 0
    return value


def mesos_cluster_cpu_usage(cluster_id, node_list, bk_biz_id=None):
    """mesos集群CPU使用率"""
    data = {"total": "0", "remain": "0"}
    return data


def mesos_cluster_memory_usage(cluster_id, node_list, bk_biz_id=None):
    """mesos集群mem使用率"""
    data = {"total": "0", "remain": "0"}
    return data


def mesos_cluster_cpu_resource_remain_range(cluster_id, start, end, bk_biz_id=None):
    """mesos集群CPU剩余量, 单位核"""
    data = {}
    return data


def mesos_cluster_cpu_resource_total_range(cluster_id, start, end, bk_biz_id=None):
    """mesos集群CPU总量, 单位核"""
    data = {}
    return data


def mesos_cluster_memory_resource_remain_range(cluster_id, start, end, bk_biz_id=None):
    """mesos集群内存剩余量, 单位MB"""
    data = {}
    return data


def mesos_cluster_memory_resource_total_range(cluster_id, start, end, bk_biz_id=None):
    """mesos集群内存总量, 单位MB"""
    data = {}
    return data


def mesos_cluster_cpu_resource_used_range(cluster_id, start, end, bk_biz_id=None):
    """mesos集群使用的CPU, 单位核"""
    data = {}
    return data


def mesos_cluster_memory_resource_used_range(cluster_id, start, end, bk_biz_id=None):
    """mesos集群使用的内存, 单位MB"""
    data = {}
    return data
