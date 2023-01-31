<script setup lang="ts">
  import { defineProps, ref, computed, watch } from 'vue'
  import { useRoute } from 'vue-router'
  import VersionList from './version-list.vue'
  import ConfigList from './config-list.vue'
  import CreateBtn from './create-btn.vue'

  const route = useRoute()

  const props = defineProps<{
    bkBizId: number
  }>()

  const appId = ref(Number(route.params.id))
  const configList = ref()

  watch(() => route.params.id, (val) => {
    appId.value = Number(val)
  })

  const updateConfigList = () => {
    configList.value.refreshConfigList()
  }

</script>
<template>
  <section class="serving-config-wrapper">
    <section class="version-list-side">
      <VersionList />
    </section>
    <section class="version-config-content">
      <section class="config-content-header">
        <section class="summary-wrapper">
          <div class="status-tag">编辑中</div>
          <div class="version-name">未命名版本</div>
        </section>
        <section class="actions-wrapper">
          <bk-button theme="primary">生成版本</bk-button>
        </section>
      </section>
      <CreateBtn :bk-biz-id="props.bkBizId" :app-id="appId" @update="updateConfigList" />
      <ConfigList ref="configList" :bk-biz-id="props.bkBizId" :app-id="appId" />
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
  .config-content-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    height: 64px;
    border-bottom: 1px solid #dcdee5;
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
      }
      .version-name {
        color: #63656e;
        font-size: 14px;
        font-weight: bold;
      }
    }
  }
</style>
