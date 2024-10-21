<template>
  <div class="more-options">
    <Ellipsis class="ellipsis" />
    <ul class="more-options-ul">
      <li class="more-options-li" @click="handleLinkTo">{{ $t('服务上线记录') }}</li>
      <bk-loading :loading="loading">
        <li
          class="more-options-li"
          v-if="
            [APPROVE_STATUS.PendApproval, APPROVE_STATUS.PendPublish].includes(props.approveStatus as APPROVE_STATUS) &&
            creator === userInfo.username
          "
          @click="handleUndo">
          {{ $t('撤销') }}
        </li>
      </bk-loading>
    </ul>
  </div>
</template>

<script setup lang="ts">
  import { ref } from 'vue';
  import { useRoute, useRouter } from 'vue-router';
  import { storeToRefs } from 'pinia';
  import useUserStore from '../../../../../store/user';
  import { Ellipsis } from 'bkui-vue/lib/icon';
  import { approve } from '../../../../../api/record';
  import BkMessage from 'bkui-vue/lib/message';
  import { APPROVE_STATUS } from '../../../../../constants/config';
  import { useI18n } from 'vue-i18n';

  const { userInfo } = storeToRefs(useUserStore());

  const props = withDefaults(
    defineProps<{
      approveStatus: string;
      creator: string;
    }>(),
    {
      approveStatus: '',
      creator: '',
    },
  );

  const emits = defineEmits(['handleUndo']);

  const route = useRoute();
  const router = useRouter();
  const { t } = useI18n();

  const loading = ref(false);

  // 跳转到服务记录页面
  const handleLinkTo = () => {
    const url = router.resolve({
      name: 'records-app',
      params: {
        appId: route.params.appId,
      },
    }).href;
    window.open(url, '_blank');
  };

  // 撤销审批
  const handleUndo = async () => {
    loading.value = true;
    try {
      const { spaceId: biz_id, appId: app_id, versionId: release_id } = route.params;
      await approve(String(biz_id), Number(app_id), Number(release_id), { publish_status: 'RevokedPublish' });
      BkMessage({
        theme: 'success',
        message: t('成功'),
      });
      emits('handleUndo');
    } catch (error) {
      console.log(error);
    } finally {
      loading.value = false;
    }
  };
</script>

<style lang="scss" scoped>
  .more-options {
    box-sizing: content-box;
    position: relative;
    margin: 0 -12px 0 0;
    width: 32px;
    height: 32px;
    cursor: pointer;
    &:hover {
      .more-options-ul {
        display: block;
      }
      .ellipsis {
        color: #3a84ff;
      }
      &::after {
        background-color: #dcdee5;
      }
    }
    &::after {
      content: '';
      position: absolute;
      left: 50%;
      top: 50%;
      transform: translate(-50%, -50%);
      width: 20px;
      height: 20px;
      border-radius: 50%;
    }
    .ellipsis {
      position: absolute;
      z-index: 1;
      top: 50%;
      left: 50%;
      transform: translate(-50%, -50%) rotate(90deg);
      font-size: 16px;
      font-weight: 700;
      color: #9a9fa9;
    }
  }
  .more-options-ul {
    position: absolute;
    z-index: 1;
    right: 7px;
    top: 32px;
    display: none;
    border: 1px solid #dcdee5;
    border-radius: 2px;
    box-shadow: 0 2px 6px 0 #0000001a;
  }
  .more-options-li {
    padding: 0 12px;
    min-width: 96px;
    line-height: 32px;
    font-size: 12px;
    white-space: nowrap;
    color: #63656e;
    background-color: #fff;
    &:hover {
      background-color: #f5f7fa;
    }
  }
</style>
