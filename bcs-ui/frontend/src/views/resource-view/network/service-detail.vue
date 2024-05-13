<template>
  <div class="detail p30">
    <!-- 基础信息 -->
    <div class="detail-title">
      {{ $t('generic.title.basicInfo') }}
    </div>
    <div class="detail-content basic-info">
      <div class="basic-info-item">
        <label>{{ $t('k8s.namespace') }}</label>
        <span>{{ data.metadata.namespace }}</span>
      </div>
      <div class="basic-info-item">
        <label>UID</label>
        <span class="bcs-ellipsis">{{ data.metadata.uid }}</span>
      </div>
      <div class="basic-info-item">
        <label>Type</label>
        <span>{{ data.spec.type }}</span>
      </div>
      <div class="basic-info-item">
        <label>ClusterIPv4</label>
        <span>{{ extData.clusterIPv4 || '--' }}</span>
      </div>
      <div class="basic-info-item">
        <label>ClusterIPv6</label>
        <span>{{ extData.clusterIPv6 || '--' }}</span>
      </div>
      <div class="basic-info-item">
        <label>External-ip</label>
        <span>{{ extData.externalIP.join(',') || '--' }}</span>
      </div>
      <div class="basic-info-item">
        <label>Port(s)</label>
        <span>{{ extData.ports.join(' ') || '--' }}</span>
      </div>
      <div class="basic-info-item">
        <label>Endpoints</label>
        <template v-if="endpoints">
          <bcs-popover class="flex-1" placement="top">
            <span>{{ endpoints.join(',') }}</span>
            <div slot="content">
              <div v-for="(item, index) in endpoints" :key="index">
                {{ item }}
              </div>
            </div>
          </bcs-popover>
        </template>
        <template v-else>
          <span>--</span>
        </template>
      </div>
      <div class="basic-info-item">
        <label>{{ $t('dashboard.network.clbUsage') }}</label>
        <span>{{ extData.clbUseType === 'useExists'
          ? $t('dashboard.network.usingExisting')
          : $t('dashboard.network.autoCreate') }}</span>
      </div>
      <div class="basic-info-item">
        <label>CLB ID</label>
        <span>{{ extData.clbID || '--' }}</span>
      </div>
      <div class="basic-info-item">
        <label>{{$t('dashboard.network.privateSubnetID')}}</label>
        <span>{{ extData.subnetID || '--' }}</span>
      </div>
      <div class="basic-info-item">
        <label>{{ $t('cluster.labels.createdAt') }}</label>
        <span>{{ extData.createTime }}</span>
      </div>
      <div class="basic-info-item">
        <label>{{ $t('k8s.age') }}</label>
        <span>{{ extData.age }}</span>
      </div>
      <div class="basic-info-item">
        <label>{{ $t('dashboard.network.maxSessionTime') }}</label>
        <span>{{ extData.stickyTime ? `${extData.stickyTime} s` : '--' }}</span>
      </div>
    </div>
    <!-- 配置、标签、注解 -->
    <bcs-tab class="mt20" type="card" :label-height="42">
      <bcs-tab-panel name="config" :label="$t('dashboard.network.config')">
        <p class="detail-title">{{ $t('dashboard.network.portmapping') }}（spec.ports）</p>
        <bk-table :data="data.spec.ports">
          <bk-table-column label="Name" prop="name">
            <template #default="{ row }">
              {{row.name || '--'}}
            </template>
          </bk-table-column>
          <bk-table-column label="Port" prop="port"></bk-table-column>
          <bk-table-column label="Protocol" prop="protocol"></bk-table-column>
          <bk-table-column label="TargetPort" prop="targetPort"></bk-table-column>
          <bk-table-column label="NodePort" prop="nodePort">
            <template #default="{ row }">
              {{ row.nodePort || '--' }}
            </template>
          </bk-table-column>
        </bk-table>
      </bcs-tab-panel>
      <bcs-tab-panel name="selector" :label="$t('k8s.selector')">
        <bk-table :data="handleTransformObjToArr(data.spec.selector)">
          <bk-table-column label="Key" prop="key"></bk-table-column>
          <bk-table-column label="Value" prop="value"></bk-table-column>
        </bk-table>
      </bcs-tab-panel>
      <bcs-tab-panel name="label" :label="$t('k8s.label')">
        <bk-table :data="handleTransformObjToArr(data.metadata.labels)">
          <bk-table-column label="Key" prop="key"></bk-table-column>
          <bk-table-column label="Value" prop="value"></bk-table-column>
        </bk-table>
      </bcs-tab-panel>
      <bcs-tab-panel name="annotations" :label="$t('k8s.annotation')">
        <bk-table :data="handleTransformObjToArr(data.metadata.annotations)">
          <bk-table-column label="Key" prop="key"></bk-table-column>
          <bk-table-column label="Value" prop="value"></bk-table-column>
        </bk-table>
      </bcs-tab-panel>
      <bcs-tab-panel name="event" :label="$t('generic.label.event')">
        <EventQueryTable
          hide-cluster-and-namespace
          :kinds="data.kind"
          :namespace="data.metadata.namespace"
          :name="data.metadata.name"
          :cluster-id="clusterId" />
      </bcs-tab-panel>
    </bcs-tab>
  </div>
</template>
<script lang="ts">
import { defineComponent, onMounted, ref } from 'vue';

import $store from '@/store';
import EventQueryTable from '@/views/project-manage/event-query/event-query-table.vue';

export default defineComponent({
  name: 'ServiceDetail',
  components: { EventQueryTable },
  props: {
    // 当前行数据
    data: {
      type: Object,
      default: () => ({}),
    },
    // 当前行对应的manifestExt数据
    extData: {
      type: Object,
      default: () => ({}),
    },
    clusterId: {
      type: String,
      required: true,
    },
  },
  setup(props) {
    const handleTransformObjToArr = (obj) => {
      if (!obj) return [];

      return Object.keys(obj).reduce<any[]>((data, key) => {
        data.push({
          key,
          value: obj[key],
        });
        return data;
      }, []);
    };

    const isLoading = ref(false);
    const endpoints = ref([]);
    const handleGetEndpoints = async () => {
      const flag = await $store.dispatch('dashboard/getNetworksEndpointsFlag', {
        $namespaces: props.data.metadata.namespace,
        $name: props.data.metadata.name,
        $clusterId: props.clusterId,
      });
      if (flag) {
        isLoading.value = true;
        const res = await $store.dispatch('dashboard/getResourceDetail', {
          $namespaceId: props.data.metadata.namespace,
          $type: 'networks',
          $category: 'endpoints',
          $name: props.data.metadata.name,
          $clusterId: props.clusterId,
        });
        endpoints.value = res.data.manifestExt?.endpoints || [];
        isLoading.value = false;
      }
    };

    onMounted(() => {
      handleGetEndpoints();
    });

    return {
      isLoading,
      endpoints,
      handleTransformObjToArr,
    };
  },
});
</script>
<style lang="postcss" scoped>
@import './network-detail.css'
</style>
