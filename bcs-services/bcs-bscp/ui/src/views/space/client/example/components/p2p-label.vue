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
            content: tipLinkComponent,
            placement: 'top',
          }" />
      </div>
    </div>
  </div>
</template>
<script lang="ts" setup>
  import { ref, h } from 'vue';
  import { Info } from 'bkui-vue/lib/icon';
  import { useI18n } from 'vue-i18n';

  const emits = defineEmits(['send-switcher']);

  const { t } = useI18n();

  const tipLinkComponent = h(
    'div',
    {
      style: { maxWidth: '570px' },
    },
    [
      t('P2P 网络加速依赖于 BCS 集群的元数据，请指定客户端即将部署的 BCS 集群 ID。'),
      h('div', [
        t('更多 P2P 前置依赖信息，请查阅白皮书 '),
        h(
          'a',
          {
            href: 'https://bk.tencent.com/docs/markdown/ZH/BSCP/1.29/UserGuide/Function/p2p_network_acceleration.md',
            target: '_blank',
          },
          t('P2P 网络加速'),
        ),
      ]),
    ],
  );

  const clusterSwitch = ref(false);
  const isShow = ref(false);

  const handleSwitcher = (val: boolean) => {
    clusterSwitch.value = val;
    sendVal(clusterSwitch.value);
  };
  const sendVal = (clusterSwitch: boolean) => {
    emits('send-switcher', clusterSwitch);
  };
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
