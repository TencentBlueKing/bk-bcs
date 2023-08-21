<script lang="ts" setup>
  import { onMounted, ref, watch } from 'vue';
  import { useRouter } from 'vue-router';
  import { storeToRefs } from 'pinia';
  import { Ellipsis, Search, Spinner } from 'bkui-vue/lib/icon'
  import { useGlobalStore } from '../../../../../../store/global';
  import { useTemplateStore } from '../../../../../../store/template';
  import { ICommonQuery } from '../../../../../../../types/index';
  import { ITemplateConfigItem, ITemplateCitedCountDetailItem } from '../../../../../../../types/template';
  import { getPackagesByTemplateIds, getCountsByTemplateIds } from '../../../../../../api/template'
  import AddToDialog from '../operations/add-to-pkgs/add-to-dialog.vue'
  import MoveOutFromPkgsDialog from '../operations/move-out-from-pkg/move-out-from-pkgs-dialog.vue'
  import PkgsTag from '../../components/packages-tag.vue'
  import AppsBoundByTemplate from '../apps-bound-by-template.vue';

  const router = useRouter()

  const { spaceId } = storeToRefs(useGlobalStore())
  const templateStore = useTemplateStore()
  const { currentTemplateSpace } = storeToRefs(templateStore)

  const props = defineProps<{
    currentTemplateSpace: number;
    currentPkg: number|string;
    selectedConfigs: ITemplateConfigItem[];
    showCitedByPkgsCol?: boolean; // 是否显示模板被套餐引用列
    showBoundByAppsCol?: boolean; // 是否显示模板被服务引用列
    getConfigList: Function;
  }>()

  const emits = defineEmits(['update:selectedConfigs'])

  const listLoading = ref(false)
  const list = ref<ITemplateConfigItem[]>([])
  const citedByPkgsLoading = ref(false)
  const citeByPkgsList = ref<{template_set_id: number; template_set_name: string;}[][]>([])
  const boundByAppsCountLoading = ref(false)
  const boundByAppsCountList = ref<ITemplateCitedCountDetailItem[]>([])
  const searchStr = ref('')
  const pagination = ref({
    current: 1,
    limit: 10,
    count: 0
  })
  const isAddToPkgsDialogShow = ref(false)
  const isMoveOutFromPkgsDialogShow = ref(false)
  const appBoundByTemplateSliderData = ref<{ open: boolean; data: { id: number; name: string; } }>({
    open: false,
    data: {
      id: 0,
      name: ''
    }
  })
  const crtConfig = ref<ITemplateConfigItem[]>([])

  watch(() => props.currentPkg, () => {
    searchStr.value = ''
    loadConfigList()
  })

  onMounted(() => {
    loadConfigList()
  })

  const loadConfigList = async () => {
    listLoading.value = true
    const params:ICommonQuery = {
      start: (pagination.value.current - 1) * pagination.value.limit,
      limit: pagination.value.limit
    }
    if (searchStr.value) {
      params.search_key = searchStr.value
    }
    const res = await props.getConfigList(params)
    list.value = res.details
    pagination.value.count = res.count
    listLoading.value = false
    const ids = list.value.map(item => item.id)
    citeByPkgsList.value = []
    boundByAppsCountList.value = []
    if (ids.length > 0) {
      if (props.showCitedByPkgsCol) {
        loadCiteByPkgsCountList(ids)
      }

      if (props.showBoundByAppsCol) {
        loadBoundByAppsList(ids)
      }
    }
  }

  const loadCiteByPkgsCountList = async(ids: number[]) => {
    citedByPkgsLoading.value = true
    const res = await getPackagesByTemplateIds(spaceId.value, currentTemplateSpace.value, ids)
    citeByPkgsList.value = res.details
    citedByPkgsLoading.value = false
  }

  const loadBoundByAppsList = async(ids: number[]) => {
    boundByAppsCountLoading.value = true
    const res = await getCountsByTemplateIds(spaceId.value, currentTemplateSpace.value, ids)
    boundByAppsCountList.value = res.details
    boundByAppsCountLoading.value = false
  }

  const refreshList = (current: number = 1) => {
    pagination.value.current = current
    loadConfigList()
  }

  // 模板移出或删除后刷新列表
  const refreshListAfterDeleted = (num: number) => {
    if (num === list.value.length && pagination.value.current > 1) {
      pagination.value.current -= 1
    }
    refreshList()
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

  const handleOpenAddToPkgsDialog = (config: ITemplateConfigItem) => {
    isAddToPkgsDialogShow.value = true
    crtConfig.value = [config]
  }

  const handleOpenMoveOutFromPkgsDialog = (config: ITemplateConfigItem) => {
    isMoveOutFromPkgsDialogShow.value = true
    crtConfig.value = [config]
  }

  const handleMovedOut = () => {
    refreshListAfterDeleted(1)
    crtConfig.value = []
    templateStore.$patch(state => {
      state.needRefreshMenuFlag = true
    })
  }

  const handleOpenAppBoundByTemplateSlider = (config: ITemplateConfigItem) => {
    appBoundByTemplateSliderData.value = {
      open: true,
      data: {
        id: config.id,
        name: config.spec.name
      }
    }
  }

  const refreshConfigList = () => {
    refreshList()
  }

  const goToVersionManage = (id: number) => {
    router.push({ name: 'template-version-manange', params: {
      templateSpaceId: props.currentTemplateSpace,
      packageId: props.currentPkg,
      templateId: id
    }})
  }

  defineExpose({
    refreshList,
    refreshListAfterDeleted
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
    <bk-loading style="min-height: 200px;" :loading="listLoading">
      <bk-table empty-text="暂无配置项" :border="['outer']" :data="list" @selection-change="handleSelectionChange">
        <bk-table-column type="selection" :min-width="40" :width="40"></bk-table-column>
        <bk-table-column label="配置项名称">
          <template #default="{ row }">
            <bk-button v-if="row.spec" text theme="primary" @click="goToVersionManage(row.id)">{{ row.spec.name }}</bk-button>
          </template>
        </bk-table-column>
        <bk-table-column label="配置项路径" prop="spec.path"></bk-table-column>
        <bk-table-column label="配置项描述" prop="spec.memo"></bk-table-column>
        <template v-if="showCitedByPkgsCol">
          <bk-table-column label="所在套餐">
            <template #default="{ index }">
              <template v-if="citedByPkgsLoading"><Spinner /></template>
              <template v-else-if="citeByPkgsList[index]">
                <PkgsTag v-if="citeByPkgsList[index].length > 0" :pkgs="citeByPkgsList[index]" />
                <span v-else>--</span>
              </template>
            </template>
          </bk-table-column>
        </template>
        <template v-if="showBoundByAppsCol">
          <bk-table-column label="被引用">
            <template #default="{ row, index }">
              <template v-if="boundByAppsCountLoading"><Spinner /></template>
              <template v-else-if="boundByAppsCountList[index]">
                <bk-button
                  v-if="boundByAppsCountList[index].bound_unnamed_app_count > 0"
                  text
                  theme="primary"
                  @click="handleOpenAppBoundByTemplateSlider(row)">
                  {{ boundByAppsCountList[index].bound_unnamed_app_count }}
                </bk-button>
                <span v-else>0</span>
              </template>
            </template>
          </bk-table-column>
        </template>
        <bk-table-column label="创建人" prop="revision.creator" :width="100"></bk-table-column>
        <bk-table-column label="更新人" prop="revision.reviser" :width="100"></bk-table-column>
        <bk-table-column label="更新时间" prop="revision.update_at" :width="180"></bk-table-column>
        <bk-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <div class="actions-wrapper">
              <slot name="columnOperations" :config="row">
                <bk-button theme='primary' text @click="goToVersionManage(row.id)">版本管理</bk-button>
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
    <AddToDialog v-model:show="isAddToPkgsDialogShow" :value="crtConfig" @added="refreshConfigList" />
    <MoveOutFromPkgsDialog
      v-model:show="isMoveOutFromPkgsDialogShow"
      :id="crtConfig.length > 0 ? crtConfig[0].id : 0"
      :name="crtConfig.length > 0 ? crtConfig[0].spec.name : ''"
      @moved-out="handleMovedOut" />
    <AppsBoundByTemplate
      v-model:show="appBoundByTemplateSliderData.open"
      :space-id="spaceId"
      :current-template-space="currentTemplateSpace"
      :config="appBoundByTemplateSliderData.data" />
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
