<template>
  <section class="client-statistics-page">
    <div class="header">
      <ClientHeader title="客户端统计" />
    </div>
    <div class="management-data-container">
      <VersionRelease :bk-biz-id="bkBizId" :app-id="appId" />
      <PullMass :bk-biz-id="bkBizId" :app-id="appId" />
      <ClientLabel :bk-biz-id="bkBizId" :app-id="appId" />
      <ComponentInfo :bk-biz-id="bkBizId" :app-id="appId" />
    </div>
  </section>
</template>

<script lang="ts" setup>
  import { ref, watch } from 'vue';
  import { useRoute } from 'vue-router';
  import ClientHeader from '../components/client-header.vue';
  import VersionRelease from './section/version-release/index.vue';
  import PullMass from './section/pull-mass/index.vue';
  import ClientLabel from './section/client-label/index.vue';
  import ComponentInfo from './section/component-info/index.vue';

  const route = useRoute();
  const bkBizId = ref(String(route.params.spaceId));
  const appId = ref(Number(route.params.appId));
  watch(
    () => route.params.appId,
    (val) => {
      if (val) {
        appId.value = Number(val);
        bkBizId.value = String(route.params.spaceId);
      }
    },
  );
</script>

<style scoped lang="scss">
  .client-statistics-page {
    padding: 0 24px;
    height: 100%;
    background: #f5f7fa;
    .header {
      padding: 24px 0;
    }
  }
</style>

<style>
  .exception-icon {
    font-size: 49px;
    color: #dee0e3;
  }
</style>
