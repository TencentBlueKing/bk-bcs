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
        <label>{{ $t('cluster.labels.createdAt') }}</label>
        <span>{{ extData.createTime }}</span>
      </div>
      <div class="basic-info-item">
        <label>{{ $t('k8s.age') }}</label>
        <span>{{ extData.age }}</span>
      </div>
    </div>
    <!-- 配置、标签、注解 -->
    <bcs-tab class="mt20" type="card" :label-height="42">
      <bcs-tab-panel name="config" :label="$t('dashboard.network.config')">
        <bcs-collapse v-model="activeCollapseName">
          <bcs-collapse-item
            v-for="(item, index) in (data.subsets || [])"
            :name="String(index)"
            :key="index"
            hide-arrow
            class="mb-[16px]">
            <div class="flex items-center rounded-sm bg-[#f5f7fa] px-3.5 text-[#313238] font-bold">
              <i
                class="bk-icon icon-down-shape !text-base mr-[5px] transition-all"
                :style="{
                  transform: activeCollapseName.includes(String(index)) ? 'rotate(0deg)' : 'rotate(-90deg)',
                }"></i>
              {{ `SubSet ${index + 1}` }}
            </div>
            <template #content>
              <div class="px-[16px] pt-[16px]">
                <!-- Addresses and notReadyAddresses -->
                <p class="detail-title">Addresses</p>
                <bk-table
                  :data="(item.addresses || [])
                    .map(item => ({
                      ...item,
                      status: 'normal'
                    }))
                    .concat(item.notReadyAddresses || [])"
                  class="mb20">
                  <bk-table-column label="IP" prop="ip" width="140"></bk-table-column>
                  <bk-table-column label="NodeName" prop="nodeName">
                    <template #default="{ row }">
                      {{ row.nodeName || '--' }}
                    </template>
                  </bk-table-column>
                  <bk-table-column label="TargetRef">
                    <template #default="{ row }">
                      <span>{{ row.targetRef ? `${row.targetRef.kind}:${row.targetRef.name}` : '--' }}</span>
                    </template>
                  </bk-table-column>
                  <bk-table-column label="Status">
                    <template #default="{ row }">
                      <StatusIcon :status="String(row.status === 'normal')">
                        {{row.status === 'normal' ? $t('generic.status.ready') : $t('generic.status.error')}}
                      </StatusIcon>
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
              </div>
            </template>
          </bcs-collapse-item>
        </bcs-collapse>
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
import { defineComponent, ref, toRefs, watch } from 'vue';

import StatusIcon from '@/components/status-icon';
import EventQueryTable from '@/views/project-manage/event-query/event-query-table.vue';

export default defineComponent({
  name: 'EndpointsDetail',
  components: { StatusIcon, EventQueryTable },
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
  setup(props) {
    const { data } = toRefs(props);
    const activeCollapseName = ref(data.value.subsets?.map((_, index) => String(index)));
    watch(data, () => {
      activeCollapseName.value = data.value.subsets?.map((_, index) => String(index) || []);
    });
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
