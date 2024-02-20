<template>
  <section class="keys-management-page">
    <bk-alert theme="info">
      <div class="alert-tips">
        <p>{{ t('密钥仅用于 SDK/API 拉取配置使用。服务管理/配置管理/分组管理等功能的权限申请，请前往') }}</p>
        <bk-button text theme="primary" @click="goToIAM">{{ t('蓝鲸权限中心') }}</bk-button>
      </div>
    </bk-alert>
    <div class="management-data-container">
      <div class="operate-area">
        <bk-button
          v-cursor="{ active: !hasManagePerm }"
          theme="primary"
          :class="{ 'bk-button-with-no-perm': !hasManagePerm }"
          :disabled="permCheckLoading"
          :loading="createPending"
          @click="getCredentialName">
          <Plus class="button-icon" />
          {{ t('新建密钥') }}
        </bk-button>
        <div class="filter-actions">
          <bk-input
            v-model="searchStr"
            class="search-group-input"
            :placeholder="t('密钥名称/说明/更新人')"
            :clearable="true"
            @clear="refreshListWithLoading()"
            @input="handleSearchInputChange">
            <template #suffix>
              <Search class="search-input-icon" />
            </template>
          </bk-input>
        </div>
      </div>
      <bk-loading style="min-height: 100px" :loading="listLoading">
        <bk-table
          class="credential-table"
          :data="tableData"
          :border="['outer']"
          :row-class="getRowCls"
          :remote-pagination="true"
          :pagination="pagination"
          @page-limit-change="handlePageLimitChange"
          @page-value-change="refreshListWithLoading">
          <bk-table-column :label="t('密钥名称')" width="188">
            <template #default="{ row, index }">
              <bk-input
                v-if="index === 0 && isCreateCredential"
                :placeholder="t('密钥名称支持中英文')"
                v-model="createCredentialName"
                @blur="testCreateCredentialName"></bk-input>
              <div v-if="row.spec" class="credential-memo">
                <div v-if="editingNameId !== row.id" class="memo-content" :title="row.spec.memo || '--'">
                  {{ row.spec.name || '--' }}
                </div>
                <div v-else class="memo-edit">
                  <div
                    ref="nameInputRef"
                    class="edit-name-input"
                    contenteditable="true"
                    @blur="handleMemoOrNameBlur(row)">
                    {{ row.spec.name }}
                  </div>
                </div>
                <div class="edit-icon">
                  <EditLine @click="handleEditName(row.id)" />
                </div>
              </div>
            </template>
          </bk-table-column>
          <bk-table-column :label="t('密钥')" width="340">
            <template #default="{ row, index }">
              <span v-if="index === 0 && isCreateCredential" style="color: #c4c6cc">{{ t('待确认') }}</span>
              <div v-if="row.spec" class="credential-text">
                <div class="text">
                  <span v-if="!row.visible">************</span>
                  <bk-overflow-title v-else type="tips">{{ row.spec.enc_credential }}</bk-overflow-title>
                </div>
                <div class="actions">
                  <Eye v-if="row.visible" class="view-icon" @click="row.visible = false" />
                  <Unvisible v-else class="view-icon" @click="row.visible = true" />
                  <Copy class="copy-icon" @click="handleCopyText(row.spec.enc_credential)" />
                </div>
              </div>
            </template>
          </bk-table-column>
          <bk-table-column :label="t('说明')" prop="memo">
            <template #default="{ row, index }">
              <bk-input
                v-if="index === 0 && isCreateCredential"
                :placeholder="t('请输入密钥说明')"
                v-model="createCredentialMemo"></bk-input>
              <div v-if="row.spec" class="credential-memo">
                <div v-if="editingMemoId !== row.id" class="memo-content" :title="row.spec.memo || '--'">
                  {{ row.spec.memo || '--' }}
                </div>
                <div v-else class="memo-edit">
                  <div
                    ref="memoInputRef"
                    class="edit-input"
                    contenteditable="true"
                    @blur="handleMemoOrNameBlur(row, false)">
                    {{ row.spec.memo }}
                  </div>
                </div>
                <div class="edit-icon">
                  <EditLine @click="handleEditMemo(row.id)" />
                </div>
              </div>
            </template>
          </bk-table-column>
          <bk-table-column :label="t('关联规则')" width="140">
            <template #default="{ row }">
              <bk-popover v-if="row.rule && row.rule.length" theme="light" :popover-delay="[300, 0]">
                <div class="table-rule">
                  {{ row.rule[0].spec.app + row.rule[0].spec.scope }}
                </div>
                <template #content>
                  <div v-for="rule in row.rule" :key="rule.id">
                    {{ rule.spec.app + rule.spec.scope }}
                  </div>
                </template>
              </bk-popover>
              <span v-else>--</span>
            </template>
          </bk-table-column>
          <bk-table-column :label="t('更新人')" width="88" prop="revision.reviser"></bk-table-column>
          <bk-table-column :label="t('更新时间')" width="154">
            <template #default="{ row }">
              <span v-if="row.revision">{{ datetimeFormat(row.revision.update_at) }}</span>
            </template>
          </bk-table-column>
          <!-- <bk-table-column label="最近使用时间" width="154">
            <template #default="{ row }">
              <span v-if="row.revision">{{ datetimeFormat(row.revision.update_at) }}</span>
            </template>
          </bk-table-column> -->
          <bk-table-column :label="t('状态')" width="110">
            <template #default="{ row }">
              <div v-if="row.spec" class="status-action">
                <bk-switcher
                  v-cursor="{ active: !hasManagePerm }"
                  size="small"
                  theme="primary"
                  :key="row.id"
                  :value="row.spec.enable"
                  :disabled="permCheckLoading"
                  :class="{ 'bk-switcher-with-no-perm': !hasManagePerm }"
                  @change="handelToggleEnable(row)" />
                <span class="text">{{ row.spec.enable ? t('已启用') : t('已禁用') }}</span>
              </div>
            </template>
          </bk-table-column>
          <bk-table-column :label="t('操作')" :width="locale === 'zh-CN' ? '160' : '260'" :fixed="'right'">
            <template #default="{ row, index }">
              <template v-if="index === 0 && isCreateCredential">
                <bk-button text theme="primary" @click="handleCreateCredential">{{ t('创建') }}</bk-button>
                <bk-button text theme="primary" style="margin-left: 8px" @click="handleCancelCreateCredential">
                  {{ t('取消') }}
                </bk-button>
              </template>
              <template v-if="row.spec">
                <bk-button text theme="primary" @click="handleOpenAssociate(row)">
                  <span :class="{ redPoint: newCredential === row.id }">{{ t('关联服务配置') }}</span>
                </bk-button>
                <div class="delete-btn" v-bk-tooltips="deleteTooltip(hasManagePerm && row.spec.enable)">
                  <bk-button
                    v-cursor="{ active: !hasManagePerm }"
                    style="margin-left: 8px"
                    text
                    theme="primary"
                    :class="{ 'bk-text-with-no-perm': !hasManagePerm }"
                    :disabled="hasManagePerm && row.spec.enable"
                    @click="handleDeleteConfirm(row)">
                    {{ t('删除') }}
                  </bk-button>
                </div>
              </template>
            </template>
          </bk-table-column>
          <template #empty>
            <table-empty :is-search-empty="isSearchEmpty" @clear="clearSearchStr" />
          </template>
        </bk-table>
      </bk-loading>
    </div>
    <AssociateConfigItems
      :show="isAssociateSliderShow"
      :id="currentCredential"
      :perm-check-loading="permCheckLoading"
      :has-manage-perm="hasManagePerm"
      @close="handleAssociateSliderClose"
      @refresh="refreshListWithLoading(pagination.current)"
      @apply-perm="checkPermBeforeOperate" />
  </section>
  <bk-dialog
    ext-cls="delete-service-dialog"
    v-model:is-show="isShowDeleteDialog"
    :theme="'primary'"
    :dialog-type="'operation'"
    header-align="center"
    footer-align="center"
    @value-change="dialogInputStr = ''">
    <div class="dialog-content">
      <div class="dialog-title">{{ t('确认删除密钥？') }}</div>
      <div class="dialog-input">
        <div class="dialog-info">
          <div>
            {{ t('删除的密钥') }}
            <span>{{ t('无法找回') }}</span>
            {{ t(',请谨慎操作！') }}
          </div>
        </div>
        <div class="tips">
          {{ t('请输入密钥名称') }} <span>{{ deleteCredentialInfo?.spec.name }}</span> {{ t('以确认删除') }}
        </div>
        <bk-input v-model="dialogInputStr" :placeholder="t('请输入')" />
      </div>
    </div>
    <template #footer>
      <div class="dialog-footer">
        <bk-button
          theme="danger"
          style="margin-right: 20px"
          :disabled="dialogInputStr !== deleteCredentialInfo?.spec.name"
          @click="handleDelete">
          {{ t('删除') }}
        </bk-button>
        <bk-button @click="isShowDeleteDialog = false">{{ t('取消') }}</bk-button>
      </div>
    </template>
  </bk-dialog>
</template>
<script setup lang="ts">
  import { ref, watch, onMounted, nextTick } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { storeToRefs } from 'pinia';
  import useGlobalStore from '../../../store/global';
  import { Plus, Search, Eye, Unvisible, Copy, EditLine } from 'bkui-vue/lib/icon';
  import BkMessage from 'bkui-vue/lib/message';
  import { InfoBox } from 'bkui-vue';
  import { permissionCheck } from '../../../api/index';
  import {
    getCredentialList,
    createCredential,
    updateCredential,
    deleteCredential,
    getCredentialScopes,
  } from '../../../api/credentials';
  import { copyToClipBoard, datetimeFormat } from '../../../utils/index';
  import { ICredentialItem } from '../../../../types/credential';
  import AssociateConfigItems from './associate-config-items/index.vue';
  import tableEmpty from '../../../components/table/table-empty.vue';
  import { debounce } from 'lodash';

  const { spaceId, permissionQuery, showApplyPermDialog } = storeToRefs(useGlobalStore());
  const { t, locale } = useI18n();

  const permCheckLoading = ref(false);
  const hasManagePerm = ref(false);
  const credentialList = ref<ICredentialItem[]>([]);
  const listLoading = ref(false);
  const createPending = ref(false);
  const createCredentialName = ref('');
  const createCredentialMemo = ref('');
  const isCreateCredential = ref(false);
  const newCredential = ref(0); // 记录新增加的密钥id，实现表格标记效果
  const searchStr = ref('');
  const editingMemoId = ref(0); // 记录当前正在编辑说明的密钥id
  const editingNameId = ref(0); // 记录当前正在编辑名称的密钥id
  const memoInputRef = ref();
  const nameInputRef = ref();
  const isAssociateSliderShow = ref(false);
  const currentCredential = ref(0);
  const isSearchEmpty = ref(false);
  const pagination = ref({
    current: 1,
    count: 0,
    limit: 10,
  });
  const tableData = ref<any>([]);
  const isShowDeleteDialog = ref(false);
  const dialogInputStr = ref('');
  const deleteCredentialInfo = ref<ICredentialItem>();

  watch(
    () => spaceId.value,
    () => {
      createPending.value = false;
      getPermData();
      refreshListWithLoading();
    },
  );

  onMounted(() => {
    getPermData();
    refreshListWithLoading();
  });

  const getPermData = async () => {
    permCheckLoading.value = true;
    const res = await permissionCheck({
      resources: [
        {
          biz_id: spaceId.value,
          basic: {
            type: 'credential',
            action: 'manage',
          },
        },
      ],
    });
    hasManagePerm.value = res.is_allowed;
    permCheckLoading.value = false;
  };

  const checkPermBeforeOperate = () => {
    if (!hasManagePerm.value) {
      permissionQuery.value = {
        resources: [
          {
            biz_id: spaceId.value,
            basic: {
              type: 'credential',
              action: 'manage',
            },
          },
        ],
      };
      showApplyPermDialog.value = true;
      return false;
    }
    return true;
  };

  // 加载密钥列表
  const loadCredentialList = async () => {
    const query: { limit: number; start: number; searchKey?: string; top_ids?: number } = {
      start: pagination.value.limit * (pagination.value.current - 1),
      limit: pagination.value.limit,
    };
    if (searchStr.value) {
      query.searchKey = searchStr.value;
    }
    if (newCredential.value) {
      query.top_ids = newCredential.value;
    }
    const res = await getCredentialList(spaceId.value, query);
    res.details.forEach((item: ICredentialItem) => (item.visible = false));
    credentialList.value = res.details;
    tableData.value = res.details;
    pagination.value.count = res.count;
    // 获取密钥关联规则
    tableData.value.forEach(async (item: any) => {
      const res = await getCredentialScopes(spaceId.value, item.id);
      item.rule = res.details;
    });
  };

  // 更新列表数据，带loading效果
  const refreshListWithLoading = async (current = 1) => {
    // 创建新密钥时，页码会切换会首页，此时不另外发请求
    if (createPending.value) {
      return;
    }
    searchStr.value ? (isSearchEmpty.value = true) : (isSearchEmpty.value = false);
    listLoading.value = true;
    pagination.value.current = current;
    await loadCredentialList();
    listLoading.value = false;
    isCreateCredential.value = false;
  };

  // 设置新增行的标记class
  const getRowCls = (data: ICredentialItem) => {
    if (newCredential.value === data.id) {
      return 'new-row-marked';
    }
    if (currentCredential.value === data.id) {
      return 'selected';
    }
    return '';
  };

  // 复制
  const handleCopyText = (text: string) => {
    copyToClipBoard(text);
    BkMessage({
      theme: 'success',
      message: t('服务密钥已复制'),
    });
  };

  // 创建密钥之前获取密钥名称
  const getCredentialName = async () => {
    if (isCreateCredential.value || !checkPermBeforeOperate()) return;
    isCreateCredential.value = true;
    tableData.value.unshift({});
  };

  // 创建密钥
  const handleCreateCredential = async () => {
    if (!createCredentialName.value) {
      BkMessage({
        theme: 'error',
        message: t('请输入密钥名称'),
      });
      return;
    }
    await testCreateCredentialName();
    try {
      createPending.value = true;
      const params = { memo: createCredentialMemo.value, name: createCredentialName.value };
      const res = await createCredential(spaceId.value, params);
      BkMessage({
        theme: 'success',
        message: t('新建服务密钥成功'),
      });
      pagination.value.current = 1;
      newCredential.value = res.id;
      await loadCredentialList();
    } catch (e) {
      console.error(e);
    } finally {
      createPending.value = false;
      handleCancelCreateCredential();
    }
  };

  // 取消创建密钥
  const handleCancelCreateCredential = () => {
    if (!tableData.value[0].id) {
      tableData.value.shift();
    }
    isCreateCredential.value = false;
    createCredentialMemo.value = '';
    createCredentialName.value = '';
  };

  // 搜索内容改变 触发搜索
  const handleSearchInputChange = debounce(() => refreshListWithLoading(), 300);

  // 密钥说明编辑
  const handleEditMemo = (id: number) => {
    editingMemoId.value = id;
    nextTick(() => {
      if (memoInputRef.value) {
        memoInputRef.value.focus();
      }
    });
  };

  // 密钥名称编辑
  const handleEditName = (id: number) => {
    editingNameId.value = id;
    nextTick(() => {
      if (nameInputRef.value) {
        nameInputRef.value.focus();
      }
    });
  };

  // 失焦时保存密钥说明或密钥名称
  const handleMemoOrNameBlur = async (credential: ICredentialItem, isEditName = true) => {
    const params = {
      id: credential.id,
      enable: credential.spec.enable,
      name: credential.spec.name,
      memo: credential.spec.memo,
    };
    if (isEditName) {
      editingNameId.value = 0;
      const name = nameInputRef.value.textContent.trim();
      if (credential.spec.name === name) {
        return;
      }
      params.name = name;
    } else {
      editingMemoId.value = 0;
      const memo = memoInputRef.value.textContent.trim();
      if (credential.spec.memo === memo) {
        return;
      }
      params.memo = memo;
    }
    await updateCredential(spaceId.value, params);
    credential.spec.memo = params.memo;
    credential.spec.name = params.name;
    BkMessage({
      theme: 'success',
      message: isEditName ? t('密钥名称修改成功') : t('密钥说明修改成功'),
    });
  };

  // 禁用/启用
  const handelToggleEnable = async (credential: ICredentialItem) => {
    if (!checkPermBeforeOperate()) {
      return;
    }
    if (credential.spec.enable) {
      InfoBox({
        title: t('确定禁用此密钥'),
        subTitle: t('禁用密钥后，使用此密钥的应用将无法正常使用 SDK/API 拉取配置'),
        'ext-cls': 'info-box-style',
        confirmText: t('禁用'),
        onConfirm: async () => {
          const params = {
            id: credential.id,
            memo: credential.spec.memo,
            enable: false,
            name: credential.spec.name,
          };
          await updateCredential(spaceId.value, params);
          BkMessage({
            theme: 'success',
            message: t('禁用成功'),
          });
          credential.spec.enable = false;
        },
      } as any);
    } else {
      const params = {
        id: credential.id,
        memo: credential.spec.memo,
        enable: true,
        name: credential.spec.name,
      };
      await updateCredential(spaceId.value, params);
      credential.spec.enable = true;
      BkMessage({
        theme: 'success',
        message: t('启用成功'),
      });
    }
  };

  // 打开关联配置文件侧滑
  const handleOpenAssociate = (credential: ICredentialItem) => {
    isAssociateSliderShow.value = true;
    currentCredential.value = credential.id;
  };

  // 关闭关联配置文件侧滑
  const handleAssociateSliderClose = () => {
    isAssociateSliderShow.value = false;
    currentCredential.value = 0;
  };

  // 删除配置文件二次确认
  const handleDeleteConfirm = async (credential: ICredentialItem) => {
    isShowDeleteDialog.value = true;
    deleteCredentialInfo.value = credential;
  };
  // 删除配置文件
  const handleDelete = async () => {
    if (!checkPermBeforeOperate()) {
      return;
    }
    await deleteCredential(spaceId.value, deleteCredentialInfo.value?.id as number);
    BkMessage({
      theme: 'success',
      message: t('删除服务密钥成功'),
    });
    if (credentialList.value.length === 1 && pagination.value.current > 1) {
      pagination.value.current = pagination.value.current - 1;
    }
    isShowDeleteDialog.value = false;
    loadCredentialList();
  };
  // 删除配置文件提示文字
  const deleteTooltip = (isShowTooltip: boolean) => {
    if (isShowTooltip) {
      return {
        content: t('已启用，不能删除'),
        placement: 'top',
      };
    }
    return {
      disabled: true,
    };
  };

  // 更改每页条数
  const handlePageLimitChange = (val: number) => {
    pagination.value.limit = val;
    refreshListWithLoading();
  };

  // 清空搜索框
  const clearSearchStr = () => {
    searchStr.value = '';
    refreshListWithLoading();
  };

  // 校验新建密钥名称
  const testCreateCredentialName = () => {
    if (!createCredentialName.value) return;
    const regex = /^[\u4e00-\u9fa5a-zA-Z0-9][\u4e00-\u9fa5a-zA-Z0-9_-]*[\u4e00-\u9fa5a-zA-Z0-9]$/;
    if (!regex.test(createCredentialName.value)) {
      BkMessage({
        theme: 'error',
        message: `${t('无效名称')}：${createCredentialName.value}，${t(
          '只允许包含中文、英文、数字、下划线 (_)、连字符 (-)，并且必须以中文、英文、数字开头和结尾。',
        )}`,
      });
      return Promise.reject();
    }
  };
  const goToIAM = () => {
    window.open(`${(window as any).BK_IAM_HOST}/apply-join-user-group`, '__blank');
  };
</script>
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
    .delete-btn {
      display: inline-block;
    }
  }
  .credential-text {
    display: flex;
    .text {
      width: 300px;
    }
    .actions {
      margin-left: 16px;
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
    display: flex;
    align-items: center;
    &:hover {
      .edit-icon {
        display: inline-block;
      }
    }
    .memo-content {
      max-width: calc(100% - 40px);
      overflow: hidden;
      white-space: nowrap;
      text-overflow: ellipsis;
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
    .edit-name-input {
      @extend .edit-input;
      min-height: 32px;
    }
    .edit-icon {
      display: none;
      padding-left: 16px;
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
  .dialog-content {
    text-align: center;
    margin-top: 48px;
    .dialog-title {
      font-size: 20px;
      color: #313238;
      line-height: 32px;
    }
    .dialog-input {
      margin-top: 16px;
      text-align: start;
      padding: 20px;
      background-color: #f4f7fa;
      .dialog-info {
        margin-bottom: 16px;
        span {
          color: red;
        }
      }
      .tips {
        margin-bottom: 8px;
        span {
          font-weight: bolder;
        }
      }
    }
  }
  .dialog-footer {
    .bk-button {
      width: 100px;
    }
  }
  .table-rule {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
</style>

<style lang="scss">
  .delete-service-dialog {
    top: 40% !important;
    .bk-modal-body {
      padding-bottom: 104px !important;
    }
    .bk-modal-header {
      display: none;
    }
    .bk-modal-footer {
      height: auto !important;
      background-color: #fff !important;
      border-top: none !important;
      padding: 24px 24px 48px !important;
    }
  }
</style>
