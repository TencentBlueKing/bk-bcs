<script setup lang="ts">
  import { ref } from 'vue'
  import VersionSimpleList from './version-simple-list.vue';
  import VersionTableList from './version-table-list.vue';

  const props = defineProps<{
    versionDetailView: Boolean,
    bkBizId: string,
    appId: number,
  }>()

  const simpleListRef = ref()

  const emits = defineEmits(['loaded'])

  const refreshVersionList = () => {
    simpleListRef.value.getVersionList()
  }

  defineExpose({
    refreshVersionList
  })

</script>
<template>
  <div class="version-list-area">
    <VersionTableList v-if="versionDetailView" :bk-biz-id="props.bkBizId" :app-id="props.appId" @loaded="emits('loaded')" />
    <VersionSimpleList v-else ref="simpleListRef" :bk-biz-id="props.bkBizId" :app-id="props.appId" @loaded="emits('loaded')" />
  </div>
</template>
<style lang="scss" scoped>
  .version-list-area {
    height: 100%;
  }
</style>