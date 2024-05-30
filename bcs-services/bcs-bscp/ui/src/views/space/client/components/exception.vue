<template>
  <bk-exception class="exception-wrap-item" type="empty" :description="$t('暂无服务')">
    <div class="exception-actions">
      <bk-button text theme="primary" @click="handleGoToCreateServiceClick">
        {{ $t('前往创建') }}
      </bk-button>
    </div>
  </bk-exception>
</template>

<script lang="ts" setup>
  import { useRouter } from 'vue-router';
  const router = useRouter();
  const handleGoToCreateServiceClick = () => {
    router.push({
      name: 'home',
    });
  };
</script>

<style scoped lang="scss">
  .exception-wrap-item {
    :deep(.bk-exception-img) {
      width: 440px;
    }
    :deep(.bk-exception-description) {
      margin-top: 8px;
      font-size: 24px;
      color: #63656e;
    }
    :deep(.bk-exception-footer){
      margin-top: 16px;
    }

  }
</style>
