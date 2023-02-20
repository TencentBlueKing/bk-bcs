<script setup lang="ts">
  import { defineProps, ref, computed, watch } from 'vue'
  import { useRoute } from 'vue-router'
  import VersionList from './version-list.vue'
  import ConfigList from './config-list/index.vue'
  import CreateConfig from './config-list/create-config.vue'
  import CreateVersion from './create-version/index.vue'
  import ReleaseVersion from './release-version/index.vue'

  const route = useRoute()

  const props = defineProps<{
    bkBizId: number
  }>()

  const appId = ref(Number(route.params.id))
  const appName = ref('') // @todo 需要调接口查询应用详情
  const versionName = ref('已提交的版本名称') // @todo 需要调接口获取
  const configList = ref()

  watch(() => route.params.id, (val) => {
    appId.value = Number(val)
  })

  const updateConfigList = () => {
    configList.value.refreshConfigList()
  }

  const handleUpdateStatus = () => {
    console.log('刷新配置当前配置状态')
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
          <CreateVersion
            :bk-biz-id="props.bkBizId"
            :app-id="appId"
            :app-name="appName"
            @confirm="handleUpdateStatus" />
          <ReleaseVersion
            :bk-biz-id="props.bkBizId"
            :app-id="appId"
            :app-name="appName"
            :version-name="versionName"
            @confirm="handleUpdateStatus" />
        </section>
      </section>
      <CreateConfig :bk-biz-id="props.bkBizId" :app-id="appId" @confirm="updateConfigList" />
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
    .actions-wrapper {
      display: flex;
      align-items: center;
    }
  }
</style>
