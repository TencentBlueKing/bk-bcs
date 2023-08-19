<script lang="ts" setup>
  import { onMounted, ref, watch } from 'vue';
  import { useRouter } from 'vue-router';
  import { Ellipsis, Search } from 'bkui-vue/lib/icon'
  import { ITemplateConfigItem } from '../../../../../../../types/template';
  import { ICommonQuery } from '../../../../../../../types/index';
  import AddToDialog from '../operations/add-to-pkgs/add-to-dialog.vue'
  import MoveOutFromPkgsDialog from '../operations/move-out-from-pkg/move-out-from-pkgs-dialog.vue'

  const router = useRouter()

  const props = defineProps<{
    currentTemplateSpace: number;
    currentPkg: number|string;
    selectedConfigs: ITemplateConfigItem[];
    getConfigList: Function;
  }>()

  const emits = defineEmits(['update:selectedConfigs'])

  const loading = ref(false)
  const list = ref<ITemplateConfigItem[]>([])
  const searchStr = ref('')
  const pagination = ref({
    current: 1,
    limit: 10,
    count: 0
  })
  const isAddToPkgsDialogShow = ref(false)
  const isMoveOutFromPkgsDialogShow = ref(false)
  const crtConfig = ref<ITemplateConfigItem[]>([])

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

  const handleSelectionChange = ({ checked, isAll, row }: { checked: boolean, isAll: boolean; row: ITemplateConfigItem }) => {
    const configs = props.selectedConfigs.slice()
    if (isAll) {
      if (checked) {
        list.value.forEach(config => {
          if (!configs.find(item => item.id === config.id)) {
            configs.push(config)
          }
        })
      } else {
        list.value.forEach(config => {
          const index = configs.findIndex(item => item.id === config.id)
          if (index > -1) {
            configs.splice(index, 1)
          }
        })
      }
    } else {
      if (checked) {
        if (!configs.find(item => item.id === row.id)) {
          configs.push(row)
        }
      } else {
        const index = configs.findIndex(item => item.id === row.id)
        if (index > -1) {
          configs.splice(index, 1)
        }
      }
    }
    emits('update:selectedConfigs', configs)
  }

  const handleEditConfig = () => {}

  const handleOpenAddToPkgsDialog = (config: ITemplateConfigItem) => {
    isAddToPkgsDialogShow.value = true
    crtConfig.value = [config]
  }

  const handleOpenMoveOutFromPkgsDialog = (config: ITemplateConfigItem) => {
    isMoveOutFromPkgsDialogShow.value = true
    crtConfig.value = [config]
  }

  const goToVersionManange = (id: number) => {
    router.push({ name: 'template-version-manange', params: {
      templateSpaceId: props.currentTemplateSpace,
      packageId: props.currentPkg,
      templateId: id
    }})
  }

  defineExpose({
    refreshList
  })

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
    <bk-loading style="min-height: 200px;" :loading="loading">
      <bk-table empty-text="暂无配置项" :border="['outer']" :data="list" @selection-change="handleSelectionChange">
        <bk-table-column type="selection" :fixed="true" :width="40"></bk-table-column>
        <bk-table-column label="配置项名称" :fixed="true" :width="280">
          <template #default="{ row }">
            <div v-if="row.spec" @click="handleEditConfig">{{ row.spec.name }}</div>
          </template>
        </bk-table-column>
        <bk-table-column label="配置项路径" prop="spec.path" :width="280"></bk-table-column>
        <bk-table-column label="配置项描述" prop="spec.memo"></bk-table-column>
        <slot name="columns"></slot>
        <bk-table-column label="创建人" prop="revision.creator" :width="100"></bk-table-column>
        <bk-table-column label="更新人" prop="revision.reviser" :width="100"></bk-table-column>
        <bk-table-column label="更新时间" prop="revision.update_at" :width="180"></bk-table-column>
        <bk-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <div class="actions-wrapper">
              <slot name="columnOperations" :config="row">
                <bk-button theme='primary' text @click="goToVersionManange(row.id)">版本管理</bk-button>
                <bk-popover
                  theme="light template-config-actions-popover"
                  placement="bottom-end"
                  :arrow="false">
                  <div class="more-actions">
                    <Ellipsis class="ellipsis-icon" />
                  </div>
                  <template #content>
                    <div class="config-actions">
                      <div class="action-item" @click="handleOpenAddToPkgsDialog(row)">添加至套餐</div>
                      <div class="action-item" @click="handleOpenMoveOutFromPkgsDialog(row)">移出套餐</div>
                    </div>
                  </template>
                </bk-popover>
              </slot>
            </div>
          </template>
        </bk-table-column>
      </bk-table>
    </bk-loading>
    <AddToDialog v-model:show="isAddToPkgsDialogShow" :value="crtConfig" />
    <MoveOutFromPkgsDialog
      v-model:show="isMoveOutFromPkgsDialogShow"
      :id="crtConfig.length > 0 ? crtConfig[0].id : 0"
      :name="crtConfig.length > 0 ? crtConfig[0].spec.name : ''" />
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
<style lang="scss">
  .template-config-actions-popover.bk-popover.bk-pop2-content {
    padding: 4px 0;
    border: 1px solid #dcdee5;
    box-shadow: 0 2px 6px 0 #0000001a;
    .config-actions {
      .action-item {
        padding: 0 12px;
        min-width: 58px;
        height: 32px;
        line-height: 32px;
        color: #63656e;
        font-size: 12px;
        cursor: pointer;
        &:hover {
          background: #f5f7fa;
        }
      }
    }
  }
</style>
