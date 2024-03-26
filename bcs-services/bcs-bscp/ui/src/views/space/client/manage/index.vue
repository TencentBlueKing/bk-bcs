<template>
  <section class="client-manege-page">
    <div class="management-data-container">
      <ClientHeader title="客户端统计" />
      <VersionRelease />
    </div>
  </section>
</template>

<script lang="ts" setup>
  import ClientHeader from '../components/client-header.vue';
  import VersionRelease from './section/version-release.vue';
</script>

<style scoped lang="scss">
  .client-manege-page {
    height: 100%;
    background: #f5f7fa;
    .management-data-container {
      padding: 24px;
    }
  }
</style>
