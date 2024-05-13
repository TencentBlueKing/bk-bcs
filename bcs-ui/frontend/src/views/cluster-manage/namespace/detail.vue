<template>
  <div class="detail p30">
    <!-- 基础信息 -->
    <div class="detail-title">
      {{ $t('generic.title.basicInfo') }}
    </div>
    <div class="detail-content basic-info">
      <div class="basic-info-item">
        <label>UID</label>
        <span class="bcs-ellipsis">{{ data.uid || '--' }}</span>
      </div>
      <div class="basic-info-item">
        <label>{{ $t('generic.label.status') }}</label>
        <span class="bcs-ellipsis">{{ data.status || '--' }}</span>
      </div>
      <div class="basic-info-item">
        <label>{{ $t('cluster.labels.createdAt') }}</label>
        <span>{{ data.createTime ? timeZoneTransForm(data.createTime, false) : '--' }}</span>
      </div>
    </div>
    <div class="detail-title mt-[20px]">
      {{ $t('logCollector.label.configInfo') }}
    </div>
    <div class="detail-content basic-info">
      <div class="basic-info-item">
        <label>{{ $t('metrics.cpuUsage') }}</label>
        <span class="bcs-ellipsis" v-if="data.quota">
          {{data.cpuUseRate.toFixed(2) * 100 }}%
          （{{`${unitConvert(data.used ? data.used.cpuLimits : '0', '', 'cpu')}${$t('units.suffix.cores')}`}}
          / {{`${unitConvert(data.quota ? data.quota.cpuLimits : '0', '', 'cpu')}${$t('units.suffix.cores')}`}}）
        </span>
        <span class="bcs-ellipsis" v-else>{{ $t('dashboard.ns.tips.notEnabledNamespaceQuota') }}</span>
      </div>
      <div class="basic-info-item">
        <label>{{ $t('metrics.memUsage') }}</label>
        <span class="bcs-ellipsis" v-if="data.quota">
          {{data.memoryUseRate.toFixed(2) * 100}}%
          （{{`${unitConvert(data.used ? data.used.memoryLimits : '0', 'Gi', 'mem')}Gi`}}
          / {{`${unitConvert(data.quota ? data.quota.memoryLimits : '0', 'Gi', 'mem')}Gi`}}）
        </span>
        <span class="bcs-ellipsis" v-else>{{ $t('dashboard.ns.tips.notEnabledNamespaceQuota') }}</span>
      </div>
    </div>
    <!-- 变量、标签、注解 -->
    <bcs-tab class="mt20" type="card" :label-height="42">
      <bcs-tab-panel name="label" :label="$t('k8s.label')">
        <bk-table :data="data.labels">
          <bk-table-column label="Key" prop="key"></bk-table-column>
          <bk-table-column label="Value" prop="value"></bk-table-column>
        </bk-table>
      </bcs-tab-panel>
      <bcs-tab-panel name="annotations" :label="$t('k8s.annotation')">
        <bk-table :data="data.annotations">
          <bk-table-column label="Key" prop="key"></bk-table-column>
          <bk-table-column label="Value" prop="value"></bk-table-column>
        </bk-table>
      </bcs-tab-panel>
      <bcs-tab-panel name="config" :label="$t('generic.label.var')">
        <bk-table :data="data.variables">
          <bk-table-column label="Key" prop="key"></bk-table-column>
          <bk-table-column label="Value" prop="value"></bk-table-column>
        </bk-table>
      </bcs-tab-panel>
    </bcs-tab>
  </div>
</template>
<script lang="ts">
/* eslint-disable camelcase */
import { defineComponent } from 'vue';

import { timeZoneTransForm } from '@/common/util';


export default defineComponent({
  name: 'NamespaceDetail',
  components: {
  },

  props: {
    data: {
      type: Object,
      default: () => {},
    },
  },
  setup() {
    const unitMap = {
      cpu: {
        list: ['m', '', 'k', 'M', 'G', 'T', 'P', 'E'],
        digit: 3,
        base: 10,
      },
      mem: {
        list: ['Ki', 'Mi', 'Gi', 'Ti', 'Pi', 'Ei'],
        digit: 10,
        base: 2,
      },
    };
    const unitConvert = (val, toUnit = '', type: 'cpu' | 'mem') => {
      const { list, digit, base } = unitMap[type];
      const num = val.match(/\d+/gi)?.[0];
      const uint = val.match(/[a-z|A-Z]+/gi)?.[0] || '';

      const index = list.indexOf(uint);
      // 没有单位直接返回
      if (index === -1) return val;

      // 要转换成的单位
      const toUnitIndex = list.indexOf(toUnit) || -1;
      const factorial = index - toUnitIndex;
      if (factorial >= 0) {
        return (num * (base ** (digit * factorial))).toFixed(2);
      }
      return (num / (base ** (Math.abs(digit) * Math.abs(factorial)))).toFixed(2);
    };
    return {
      unitConvert,
      timeZoneTransForm,
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
