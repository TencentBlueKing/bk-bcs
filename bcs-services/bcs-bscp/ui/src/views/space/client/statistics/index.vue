<template>
  <section class="client-manege-page">
    <div class="header">
      <ClientHeader title="客户端统计" />
    </div>
    <div class="management-data-container">
      <VersionRelease />
      <PullMass />
      <ClientLabel />
      <ComponentInfo />
    </div>
  </section>
</template>

<script lang="ts" setup>
  import ClientHeader from '../components/client-header.vue';
  import VersionRelease from './section/version-release.vue';
  import PullMass from './section/pull-mass/index.vue';
  import ClientLabel from './section/client-label/index.vue';
  import ComponentInfo from './section/component-info/index.vue';
</script>

<style scoped lang="scss">
  .client-manege-page {
    padding: 0 24px;
    height: 100%;
    background: #f5f7fa;
    .header {
      padding-top: 24px;
    }
  }
</style>
