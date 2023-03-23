<script setup lang="ts">
  import VersionSimpleList from './version-simple-list.vue';
  import VersionTableList from './version-table-list.vue';

  const props = defineProps<{
    versionDetailView: Boolean,
    bkBizId: string,
    appId: number,
  }>()

  const emits = defineEmits(['loaded'])

</script>
<template>
  <div class="version-list-area">
    <VersionTableList v-if="versionDetailView" :bk-biz-id="props.bkBizId" :app-id="props.appId" @loaded="emits('loaded')" />
    <VersionSimpleList v-else :bk-biz-id="props.bkBizId" :app-id="props.appId" @loaded="emits('loaded')" />
  </div>
</template>
<style lang="scss" scoped>
  .version-list-area {
    height: 100%;
  }
</style>