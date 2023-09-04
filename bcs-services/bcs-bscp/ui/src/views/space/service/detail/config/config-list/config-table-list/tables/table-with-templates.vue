<script lang="ts" setup>
  import { ref, computed, watch, onMounted } from 'vue'
  import { storeToRefs } from 'pinia'
  import { InfoBox, Message } from "bkui-vue/lib";
  import { DownShape, Close } from 'bkui-vue/lib/icon';
  import { useConfigStore } from '../../../../../../../../store/config'
  import { ICommonQuery } from '../../../../../../../../../types/index';
  import { IConfigItem, IConfigListQueryParams, IBoundTemplateDetail } from '../../../../../../../../../types/config'
  import { getConfigList, getBoundTemplates, getBoundTemplatesByAppVersion, deleteServiceConfigItem, deleteBoundPkg } from '../../../../../../../../api/config'
  import { getAppPkgBindingRelations } from '../../../../../../../../api/template'
  import StatusTag from './status-tag'
  import EditConfig from '../edit-config.vue'
  import ViewConfig from '../view-config.vue'
  import VersionDiff from '../../../components/version-diff/index.vue'
  import ReplaceTemplateVersion from '../replace-template-version.vue';

  interface IConfigsGroupData {
    id: number;
    name: string;
    expand: boolean;
    configs: IConfigTableItem[];
  }

  interface IConfigTableItem {
    id: number;
    name: string;
    versionId: number;
    versionName: string;
    path: string;
    creator: string;
    reviser: string;
    update_at: string;
    file_state: string;
  }

  const configStore = useConfigStore()
  const { versionData } = storeToRefs(configStore)

  const props = defineProps<{
    bkBizId: string;
    appId: number;
    searchStr: string;
  }>()

  const loading = ref(false)
  const commonConfigListLoading = ref(false)
  const bindingId = ref(0)
  const configList = ref<IConfigItem[]>([]) // 非模板配置项
  const boundTemplateListLoading = ref(false)
  const templateList = ref<IBoundTemplateDetail[]>([]) // 配置项模板
  const tableGroupsData = ref<IConfigsGroupData[]>([])
  const editPanelShow = ref(false)
  const activeConfig = ref(0)
  const isDiffPanelShow = ref(false)
  const viewConfigSliderData = ref({
    open: false,
    data: { id: 0, type: '' }
  })
  const replaceDialogData = ref({
    open: false,
    data: {
      pkgId: 0,
      templateId: 0,
      versionId: 0,
      versionName: ''
    }
  })
  const diffConfig = ref(0)

  // 是否为未命名版本
  const isUnNamedVersion = computed(() => {
    return versionData.value.id === 0
  })

  watch(() => versionData.value.id, async() => {
    await getBindingId()
    getAllConfigList()
  })

  watch(() => props.searchStr, () => {
    getAllConfigList()
  })

  onMounted(async() => {
    await getBindingId()
    getAllConfigList()
  })

  const getBindingId = async () => {
    const res = await getAppPkgBindingRelations(props.bkBizId, props.appId)
    bindingId.value = res.details.length === 1 ? res.details[0].id : 0
  }

  const getAllConfigList = async() => {
    await Promise.all([
      getCommonConfigList(),
      getBoundTemplateList()
    ])
    tableGroupsData.value = transListToTableData()
  }

  // 获取非模板配置项列表
  const getCommonConfigList = async() => {
    commonConfigListLoading.value = true
    try {
      const params: IConfigListQueryParams = {
        start: 0,
        all: true
      }
      if (props.searchStr) {
        params.searchKey = props.searchStr
      }
      if (!isUnNamedVersion.value) {
        params.release_id = versionData.value.id
      }
      const res = await getConfigList(props.bkBizId, props.appId, params)
      configList.value = res.details
    } catch (e) {
      console.error(e)
    } finally {
      commonConfigListLoading.value = false
    }
  }

  // 获取模板配置项列表
  const getBoundTemplateList = async() => {
    boundTemplateListLoading.value = true
    try {
      const params: ICommonQuery = {
        start: 0,
        all: true
      }
      if (props.searchStr) {
        params.search_fields = 'revision_name,revision_memo,name,path,creator'
        params.search_value = props.searchStr
      }

      let res
      if (isUnNamedVersion.value) {
        res = await getBoundTemplates(props.bkBizId, props.appId, params)
      } else {
        res = await getBoundTemplatesByAppVersion(props.bkBizId, props.appId, versionData.value.id)
      }
      templateList.value = res.details
    } catch (e) {
      console.error(e)
    } finally {
      boundTemplateListLoading.value = false
    }
  }

  const transListToTableData = () => {
    const pkgsGroups = groupTplsByPkg(templateList.value)
    return [
      { id: 0, name: '非模板配置', expand: true, configs: transConfigsToTableItemData(configList.value) },
      ...pkgsGroups
    ]
  }

  // 将非模板配置项数据转为表格数据
  const transConfigsToTableItemData = (list: IConfigItem[]) => {
    return list.map((item: IConfigItem) => {
        const { id, spec, revision, file_state } = item
        const { name, path } = spec
        const { creator, reviser, update_at } = revision
        return { id, name, versionId: 0, versionName: '--', path, creator, reviser, update_at, file_state }
      })
  }

  // 将模板按套餐分组，并将模板数据格式转为表格数据
  const groupTplsByPkg = (list: IBoundTemplateDetail[]) => {
    const groups: IConfigsGroupData[] = []
    list.forEach(tpl => {
      const {
        template_space_name, template_set_id, template_set_name,
        template_id: id, name, template_revision_id: versionId,
        template_revision_name: versionName, path, creator
      } = tpl
      const config = { id, name, versionId, versionName, path, creator, reviser: '--', update_at: '--', file_state: '' }
      const group = groups.find(item => item.id === template_set_id)
      if (group) {
        group.configs.push(config)
      } else {
        groups.push({
          id: template_set_id,
          name: `${template_space_name} - ${template_set_name}`,
          expand: false,
          configs: [config]
        })
      }
    })
    return groups
  }

  const handleEditOpen = (config: IConfigTableItem) => {
    activeConfig.value = config.id
    editPanelShow.value = true
  }

  // 编辑模板
  const handleEditConfigConfirm = async() => {
    await getCommonConfigList()
    tableGroupsData.value = transListToTableData()
  }

  // 查看配置项或模板版本
  const handleViewConfig = (id: number, type: string) => {
    viewConfigSliderData.value = {
      open: true,
      data: { id, type }
    }
  }

  const handleOpenReplaceVersionDialog = (config: IConfigTableItem) => {
    const { id: templateId, versionId, versionName } = config
    const pkgId = templateList.value.find(item => item.template_id === templateId)?.template_set_id || 0
    replaceDialogData.value = {
      open: true,
      data: { pkgId, templateId, versionId, versionName }
    }
  }

  // 删除模板套餐
  const handleDeletePkg = async(id: number) => {
    const tpl = templateList.value.find(item => item.template_set_id === id)
    if (tpl) {
      InfoBox({
        title: `确认是否删除模板套餐【${tpl.template_space_name} - ${tpl.template_set_name}】?`,
        infoType: "danger",
        headerAlign: "center" as const,
        footerAlign: "center" as const,
        onConfirm: async () => {
          await deleteBoundPkg(props.bkBizId, props.appId, bindingId.value, [tpl.template_set_id])
          await getBoundTemplateList()
          tableGroupsData.value = transListToTableData()
          Message({
            theme: 'success',
            message: '删除模板套餐成功'
          })
        },
      } as any);
    }
  }

  // 非模板配置项diff
  const handleConfigDiff = (config: IConfigTableItem) => {
    diffConfig.value = config.id
    isDiffPanelShow.value = true
  }

  // 模板配置项diff
  const handleTplDiff = (config: IConfigTableItem) => {

  }

  // 删除配置项
  const handleDel = (config: IConfigTableItem) => {
    InfoBox({
      title: `确认是否删除配置项【${config.name}】?`,
      infoType: "danger",
      headerAlign: "center" as const,
      footerAlign: "center" as const,
      onConfirm: async () => {
        await deleteServiceConfigItem(config.id, props.bkBizId, props.appId)
        await getCommonConfigList()
        tableGroupsData.value = transListToTableData()
      },
    } as any);
  }

  defineExpose({
    refresh: getAllConfigList
  })

</script>
<template>
  <bk-loading :loading="loading" style="min-height: 200px">
    <table class="config-groups-table">
      <thead>
        <tr class="config-groups-table-tr">
          <th class="name">配置项名称</th>
          <th class="version">配置模板版本</th>
          <th class="path">配置路径</th>
          <th class="user">创建人</th>
          <th class="user">修改人</th>
          <th class="datetime">修改时间</th>
          <th class="status">变更状态</th>
          <th class="operation">操作</th>
        </tr>
      </thead>
      <tbody>
        <template v-for="group in tableGroupsData" :key="group.id">
          <tr class="config-groups-table-tr group-title-row">
            <td colspan="8" class="config-groups-table-td">
              <div class="configs-group">
                <div class="name-wrapper" @click="group.expand = !group.expand">
                  <DownShape :class="['fold-icon', { fold: !group.expand }]" />
                  {{ group.name }}
                </div>
                <div v-if="group.id !== 0" class="delete-btn" @click="handleDeletePkg(group.id)">
                  <Close class="close-icon" />
                  删除套餐
                </div>
              </div>
            </td>
          </tr>
          <template v-if="group.expand">
            <tr class="config-groups-table-tr">
              <td colspan="8" class="config-groups-table-td">
                <div class="configs-list-wrapper">
                  <table class="config-list-table">
                    <tbody>
                      <tr v-for="config in group.configs" :key="config.id" class="config-row">
                        <td>
                          <template v-if="group.id === 0">
                            <bk-button
                              v-if="isUnNamedVersion"
                              text
                              theme="primary"
                              :disabled="config.file_state === 'DELETE'"
                              @click="handleEditOpen(config)">
                              {{ config.name }}
                            </bk-button>
                            <bk-button
                              v-else
                              text
                              theme="primary"
                              :disabled="config.file_state === 'DELETE'"
                              @click="handleViewConfig(config.id, 'config')">
                              {{ config.name }}
                            </bk-button>
                          </template>
                          <bk-button
                            v-else
                            text
                            theme="primary"
                            :disabled="config.file_state === 'DELETE'"
                            @click="handleViewConfig(config.versionId, 'template')">
                            {{ config.name }}
                          </bk-button>
                        </td>
                        <td>{{ config.versionName }}</td>
                        <td>{{ config.path }}</td>
                        <td class="user">{{ config.creator }}</td>
                        <td class="user">{{ config.reviser }}</td>
                        <td class="datetime">{{ config.update_at }}</td>
                        <td class="status"><StatusTag :status="config.file_state" /></td>
                        <td class="operation">
                          <div class="config-actions">
                            <!-- 非套餐配置项 -->
                            <template v-if="group.id === 0">
                              <template v-if="isUnNamedVersion">
                                <bk-button text theme="primary" @click="handleEditOpen(config)">编辑</bk-button>
                                <bk-button text theme="primary" @click="handleDel(config)">删除</bk-button>
                              </template>
                              <template v-else>
                                <bk-button text theme="primary" @click="handleViewConfig(config.id, 'config')">查看</bk-button>
                                <bk-button v-if="versionData.status.publish_status !== 'editing'" text theme="primary" @click="handleConfigDiff(config)">对比</bk-button>
                              </template>
                            </template>
                            <!-- 套餐模板 -->
                            <template v-else>
                              <bk-button v-if="isUnNamedVersion" text theme="primary" @click="handleOpenReplaceVersionDialog(config)">替换版本</bk-button>
                              <template v-else>
                                <bk-button text theme="primary" @click="handleViewConfig(config.versionId, 'template')">查看</bk-button>
                                <bk-button v-if="versionData.status.publish_status !== 'editing'" text theme="primary" @click="handleTplDiff(config)">对比</bk-button>
                              </template>
                            </template>
                          </div>
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>
                <bk-exception v-if="group.configs.length === 0" class="exception-tips" scene="part" type="empty">暂无配置项</bk-exception>
              </td>
            </tr>
          </template>
        </template>
      </tbody>
    </table>
  </bk-loading>
  <edit-config
    v-model:show="editPanelShow"
    :config-id="activeConfig"
    :bk-biz-id="props.bkBizId"
    :app-id="props.appId"
    @confirm="handleEditConfigConfirm" />
  <ViewConfig
    v-model:show="viewConfigSliderData.open"
    v-bind="viewConfigSliderData.data"
    :bk-biz-id="props.bkBizId"
    :app-id="props.appId" />
  <VersionDiff
    v-model:show="isDiffPanelShow"
    :current-version="versionData"
    :current-config="diffConfig" />
  <ReplaceTemplateVersion
    v-model:show="replaceDialogData.open"
    v-bind="replaceDialogData.data"
    :binding-id="bindingId"
    :bk-biz-id="props.bkBizId"
    :app-id="props.appId"
    @updated="getAllConfigList" />
</template>
<style lang="scss" scoped>
  .config-groups-table {
    width: 100%;
    border: 1px solid #dedee5;
    border-collapse: collapse;
    table-layout: fixed;
    .config-groups-table-tr {
      th {
        padding: 11px 16px;
        color: #313238;
        font-weight: normal;
        font-size: 12px;
        line-height: 20px;
        text-align: left;
        background: #fafbfd;
        border-bottom: 1px solid #dedee5;
      }
    }
    .group-title-row:hover {
      background: #f5f7fa;
    }
    .config-groups-table-td {
      padding: 0;
      font-size: 12px;
      line-height: 20px;
      text-align: left;
      border-bottom: 1px solid #dedee5;
      color: #63656e;
    }
    .configs-group {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 0 16px;
      height: 42px;
      .name-wrapper {
        display: flex;
        align-items: center;
        height: 100%;
        cursor: pointer;
        .fold-icon {
          margin-right: 8px;
          font-size: 14px;
          color: #3a84ff;
          transition: transform 0.2s ease-in-out;
          &.fold {
            color: #c4c6cc;
            transform: rotate(-90deg);
          }
        }
      }
      .delete-btn {
        display: flex;
        align-items: center;
        margin-right: 20px;
        color: #3a84ff;
        cursor: pointer;
        .close-icon {
          margin-right: 4px;
          font-size: 14px;
        }
      }
    }
    .user {
      width: 120px;
    }
    .datetime {
      width: 158px;
    }
    .status {
      width: 100px;
    }
    .operation {
      width: 100px;
    }
    .exception-tips {
      margin: 20px 0;
    }
  }
  .configs-list-wrapper {
    max-height: 420px;
    overflow: auto;
  }
  .config-list-table {
    width: 100%;
    border-collapse: collapse;
    table-layout: fixed;
    .config-row {
      &:hover {
        background: #f5f7fa;
      }
      &:last-child td {
        border: none;
      }
    }
    td {
      padding: 11px 16px;
      height: 42px;
      border-bottom: 1px solid #dedee5;
      font-size: 12px;
      line-height: 20px;
      color: #63656e;
      white-space: nowrap;
      text-overflow: ellipsis;
      overflow: hidden;
    }
    .config-actions {
      .bk-button:not(:last-child) {
        margin-right: 8px;
      }
    }
  }
</style>
