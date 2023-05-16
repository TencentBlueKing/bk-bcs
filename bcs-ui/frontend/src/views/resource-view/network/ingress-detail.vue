<template>
  <div class="detail p30">
    <!-- 基础信息 -->
    <div class="detail-title">
      {{ $t('基础信息') }}
    </div>
    <div class="detail-content basic-info">
      <div class="basic-info-item">
        <label>{{ $t('命名空间') }}</label>
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
        <label>{{ $t('创建时间') }}</label>
        <span>{{ extData.createTime }}</span>
      </div>
      <div class="basic-info-item">
        <label>{{ $t('存在时间') }}</label>
        <span>{{ extData.age }}</span>
      </div>
      <div class="basic-info-item">
        <label>{{ $t('控制器') }}</label>
        <span>{{ extData.controller || '--' }}</span>
      </div>
      <div class="basic-info-item">
        <label>{{ $t('CLB 使用方式') }}</label>
        <span>{{ extData.clbUseType === 'useExists' ? $t('使用已有') : $t('自动创建') }}</span>
      </div>
      <div class="basic-info-item">
        <label>CLB ID</label>
        <span>{{ extData.clbID || '--' }}</span>
      </div>
      <div class="basic-info-item">
        <label>{{ $t('内网子网 ID') }}</label>
        <span>{{ extData.subNetID || '--' }}</span>
      </div>
      <div class="basic-info-item">
        <label>{{ $t('是否开启自动重定向') }}</label>
        <span>{{ extData.autoRewrite ? $t('是') : $t('否') }}</span>
      </div>
    </div>
    <!-- 配置、标签、注解 -->
    <bcs-tab class="mt20" type="card" :label-height="42">
      <bcs-tab-panel name="rules" :label="$t('规则')">
        <bk-table :data="extData.rules">
          <bk-table-column label="Host" prop="host"></bk-table-column>
          <bk-table-column label="Path" prop="path"></bk-table-column>
          <bk-table-column label="ServiceName" prop="serviceName"></bk-table-column>
          <bk-table-column label="ServicePort" prop="port"></bk-table-column>
        </bk-table>
      </bcs-tab-panel>
      <bcs-tab-panel name="tls" :label="$t('证书')">
        <bk-table :data="data.spec.tls" class="mb20">
          <bk-table-column label="Hosts" prop="hosts">
            <template #default="{ row }">
              {{ (row.hosts || []).join(', ') || '--' }}
            </template>
          </bk-table-column>
          <bk-table-column label="SecretName" prop="secretName"></bk-table-column>
        </bk-table>
      </bcs-tab-panel>
      <bcs-tab-panel name="label" :label="$t('标签')">
        <bk-table :data="handleTransformObjToArr(data.metadata.labels)">
          <bk-table-column label="Key" prop="key"></bk-table-column>
          <bk-table-column label="Value" prop="value"></bk-table-column>
        </bk-table>
      </bcs-tab-panel>
      <bcs-tab-panel name="annotations" :label="$t('注解')">
        <bk-table :data="handleTransformObjToArr(data.metadata.annotations)">
          <bk-table-column label="Key" prop="key"></bk-table-column>
          <bk-table-column label="Value" prop="value"></bk-table-column>
        </bk-table>
      </bcs-tab-panel>
      <bcs-tab-panel name="event" :label="$t('事件')">
        <EventQueryTableVue
          hide-cluster-and-namespace
          :kinds="data.kind"
          :namespace="data.metadata.namespace"
          :name="data.metadata.name" />
      </bcs-tab-panel>
    </bcs-tab>
  </div>
</template>
<script lang="ts">
import { defineComponent } from 'vue';
import EventQueryTableVue from '@/views/project-manage/event-query/event-query-table.vue';

export default defineComponent({
  name: 'IngressDetail',
  components: { EventQueryTableVue },
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
