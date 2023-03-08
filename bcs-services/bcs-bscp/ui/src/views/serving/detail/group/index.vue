<script setup lang="ts">
  import { ref } from 'vue'
  import { Plus, Search } from 'bkui-vue/lib/icon'
  import CreateGroup from './create-group.vue'

  const currentTab = ref('normal')
  const isCreateDialogShow = ref(false)

</script>
<template>
  <section class="app-group-page">
    <div class="operate-area">
      <div class="action-btns">
        <bk-button theme="primary" @click="isCreateDialogShow = true"><Plus class="button-icon" />创建分组</bk-button>
        <div class="group-tabs">
          <div :class="['tab-item', { active: currentTab === 'normal' }]" @click="currentTab='normal'">
            普通分组
            <span class="count">0</span>
          </div>
          <div :class="['tab-item', { active: currentTab === 'debug' }]" @click="currentTab='debug'">
            调试用分组
            <span class="count">0</span>
          </div>
        </div>
      </div>
      <bk-input class="search-group-input" placeholder="分组名称">
         <template #suffix>
            <Search class="search-input-icon" />
         </template>
      </bk-input>
    </div>
    <div class="group-list-wrapper">
      <bk-exception type="empty">此服务下暂无分组</bk-exception>
    </div>
    <CreateGroup v-model:show="isCreateDialogShow" />
  </section>
</template>
<style lang="scss" scoped>
  .app-group-page {
    padding: 24px;
    height: 100%;
    background: #f5f7fa;
  }
  .operate-area {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
  .action-btns {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
  .group-tabs {
    display: flex;
    align-items: center;
    margin-left: 24px;
    padding: 3px 4px;
    background: #f0f1f5;
    border-radius: 4px;
    .tab-item {
      padding: 6px 14px;
      font-size: 12px;
      line-height: 14px;
      color: #63656e;
      border-radius: 4px;
      cursor: pointer;
      &.active {
        color: #3a84ff;
        background: #ffffff;
        .count {
          background: #a3c5fd;
          color: #ffffff;
          line-height: 16px;
        }
      }
      .count {
        padding: 0 8px;
        border-radius: 2px;
        color: #979ba5;
      }
    }
  }
  .button-icon {
    font-size: 18px;
  }
  .search-group-input {
    width: 320px;
  }
  .search-input-icon {
    padding-right: 10px;
    color: #979ba5;
    background: #ffffff;
  }
</style>
