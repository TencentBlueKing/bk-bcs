<template>
  <bcs-exception type="empty" scene="page">
    <div class="flex justify-center text-[24px]">{{ $t('暂无集群') }}</div>
    <div class="mt-[16px] text-[14px] text-[#979BA5]">{{ $t('推荐使用共享集群。如有需要也可创建新的集群') }}</div>
    <div class="flex justify-center mt-[24px]">
      <bk-button class="min-w-[116px]" theme="primary" @click="handleGotoResourceView">{{ $t('使用共享集群') }}</bk-button>
      <bk-button class="min-w-[116px]" @click="handleCreateCluster">{{ $t('新建集群') }}</bk-button>
    </div>
    <div class="flex justify-center mt-[16px]">
      <bk-button text @click="handleGotoDoc">
        <span class="relative top-[-1px]">
          <i class="bcs-icon bcs-icon-question-circle"></i>
        </span>
        {{ $t('游戏接入指南') }}
      </bk-button>
    </div>
  </bcs-exception>
</template>
<script lang="ts">
import { defineComponent } from 'vue';
import $router from '@/router';

export default defineComponent({
  name: 'ClusterGuide',
  setup() {
    const handleGotoResourceView = () => {
      $router.push({ name: 'dashboardNamespace' });
    };
    const handleCreateCluster = () => {
      $router.push({ name: 'clusterCreate' });
    };
    const handleGotoDoc = () => {
      window.open(window.BCS_CONFIG?.help);
    };

    return {
      handleGotoResourceView,
      handleCreateCluster,
      handleGotoDoc,
    };
  },
});
</script>
