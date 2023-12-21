<template>
  <section class="variables-management-page">
    <bk-alert theme="info">{{ headInfo }}<span @click="goVariablesDoc" class="hyperlink">配置模板与变量</span>
    </bk-alert>
    <div class="operation-area">
      <div class="button">
        <bk-button theme="primary" @click="isCreateSliderShow = true"><Plus class="button-icon" />新增变量</bk-button>
        <bk-button @click="isImportVariableShow = true">导入变量</bk-button>
      </div>
      <SearchInput v-model="searchStr" placeholder="请输入变量名称" :width="320" @search="refreshList()" />
    </div>
    <div class="variable-table">
      <bk-table
        :border="['outer']"
        :data="list"
        :remote-pagination="true"
        :pagination="pagination"
        @page-limit-change="handlePageLimitChange"
        @page-value-change="refreshList"
      >
        <bk-table-column label="变量名称">
          <template #default="{ row }">
            <bk-button v-if="row.spec" text theme="primary" @click="handleEditVar(row)">{{ row.spec.name }}</bk-button>
          </template>
        </bk-table-column>
        <bk-table-column label="类型" prop="spec.type" width="120"></bk-table-column>
        <bk-table-column label="默认值" prop="spec.default_val"></bk-table-column>
        <bk-table-column label="描述">
          <template #default="{ row }">
            <span v-if="row.spec">{{ row.spec.memo || '--' }}</span>
          </template>
        </bk-table-column>
        <bk-table-column label="操作" width="200">
          <template #default="{ row }">
            <div class="action-btns">
              <bk-button text theme="primary" @click="handleEditVar(row)">编辑</bk-button>
              <bk-button text theme="primary" @click="handleDeleteVar(row)">删除</bk-button>
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
      @edited="refreshList"
    />
    <VariableImport v-model:show="isImportVariableShow" @edited="refreshList" />
  </section>
  <DeleteConfirmDialog
    v-model:isShow="isDeleteVariableDialogShow"
    title="确认删除该全局变量？"
    @confirm="handleDeleteVarConfirm"
  >
    <div style="margin-bottom: 8px;">
      全局变量: <span style="color: #313238;font-weight: 600;">{{ deleteVariableItem?.spec.name }}</span>
    </div>
    <div>一旦删除，该操作将无法撤销，服务配置文件中不可再引用该全局变量，请谨慎操作</div>
  </DeleteConfirmDialog>
</template>
<script lang="ts" setup>
import { onMounted, ref, watch } from 'vue';
import { storeToRefs } from 'pinia';
import { Plus } from 'bkui-vue/lib/icon';
import useGlobalStore from '../../../store/global';
import { ICommonQuery, IPagination } from '../../../../types/index';
import { IVariableEditParams, IVariableItem } from '../../../../types/variable';
import { getVariableList, deleteVariable } from '../../../api/variable';
import VariableCreate from './variable-create.vue';
import VariableEdit from './variable-edit.vue';
import VariableImport from './variable-import.vue';
import SearchInput from '../../../components/search-input.vue';
import TableEmpty from '../../../components/table/table-empty.vue';
import DeleteConfirmDialog from '../../../components/delete-confirm-dialog.vue';
import Message from 'bkui-vue/lib/message';

const { spaceId } = storeToRefs(useGlobalStore());

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
const headInfo = `定义全局变量后可供业务下所有的服务配置文件引用，使用go template语法引用，例如{{ .bk_bscp_appid }},
      变量使用详情请参考：`;
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

// 删除变量
const handleDeleteVar = (variable: IVariableItem) => {
  isDeleteVariableDialogShow.value = true;
  deleteVariableItem.value = variable;
};

const handleDeleteVarConfirm = async () => {
  await deleteVariable(spaceId.value, deleteVariableItem.value!.id);
  Message({
    message: '删除变量成功',
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
  searchStr.value ? (isSearchEmpty.value = true) : (isSearchEmpty.value = false);
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
.operation-area {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 24px;
  padding: 0 24px;
  .button {
    width: 200px;
    display: flex;
    align-items: center;
    justify-content: space-between;
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
