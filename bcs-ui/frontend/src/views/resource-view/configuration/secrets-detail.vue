<template>
  <div class="detail p30">
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
      <div class="basic-info-item">
        <label>{{ $t('不可变更') }}</label>
        <span>{{ extData.immutable ? $t('是') : $t('否') }}</span>
      </div>
    </div>
    <bcs-tab class="mt20" type="card" :label-height="42">
      <bcs-tab-panel name="data" label="Data">
        <bk-table
          :data="handleTransformObjToArr(data.data)"
          @row-mouse-enter="handleMouseEnter"
          @row-mouse-leave="handleMouseLeave">
          <bk-table-column label="Key" prop="key"></bk-table-column>
          <bk-table-column label="Value" prop="value"></bk-table-column>
          <bk-table-column label="" width="40">
            <template #default="{ row, $index }">
              <span
                v-bk-tooltips.top="$t('复制')"
                v-show="$index === ativeIndex"
                @click="handleCopyContent(row.value)">
                <i class="bcs-icon bcs-icon-copy"></i>
              </span>
            </template>
          </bk-table-column>
        </bk-table>
      </bcs-tab-panel>
      <bcs-tab-panel name="lebel" :label="$t('标签')">
        <bk-table :data="handleTransformObjToArr(data.metadata.labels)">
          <bk-table-column label="Key" prop="key"></bk-table-column>
          <bk-table-column label="Value" prop="value"></bk-table-column>
        </bk-table>
      </bcs-tab-panel>
      <bcs-tab-panel name="annotation" :label="$t('注解')">
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
import useTableHover from '../../../composables/use-table-hover';
import EventQueryTableVue from '@/views/project-manage/event-query/event-query-table.vue';

export default defineComponent({
  name: 'SecretsDetail',
  components: { EventQueryTableVue },
  props: {
    // 当前行数据
    data: {
      type: Object,
      default: () => ({}),
    },
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

    const {
      ativeIndex,
      handleMouseEnter,
      handleMouseLeave,
      handleCopyContent,
    } = useTableHover();

    return {
      ativeIndex,
      handleTransformObjToArr,
      handleMouseEnter,
      handleMouseLeave,
      handleCopyContent,
    };
  },
});
</script>
<style lang="postcss" scoped>
@import './config-detail.css';
</style>
