<template>
  <bcs-form class="form-content" :label-width="210">
    <bcs-form-item
      v-for="item in list"
      :key="item.prop"
      :label="item.name"
      class="form-content-item"
      :style="{ width: width }"
      desc-icon="bk-icon icon-info-circle"
      :desc="item.desc">
      <template v-if="item.prop === 'status'">
        <LoadingIcon v-if="autoscalerData[item.prop] === 'UPDATING'">
          {{scalerStatusMap[autoscalerData[item.prop]]}}
        </LoadingIcon>
        <StatusIcon
          :status="autoscalerData[item.prop]"
          :status-color-map="scalerColorMap"
          v-else>
          <span
            :class="{ 'error-tips': autoscalerData[item.prop] === 'UPDATE-FAILURE' }"
            v-bk-tooltips="{
              content: autoscalerData.errorMessage,
              disabled: autoscalerData[item.prop] !== 'UPDATE-FAILURE' || !autoscalerData.errorMessage,
              width: 400
            }">
            {{scalerStatusMap[autoscalerData[item.prop]] || $t('generic.status.unknown')}}
          </span>
        </StatusIcon>
        <template v-if="autoscalerData[item.prop] === 'UPDATE-FAILURE'">
          <span
            class="ml10"
            v-bk-tooltips="$t('cluster.nodeTemplate.sops.status.running.detailBtn')"
            @click="handleGotoHelmRelease">
            <i class="bcs-icon bcs-icon-fenxiang"></i>
          </span>
        </template>
      </template>
      <span v-else-if="typeof autoscalerData[item.prop] === 'boolean'">
        {{(item.invert ? !autoscalerData[item.prop] : autoscalerData[item.prop])
          ? $t('units.boolean.true')
          : $t('units.boolean.false')}}
      </span>
      <span v-else>
        {{`${autoscalerData[item.prop]} ${item.unit || ''}`}}
        <span v-if="item.suffix" class="ml10">{{item.suffix}}</span>
      </span>
      <slot name="suffix" :data="item"></slot>
    </bcs-form-item>
  </bcs-form>
</template>
<script lang="ts">
import { defineComponent, PropType } from 'vue';

import LoadingIcon from '@/components/loading-icon.vue';
import StatusIcon from '@/components/status-icon';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';

export default defineComponent({
  name: 'AutoScalerFormItem',
  components: { StatusIcon, LoadingIcon },
  props: {
    list: {
      type: Array as PropType<any[]>,
      default: () => [],
    },
    autoscalerData: {
      type: Object,
      default: () => ({}),
    },
    width: {
      type: String,
      default: '50%',
    },
    clusterId: {
      type: String,
    },
  },
  setup(props) {
    // 获取自动扩缩容配置
    const scalerStatusMap = { // 自动扩缩容状态
      NORMAL: $i18n.t('generic.status.ready'),
      UPDATING: $i18n.t('generic.status.updating'),
      'UPDATE-FAILURE': $i18n.t('generic.status.updateFailed'),
      STOPPED: $i18n.t('generic.status.terminated'),
    };
    const scalerColorMap = {
      NORMAL: 'green',
      UPDATING: 'green',
      'UPDATE-FAILURE': 'red',
      STOPPED: 'gray',
    };
    const handleGotoHelmRelease = () => {
      // todo 记录集群信息
      $router.push({
        name: 'releaseList',
        query: {
          clusterId: props.clusterId,
        },
      });
    };
    return {
      scalerStatusMap,
      scalerColorMap,
      handleGotoHelmRelease,
    };
  },
});
</script>
<style lang="postcss" scoped>
>>> .form-content {
  display: flex;
  flex-wrap: wrap;
  &-item {
      margin-top: 0;
      font-size: 12px;
      width: 100%;
  }
  .bk-label {
      font-size: 12px;
      color: #979BA5;
      text-align: left;
  }
  .bk-form-content {
      font-size: 12px;
      color: #313238;
      display: flex;
      align-items: center;
  }
}
>>> .bcs-icon-fenxiang {
    color: #3a84ff;
    cursor: pointer;
}
>>> .error-tips {
    line-height: 1;
    padding: 2px 0;
    border-bottom: 1px dashed #979BA5;
}
</style>
