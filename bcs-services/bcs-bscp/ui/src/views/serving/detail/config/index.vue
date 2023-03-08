<script setup lang="ts">
  import { defineProps, ref, computed, watch } from 'vue'
  import { AngleDoubleRight } from 'bkui-vue/lib/icon'
  import VersionList from './version-list.vue'
  import ConfigDetailTable from './config-detail-table/index.vue'
  import ConfigDetailList from './config-detail-list.vue'
  import VersionDetailTable from './version-detail-table.vue'

  const props = defineProps<{
    bkBizId: string,
    appId: number
  }>()

  const versionListRef = ref()
  const releaseId = ref<number|null>(null) // 当前选中的版本ID
  const versionDetailView = ref(false)

  const updateVersionList = () => {
    versionListRef.value.getVersionList()
  }

  const handleUpdateReleaseId = (id: number) => {
    releaseId.value = id
  }

</script>
<template>
  <section :class="['serving-config-wrapper', { 'version-detail-view': versionDetailView}]">
    <section class="version-list-side">
      <VersionDetailTable v-if="versionDetailView" :bk-biz-id="props.bkBizId" :app-id="props.appId" />
      <VersionList v-else ref="versionListRef" :bk-biz-id="props.bkBizId" :app-id="props.appId" :release-id="releaseId" @updateReleaseId="handleUpdateReleaseId" />
      <div :class="['view-change-trigger', { extend: versionDetailView }]" @click="versionDetailView = !versionDetailView">
        <AngleDoubleRight class="arrow-icon" />
        <span class="text">版本详情</span>
      </div>
    </section>
    <section class="version-config-content">
      <ConfigDetailList :bk-biz-id="props.bkBizId" :app-id="props.appId" v-if="versionDetailView" />
      <ConfigDetailTable v-else ref="configListRef" :bk-biz-id="props.bkBizId" :app-id="props.appId" :release-id="releaseId" @updateVersionList="updateVersionList" />
    </section>
  </section>
</template>
<style lang="scss" scoped>
  .serving-config-wrapper {
    display: flex;
    align-content: center;
    justify-content: space-between;
    height: 100%;
    &.version-detail-view {
      .version-list-side {
        width: calc(100% - 366px);
      }
      .version-config-content {
        width: 366px;
      }
    }
  }
  .version-list-side {
    position: relative;
    width: 280px;
    height: 100%;
    background: #fafbfd;
    box-shadow: 0 2px 2px 0 rgba(0,0,0,0.15);
    z-index: 1;
    transition: width 0.3 ease-in-out;
  }
  .view-change-trigger {
    position: absolute;
    top: 37%;
    right: -16px;
    padding-top: 8px;
    width: 16px;
    color: #ffffff;
    background: #c4c6cc;
    border-radius: 0 4px 4px 0;
    text-align: center;
    cursor: pointer;
    &.extend {
      background: #a3c5fd;
      .arrow-icon {
        transform: rotate(180deg);
      }
    }
    .text {
      display: inline-block;
      margin-top: -8px;
      font-size: 12px;
      transform: scale(0.833);
    }
    .arrow-icon {
      font-size: 14px;
    }
  }
  .version-config-content {
    width: calc(100% - 280px);
    height: 100%;
    background: #ffffff;
    transition: width 0.3 ease-in-out;
  }
</style>
