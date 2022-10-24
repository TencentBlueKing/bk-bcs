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
        <label>{{ $t('创建时间') }}</label>
        <span>{{ extData.createTime }}</span>
      </div>
      <div class="basic-info-item">
        <label>{{ $t('存在时间') }}</label>
        <span>{{ extData.age }}</span>
      </div>
    </div>
    <!-- 配置、标签、注解 -->
    <bcs-tab class="mt20" type="card" :label-height="42">
      <bcs-tab-panel name="config" :label="$t('配置')">
        <bcs-collapse v-model="activeCollapseName">
          <bcs-collapse-item
            v-for="(item, index) in (data.subsets || [])"
            :name="String(index)"
            :key="index">
            {{ `subset ${index + 1}` }}
            <template #content>
              <!-- Addresses and notReadyAddresses -->
              <p class="detail-title">Addresses</p>
              <bk-table
                :data="item.addresses
                  .map(item => ({
                    ...item,
                    status: 'normal'
                  }))
                  .concat(item.notReadyAddresses || [])"
                class="mb20">
                <bk-table-column label="IP" prop="ip" width="140"></bk-table-column>
                <bk-table-column label="NodeName" prop="nodeName"></bk-table-column>
                <bk-table-column label="TargetRef">
                  <template #default="{ row }">
                    <span>{{ row.targetRef ? `${row.targetRef.kind}:${row.targetRef.name}` : '--' }}</span>
                  </template>
                </bk-table-column>
                <bk-table-column label="Status">
                  <template #default="{ row }">
                    {{row.status === 'normal' ? $t('正常') : $t('异常')}}
                  </template>
                </bk-table-column>
              </bk-table>
              <!-- Ports -->
              <p class="detail-title">Ports</p>
              <bk-table :data="item.ports">
                <bk-table-column label="Name" prop="name">
                  <template #default="{ row }">
                    {{row.name || '--'}}
                  </template>
                </bk-table-column>
                <bk-table-column label="Protocol" prop="protocol"></bk-table-column>
                <bk-table-column label="Port" prop="port"></bk-table-column>
              </bk-table>
            </template>
          </bcs-collapse-item>
        </bcs-collapse>
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
    </bcs-tab>
  </div>
</template>
<script lang="ts">
import { defineComponent, ref } from '@vue/composition-api';

export default defineComponent({
  name: 'EndpointsDetail',
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
    const activeCollapseName = ref(['0'])
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
      activeCollapseName,
      handleTransformObjToArr,
    };
  },
});
</script>
<style lang="postcss" scoped>
@import './network-detail.css';
</style>
