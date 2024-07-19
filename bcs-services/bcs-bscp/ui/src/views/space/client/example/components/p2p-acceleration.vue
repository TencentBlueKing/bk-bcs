<template>
  <div class="p2p-container">
    <!-- 标签 -->
    <div class="p2p-wrap">
      <span class="label-span">{{ $t('启用 P2P 网络加速') }}</span>
      <bk-switcher class="label-switch" :value="clusterSwitch" size="small" theme="primary" @change="handleSwitcher" />
      <bk-popover width="525" :is-show="isShow" trigger="manual" placement="right-start" theme="light">
        <span class="label-span em" @click="isShow = true">{{ $t('查看说明') }}</span>
        <template #content>
          <div class="popover-wrap">
            <div class="popover-block">
              {{
                $t('启用 P2P 网络加速主要适用于业务但配置文件较大及大量节点拉取配置的场景，以实现更优的文件传输速度。')
              }}
            </div>
            <div class="popover-block popover-block-gap">
              {{ $t('以下是启用 P2P 网络加速的基本条件，已确保实现有效的网络加速：') }}
            </div>
            <div class="popover-block">{{ $t('⒈ 单个配置文件的大小应超过 50MB') }}</div>
            <div class="popover-block">{{ $t('⒉ 客户端实例数量应超过 50 个') }}</div>
            <bk-button class="popover-btn" theme="primary" text @click="isShow = false">{{ $t('我知道了') }}</bk-button>
          </div>
        </template>
      </bk-popover>
    </div>
    <div class="p2p-content" v-if="clusterSwitch">
      <div class="label-wrap">
        <span class="label-span required">{{ $t('BCS 集群 ID') }}</span>
        <info
          class="icon-info"
          v-bk-tooltips="{
            content: $t(
              'P2P 网络加速依赖于 BCS 集群的元数据，请指定客户端即将部署的 BCS 集群 ID。\n更多 P2P 前置依赖信息，请查阅白皮书 P2P 网络加速',
            ),
            placement: 'top',
          }" />
      </div>
      <cluster-selector
        ref="clusterSelectorRef"
        class="cluster-selector"
        @current-cluster="(val) => (clusterInfo = val)" />
    </div>
  </div>
</template>
<script lang="ts" setup>
  import { ref, watch } from 'vue';
  import { Info } from 'bkui-vue/lib/icon';
  import ClusterSelector from '../components/cluster-selector.vue';

  interface IclusterInfo {
    name: string;
    value: string;
  }

  const emits = defineEmits(['send-cluster']);

  const clusterSelectorRef = ref();
  const clusterSwitch = ref(false);
  const isShow = ref(false);
  const clusterInfo = ref<IclusterInfo>({
    name: '',
    value: '',
  });

  watch(clusterInfo, () => {
    sendVal({ clusterSwitch: clusterSwitch.value, clusterInfo: clusterInfo.value });
  });

  // 验证是否打开p2p网络加速开关
  const isValid = () => {
    if (clusterSwitch.value && !(clusterInfo.value.name || clusterInfo.value.value)) {
      clusterSelectorRef.value.validateCluster();
      return false;
    }
    return true;
  };

  const handleSwitcher = (val: boolean) => {
    clusterSwitch.value = val;
    // 关闭开关清空值
    if (!val) {
      clusterInfo.value.name = '';
      clusterInfo.value.value = '';
    }
    sendVal({ clusterSwitch: clusterSwitch.value, clusterInfo: clusterInfo.value });
  };
  const sendVal = ({ clusterSwitch, clusterInfo }: { clusterSwitch: boolean; clusterInfo: IclusterInfo }) => {
    emits('send-cluster', { clusterSwitch, clusterInfo });
  };

  defineExpose({
    isValid,
  });
</script>

<style scoped lang="scss">
  .p2p-wrap {
    display: flex;
    justify-content: flex-start;
    align-items: center;
  }
  .label-span {
    font-size: 12px;
    line-height: 20px;
    color: #63656e;
    &.em {
      cursor: pointer;
      color: #3a84ff;
    }
    &.required {
      padding-right: 10px;
      position: relative;
      &::after {
        content: '*';
        position: absolute;
        right: 0;
        top: 50%;
        transform: translateY(-50%);
        font-size: 12px;
        color: #ea3636;
      }
    }
  }
  .label-switch {
    margin: 0 16px 0 12px;
  }
  .p2p-content {
    margin-top: 24px;
  }
  .label-wrap {
    display: flex;
    justify-content: flex-start;
    align-items: center;
  }
  .icon-info {
    margin-left: 13px;
    font-size: 14px;
    color: #979ba5;
    cursor: pointer;
  }
  .cluster-selector {
    margin-top: 8px;
  }
  .label-item {
    position: relative;
    width: 100%;
    display: flex;
    justify-content: flex-start;
    align-items: center;
    .label-item-icon {
      margin: 0 4px;
      font-size: 12px;
    }
    & + .label-item {
      margin-top: 18px;
    }
    .bk-input-wrap {
      flex: 1;
      &.is-error {
        border-color: #ea3636;
        &:focus-within {
          border-color: #3a84ff;
        }
      }
    }
    .error-msg {
      position: absolute;
      left: 0;
      bottom: -14px;
      font-size: 12px;
      line-height: 1;
      white-space: nowrap;
      color: #ea3636;
      animation: form-error-appear-animation 0.15s;
      &.is--value {
        left: 50%;
      }
    }
  }
  @keyframes form-error-appear-animation {
    0% {
      opacity: 0;
      transform: translateY(-30%);
    }
    100% {
      opacity: 1;
      transform: translateY(0);
    }
  }
  .popover-wrap {
    font-size: 14px;
  }
  .popover-block {
    line-height: 22px;
    color: #63656e;
    &-gap {
      margin-top: 20px;
    }
  }
  .popover-btn {
    display: block;
    margin: 8px 0 0 auto;
  }
</style>
