<template>
  <bk-dialog
    ref="dialog"
    ext-cls="confirm-dialog"
    footer-align="center"
    dialog-type="operation"
    :is-loading="btnLoading"
    :cancel-text="t('取消')"
    :confirm-text="dialogType === 'publish' ? t('上线') : t('撤销')"
    :is-show="show"
    :close-icon="true"
    :show-mask="true"
    :quick-close="false"
    :multi-instance="false"
    @confirm="handleConfirm"
    @closed="handleClose">
    <template #header>
      <div class="headline">
        {{ dialogType === 'publish' ? `${t('确认上线该版本')}?` : `${t('确认撤销该上线任务')}？` }}
      </div>
    </template>
    <ul :class="['content-info', { 'li-en': locale !== 'zh-cn' }]">
      <li class="content-info__li">
        <span class="content-info__hd"> {{ $t('服务') }}： </span>
        <span class="content-info__bd"> {{ data.service || '--' }} </span>
      </li>
      <li class="content-info__li">
        <span class="content-info__hd"> {{ t('待上线版本') }}： </span>
        <span class="content-info__bd"> {{ data.version || '--' }} </span>
      </li>
      <li class="content-info__li">
        <span class="content-info__hd"> {{ t('上线范围') }}： </span>
        <span class="content-info__bd"> {{ data.group || '--' }} </span>
      </li>
    </ul>
    <div v-if="dialogType !== 'publish'">
      <div class="textarea-title">{{ t('说明') }}</div>
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
  import { computed, ref } from 'vue';
  import { approve } from '../../../../api/record';
  import { IDialogData } from '../../../../../types/record';
  import { APPROVE_STATUS } from '../../../../constants/record';
  import BkMessage from 'bkui-vue/lib/message';
  import { useI18n } from 'vue-i18n';

  const { t, locale } = useI18n();

  const props = defineProps<{
    show: boolean;
    spaceId: string;
    appId: number;
    releaseId: number;
    dialogType: string;
    data: IDialogData;
  }>();

  const emits = defineEmits(['update:show', 'refreshList']);

  const btnLoading = ref(false);
  const reason = ref('');

  const submitType = computed(() => {
    return props.dialogType === 'publish' ? APPROVE_STATUS.AlreadyPublish : APPROVE_STATUS.RevokedPublish;
  });

  const handleClose = () => {
    emits('update:show', false);
  };
  const handleConfirm = async () => {
    btnLoading.value = true;
    try {
      await approve(props.spaceId, props.appId, props.releaseId, {
        publish_status: submitType.value,
        reason: reason.value,
      });
      emits('update:show', false);
      emits('refreshList');
      BkMessage({
        theme: 'success',
        message: t('操作成功'),
      });
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
  .content-info {
    margin-top: 4px;
    padding: 21px 12px;
    font-size: 14px;
    line-height: 22px;
    background-color: #f5f6fa;
    &.li-en .content-info__hd {
      width: 130px;
    }
    &__li {
      display: flex;
      justify-content: flex-start;
      align-items: flex-start;
      & + .content-info__li {
        margin-top: 18px;
      }
    }
    &__hd {
      width: 96px;
      text-align: right;
      color: #63656e;
    }
    &__bd {
      padding-left: 8px;
      flex: 1;
      word-wrap: break-word;
      word-break: break-all;
      color: #313238;
    }
  }
  .textarea-title {
    margin-top: 24px;
    font-size: 14px;
    color: #333;
  }
  .textarea-content {
    margin-top: 7px;
  }
</style>
