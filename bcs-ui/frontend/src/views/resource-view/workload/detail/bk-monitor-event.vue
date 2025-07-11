<template>
  <div>
    <iframe class="w-full min-h-[800px]" :src="bkMonitorEventSourceURL" :key="refreshIframe"></iframe>
  </div>
</template>

<script lang="ts">
import { computed, defineComponent, onBeforeMount, ref, toRefs, watch } from 'vue';

import { getEventDataID } from '@/api/modules/monitor';
import { useProject } from '@/composables/use-app';
import $i18n from '@/i18n/i18n-setup';

interface ISearchCondition {
  condition: 'and' | 'or'
  key: string
  method: 'eq' | 'include'
  value: string[]
}

export default defineComponent({
  name: 'BkMonitorEventQuery',
  props: {
    // 资源类型
    kinds: {
      type: [String, Array<string>],
      default: '',
    },
    // 集群ID
    clusterId: {
      type: String,
      default: '',
      required: true,
    },
    // 命名空间
    namespace: {
      type: String,
      default: '',
    },
    // 资源名称
    name: {
      type: [String, Array<string>],
      default: '',
    },
  },
  setup(props) {
    const {
      kinds,
      clusterId,
      namespace,
      name,
    } = toRefs(props);

    // table id
    const tableID = ref('');
    // 快捷时间
    const shortcuts = ref([
      {
        text: $i18n.t('projects.eventQuery.lastHour'),
        value() {
          const end = new Date();
          const start = new Date();
          start.setTime(start.getTime() - 3600 * 1000);
          return [start, end];
        },
      },
      {
        text: $i18n.t('projects.eventQuery.last6Hours'),
        value() {
          const end = new Date();
          const start = new Date();
          start.setTime(start.getTime() - 3600 * 1000 * 6);
          return [start, end];
        },
      },
      {
        text: $i18n.t('projects.eventQuery.last24Hours'),
        value() {
          const end = new Date();
          const start = new Date();
          start.setTime(start.getTime() - 3600 * 1000 * 24);
          return [start, end];
        },
      },
      {
        text: $i18n.t('projects.eventQuery.last3Days'),
        value() {
          const end = new Date();
          const start = new Date();
          start.setTime(start.getTime() - 3600 * 1000 * 24 * 3);
          return [start, end];
        },
      },
      {
        text: $i18n.t('units.time.lastDays'),
        value() {
          const end = new Date();
          const start = new Date();
          start.setTime(start.getTime() - 3600 * 1000 * 24 * 7);
          return [start, end];
        },
      },
    ]);
    // 可选组件列表
    const componentList = [
      'kubelet',
      'default-scheduler',
      'replicaset-controller',
      'deployment-controller',
      'service-controller',
      'endpoint-controller',
      'horizontal-pod-autoscaler',
      'statefulset-controller',
      'daemonset-controller',
      'tke-eni-ipamd',
      'loadbalancer-controller',
      'job-controller',
      'cronjob-controller',
      'bcs-ingress-controller',
      'node-controller',
      'endpoint-slice-controller',
      'persistentvolume-controller',
      'kube-proxy',
      'taint-controller',
      'egress-controller',
      'nginx-ingress-controller',
      'pod-autoscaler',
      'controllermanager',
      'cluster-autoscaler',
      'loadbalance-controller',
      'admission-controller',
      'kube-controller-manager',
      'bcs-gamestatefulset-operator',
      'attachdetach-controller',
      'bcs-cloud-netcontroller',
      'kyverno-admission',
      'volume_expand',
      'cidrAllocator',
      'sspilot/shardedservice-controller.routing.red.tencent.com',
      'bcs-gamedeployment-operator',
      'route_controller',
      'bcs-hook-operator',
      'elasticsearch-controller',
      'eklet',
      'clusternet-hub',
      'oombeat',
      'clusternet-scheduler',
      'clusternet-agent',
      'oom-guard',
      'StatefulSet',
      'sspilot/headlessservice-controller.routing.red.tencent.com',
    ];
    // searchselect数据源
    const filterData = computed(() => [
      {
        name: $i18n.t('projects.eventQuery.module'),
        id: 'component',
        children: componentList.map(item => ({ id: item, name: item })),
      },
      {
        name: $i18n.t('projects.eventQuery.resourceName'),
        id: 'name',
      },
      {
        name: $i18n.t('projects.eventQuery.level'),
        id: 'level',
        children: [{ id: 'Normal', name: 'Normal' }, { id: 'Warning', name: 'Warning' }],
      },
    ].filter(item => !params.value.searchSelect.some(data => data.id === item.id)));
    // 搜索参数
    const params = ref<{
      clusterId: string
      namespace: string
      searchSelect: any[]
      date: Date[]
    }>({
      clusterId: clusterId.value,
      namespace: namespace.value,
      searchSelect: [],
      date: [],
    });
    // 解析搜索数据格式成监控需要的格式
    const parseToBkMonitorSearchValue = computed(() => {
      const where: Array<ISearchCondition> = [
        {
          key: 'bcs_cluster_id',
          condition: 'and',
          method: 'eq',
          value: [clusterId.value],
        },
        {
          key: 'kind',
          condition: 'and',
          method: 'eq',
          value: Array.isArray(kinds.value) ? kinds.value : [kinds.value],
        },
        {
          key: 'namespace',
          condition: 'and',
          method: 'eq',
          value: [namespace.value],
        },
        {
          key: 'name',
          condition: 'and',
          method: 'eq',
          value: Array.isArray(name.value) ? name.value : [name.value],
        },
      ];
      params.value.searchSelect.forEach((item) => {
        const valueIds = item.values?.map(v => v.id);
        if (item.id === 'level') {
          where.push({
            key: 'type',
            method: 'eq',
            condition: 'and',
            value: [valueIds?.[0]],
          });
        } else if (item.id === 'component') {
          where.push({
            key: 'target',
            condition: 'and',
            method: 'eq',
            value: [valueIds?.[0]],
          });
        } else if (item.id === 'name') {
          // 在给定name子集中搜索
          const data = where.find(item => item.key === 'name');
          if (data) {
            data.value = valueIds;
          }
        }
      });
      return [{
        data: {
          query_configs: [{
            data_source_label: 'custom',
            data_type_label: 'event',
            result_table_id: tableID.value,
            where,
          }],
        },
      }];
    });

    // 拼接监控事件查询参数
    const { curProject } = useProject();
    const columns = [
      {
        id: 'dimensions.name',
        name: $i18n.t('projects.eventQuery.resourceName'),
      },
      {
        id: 'dimensions.type',
        name: $i18n.t('projects.eventQuery.level'),
      },
    ];
    const hideFeatures = [
      /** 收藏 */
      'favorite',
      /** 数据ID */
      'dataId',
      /** 时间范围 */
      // 'dateRange',
      /** 维度筛选 */
      // 'dimensionFilter',
      /** 标题 */
      'title',
      /** 表头 */
      // 'header',
    ];
    function encodeURL(data) {
      return encodeURIComponent(JSON.stringify(data));
    }
    const bkMonitorEventSourceURL = computed(() => {
      const [from, to] = params.value.date;
      const fromTimestamp = from ? new Date(from)?.getTime() : '';
      const toTimestamp = to ? new Date(to)?.getTime() : '';
      return `${window.BKMONITOR_HOST}/?space_uid=bkci__${curProject.value.projectCode}&bizId=${curProject.value.businessID}&needMenu=false&onlyShowView=true#/event-explore?targets=${encodeURL(parseToBkMonitorSearchValue.value)}&from=${fromTimestamp}&to=${toTimestamp}&timezone=Asia/Shanghai&type=event&columns=${encodeURL(columns)}&hideFeatures=${encodeURL(hideFeatures)}&prop=time&order=descending`;
    });

    // 获取表名称
    async function handleGetTableID() {
      const data = await getEventDataID({ $clusterId: clusterId.value }).catch(() => ({
        result_table_id: '',
        bk_data_id: '',
        data_name: '',
        vm_result_table_id: '',
      }));
      tableID.value = data?.result_table_id;
    }

    const refreshIframe = ref<string>('');

    watch(bkMonitorEventSourceURL, () => {
      refreshIframe.value = new Date().getTime()
        .toString();
    });

    watch(clusterId, () => {
      if (clusterId.value) {
        handleGetTableID();
      }
    }, { immediate: true });

    onBeforeMount(() => {
      const end = new Date();
      const start = new Date();
      start.setTime(start.getTime() - 3600 * 1000);
      params.value.date = [
        start,
        end,
      ];
    });

    return {
      refreshIframe,
      filterData,
      params,
      shortcuts,
      bkMonitorEventSourceURL,
    };
  },
});
</script>
