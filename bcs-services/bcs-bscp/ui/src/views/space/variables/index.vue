<script lang="ts" setup>
  import { Plus, Search } from 'bkui-vue/lib/icon'

  const handleSearch = () => {}

  const openCreateSlider = () => {}
</script>
<template>
  <section class="variables-management-page">
    <bk-alert theme="info">可为配置文件内的变量预先定义默认值。默认变量值只在变量首次使用时生效，之后更新配置文件，变量值将使用上一次生成版本是的变量值填充。</bk-alert>
    <div class="operation-area">
      <bk-button theme="primary" @click="openCreateSlider"><Plus class="button-icon" />新增变量</bk-button>
      <bk-input class="search-input" placeholder="请输入变量名称" @enter="handleSearch">
          <template #suffix>
            <Search class="search-input-icon" />
          </template>
      </bk-input>
    </div>
    <div class="variable-table">
      <bk-table>
        <bk-table-column label="变量名称"></bk-table-column>
        <bk-table-column label="类型"></bk-table-column>
        <bk-table-column label="默认值"></bk-table-column>
        <bk-table-column label="描述"></bk-table-column>
        <bk-table-column label="操作"></bk-table-column>
      </bk-table>
    </div>
  </section>
</template>
<style lang="scss" scoped>
  .variables-management-page {
    height: 100%;
    background: #f5f7fa;
  }
  .operation-area {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-top: 24px;
    padding: 0 24px;
    .button-icon {
      font-size: 18px;
    }
    .search-input {
      width: 320px;
    }
    .search-input-icon {
      padding-right: 10px;
      color: #979ba5;
      background: #ffffff;
    }
  }
  .variable-table {
    padding: 16px 24px 24px;
  }
</style>
