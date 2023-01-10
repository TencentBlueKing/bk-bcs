<script setup lang="ts">
  import { ref } from 'vue'
  import VersionList from './version-list.vue'
  import ConfigList from './config-list.vue'
  import ConfigItemEdit from './config-item-edit.vue'

  const configEditData = ref({
    show: false,
    type: 'new',
    data: {}
  })

  const handleCreateConfigItem = () => {
    configEditData.value.show = true
    configEditData.value.data = {
      biz_id: 0,
      app_id: 0,
      name: '',
      path: '',
      file_type: 'text',
      user_group: '',
      privilege: ''
    }
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
      <bk-button style="margin: 16px 0;" outline theme="primary" @click="handleCreateConfigItem">新增配置项</bk-button>
      <ConfigList />
    </section>
    <ConfigItemEdit v-model:show="configEditData.show" :type="configEditData.type" :config="configEditData.data"/>
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
