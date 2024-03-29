<template>
  <section class="variables-management-page">
    <bk-alert theme="info">
      {{
        t('定义全局变量后可供业务下所有的服务配置文件引用，使用go template语法引用，例如,变量使用详情请参考：', {
          var: '\{\{ .bk_bscp_appid \}\}',
        })
      }}
      <span @click="goVariablesDoc" class="hyperlink">{{ t('配置模板与变量') }}</span>
    </bk-alert>
    <div class="operation-area">
      <div class="button">
        <bk-button theme="primary" @click="isCreateSliderShow = true">
          <Plus class="button-icon" />
          {{ t('新增变量') }}
        </bk-button>
        <bk-button @click="isImportVariableShow = true">{{ t('导入变量') }}</bk-button>
      </div>
      <SearchInput v-model="searchStr" :placeholder="t('请输入变量名称')" :width="320" @search="refreshList()" />
    </div>
    <div class="variable-table">
      <bk-table
        :border="['outer']"
        :data="list"
        :remote-pagination="true"
        :pagination="pagination"
        @page-limit-change="handlePageLimitChange"
        @page-value-change="refreshList">
        <bk-table-column :label="t('变量名称')" width="300">
          <template #default="{ row }">
            <div v-if="row.spec" class="var-name-wrapper">
              <bk-overflow-title class="name-text" type="tips" :key="row.id" @click="handleEditVar(row)">
                {{ row.spec.name }}
              </bk-overflow-title>
              <Copy class="copy-icon" @click="handleCopyText(row.spec.name)" />
            </div>
          </template>
        </bk-table-column>
        <bk-table-column :label="t('类型')" prop="spec.type" width="180"></bk-table-column>
        <bk-table-column :label="t('默认值')" prop="spec.default_val"></bk-table-column>
        <bk-table-column :label="t('描述')">
          <template #default="{ row }">
            <span v-if="row.spec">{{ row.spec.memo || '--' }}</span>
          </template>
        </bk-table-column>
        <bk-table-column :label="t('操作')" width="240">
          <template #default="{ row }">
            <div class="action-btns">
              <bk-button text theme="primary" @click="handleEditVar(row)">{{ t('编辑') }}</bk-button>
              <bk-button text theme="primary" @click="handleDeleteVar(row)">{{ t('删除') }}</bk-button>
            </div>
          </template>
        </bk-table-column>
        <template #empty>
          <TableEmpty :is-search-empty="isSearchEmpty" @clear="clearSearchStr"></TableEmpty>
        </template>
      </bk-table>
    </div>
    <VariableCreate v-model:show="isCreateSliderShow" @created="refreshList" />
    <VariableEdit
      v-model:show="editSliderData.open"
      :id="editSliderData.id"
      :data="editSliderData.data"
      @edited="refreshList" />
    <VariableImport v-model:show="isImportVariableShow" @edited="refreshList" />
  </section>
  <DeleteConfirmDialog
    v-model:isShow="isDeleteVariableDialogShow"
    :title="t('确认删除该全局变量？')"
    @confirm="handleDeleteVarConfirm">
    <div style="margin-bottom: 8px">
      {{ t('全局变量') }}: <span style="color: #313238; font-weight: 600">{{ deleteVariableItem?.spec.name }}</span>
    </div>
    <div>{{ t('一旦删除，该操作将无法撤销，服务配置文件中不可再引用该全局变量，请谨慎操作') }}</div>
  </DeleteConfirmDialog>
</template>
<script lang="ts" setup>
  import { onMounted, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { storeToRefs } from 'pinia';
  import { Plus, Copy } from 'bkui-vue/lib/icon';
  import BkMessage from 'bkui-vue/lib/message';
  import useGlobalStore from '../../../store/global';
  import { ICommonQuery, IPagination } from '../../../../types/index';
  import { IVariableEditParams, IVariableItem } from '../../../../types/variable';
  import { getVariableList, deleteVariable } from '../../../api/variable';
  import { copyToClipBoard } from '../../../utils/index';
  import VariableCreate from './variable-create.vue';
  import VariableEdit from './variable-edit.vue';
  import VariableImport from './variable-import.vue';
  import SearchInput from '../../../components/search-input.vue';
  import TableEmpty from '../../../components/table/table-empty.vue';
  import DeleteConfirmDialog from '../../../components/delete-confirm-dialog.vue';

  const { spaceId } = storeToRefs(useGlobalStore());
  const { t } = useI18n();

  const loading = ref(false);
  const list = ref<IVariableItem[]>([]);
  const searchStr = ref('');
  const pagination = ref<IPagination>({
    current: 1,
    count: 0,
    limit: 10,
  });
  const isCreateSliderShow = ref(false);
  const isImportVariableShow = ref(false);
  const isDeleteVariableDialogShow = ref(false);
  const deleteVariableItem = ref<IVariableItem>();
  const editSliderData = ref<{ open: boolean; id: number; data: IVariableEditParams }>({
    open: false,
    id: 0,
    data: {
      name: '',
      type: '',
      default_val: '',
      memo: '',
    },
  });
  const isSearchEmpty = ref(false);
  watch(
    () => spaceId.value,
    () => {
      refreshList();
    },
  );

  onMounted(() => {
    getVariables();
  });

  const getVariables = async () => {
    loading.value = true;
    const params: ICommonQuery = {
      start: (pagination.value.current - 1) * pagination.value.limit,
      limit: pagination.value.limit,
    };
    if (searchStr.value) {
      params.search_fields = 'name';
      params.search_value = searchStr.value;
    }
    const res = await getVariableList(spaceId.value, params);
    list.value = res.details;
    pagination.value.count = res.count;
    loading.value = false;
  };

  const handleEditVar = (variable: IVariableItem) => {
    const { id, spec } = variable;
    editSliderData.value = {
      open: true,
      id,
      data: { ...spec },
    };
  };

  // 复制
  const handleCopyText = (name: string) => {
    copyToClipBoard(`{{ .${name} }}`);
    BkMessage({
      theme: 'success',
      message: `${t('引用方式')} {{ .${name} }} ${t('已成功复制到剪贴板')}`,
    });
  };

  // 删除变量
  const handleDeleteVar = (variable: IVariableItem) => {
    isDeleteVariableDialogShow.value = true;
    deleteVariableItem.value = variable;
  };

  const handleDeleteVarConfirm = async () => {
    await deleteVariable(spaceId.value, deleteVariableItem.value!.id);
    BkMessage({
      message: t('删除变量成功'),
      theme: 'success',
    });
    if (list.value.length === 1 && pagination.value.current > 1) {
      pagination.value.current = pagination.value.current - 1;
    }
    isDeleteVariableDialogShow.value = false;
    getVariables();
  };

  const handlePageLimitChange = (val: number) => {
    pagination.value.limit = val;
    refreshList();
  };

  const refreshList = (current = 1) => {
    isSearchEmpty.value = searchStr.value !== '';
    pagination.value.current = current;
    getVariables();
  };

  const clearSearchStr = () => {
    searchStr.value = '';
    refreshList();
  };

  // @ts-ignore
  // eslint-disable-next-line
  const goVariablesDoc = () => window.open(BSCP_CONFIG.variable_template_doc);
</script>
<style lang="scss" scoped>
  .variables-management-page {
    height: 100%;
    background: #f5f7fa;
    .hyperlink {
      color: #3a84ff;
      cursor: pointer;
    }
  }
  .var-name-wrapper {
    position: relative;
    padding-right: 20px;
    .name-text {
      color: #3a84ff;
      cursor: pointer;
    }
    .copy-icon {
      position: absolute;
      top: 15px;
      right: 4px;
      font-size: 12px;
      color: #979ba5;
      cursor: pointer;
      &:hover {
        color: #3a84ff;
      }
    }
  }
  .operation-area {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-top: 24px;
    padding: 0 24px;
    .button {
      display: flex;
      align-items: center;
      justify-content: space-between;
      .bk-button {
        margin-right: 8px;
      }
      .button-icon {
        font-size: 18px;
      }
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
  .action-btns {
    .bk-button:not(:last-of-type) {
      margin-right: 8px;
    }
  }
</style>
