<template>
  <section class="client-statistics-page">
    <div class="header">
      <ClientHeader :title="$t('客户端统计')" />
    </div>
    <div v-if="appId" class="management-data-container">
      <VersionRelease :bk-biz-id="bkBizId" :app-id="appId" />
      <PullMass :bk-biz-id="bkBizId" :app-id="appId" />
      <LabelAndAnnotations :bk-biz-id="bkBizId" :app-id="appId" />
      <ComponentInfo :bk-biz-id="bkBizId" :app-id="appId" />
    </div>
    <Exception v-else />
  </section>
</template>

<script lang="ts" setup>
  import { ref, watch } from 'vue';
  import { useRoute } from 'vue-router';
  import ClientHeader from '../components/client-header.vue';
  import VersionRelease from './section/version-release/index.vue';
  import PullMass from './section/pull-mass/index.vue';
  import LabelAndAnnotations from './section/label-and-annotations/index.vue';
  import ComponentInfo from './section/component-info/index.vue';
  import Exception from '../components/exception.vue';

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
    background: #f5f7fa;
    .header {
      padding: 24px 0;
    }
  }
  .management-data-container > div {
    padding-bottom: 24px;
  }
</style>

<style>
  .exception-icon {
    font-size: 49px;
    color: #dee0e3;
  }
</style>
