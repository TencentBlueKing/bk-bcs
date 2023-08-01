<script setup lang="ts">
  import { ref, onMounted, nextTick } from 'vue'
  import { storeToRefs } from 'pinia'
  import { useGlobalStore } from '../../../store/global'
  import { Plus, Search, Eye, Unvisible, Copy, EditLine } from 'bkui-vue/lib/icon'
  import BkMessage from 'bkui-vue/lib/message'
  import { InfoBox } from 'bkui-vue/lib'
  import { getCredentialList, createCredential, updateCredential, deleteCredential } from '../../../api/credentials'
  import { copyToClipBoard, datetimeFormat } from '../../../utils/index'
  import { ICredentialItem } from '../../../../types/credential'
  import AssociateConfigItems from './associate-config-items/index.vue'

  const { spaceId } = storeToRefs(useGlobalStore())

  const credentialList = ref<ICredentialItem[]>([])
  const listLoading = ref(false)
  const createPending = ref(false)
  const newCredentials = ref<number[]>([]) // 记录新增加的密钥id，实现表格标记效果
  const searchStr = ref('')
  const editingMemoId = ref(0) // 记录当前正在编辑说明的密钥id
  const memoInputRef = ref()
  const isAssociateSliderShow = ref(false)
  const currentCredential = ref(0)
  const pagination = ref({
    current: 1,
    count: 0,
    limit: 10,
  })

  onMounted(() => {
    refreshListWithLoading()
  })

  // 加载密钥列表
  const loadCredentialList = async () => {
    const query: { limit: number, start: number, searchKey?: string } = {
      start: pagination.value.limit * (pagination.value.current - 1),
      limit: pagination.value.limit
    }
    if (searchStr.value) {
      query.searchKey = searchStr.value
    }
    const res = await getCredentialList(spaceId.value, query)
    res.details.forEach((item: ICredentialItem) => item.visible = false)
    credentialList.value = res.details
    pagination.value.count = res.count
  }

  // 更新列表数据，带loading效果
  const refreshListWithLoading = async (current: number = 1) => {
    // 创建新密钥时，页码会切换会首页，此时不另外发请求
    if (createPending.value) {
      return
    }
    listLoading.value = true
    pagination.value.current = current
    await loadCredentialList()
    listLoading.value = false
  }

  // 设置新增行的标记class
  const getRowCls = (data: ICredentialItem) => {
    if (newCredentials.value.includes(data.id)) {
      return 'new-row-marked'
    }
    if (currentCredential.value === data.id) {
      return 'selected'
    }
    return ''
  }

  // 复制
  const handleCopyText = (text: string) => {
    copyToClipBoard(text)
    BkMessage({
      theme: 'success',
      message: '服务密钥已复制'
    })
  }

  // 创建密钥
  const handleCreateCredential = async () => {
    try {
      createPending.value = true
      const params = { memo: '' }
      const res = await createCredential(spaceId.value, params)
      pagination.value.current = 1
      await loadCredentialList()
      newCredentials.value.push(res.id)
      setTimeout(() => {
        const index = newCredentials.value.indexOf(res.id)
        newCredentials.value.splice(index, 1)
      }, 3000)
    } catch (e) {
      console.error(e)
    } finally {
      createPending.value = false
    }
  }

  // 搜索框输入事件处理，内容为空时触发一次搜索
  const handleSearchInputChange = (val: string) => {
    if (!val) {
      refreshListWithLoading()
    }
  }

  // 密钥说明编辑
  const handleEditMemo = (id: number) => {
    editingMemoId.value = id
    nextTick(() => {
      if (memoInputRef.value) {
        memoInputRef.value.focus()
      }
    })
  }

  // 失焦时保存密钥说明
  const handleMemoBlur = async (credential: ICredentialItem) => {
    editingMemoId.value = 0
    const memo = memoInputRef.value.textContent.trim()
    if (credential.spec.memo === memo) {
      return
    }

    const params = {
      id: credential.id,
      enable: credential.spec.enable,
      memo
    }
    await updateCredential(spaceId.value, params)
    credential.spec.memo = memo
    BkMessage({
      theme: 'success',
      message: '密钥说明修改成功'
    })
  }

  // 禁用/启用
  const handelToggleEnable = async(credential: ICredentialItem) => {
    if (credential.spec.enable) {
      InfoBox({
        title: '确定禁用此密钥',
        subTitle: '禁用密钥后，使用此密钥的应用将无法正常使用 SDK/API 拉取配置',
        infoType: 'warning',
        confirmText: '禁用',
        onConfirm: async () => {
          const params = {
            id: credential.id,
            memo: credential.spec.memo,
            enable: false
          }
          await updateCredential(spaceId.value, params)
          credential.spec.enable = false
        },
      } as any)
    } else {
      const params = {
        id: credential.id,
        memo: credential.spec.memo,
        enable: true
      }
      await updateCredential(spaceId.value, params)
      credential.spec.enable = true
      BkMessage({
        theme: 'success',
        message: '启用成功'
      })
    }
  }

  // 打开关联配置项侧滑
  const handleOpenAssociate = (credential: ICredentialItem) => {
    isAssociateSliderShow.value = true
    currentCredential.value = credential.id
  }

  // 关闭关联配置项侧滑
  const handleAssociateSliderClose = () => {
    isAssociateSliderShow.value = false
    currentCredential.value = 0
  }

  // 删除配置项
  const handleDelete = (credential: ICredentialItem) => {
    InfoBox({
      title: '确定删除此密钥',
      subTitle: '删除密钥后，使用此密钥的应用将无法正常使用 SDK/API 拉取配置，且密钥无法恢复',
      confirmText: '删除',
      infoType: 'warning',
      onConfirm: async () => {
        await deleteCredential(spaceId.value, credential.id)
        if (credentialList.value.length === 1 && pagination.value.current > 1) {
          pagination.value.current = pagination.value.current - 1
        }
        loadCredentialList()
      },
    } as any)
  }

  // 更改每页条数
  const handlePageLimitChange = (val: number) => {
    pagination.value.limit = val
    refreshListWithLoading()
  }

  const goToIAM = () => {
    window.open((<any>window).BK_IAM_HOST + '/apply-join-user-group', '__blank')
  }

</script>
<template>
  <section class="keys-management-page">
    <bk-alert theme="info">
      <div class="alert-tips">
        <p>密钥仅用于 SDK/API 拉取配置使用。服务管理/配置管理/分组管理等功能的权限申请，请前往</p>
        <bk-button text theme="primary" @click="goToIAM">蓝鲸权限中心</bk-button>
      </div>
    </bk-alert>
    <div class="management-data-container">
      <div class="operate-area">
        <bk-button theme="primary" :loading="createPending" @click="handleCreateCredential"><Plus class="button-icon" />新建密钥</bk-button>
        <div class="filter-actions">
          <bk-input
            v-model="searchStr"
            class="search-group-input"
            placeholder="状态/说明/更新人/更新时间"
            :clearable="true"
            @enter="refreshListWithLoading"
            @clear="refreshListWithLoading"
            @change="handleSearchInputChange">
            <template #suffix>
                <Search class="search-input-icon" />
            </template>
          </bk-input>
        </div>
      </div>
      <bk-loading style="min-height: 300px;" :loading="listLoading">
        <bk-table class="credential-table" :data="credentialList" :border="['outer']" :row-class="getRowCls">
          <bk-table-column label="密钥" width="340">
            <template #default="{ row }">
              <div v-if="row.spec" class="credential-text">
                <div class="text">{{ row.visible ? row.spec.enc_credential : '********************************' }}</div>
                <div class="actions">
                  <Eye v-if="!row.visible" class="view-icon" @click="row.visible = true"/>
                  <Unvisible v-else class="view-icon" @click="row.visible = false" />
                  <Copy class="copy-icon" @click="handleCopyText(row.spec.enc_credential)" />
                </div>
              </div>
            </template>
          </bk-table-column>
          <bk-table-column label="说明" prop="memo">
            <template #default="{ row }">
              <div v-if="row.spec" class="credential-memo">
                <div v-if="editingMemoId !== row.id" class="memo-content" :title="row.spec.memo || '--'" >{{ row.spec.memo || '--' }}</div>
                <div v-else class="memo-edit">
                  <div ref="memoInputRef" class="edit-input" contenteditable="true" @blur="handleMemoBlur(row)">{{ row.spec.memo }}</div>
                </div>
                <div class="edit-icon">
                  <EditLine @click="handleEditMemo(row.id)" />
                </div>
              </div>
            </template>
          </bk-table-column>
          <bk-table-column label="更新人" width="160" prop="revision.reviser"></bk-table-column>
          <bk-table-column label="更新时间" width="220">
            <template #default="{ row }">
              <span v-if="row.revision">{{ datetimeFormat(row.revision.update_at) }}</span>
            </template>
          </bk-table-column>
          <bk-table-column label="状态" width="110">
            <template #default="{ row }">
              <div v-if="row.spec" class="status-action">
                <bk-switcher size="small" theme="primary" :key="row.id" :value="row.spec.enable" @change="handelToggleEnable(row)"></bk-switcher>
                <span class="text">{{ row.spec.enable ? '已启用' : '已禁用' }}</span>
              </div>
            </template>
          </bk-table-column>
          <bk-table-column label="操作" width="140">
            <template #default="{ row }">
              <template v-if="row.spec">
                <bk-button text theme="primary" @click="handleOpenAssociate(row)">关联配置项</bk-button>
                <bk-button
                  style="margin-left: 8px;"
                  text
                  theme="primary"
                  :disabled="row.spec.enable"
                  @click="handleDelete(row)">
                  删除
                </bk-button>
              </template>
            </template>
          </bk-table-column>
        </bk-table>
        <bk-pagination
          class="table-list-pagination"
          v-model="pagination.current"
          location="left"
          :layout="['total', 'limit', 'list']"
          :count="pagination.count"
          :limit="pagination.limit"
          @change="refreshListWithLoading"
          @limit-change="handlePageLimitChange" />
      </bk-loading>
    </div>
    <AssociateConfigItems
      :show="isAssociateSliderShow"
      :id="currentCredential"
      @close="handleAssociateSliderClose"
      @refresh="refreshListWithLoading(pagination.current)" />
  </section>
</template>
<style lang="scss" scoped>
  .alert-tips {
    display: flex;
    > p {
      margin: 0;
      line-height: 20px;
    }
  }
  .keys-management-page {
    height: 100%;
    background: #f5f7fa;
  }
  .management-data-container {
    padding: 16px 24px 24px;
    height: calc(100% - 38px);
  }
  .operate-area {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 16px;
    .button-icon {
      font-size: 22px;
    }
  }
  .filter-actions {
    display: flex;
    align-items: center;
    .actions {
      > span {
        font-size: 14px;
      }
    }
  }
  .search-group-input {
    width: 320px;
  }
  .search-input-icon {
    padding-right: 10px;
    color: #979ba5;
    background: #ffffff;
  }
  .credential-table {
    overflow: visible !important;
    :deep(.bk-table-body) {
      tr.new-row-marked td {
        background: #f2fff4 !important;
      }
      tr.selected td {
        background: #e1ecff !important;
      }
    }
    :deep(.bk-table-body) {
      overflow: visible !important;
      tbody td .cell {
        overflow: visible !important;
      }
    }
  }
  .credential-text {
    display: flex;
    align-items: center;
    justify-content: space-between;
    .text {
      width: calc(100% - 80px);
    }
    .actions {
      display: flex;
      align-items: center;
      color: #979ba5;
      .view-icon {
        margin-right: 8px;
        font-size: 14px;
      }
      .copy-icon {
        font-size: 12px;
      }
      .view-icon,
      .copy-icon {
        cursor: pointer;
        &:hover {
          color: #3a84ff;
        }
      }
    }
  }
  .credential-memo {
    position: relative;
    padding-right: 20px;
    &:hover {
      .edit-icon {
        display: inline-block;
      }
    }
    .memo-content {
      width: 100%;
      overflow: hidden;
      white-space: nowrap;
      text-overflow: ellipsis;
    }
    .memo-edit {
      position: absolute;
      top: 2px;
      right: 0;
      bottom: 0;
      left: 0;
      z-index: 1;
    }
    .edit-input {
      padding: 6px 10px;
      min-height: 60px;
      line-height: 20px;
      font-size: 12px;
      color: #63656e;
      border: 1px solid #c4c6cc;
      border-radius: 2px;
      background: #ffffff;
      outline: none;
      -webkit-user-modify: read-write-plaintext-only;
      &:focus {
        border-color: #3a84ff;
        box-shadow: 0 0 3px #a3c5fd;
      }
    }
    .edit-icon  {
      position: absolute;
      top: 4px;
      right: 0;
      display: none;
      font-size: 14px;
      color: #979ba5;
      cursor: pointer;
      &:hover {
        color: #3a84ff;
      }
    }
  }
  .status-action {
    display: flex;
    align-items: center;
    .text {
      margin-left: 9px;
    }
  }
  .table-list-pagination {
    padding: 12px;
    border: 1px solid #dcdee5;
    border-top: none;
    border-radius: 0 0 2px 2px;
    background: #ffffff;
    :deep(.bk-pagination-list.is-last) {
      margin-left: auto;
    }
  }
</style>
