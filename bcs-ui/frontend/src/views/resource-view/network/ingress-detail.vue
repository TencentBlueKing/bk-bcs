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
        <label>Hosts</label>
        <span>{{ extData.hosts.join(',') || '*' }}</span>
      </div>
      <div class="basic-info-item">
        <label>Address</label>
        <span>{{ extData.addresses.join(',') || '--' }}</span>
      </div>
      <div class="basic-info-item">
        <label>Port(s)</label>
        <span>{{ extData.defaultPorts || '--' }}</span>
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
        <label>{{ $t('dashboard.network.controller') }}</label>
        <span>{{ extData.controller || '--' }}</span>
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
        <label>{{ $t('dashboard.network.privateSubnetID') }}</label>
        <span>{{ extData.subNetID || '--' }}</span>
      </div>
      <div class="basic-info-item">
        <label>{{ $t('dashboard.network.enableAutoRedirect') }}</label>
        <span>{{ extData.autoRewrite ? $t('units.boolean.true') : $t('units.boolean.false') }}</span>
      </div>
    </div>
    <!-- 配置、标签、注解 -->
    <bcs-tab class="mt20" type="card" :label-height="42">
      <bcs-tab-panel name="rules" :label="$t('generic.label.rule')">
        <bk-table :data="extData.rules">
          <bk-table-column label="Host" prop="host"></bk-table-column>
          <bk-table-column label="Path" prop="path"></bk-table-column>
          <bk-table-column label="ServiceName" prop="serviceName"></bk-table-column>
          <bk-table-column label="ServicePort" prop="port"></bk-table-column>
        </bk-table>
      </bcs-tab-panel>
      <bcs-tab-panel name="tls" :label="$t('dashboard.network.certificate')">
        <bk-table :data="data.spec.tls" class="mb20">
          <bk-table-column label="Hosts" prop="hosts">
            <template #default="{ row }">
              {{ (row.hosts || []).join(', ') || '--' }}
            </template>
          </bk-table-column>
          <bk-table-column label="SecretName" prop="secretName"></bk-table-column>
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
          :cluster-id="extData.clusterID" />
      </bcs-tab-panel>
    </bcs-tab>
  </div>
</template>
<script lang="ts">
import { defineComponent } from 'vue';

import EventQueryTable from '@/views/project-manage/event-query/event-query-table.vue';

export default defineComponent({
  name: 'IngressDetail',
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
  },
  setup() {
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

    return {
      handleTransformObjToArr,
    };
  },
});
</script>
<style lang="postcss" scoped>
@import './network-detail.css';
</style>
