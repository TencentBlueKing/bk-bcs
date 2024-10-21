<template>
  <bk-dialog
    ref="dialog"
    ext-cls="confirm-dialog"
    footer-align="center"
    dialog-type="operation"
    :is-show="show"
    :is-loading="btnLoading"
    :cancel-text="t('取消')"
    :confirm-text="t('驳回')"
    :close-icon="true"
    :show-mask="true"
    :quick-close="false"
    :multi-instance="false"
    @confirm="handleConfirm"
    @closed="handleClose">
    <template #header>
      <div class="tip-icon__wrap">
        <exclamation-circle-shape class="tip-icon" />
      </div>
      <div class="headline">{{ t('确认驳回该上线任务') }}?</div>
    </template>
    <ul class="content-info">
      <li class="content-info__li">
        <span class="content-info__hd"> {{ t('待上线版本') }}： </span>
        <span class="content-info__bd"> {{ releaseName || '--' }} </span>
      </li>
    </ul>
    <div>
      <div class="textarea-title is-required">{{ t('驳回理由') }}</div>
      <bk-input
        v-model="reason"
        class="textarea-content"
        :maxlength="100"
        :over-max-length-limit="false"
        :rows="2"
        type="textarea" />
    </div>
  </bk-dialog>
</template>

<script setup lang="ts">
  import { ref } from 'vue';
  import { approve } from '../../../../api/record';
  import BkMessage from 'bkui-vue/lib/message';
  import { ExclamationCircleShape } from 'bkui-vue/lib/icon';
  import { useI18n } from 'vue-i18n';
  import { APPROVE_STATUS } from '../../../../constants/record';

  const props = defineProps<{
    show: boolean;
    spaceId: string;
    appId: number;
    releaseId: number;
    releaseName: string;
  }>();

  const emits = defineEmits(['update:show', 'reject']);

  const { t } = useI18n();

  const btnLoading = ref(false);
  const reason = ref('');

  const handleClose = () => {
    emits('update:show', false);
  };
  const handleConfirm = async () => {
    if (!reason.value) {
      BkMessage({
        theme: 'error',
        message: t('请输入驳回理由'),
      });
      return;
    }
    btnLoading.value = true;
    try {
      await approve(props.spaceId, props.appId, props.releaseId, {
        publish_status: APPROVE_STATUS.RejectedApproval,
        reason: reason.value,
      });
      BkMessage({
        theme: 'success',
        message: t('操作成功'),
      });
      emits('reject');
    } catch (e) {
      console.log(e);
    } finally {
      btnLoading.value = false;
    }
  };
</script>

<style lang="scss" scoped>
  :deep(.confirm-dialog) {
    .bk-modal-body {
      padding-bottom: 0;
    }
    .bk-modal-content {
      padding: 0 32px;
      height: auto;
      max-height: none;
      min-height: auto;
      border-radius: 2px;
    }
    .bk-modal-footer {
      position: relative;
      padding: 24px 0;
      height: auto;
      border: none;
    }
    .bk-dialog-footer .bk-button {
      min-width: 88px;
    }
  }
  .headline {
    margin-top: 16px;
    text-align: center;
  }
  .tip-icon__wrap {
    margin: 0 auto;
    width: 42px;
    height: 42px;
    position: relative;
    &::after {
      content: '';
      position: absolute;
      z-index: -1;
      top: 50%;
      left: 50%;
      transform: translate3d(-50%, -50%, 0);
      width: 30px;
      height: 30px;
      border-radius: 50%;
      background-color: #ff9c01;
    }
    .tip-icon {
      font-size: 42px;
      line-height: 42px;
      vertical-align: middle;
      color: #ffe8c3;
    }
  }
  .content-info {
    margin-top: 4px;
    padding: 13px 21px;
    font-size: 14px;
    line-height: 22px;
    background-color: #f5f6fa;
    &__li {
      display: flex;
      justify-content: flex-start;
      align-items: flex-start;
      & + .content-info__li {
        margin-top: 18px;
      }
    }
    &__hd {
      text-align: left;
      color: #63656e;
    }
    &__bd {
      flex: 1;
      word-wrap: break-word;
      word-break: break-all;
      color: #313238;
    }
  }
  .textarea-title {
    position: relative;
    margin-top: 24px;
    display: inline-block;
    vertical-align: middle;
    font-size: 14px;
    color: #333;
    &.is-required {
      padding-right: 14px;
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
  .textarea-content {
    margin-top: 7px;
  }
</style>
