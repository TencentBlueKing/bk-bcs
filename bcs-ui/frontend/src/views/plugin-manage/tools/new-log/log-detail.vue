<template>
  <div class="p30" slot="content">
    <p class="data-title">
      {{$t('基础信息')}}
    </p>
    <div class="biz-metadata-box vertical mb20">
      <div class="data-item">
        <p class="key">{{$t('所属集群')}}：</p>
        <p class="value">{{cluster.clusterName || '--'}}</p>
      </div>
      <div class="data-item">
        <p class="key">{{$t('命名空间')}}：</p>
        <p class="value">{{data.namespace}}</p>
      </div>
      <div class="data-item">
        <p class="key">{{$t('规则名称')}}：</p>
        <p class="value">{{data.name || '--'}}</p>
      </div>
    </div>
    <p class="data-title">
      {{$t('日志源信息')}}
    </p>

    <div class="biz-metadata-box vertical mb0">
      <div class="data-item">
        <p class="key">{{$t('日志源类型')}}：</p>
        <p class="value">{{logSourceTypeMap[data.config_selected]}}</p>
      </div>
      <template v-if="data.config_selected === 'SelectedContainers'">
        <div class="data-item">
          <p class="key">{{$t('应用类型')}}：</p>
          <p class="value">{{data.config.workload.kind || '--'}}</p>
        </div>
        <div class="data-item">
          <p class="key">{{$t('应用名称')}}：</p>
          <p class="value">{{data.config.workload.name || '--'}}</p>
        </div>

        <div class="data-item">
          <p class="key">{{$t('采集路径')}}：</p>
          <div class="value"></div>
        </div>
        <bcs-table :data="data.config.workload.containers">
          <bcs-table-column :label="$t('容器名')" prop="container_name"></bcs-table-column>
          <bcs-table-column :label="$t('标准输出')">
            <template #default="{ row }">
              {{row.enable_stdout ? $t('是') : $t('否')}}
            </template>
          </bcs-table-column>
          <bcs-table-column :label="$t('文件路径')">
            <template #default="{ row }">
              {{row.paths.join(';') || '--'}}
            </template>
          </bcs-table-column>
        </bcs-table>
      </template>
      <template v-else-if="data.config_selected === 'AllContainers'">
        <div class="data-item">
          <p class="key">{{$t('是否采集')}}：</p>
          <p class="value">{{data.config.all_containers.enable_stdout ? $t('是') : $t('否')}}</p>
        </div>
        <div class="data-item">
          <p class="key">{{$t('文件路径')}}：</p>
          <p class="value">{{data.config.all_containers.paths.join(';') || '--'}}</p>
        </div>
      </template>
      <template v-else-if="data.config_selected === 'SelectedLabels'">
        <div class="data-item">
          <p class="key">{{$t('是否采集')}}：</p>
          <p class="value">{{data.config.label_selector.enable_stdout ? $t('是') : $t('否')}}</p>
        </div>
        <div class="data-item">
          <p class="key">{{$t('匹配标签')}}：</p>
          <p class="value">
            <ul class="key-list" v-if="Object.keys(data.config.label_selector.match_labels).length">
              <li v-for="(label, index) of data.config.label_selector.match_labels" :key="index">
                <span class="key f12 m0" style="cursor: default;">{{label.key || '--'}}</span>
                <span class="value f12 m0" style="cursor: default;">{{label.value || '--'}}</span>
              </li>
            </ul>
            <span v-else>--</span>
          </p>
        </div>
        <div class="data-item">
          <p class="key">{{$t('匹配表达式')}}：</p>
          <p class="value">
            <ul class="key-list" v-if="data.config.label_selector.match_expressions.length">
              <li v-for="(expression, index) of data.config.label_selector.match_expressions" :key="index">
                <span class="key f12 m0">{{expression.key || '--'}}</span>
                <span class="value f12 m0">{{expression.operator || '--'}}</span>
                <span class="value f12 m0" v-if="expression.values">{{expression.values || '--'}}</span>
              </li>
            </ul>
            <span v-else>--</span>
          </p>
        </div>
        <div class="data-item">
          <p class="key">{{$t('文件路径')}}：</p>
          <p class="value">{{data.config.label_selector.paths.join(';') || '--'}}</p>
        </div>
      </template>
    </div>
  </div>
</template>
<script lang="ts">
import { defineComponent } from 'vue';

export default defineComponent({
  props: {
    logSourceTypeMap: {
      type: Object,
      default: () => ({}),
    },
    data: {
      type: Object,
      default: () => ({}),
      required: true,
    },
    cluster: {
      type: Object,
      default: () => ({}),
    },
  },
  setup() {

  },
});
</script>
<style lang="postcss" scoped>
.data-title {
  font-weight: normal;
  font-size: 14px;
  color: #313238;
  margin-bottom: 13px;
}

.biz-metadata-box {
  &.vertical {
      display: block;
      border: none;

      .data-item {
          display: flex;
          border: none;
          padding: 0 0 10px 0;

          > .key {
              text-align: right;
              font-size: 14px;
              color: #979BA5;
              display: inline-block;
              margin-bottom: 0;
              min-width: 120px;
          }

          .value {
              color: #63656E;
              font-size: 14px;
              flex: 1;
          }
      }
  }
}
</style>
