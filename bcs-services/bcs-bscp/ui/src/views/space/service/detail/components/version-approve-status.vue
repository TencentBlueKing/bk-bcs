<template>
  <div class="version-approve-status" v-if="approverList && route.params.versionId">
    <Spinner v-show="approveStatus === 0" class="spinner" />
    <div
      v-show="approveStatus !== 0"
      :class="['dot', { online: approveStatus === 1, offline: approveStatus === 2 }]"></div>
    <span class="approve-status-text">{{ approveText }}</span>
    <text-file
      v-show="approveStatus > -1"
      v-bk-tooltips="{
        content: `${approveStatus === 3 ? t('撤销人') : t('审批人')}：${approverList}`,
        placement: 'bottom',
      }"
      class="text-file" />
  </div>
</template>

<script setup lang="ts">
  import { ref, watch, onMounted } from 'vue';
  import { Spinner, TextFile } from 'bkui-vue/lib/icon';
  import { storeToRefs } from 'pinia';
  import useConfigStore from '../../../../../store/config';
  import { useRoute } from 'vue-router';
  import { useI18n } from 'vue-i18n';
  import { versionStatusQuery } from '../../../../../api/config';
  import { APPROVE_TYPE } from '../../../../../constants/config';

  const configStore = useConfigStore();
  const { publishedVersionId } = storeToRefs(configStore);

  const emits = defineEmits(['sendData']);

  const route = useRoute();
  const { t } = useI18n();

  const approverList = ref(''); // 审批人
  const approveStatus = ref(-1); // 审批图标状态展示
  const approveText = ref(''); // 审批文案

  watch([() => route.params, publishedVersionId], () => {
    loadStatus();
  });

  onMounted(() => {
    loadStatus();
  });

  const loadStatus = async () => {
    if (route.params.versionId) {
      const { spaceId, appId, versionId } = route.params;
      try {
        const resp = await versionStatusQuery(String(spaceId), Number(appId), Number(versionId));
        const { spec } = resp.data;
        approverList.value = spec.approver_progress; // 审批人
        approveText.value = publishStatusText(spec.publish_status);
        // console.log(resp.data, 'dadada');
        sendData(resp.data);
      } catch (error) {
        console.log(error);
      }
    }
  };

  const publishStatusText = (type: string) => {
    switch (type) {
      case 'PendApproval':
        approveStatus.value = APPROVE_TYPE.PendApproval;
        return t('待审批');
      case 'RejectedApproval':
        approveStatus.value = APPROVE_TYPE.Rejected;
        return t('审批驳回');
      case 'RevokedPublish':
        approveStatus.value = APPROVE_TYPE.Revoke;
        return t('撤销上线');
      case 'PendPublish':
        approveStatus.value = APPROVE_TYPE.PendPublish;
        return t('审批通过');
      case 'AlreadyPublish':
      default:
        approveStatus.value = -1;
        return '';
    }
  };

  const sendData = (data: any) => {
    const { spec, revision } = data;
    const approveData = {
      status: spec.publish_status,
      time: spec.publish_time,
      type: spec.publish_type,
    };
    emits('sendData', approveData, revision?.creator || '');
  };

  defineExpose({
    loadStatus,
  });
</script>

<style lang="scss" scoped>
  .version-approve-status {
    margin: 0 16px 0 8px;
    display: flex;
    justify-content: center;
    align-items: center;
    .spinner {
      margin-right: 8px;
    }
    .text-file {
      font-size: 14px;
      color: #63656e;
    }
  }
  .approve-status-text {
    margin-right: 8px;
    font-size: 12px;
    color: #63656e;
  }
  .dot {
    margin-right: 8px;
    width: 13px;
    height: 13px;
    border-radius: 50%;
    &.online {
      border: 3px solid #e0f5e7;
      background-color: #3fc06d;
    }
    &.offline {
      border: 3px solid #eeeef0;
      background-color: #979ba5;
    }
  }
</style>
