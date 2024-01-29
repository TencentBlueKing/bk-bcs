<template>
  <div>
    <bcs-alert type="info" class="mb20" :title="$t('projects.eventQuery.info')"></bcs-alert>
    <div class="flex justify-end mb-[20px]">
      <template v-if="!hideClusterAndNamespace">
        <ClusterSelect
          v-model="params.clusterId"
          cluster-type="all"
          searchable
          @change="handleClusterChange">
        </ClusterSelect>
        <NamespaceSelect
          :cluster-id="params.clusterId"
          :clearable="nsClearable"
          :required="nsRequired"
          :list="namespaceList"
          :loading="namespaceLoading"
          v-model="params.namespace"
          class="w-[180px] ml-[5px]"
          @change="handleInitEventData">
        </NamespaceSelect>
      </template>
      <bcs-date-picker
        :placeholder="$t('generic.placeholder.searchDate')"
        :shortcuts="shortcuts"
        class="ml-[5px] max-w-[320px]"
        type="datetimerange"
        placement="bottom"
        shortcut-close
        transfer
        v-model="params.date"
        @change="handleInitEventData">
      </bcs-date-picker>
      <bcs-search-select
        class="flex-1 ml-[5px] bg-[#fff]"
        clearable
        filter
        :show-condition="false"
        :data="filterData"
        :placeholder="hideClusterAndNamespace ? $t('projects.eventQuery._search') : $t('projects.eventQuery.search')"
        :show-popover-tag-change="false"
        :popover-zindex="9999"
        selected-style="checkbox"
        v-model="params.searchSelect"
        @change="handleInitEventData">
      </bcs-search-select>
    </div>
    <bcs-table
      :data="events"
      :pagination="pagination"
      v-bkloading="{ isLoading: eventLoading }"
      @page-change="handlePageChange"
      @page-limit-change="handlePageLimitChange">
      <bcs-table-column :label="$t('generic.label.time')" prop="eventTime" width="180">
        <template #default="{ row }">
          {{formatDate(row.eventTime)}}
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('projects.eventQuery.module')" prop="component" width="210" show-overflow-tooltip>
        <template #default="{ row }">
          {{ row.component || '--' }}
        </template>
      </bcs-table-column>
      <bcs-table-column
        :label="$t('projects.eventQuery.resourceName')"
        prop="extraInfo.name"
        width="200"
        show-overflow-tooltip>
      </bcs-table-column>
      <bcs-table-column :label="$t('projects.eventQuery.level')" prop="level" width="100"></bcs-table-column>
      <bcs-table-column
        :label="$t('projects.eventQuery.content')"
        prop="describe"
        min-width="100">
        <template #default="{ row }">
          <div class="overflow-auto leading-normal tracking-wide py-[8px]">
            {{ row.describe || '--' }}
          </div>
        </template>
      </bcs-table-column>
      <template #empty>
        <BcsEmptyTableStatus :type="searchEmpty ? 'search-empty' : 'empty'" @clear="handleClearSearchData" />
      </template>
    </bcs-table>
  </div>
</template>

<script lang="ts">
import { computed, defineComponent, onMounted, ref, toRefs, watch } from 'vue';

import { storageEvents } from '@/api/modules/storage';
import { formatDate } from '@/common/util';
import ClusterSelect from '@/components/cluster-selector/cluster-select.vue';
import NamespaceSelect from '@/components/namespace-selector/namespace-select.vue';
import { useCluster } from '@/composables/use-app';
import $i18n from '@/i18n/i18n-setup';
import { useSelectItemsNamespace } from '@/views/resource-view/namespace/use-namespace';

export default defineComponent({
  name: 'EventQuery',
  components: { ClusterSelect, NamespaceSelect },
  props: {
    // 资源类型
    kinds: {
      type: [String, Array],
      default: '',
    },
    // 集群ID
    clusterId: {
      type: String,
      default: '',
    },
    // 命名空间
    namespace: {
      type: String,
      default: '',
    },
    // 资源名称
    name: {
      type: [String, Array],
      default: '',
    },
    // 事件级别
    level: {
      type: String,
      default: '',
    },
    // 组件
    component: {
      type: String,
      default: '',
    },
    // 隐藏集群和namespace选择, eg: Pod、Deployment
    hideClusterAndNamespace: {
      type: Boolean,
      default: false,
    },
    nsClearable: {
      type: Boolean,
      default: false,
    },
    nsRequired: {
      type: Boolean,
      default: true,
    },
  },
  setup(props) {
    const {
      clusterId,
      namespace,
      kinds,
      name,
      level,
      component,
      hideClusterAndNamespace,
      nsRequired,
    } = toRefs(props);

    const { curClusterId, clusterList } = useCluster();
    const { namespaceList, namespaceLoading, getNamespaceData } = useSelectItemsNamespace();
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
    const kindList = [
      'Node',
      'Deployment',
      'ReplicaSet',
      'DaemonSet',
      'StatefulSet',
      'CronJob',
      'Job',
      'Pod',
      'Ingress',
      'Service',
      'Endpoints',
      'PersistentVolumeClaim',
      'PersistentVolume',
      'HorizontalPodAutoscaler',
      'GameStatefulSet',
      'GameDeployment',
      'HookTemplate',
    ];
    const params = ref<{
      clusterId: string
      namespace: string
      searchSelect: any[]
      date: string[]
    }>({
      clusterId: clusterId.value,
      namespace: namespace.value,
      searchSelect: [],
      date: [],
    });
    const searchEmpty = computed(() => !!params.value.searchSelect?.length);
    const parseSearchSelectValue = computed(() => {
      const data = {
        level: '',
        component: '',
        names: [],
        kinds: [],
      };
      params.value.searchSelect.forEach((item) => {
        const valueIds = item.values?.map(v => v.id);
        if (item.id === 'level') {
          data.level = valueIds?.[0] || '';
        } else if (item.id === 'component') {
          data.component = valueIds?.[0] || '';
        } else if (item.id === 'name') {
          data.names = valueIds;
        } else if (item.id === 'kind') {
          data.kinds = valueIds;
        }
      });
      return data;
    });
    const filterData = computed(() => [
      {
        name: $i18n.t('projects.eventQuery.module'),
        id: 'component',
        children: componentList.map(item => ({ id: item, name: item })),
      },
      {
        name: $i18n.t('k8s.kind'),
        id: 'kind',
        multiable: true,
        children: kindList.map(item => ({ id: item, name: item })),
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
    ].filter((item) => {
      // 具体资源时，不展示kind搜索
      if (hideClusterAndNamespace.value && item.id === 'kind') return false;
      return !params.value.searchSelect.some(data => data.id === item.id);
    }));

    // 获取事件信息
    const handleInitEventData = () => {
      events.value = [];
      pagination.value.current = 1;
      handleGetEventList();
    };

    watch(name, (newValue, oldValue) => {
      if (JSON.stringify(newValue) === JSON.stringify(oldValue)) return;
      if (!hideClusterAndNamespace.value) {
        const values = (Array.isArray(name.value) ? name.value : [name.value]).map(item => ({ id: item, name: item }));
        const data = params.value.searchSelect.find(item => item.id === 'name');
        if (data) {
          data.values = values;
        } else {
          params.value.searchSelect.push({
            id: 'name',
            name: $i18n.t('projects.eventQuery.resourceName'),
            values,
          });
        }
      } else {
        // 具体某一个资源时，不展示name过滤信息
        handleInitEventData();
      }
    });

    // cluster change
    const handleClusterChange = async () => {
      eventLoading.value = true;
      await getNamespaceData({ clusterId: params.value.clusterId });
      params.value.namespace = '';
      await handleInitEventData();
      eventLoading.value = false;
    };
    // 事件列表
    const events = ref([]);
    const eventLoading = ref(false);
    const pagination = ref({
      current: 1,
      count: 0,
      limit: 10,
    });
    const handleGetEventList = async () => {
      const clusterId = params.value.clusterId || curClusterId.value;
      if (!clusterId) return;

      const cluster = clusterList.value.find(item => item.clusterID === clusterId);
      if (cluster?.is_shared && !params.value.namespace) return; // 共享集群没有命名空间时，不请求
      eventLoading.value = true;
      const [start, end] = params.value.date;
      const { data = [], total = 0 } = await storageEvents({
        offset: (pagination.value.current - 1) * pagination.value.limit,
        length: pagination.value.limit,
        clusterId,
        env: 'k8s',
        kind: parseSearchSelectValue.value.kinds.join(',') || (Array.isArray(kinds.value) ? kinds.value : [kinds.value]).join(','), // 对象
        'extraInfo.namespace': params.value.namespace, // 命名空间
        'extraInfo.name': parseSearchSelectValue.value.names.join(',') || (Array.isArray(name.value) ? name.value : [name.value]).join(','),
        timeBegin: start ? parseInt(`${new Date(start).getTime() / 1000}`) : '', // 开始时间
        timeEnd: end ? parseInt(`${new Date(end).getTime() / 1000}`) : '', // 结束时间
        level: parseSearchSelectValue.value.level, // 事件级别
        component: parseSearchSelectValue.value.component, // 组件
      }, { needRes: true }).catch(() => ({ data: [], total: 0 }));
      pagination.value.count = total;
      events.value = data || [];
      eventLoading.value = false;
    };
    const handlePageChange = (page) => {
      pagination.value.current = page;
      handleGetEventList();
    };
    const handlePageLimitChange = (limit) => {
      pagination.value.current = 1;
      pagination.value.limit = limit;
      handleGetEventList();
    };

    onMounted(() => {
      if (level.value) {
        params.value.searchSelect.push({
          id: 'level',
          name: $i18n.t('projects.eventQuery.level'),
          values: [{ id: level.value, name: level.value }],
        });
      }
      if (kinds.value && !hideClusterAndNamespace.value) {
        params.value.searchSelect.push({
          id: 'kind',
          name: $i18n.t('k8s.kind'),
          values: (Array.isArray(kinds.value) ? kinds.value : [kinds.value]).map(item => ({ id: item, name: item })),
        });
      }
      if (name.value && !hideClusterAndNamespace.value) {
        params.value.searchSelect.push({
          id: 'name',
          name: $i18n.t('projects.eventQuery.resourceName'),
          values: (Array.isArray(name.value) ? name.value : [name.value]).map(item => ({ id: item, name: item })),
        });
      }
      if (component.value) {
        params.value.searchSelect.push({
          id: 'component',
          name: $i18n.t('projects.eventQuery.module'),
          values: [{ id: component.value, name: component.value }],
        });
      }
    });

    const handleClearSearchData = () => {
      params.value.searchSelect = [];
    };

    onMounted(async () => {
      eventLoading.value = true;
      await getNamespaceData({ clusterId: params.value.clusterId || curClusterId.value }, nsRequired.value);
      await handleGetEventList();
      eventLoading.value = false;
    });

    return {
      kindList,
      filterData,
      params,
      events,
      pagination,
      eventLoading,
      shortcuts,
      searchEmpty,
      namespaceList,
      namespaceLoading,
      handleClearSearchData,
      formatDate,
      handlePageChange,
      handlePageLimitChange,
      handleClusterChange,
      handleInitEventData,
    };
  },
});
</script>
