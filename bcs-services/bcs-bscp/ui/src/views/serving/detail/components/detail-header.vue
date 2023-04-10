<script setup lang="ts">
  import { ref } from 'vue'
  import { useRoute, useRouter } from 'vue-router'
  import ServingSelector from './serving-selector.vue'
  import { Help } from "bkui-vue/lib/icon";

  const route = useRoute()
  const router = useRouter()

  const props = defineProps<{
    bkBizId: string,
    appId: number,
  }>()

  const tabs = ref([
    { name: 'config', label: '配置管理', routeName: 'serving-config' },
    // { name: 'group', label: '分组管理', routeName: 'serving-group' },
    // { name: 'client', label: '客户端统计', routeName: 'serving-client' },
  ])

  const getDefaultTab = () => {
    const tab = tabs.value.find(item => item.routeName === route.name)
    return tab ? tab.name : 'config'
  }
  const activeTab = ref(getDefaultTab())

  const handleTabChange = (val: string) => {
    const tab = tabs.value.find(item => item.name === val)
    if (tab) {
      router.push({ name: tab.routeName })
    }
  }

</script>
<template>
  <div class="serving-detail-header">
    <div class="serving-list-wrapper">
      <ServingSelector :value="props.appId" />
    </div>
    <div class="detail-header-tabs">
      <BkTab type="unborder-card" v-model:active="activeTab" @change="handleTabChange">
        <BkTabPanel v-for="tab in tabs" :key="tab.name" :name="tab.name" :label="tab.label"></BkTabPanel>
      </BkTab>
    </div>
    <div class="access-guide">
      <Help /><span style="margin-left: 6px;">接入指引</span>
    </div>
  </div>
</template>
<style lang="scss" scoped>
  .serving-detail-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 24px;
    height: 52px;
    min-width: 1366px;
    .serving-list-wrapper {
      width: 240px;
    }
    .detail-header-tabs {
      :deep(.bk-tab-header) {
        border-bottom: none;
      }
      :deep(.bk-tab-content) {
        display: none;
      }
    }
    .access-guide {
      display: flex;
      align-items: center;
      color: #3a84ff;
      font-size: 14px;
      cursor: pointer;
    }
  }
</style>