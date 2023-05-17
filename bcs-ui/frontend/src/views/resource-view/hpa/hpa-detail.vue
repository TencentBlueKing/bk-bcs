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
        <label>Reference</label>
        <span class="bcs-ellipsis">{{ extData.reference }}</span>
      </div>
      <div class="basic-info-item">
        <label>Targets</label>
        <span>{{ extData.targets }}</span>
      </div>
      <div class="basic-info-item">
        <label>MinPods</label>
        <span>{{ extData.minPods }}</span>
      </div>
      <div class="basic-info-item">
        <label>MaxPods</label>
        <span>{{ extData.maxPods }}</span>
      </div>
      <div class="basic-info-item">
        <label>Replicas</label>
        <span>{{ extData.replicas }}</span>
      </div>
      <div class="basic-info-item">
        <label>UID</label>
        <span>{{ data.metadata.uid }}</span>
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
    <!-- 状态、标签、注解 -->
    <bcs-tab class="mt20" :label-height="42">
      <bcs-tab-panel name="conditions" :label="$t('状态（Conditions）')">
        <bk-table :data="data.status.conditions || []">
          <bk-table-column :label="$t('类别')" prop="type"></bk-table-column>
          <bk-table-column :label="$t('状态')" prop="status">
            <template #default="{ row }">
              <StatusIcon :status="row.status"></StatusIcon>
            </template>
          </bk-table-column>
          <bk-table-column :label="$t('最后迁移时间')" prop="lastTransitionTime">
            <template #default="{ row }">
              {{ formatTime(row.lastTransitionTime, 'yyyy-MM-dd hh:mm:ss') }}
            </template>
          </bk-table-column>
          <bk-table-column :label="$t('原因')">
            <template #default="{ row }">
              {{ row.reason || '--' }}
            </template>
          </bk-table-column>
          <bk-table-column :label="$t('消息')">
            <template #default="{ row }">
              {{ row.message || '--' }}
            </template>
          </bk-table-column>
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
import StatusIcon from '../../../components/status-icon';
import { formatTime } from '@/common/util';
import EventQueryTableVue from '@/views/project-manage/event-query/event-query-table.vue';

export default defineComponent({
  name: 'HPADetail',
  components: { StatusIcon, EventQueryTableVue },
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

    return {
      formatTime,
      handleTransformObjToArr,
    };
  },
});
</script>
<style lang="postcss" scoped>
.detail {
    font-size: 14px;
    /deep/ .bk-tab-label-item {
        background-color: #FAFBFD;
        border-bottom: 1px solid #dcdee5;
        line-height: 41px !important;
        height: 41px;
        &.active {
            border-bottom: none;
        }
    }
    /deep/ .bk-tab-label-wrapper {
        overflow: unset !important;
    }
    &-title {
        margin-bottom: 10px;
        color: #313238;
    }
    &-content {
        &.basic-info {
            border: 1px solid #dfe0e5;
            border-radius: 2px;
            .basic-info-item {
                display: flex;
                align-items: center;
                height: 32px;
                padding: 0 15px;
                &:nth-of-type(even) {
                    background: #F7F8FA;
                }
                label {
                    line-height: 32px;
                    border-right: 1px solid #dfe0e5;
                    width: 200px;
                }
                span {
                    padding: 0 15px;
                    flex: 1;
                    overflow: hidden;
                    text-overflow: ellipsis;
                    white-space: normal;
                    word-break: break-all;
                    display: -webkit-box;
                    -webkit-line-clamp: 1;
                    -webkit-box-orient: vertical;
                }
            }
        }
    }
}
</style>
