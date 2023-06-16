<script setup lang="ts">
  import { ref, watch } from 'vue'
  import { useRoute, useRouter } from 'vue-router'
  import { InfoLine } from 'bkui-vue/lib/icon';
  import { storeToRefs } from 'pinia'
  import { useConfigStore } from '../../../../../store/config'
  import { VERSION_STATUS_MAP } from '../../../../../constants/config'

  const route = useRoute()
  const router = useRouter()

  const { versionData } = storeToRefs(useConfigStore())

  const props = defineProps<{
    versionDetailView: Boolean;
  }>()

  const tabs = ref([
    { name: 'config', label: '配置管理', routeName: 'service-config' },
    { name: 'script', label: '初始化脚本', routeName: 'init-script' },
  ])

  const getDefaultTab = () => {
    const tab = tabs.value.find(item => item.routeName === route.name)
    return tab ? tab.name : 'config'
  }
  const activeTab = ref(getDefaultTab())

  watch(() => route.name, () => {
    activeTab.value = getDefaultTab()
  })

  const handleTabChange = (val: string) => {
    const tab = tabs.value.find(item => item.name === val)
    if (tab) {
      router.push({ name: tab.routeName })
    }
  }

</script>
<template>
  <div class="service-detail-header">
    <section class="summary-wrapper">
      <div :class="['status-tag', versionData.status.publish_status]">{{ VERSION_STATUS_MAP[versionData.status.publish_status as keyof typeof VERSION_STATUS_MAP] || '编辑中' }}</div>
      <div class="version-name" :title="versionData.spec.name">{{ versionData.spec.name }}</div>
      <InfoLine
        v-if="versionData.spec.memo"
        v-bk-tooltips="{
          content: versionData.spec.memo,
          placement: 'bottom-start',
          theme: 'light'
        }"
        class="version-desc" />
    </section>
    <div v-if="!props.versionDetailView" class="detail-header-tabs">
      <BkTab type="unborder-card" v-model:active="activeTab" :label-height="41" @change="handleTabChange">
        <BkTabPanel v-for="tab in tabs" :key="tab.name" :name="tab.name" :label="tab.label"></BkTabPanel>
      </BkTab>
    </div>
  </div>
</template>
<style lang="scss" scoped>
  .service-detail-header {
    position: relative;
    display: flex;
    align-items: center;
    padding: 0 24px;
    height: 41px;
    box-shadow: 0 3px 4px 0 #0000000a;
    z-index: 1;


    .summary-wrapper {
      display: flex;
      align-items: center;
      justify-content: space-between;
      .status-tag {
        margin-right: 8px;
        padding: 0 10px;
        height: 22px;
        line-height: 20px;
        font-size: 12px;
        color: #63656e;
        border: 1px solid rgba(151,155,165,0.30);
        border-radius: 11px;
        &.not_released {
        color: #fe9000;
        background: #ffe8c3;
        border-color: rgba(254, 156, 0, 0.3);
        }
        &.full_released,
        &.partial_released {
          color: #14a568;
          background: #e4faf0;
          border-color: rgba(20, 165, 104, 0.3);
        }
      }
      .version-name {
        max-width: 220px;
        color: #63656e;
        font-size: 14px;
        font-weight: bold;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }
      .version-desc {
        margin-left: 8px;
        font-size: 15px;
        color: #979ba5;
        cursor: pointer;
      }
    }
    .detail-header-tabs {
      position: absolute;
      top: 0;
      left: 50%;
      transform: translateX(-50%);
      :deep(.bk-tab-header) {
        border-bottom: none;
      }
      :deep(.bk-tab-content) {
        display: none;
      }
    }
  }
</style>