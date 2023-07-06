<script setup lang="ts">
  import { onBeforeUnmount } from 'vue'
  import { storeToRefs } from 'pinia'
  import { useConfigStore } from '../../../../../store/config'
  import { AngleDoubleRight } from 'bkui-vue/lib/icon'
  import VersionList from './version-list/index.vue'
  import ConfigList from './config-list/index.vue'

  const store = useConfigStore()
  const { versionDetailView } = storeToRefs(store)

  const props = defineProps<{
    bkBizId: string,
    appId: number
  }>()

  // 切换视图
  const handleToggleView = () => {
    versionDetailView.value = !versionDetailView.value
  }

</script>
<template>
  <section :class="['service-config-wrapper', { 'version-detail-view': versionDetailView}]">
    <section class="version-list-side">
      <VersionList
        :version-detail-view="versionDetailView"
        :bk-biz-id="props.bkBizId"
        :app-id="props.appId"/>
      <div :class="['view-change-trigger', { extend: versionDetailView }]" @click="handleToggleView">
        <AngleDoubleRight class="arrow-icon" />
        <span class="text">版本详情</span>
      </div>
    </section>
    <section class="version-config-content">
      <ConfigList
        :version-detail-view="versionDetailView"
        :bk-biz-id="props.bkBizId"
        :app-id="props.appId" />
    </section>
  </section>
</template>
<style lang="scss" scoped>
  .service-config-wrapper {
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
    &:hover {
      background: #a3c5fd;
    }
    &.extend {
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
