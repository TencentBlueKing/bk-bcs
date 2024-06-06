<template>
  <bk-dialog
    class="client-retry-dialog"
    footer-align="center"
    :confirm-text="t('重试')"
    :is-show="isShow"
    :loading="pending"
    :quick-close="false"
    @confirm="handleRetry"
    @closed="close">
    <template #header>
      <div class="dialog-title">
        <ExclamationCircleShape class="warning-icon" />
        <div class="title">{{ title }}</div>
      </div>
    </template>
    <div v-if="props.selections.length > 0" class="dialog-content">
      <bk-table
        v-if="props.isBatch"
        empty-cell-text="--"
        :data="props.selections"
        :border="['outer']"
        max-height="300"
        show-overflow-tooltip>
        <bk-table-column label="UID" field="uid" width="260" />
        <bk-table-column :label="t('当前配置版本')" field="release_name" />
      </bk-table>
      <div v-else class="client-info">
        <div class="client-info-item">
          <div class="label">UID：</div>
          <div class="value">{{ props.selections[0].uid }}</div>
        </div>
        <div class="client-info-item">
          <div class="label">{{ t('当前配置版本') }}：</div>
          <div class="value">{{ props.selections[0].release_name || '--' }}</div>
        </div>
      </div>
    </div>
  </bk-dialog>
</template>
<script lang="ts" setup>
  import { ref, computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import BkMessage from 'bkui-vue/lib/message';
  import { ExclamationCircleShape } from 'bkui-vue/lib/icon';
  import { retryClients } from '../../../../../api/client';

  const { t } = useI18n();

  const props = defineProps<{
    bkBizId: string;
    appId: number;
    isShow: boolean;
    isBatch: boolean;
    selections: { id: number; uid: string; release_name: string }[];
  }>();

  const emits = defineEmits(['close', 'retried']);

  const pending = ref(false);

  const title = computed(() => {
    return props.isBatch ? t('确定批量重试拉取客户端配置？') : t('确定重试拉取客户端配置？');
  });

  const handleRetry = async () => {
    pending.value = true;
    try {
      const ids = props.selections.map((item) => item.id);
      await retryClients(props.bkBizId, props.appId, ids);
      close();
      BkMessage({
        theme: 'success',
        message: t('重试成功'),
      });
      // 弹窗组件关闭使用setTimeout了延时，在外层需要再包一层定时器，否则先触发列表刷新会导致弹窗无法销毁
      setTimeout(() => {
        emits('retried', ids);
      }, 300);
    } catch (e) {
      console.error(e);
    } finally {
      pending.value = false;
    }
  };

  const close = () => {
    emits('close');
  };
</script>
<style lang="scss" scoped>
  .dialog-title {
    text-align: center;
    .warning-icon {
      font-size: 42px;
      color: #ff9c01;
    }
    .title {
      margin-top: 16px;
      color: #313238;
      line-height: 32px;
      font-size: 20px;
    }
  }
  .client-info {
    padding: 16px;
    background: #f5f7fa;
    border-radius: 2px;
    .client-info-item {
      display: flex;
      align-items: flex-start;
      font-size: 14px;
      line-height: 22px;
      &:not(:last-child) {
        margin-bottom: 18px;
      }
      .label {
        flex-shrink: 0;
        width: 100px;
        color: #63656e;
        text-align: right;
      }
      .value {
        color: #313238;
        word-break: break-all;
      }
    }
  }
</style>
<style lang="scss">
  .client-retry-dialog.bk-dialog-wrapper {
    .bk-modal-wrapper {
      .bk-modal-body {
        padding-bottom: 0;
      }
      .bk-modal-content {
        height: auto;
        min-height: unset;
      }
      .bk-modal-footer {
        position: unset;
        padding: 24px;
        height: auto;
        background: #ffffff;
        border-top: none;
        .bk-button {
          min-width: 88px;
        }
      }
    }
  }
</style>
