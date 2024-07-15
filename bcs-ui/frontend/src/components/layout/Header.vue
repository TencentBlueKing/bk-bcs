<template>
  <div class="content-header">
    <div class="content-header-title">
      <slot>
        <i class="bcs-icon bcs-icon-arrows-left back mr-[4px]" v-if="!hideBack" @click="goBack"></i>
        {{ title }}
        <span class="desc ml-[4px]" v-if="desc">{{ desc }}</span>
        <template v-if="clusterId || namespace">
          <bcs-divider direction="vertical"></bcs-divider>
          <bcs-tag class="!ml-[0px]">
            <template v-if="clusterId && namespace">
              {{ $t('view.labels.clusterAndNs') }}: {{ `${clusterNameMap[clusterId]}/${namespace}` }}
            </template>
            <template v-else-if="clusterId && !namespace">
              {{ $t('generic.label.cluster') }}: {{ clusterNameMap[clusterId] }}
            </template>
            <template v-else-if="!clusterId && namespace">
              {{ $t('k8s.namespace') }}: {{ namespace }}
            </template>
          </bcs-tag>
        </template>
      </slot>
    </div>
    <div class="content-header-doc">
      <slot name="right"></slot>
    </div>
    <div class="content-header-tab">
      <div
        v-for="item in tabs"
        :key="item.name"
        :class="['tab-item',{ active: active === item.name }]"
        @click="handleChangeTab(item)">
        {{ item.displayName }}
      </div>
    </div>
  </div>
</template>
<script lang="ts">
import { defineComponent, PropType } from 'vue';

import { useCluster } from '@/composables/use-app';
import $router from '@/router';

export default defineComponent({
  name: 'ContentHeader',
  props: {
    title: {
      type: String,
      default: '',
    },
    desc: {
      type: String,
      default: '',
    },
    hideBack: {
      type: Boolean,
      default: false,
    },
    tabs: {
      type: Array as PropType<{name: string;displayName: string}[]>,
      default: () => ([]),
    },
    active: {
      type: String,
      default: '',
    },
    clusterId: {
      type: String,
      default: '',
    },
    namespace: {
      type: String,
      default: '',
    },
  },
  emits: ['tab-change'],
  setup(props, ctx) {
    const { clusterNameMap } = useCluster();
    const goBack = () => {
      $router.back();
    };
    const handleChangeTab = (item: {name: string;displayName: string}) => {
      if (props.active === item.name) return;

      ctx.emit('tab-change', item);
    };

    return {
      goBack,
      handleChangeTab,
      clusterNameMap,
    };
  },
});
</script>
<style lang="postcss" scoped>
.content-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    height: 52px;
    border-bottom: 1px solid #dde4eb;
    background: #fff;
    box-shadow: 0 3px 4px 0 rgba(0,0,0,0.04);
    padding: 0 24px;
    position: relative;
    &-title {
        display: flex;
        align-items: center;
        font-size: 16px;
        i {
            cursor: pointer;
            font-weight: 700;
            color: #3a84ff;
        }
        .desc {
            display: inline-flex;
            line-height: 22px;
            background-color: #F0F1F5;
            padding: 0 8px;
            font-size: 12px;
            border-radius: 2px;
        }
    }
    &-tab {
      font-size: 14px;
      position: absolute;
      left: 50%;
      transform: translateX(-50%);
      display: flex;
      bottom: 0;
      .tab-item {
        padding: 10px 16px;
        cursor: pointer;
        &:hover {
          color: #3a84ff;
        }
        &.active {
          color: #3a84ff;
          border-bottom: 2px solid #3a84ff;
        }
      }
    }
    &-doc {
      font-size: 14px;
      flex: 1;
      display: flex;
      align-items: center;
      justify-content: flex-end;
    }
}
</style>
