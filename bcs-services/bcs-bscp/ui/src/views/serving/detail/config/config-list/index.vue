<script setup lang="ts">
  import ConfigSimpleList from './config-simple-list.vue';
  import ConfigTableList from './config-table-list/index.vue'

  const props = defineProps<{
    versionDetailView: boolean,
    bkBizId: string,
    appId: number,
  }>()

  const emits = defineEmits(['updateVersionList'])

</script>
<template>
  <div class="config-list-area">
    <ConfigSimpleList v-if="versionDetailView" :bk-biz-id="props.bkBizId" :app-id="props.appId" />
    <ConfigTableList v-else :bk-biz-id="props.bkBizId" :app-id="props.appId" @updateVersionList="emits('updateVersionList')" />
  </div>
</template>
<style lang="scss" scoped>
  .config-list-area {
    height: 100%;
  }
</style>