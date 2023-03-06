<script setup lang="ts">
  import { defineProps, ref, computed, watch } from 'vue'
  import { useStore } from 'vuex'
  import VersionList from './version-list.vue'
  import ConfigList from './config-list/index.vue'

  const props = defineProps<{
    bkBizId: string,
    appId: number
  }>()

  const versionListRef = ref()
  const releaseId = ref<number|null>(null) // 当前选中的版本ID

  const updateVersionList = () => {
    versionListRef.value.getVersionList()
  }

  const handleUpdateReleaseId = (id: number) => {
    releaseId.value = id
  }

</script>
<template>
  <section class="serving-config-wrapper">
    <section class="version-list-side">
      <VersionList ref="versionListRef" :bk-biz-id="props.bkBizId" :app-id="props.appId" :release-id="releaseId" @updateReleaseId="handleUpdateReleaseId" />
    </section>
    <section class="version-config-content">
      <ConfigList ref="configListRef" :bk-biz-id="props.bkBizId" :app-id="props.appId" :release-id="releaseId" @updateVersionList="updateVersionList" />
    </section>
  </section>
</template>
<style lang="scss" scoped>
  .serving-config-wrapper {
    display: flex;
    align-content: center;
    justify-content: space-between;
    height: 100%;
  }
  .version-list-side {
    width: 280px;
    height: 100%;
    background: #fafbfd;
    box-shadow: 0 2px 2px 0 rgba(0,0,0,0.15);
    z-index: 1;
  }
  .version-config-content {
    padding: 0 24px;
    width: calc(100% - 280px);
    height: 100%;
    background: #ffffff;
  }
</style>
