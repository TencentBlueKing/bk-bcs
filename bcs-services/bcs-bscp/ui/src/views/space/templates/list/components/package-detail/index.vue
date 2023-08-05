<script lang="ts" setup>
  import { ref, computed } from 'vue';
  import { storeToRefs } from 'pinia'
  import { Search } from 'bkui-vue/lib/icon'
  import { useTemplateStore } from '../../../../../../store/template'
  import { getDefaultPackageConfig } from '../../../../../../utils/template'
  import PackageConfigTable from './table.vue'

  const { packageList, currentTemplateSpace } = storeToRefs(useTemplateStore())

  const pkg = computed(() => {
    const pkgDetail = packageList.value.find(item => item.id === currentTemplateSpace.value)
    if (pkgDetail) {
      return pkgDetail
    }
    return getDefaultPackageConfig()
  })
</script>
<template>
  <div class="header-wrapper">
    <div class="tag">公开</div>
    <h4 class="package-name">{{ pkg.spec.name }}</h4>
    <p class="package-desc">{{ pkg.spec.memo }}</p>
  </div>
  <div class="operation-wrapper">
    <div class="config-actions">
      <bk-button theme="primary" class="create-config-btn">添加配置项</bk-button>
      <bk-button>批量添加至</bk-button>
      <bk-button>批量删除</bk-button>
    </div>
    <div class="config-search">
      <bk-input placeholder="配置项名称/路径/描述/创建人/更新人">
        <template #suffix>
          <Search class="search-icon" />
        </template>
      </bk-input>
    </div>
  </div>
  <div class="table-list-wrapper">
    <PackageConfigTable />
  </div>
</template>
<style lang="scss" scoped>
  .header-wrapper {
    display: flex;
    align-items: center;
    .tag {
      margin-right: 8px;
      padding: 0 8px;
      height: 22px;
      line-height: 22px;
      font-size: 12px;
      color: #4193e5;
      background: #dae9fd;
      border-radius: 2px;
      text-align: center;
    }
    .package-name {
      font-size: 14px;
      font-weight: 700;
      color: #63656e;
    }
    .package-desc {
      margin-left: 16px;
      height: 20px;
      line-height: 20px;
      color: #979ba5;
      font-size: 12px;
    }
  }
  .operation-wrapper {
    display: flex;
    align-items: center;
    justify-content: space-between;
    .create-config-btn {
      margin-right: 8px;
    }
    .config-search {
      width: 320px;
      background-color: #ffffff;
      .search-icon {
        margin-right: 10px;
        color: #979ba5;
      }
    }
  }
  .table-list-wrapper {
    margin-top: 16px;
  }
</style>
