<script lang="ts" setup>
  import { onMounted, ref, watch } from 'vue';
  import { Ellipsis, Search } from 'bkui-vue/lib/icon'
  import { ITemplateConfigItem } from '../../../../../../../types/template';
  import { ICommonQuery } from '../../../../../../../types/index';

  const props = defineProps<{
    currentPkg: number|string;
    getConfigList: Function;
  }>()

  const loading = ref(false)
  const list = ref<ITemplateConfigItem[]>([])
  const searchStr = ref('')
  const pagination = ref({
    current: 1,
    limit: 10,
    count: 0
  })

  watch(() => props.currentPkg, () => {
    searchStr.value = ''
    loadConfigList()
  })

  onMounted(() => {
    loadConfigList()
  })

  const loadConfigList = async () => {
    loading.value = true
    const params:ICommonQuery = {
      start: (pagination.value.current - 1) * pagination.value.limit,
      limit: pagination.value.limit
    }
    if (searchStr.value) {
      params.search_key = searchStr.value
    }
    const res = await props.getConfigList(params)
    list.value = res.details
    loading.value = false
  }

  const refreshList = (current: number = 1) => {
    pagination.value.current = current
    loadConfigList()
  }

  const handleSearchInputChange = () => {
    if (!searchStr.value) {
      refreshList()
    }
  }
</script>
<template>
  <div class="package-config-table">
    <div class="operate-area">
      <div class="table-operate-btns">
        <slot name="tableOperations">
        </slot>
      </div>
      <bk-input
        v-model="searchStr"
        class="search-script-input"
        placeholder="配置项名称/路径/描述/创建人/更新人"
        :clearable="true"
        @enter="refreshList()"
        @clear="refreshList()"
        @input="handleSearchInputChange">
          <template #suffix>
            <Search class="search-input-icon" />
          </template>
      </bk-input>
    </div>
    <bk-table empty-text="暂无配置项" :border="['outer']" :data="list">
      <bk-table-column label="配置项名称">
        <template #default="{ row }">
          <div v-if="row.spec">{{ row.spec.name }}</div>
        </template>
      </bk-table-column>
      <bk-table-column label="配置项路径" prop="spec.path"></bk-table-column>
      <bk-table-column label="配置项描述" prop="spec.memo"></bk-table-column>
      <slot name="columns"></slot>
      <bk-table-column label="创建人" prop="revision.creator"></bk-table-column>
      <bk-table-column label="更新人" prop="revision.reviser"></bk-table-column>
      <bk-table-column label="更新时间" prop="revision.update_at"></bk-table-column>
      <bk-table-column label="操作">
        <template #default="{ row }">
          <div class="actions-wrapper">
            <bk-button theme='primary' text>版本管理</bk-button>
            <div class="more-actions">
              <Ellipsis class="ellipsis-icon" />
            </div>
          </div>
        </template>
      </bk-table-column>
    </bk-table>
  </div>
</template>
<style lang="scss" scoped>
  .operate-area {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 16px;
    .table-operate-btns {
      display: flex;
      align-items: center;
      :deep(.bk-button) {
        margin-right: 8px;
      }
    }
  }
  .search-script-input {
    width: 320px;
  }
  .search-input-icon {
    padding-right: 10px;
    color: #979ba5;
    background: #ffffff;
  }
  .actions-wrapper {
    display: flex;
    align-items: center;
    height: 100%;
    .more-actions {
      display: flex;
      align-items: center;
      justify-content: center;
      margin-left: 16px;
      width: 16px;
      height: 16px;
      border-radius: 50%;
      cursor: pointer;
      &:hover {
        background: #dcdee5;
        color: #3a84ff;
      }
    }
    .ellipsis-icon {
      transform: rotate(90deg);
    }
  }
</style>
