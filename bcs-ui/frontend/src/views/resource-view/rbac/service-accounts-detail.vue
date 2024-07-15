<template>
  <div class="detail p30">
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
    <div class="mt20 mb10 scerets-content">
      <bcs-tab :label-height="42">
        <bcs-tab-panel name="secrets" label="Secrets">
          <bk-table :data="data.secrets">
            <bk-table-column label="name" prop="name"></bk-table-column>
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
  </div>
</template>
<script lang="ts">
import { defineComponent } from 'vue';

import EventQueryTable from '@/views/project-manage/event-query/event-query-table.vue';

export default defineComponent({
  name: 'ConfigMapsDetail',
  components: { EventQueryTable },
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
