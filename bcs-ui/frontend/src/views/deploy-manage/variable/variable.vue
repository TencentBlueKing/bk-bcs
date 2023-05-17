<template>
  <LayoutContent hide-back :title="$tc('变量管理')">
    <template #header-right>
      <bcs-button text @click="showSideslider = true">{{$t('如何从文件导入变量？')}}</bcs-button>
    </template>
    <LayoutRow class="mb15">
      <template #left>
        <bcs-button icon="plus" theme="primary" @click="handleAddVariable">{{$t('新增变量')}}</bcs-button>
        <bcs-button class="ml10" :loading="fileLoading">
          {{$t('文件导入')}}
          <input type="file" accept=".json" class="file-input" @change="handleFileChange" />
        </bcs-button>
        <bcs-button class="ml10" :disabled="!selections.length" @click="handleBatchDelete">{{$t('批量删除')}}</bcs-button>
      </template>
      <template #right>
        <bcs-select class="mw200" :clearable="false" :placeholder="' '" v-model="scope">
          <bcs-option id="" :name="$t('全部')"></bcs-option>
          <bcs-option v-for="item in scopeList" :key="item.id" :id="item.id" :name="item.name"></bcs-option>
        </bcs-select>
        <bcs-input
          right-icon="bk-icon icon-search"
          class="ml5 mw320"
          :placeholder="$t('输入Key搜索')"
          clearable
          v-model="searchKey">
        </bcs-input>
      </template>
    </LayoutRow>
    <bcs-table
      v-bkloading="{ isLoading }"
      :data="variableList"
      :pagination="pagination"
      @page-change="handlePageChange"
      @page-limit-change="handlePageLimitChange"
      @selection-change="handleSelectionChange">
      <bcs-table-column type="selection" width="60" :selectable="selectable"></bcs-table-column>
      <bcs-table-column :label="$t('变量名称')" prop="name" width="200"></bcs-table-column>
      <bcs-table-column label="KEY" prop="key"></bcs-table-column>
      <bcs-table-column :label="$t('默认值')" prop="defaultValue">
        <template #default="{ row }">
          {{row.defaultValue || '--'}}
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('描述')" prop="desc">
        <template #default="{ row }">
          {{row.desc || '--'}}
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('类型')" prop="categoryName" width="120">
        <template #default="{ row }">
          {{row.categoryName || '--'}}
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('作用范围')" prop="scopeName" width="140">
        <template #default="{ row }">
          {{row.scopeName || '--'}}
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('操作')" width="160">
        <template #default="{ row }">
          <span
            class="mr10"
            v-bk-tooltips="{
              content: $t('系统内置变量'),
              disabled: row.category !== 'sys'
            }">
            <bcs-button
              text
              :disabled="row.category === 'sys' || row.scope === 'global'"
              @click="handleSetVariable(row)">
              {{$t('设置')}}
            </bcs-button>
          </span>
          <span
            class="mr10"
            v-bk-tooltips="{
              content: $t('系统内置变量'),
              disabled: row.category !== 'sys'
            }">
            <bcs-button text :disabled="row.category === 'sys'" @click="handleEdit(row)">{{$t('编辑')}}</bcs-button>
          </span>
          <span
            v-bk-tooltips="{
              content: $t('系统内置变量'),
              disabled: row.category !== 'sys'
            }">
            <bcs-button text :disabled="row.category === 'sys'" @click="handleDelete(row)">{{$t('删除')}}</bcs-button>
          </span>
        </template>
      </bcs-table-column>
      <template #empty>
        <BcsEmptyTableStatus :type="searchKey ? 'search-empty' : 'empty'" @clear="searchKey = ''" />
      </template>
    </bcs-table>
    <!-- 新增变量 OR 编辑 -->
    <bcs-dialog
      width="640"
      v-model="showCreateOrEdit"
      :title="currentRow ? $t('编辑变量') : $t('新增变量')"
      :auto-close="false"
      header-position="left"
      :mask-close="false"
      :loading="dialogLoading"
      render-directive="if"
      :confirm-fn="handleConfirmDialog">
      <BkForm
        :label-width="100"
        :model="formData"
        :rules="formRules"
        ref="formRef">
        <BkFormItem :label="$t('作用域范围')">
          <bk-radio-group v-model="formData.scope">
            <bk-radio
              v-for="item in scopeList"
              :key="item.id"
              :disabled="!!currentRow"
              :value="item.id">
              {{item.name}}
            </bk-radio>
          </bk-radio-group>
        </BkFormItem>
        <BkFormItem :label="$t('名称')" property="name" required>
          <bcs-input v-model="formData.name"></bcs-input>
        </BkFormItem>
        <BkFormItem :label="$t('KEY')" property="key" required>
          <bcs-input v-model="formData.key" :disabled="!!currentRow"></bcs-input>
        </BkFormItem>
        <BkFormItem :label="$t('默认值')">
          <bcs-input v-model="formData.default"></bcs-input>
        </BkFormItem>
        <BkFormItem :label="$t('描述')">
          <bcs-input v-model="formData.desc" type="textarea" maxlength="255"></bcs-input>
        </BkFormItem>
      </BkForm>
    </bcs-dialog>
    <!-- 导入变量说明 -->
    <bcs-sideslider
      :is-show.sync="showSideslider"
      quick-close
      :width="800"
      :title="$t('如何从文件导入变量？')">
      <template #content>
        <div class="p20">
          <bcs-alert type="info" class="mb10">
            <template #title>
              <p>{{$t('按上面的模板创建你的json文件，选择“文件导入”操作')}}</p>
              <p>{{$t('scope值含义，global表示全局变量，cluster表示集群变量，namespace表示命名空间变量')}}</p>
              <p>{{$t('cluster和namespace变量需要提供vars关键字，cluster变量的vars需要包含集群ID cluster_id和变量值 value')}}</p>
              <p>{{$t('namespace变量的vars需要包含集群ID cluster_id、命名空间名称 namespace 和变量值 value')}}</p>
            </template>
          </bcs-alert>
          <div class="code-wrapper">
            <CodeEditor
              :value="JSON.stringify(exampleData, null, 4)"
              width="100%"
              height="100%"
              readonly>
            </CodeEditor>
          </div>
        </div>
      </template>
    </bcs-sideslider>
    <!-- 变量设置 -->
    <bcs-sideslider
      :is-show.sync="showSetSlider"
      :width="700"
      :before-close="handleBeforeClose"
      quick-close>
      <template #header>
        <LayoutRow>
          <template #left>{{currentRow && currentRow.name}}</template>
          <template #right>
            <bcs-button text class="switch-mode-btn" @click="mode = mode === 'form' ? 'json' : 'form'">
              <i class="bcs-icon bcs-icon-qiehuan"></i>
              {{mode === 'form' ? $t('切换为文本模式') : $t('切换为表单模式')}}
            </bcs-button>
          </template>
        </LayoutRow>
      </template>
      <template #content>
        <div class="slider-wrapper p20" v-bkloading="{ isLoading: setSliderLoading }">
          <bcs-table :data="setData" class="slider-wrapper-table" v-if="mode === 'form'">
            <bcs-table-column :label="$t('所属集群')" prop="clusterName"></bcs-table-column>
            <bcs-table-column
              :label="$t('所属命名空间')"
              prop="namespace"
              v-if="currentRow && currentRow.scope === 'namespace'">
            </bcs-table-column>
            <bcs-table-column :label="$t('值')">
              <template #default="{ row }">
                <bcs-input v-model="row.value" @change="setChanged(true)"></bcs-input>
              </template>
            </bcs-table-column>
          </bcs-table>
          <div class="code-wrapper" v-else>
            <CodeEditor
              :value="JSON.stringify(setData, null, 4)"
              @change="handleJsonDataChange">
            </CodeEditor>
          </div>
          <div class="mt15">
            <bcs-button theme="primary" :disabled="!currentRow" @click="handleSave">{{$t('保存')}}</bcs-button>
            <bcs-button @click="showSetSlider = false">{{$t('取消')}}</bcs-button>
          </div>
        </div>
      </template>
    </bcs-sideslider>
  </LayoutContent>
</template>
<script lang="ts">
import { computed, defineComponent, onMounted, ref, watch } from 'vue';
import LayoutContent from '@/components/layout/Content.vue';
import LayoutRow from '@/components/layout/Row.vue';
import useVariable, { IParams, Pick } from './use-variable';
import $i18n from '@/i18n/i18n-setup';
import BkFormItem from 'bk-magic-vue/lib/form-item';
import BkForm from 'bk-magic-vue/lib/form';
import CodeEditor from '@/components/monaco-editor/new-editor.vue';
import exampleData from './variable.json';
import useDebouncedRef from '@/composables/use-debounce';
import useSideslider from '@/composables/use-sideslider';
import $bkMessage from '@/common/bkmagic';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';

export default defineComponent({
  name: 'VariableManager',
  components: {
    LayoutContent,
    LayoutRow,
    BkForm,
    BkFormItem,
    CodeEditor,
  },
  setup() {
    const { reset, setChanged, handleBeforeClose } = useSideslider();
    const showSideslider = ref(false);
    const scopeList = ref([
      {
        id: 'global',
        name: $i18n.t('全局变量'),
      },
      {
        id: 'cluster',
        name: $i18n.t('集群变量'),
      },
      {
        id: 'namespace',
        name: $i18n.t('命名空间变量'),
      },
    ]);
    const searchKey = useDebouncedRef<string>('', 360);
    const scope = ref<Pick<IParams, 'scope'>>('');
    const params = computed<IParams>(() => ({
      limit: pagination.value.limit,
      offset: pagination.value.limit * (pagination.value.current - 1),
      searchKey: searchKey.value,
      scope: scope.value,
      all: false,
    }));
    const selections = ref<any[]>([]);

    const {
      isLoading,
      variableList,
      pagination,
      handlePageChange,
      handlePageLimitChange,
      getVariableDefinitions,
      handleCreateVariable,
      handleUpdateVariable,
      handleDeleteDefinitions,
      handleImportVariable,
      getClusterVariable,
      getNamespaceVariable,
      handleUpdateClusterVariable,
      handleUpdateNamespaceVariable,
    } = useVariable();

    watch(() => [searchKey.value, scope.value], () => {
      pagination.value.current = 1;
    });

    watch(params, () => {
      getVariableDefinitions(params.value);
    });

    function selectable(row) {
      return row.category !== 'sys';
    }

    function handleSelectionChange(selection: any[]) {
      selections.value = selection;
    }

    // 变量操作
    const showCreateOrEdit = ref(false);
    const dialogLoading = ref(false);
    const currentRow = ref<Record<string, any>|null>(null);
    const formRef = ref<any>(null);
    const formData = ref({
      name: '',
      key: '',
      scope: 'global',
      default: '',
      desc: '',
    });
    const formRules = ref({
      name: [
        {
          required: true,
          message: $i18n.t('必填项'),
          trigger: 'blur',
        },
        {
          validator(val) {
            return val.length <= 32;
          },
          message: $i18n.t('请输入32个字符以内的变量名称'),
          trigger: 'blur',
        },
      ],
      key: [
        {
          required: true,
          message: $i18n.t('必填项'),
          trigger: 'blur',
        },
        {
          validator(val) {
            return /^[A-Za-z][A-Za-z0-9_]{0,63}$/.test(val);
          },
          message: $i18n.t('只能包含字母、数字和下划线，且以字母开头，最大长度为64个字符'),
          trigger: 'blur',
        },
      ],
    });
    watch(showCreateOrEdit, () => {
      if (!showCreateOrEdit.value) {
        currentRow.value = null;
      }
    });

    const fileLoading = ref(false);
    function handleFileChange(event) {
      const [file] = event.target.files;
      if (!file) return;
      fileLoading.value = true;
      const reader = new FileReader();
      reader.readAsText(file, 'UTF-8');
      reader.onload = async (e) => {
        try {
          const data = e?.target?.result as string || '';
          event.target.value = '';
          const result = await handleImportVariable({ data: JSON.parse(data) });
          if (result) {
            $bkMessage({
              theme: 'success',
              message: $i18n.t('导入成功'),
            });
            getVariableDefinitions(params.value);
          }
        } catch (error) {
          console.error(error);
        }
        fileLoading.value = false;
      };
    }
    function handleAddVariable() {
      formData.value = {
        name: '',
        key: '',
        scope: 'global',
        default: '',
        desc: '',
      };
      showCreateOrEdit.value = true;
    }
    function handleEdit(row) {
      currentRow.value = row;
      formData.value = {
        name: row.name,
        key: row.key,
        scope: row.scope,
        default: row.default,
        desc: row.desc,
      };
      showCreateOrEdit.value = true;
    }
    async function handleConfirmDialog() {
      const validate = await formRef.value?.validate();
      if (!validate) return;

      dialogLoading.value = true;
      let result = false;
      if (currentRow.value) {
        result = await handleUpdateVariable({
          $variableID: currentRow.value.id,
          ...formData.value,
        });
      } else {
        result = await handleCreateVariable(formData.value);
      }
      dialogLoading.value = false;

      if (result) {
        $bkMessage({
          theme: 'success',
          message: $i18n.t('操作成功'),
        });
        getVariableDefinitions(params.value);
        showCreateOrEdit.value = false;
      }
    }
    function handleDelete(row) {
      $bkInfo({
        type: 'warning',
        clsName: 'custom-info-confirm',
        title: $i18n.t('确认删除节点'),
        subTitle: $i18n.t('确认删除变量 {name}', { name: row.name }),
        defaultInfo: true,
        confirmFn: async () => {
          const result =  await handleDeleteDefinitions({
            idList: row.id,
          });
          if (result) {
            $bkMessage({
              theme: 'success',
              message: $i18n.t('删除成功'),
            });
            pagination.value.current = 1;
            getVariableDefinitions(params.value);
          }
        },
      });
    }
    function handleBatchDelete() {
      $bkInfo({
        type: 'warning',
        clsName: 'custom-info-confirm',
        title: $i18n.t('确认删除节点'),
        subTitle: $i18n.t(
          '确认删除 {name} 等 {count} 个变量',
          {
            name: selections.value[0]?.name,
            count: selections.value.length,
          },
        ),
        defaultInfo: true,
        confirmFn: async () => {
          const result =  await handleDeleteDefinitions({
            idList: selections.value.map(item => item.id).join(','),
          });
          if (result) {
            $bkMessage({
              theme: 'success',
              message: $i18n.t('删除成功'),
            });
            pagination.value.current = 1;
            getVariableDefinitions(params.value);
          }
        },
      });
    }
    // 变量设置
    const showSetSlider = ref(false);
    const setSliderLoading = ref(false);
    const setData = ref<any[]>([]);
    const mode = ref<'form' | 'json'>('form');
    watch(showSetSlider, () => {
      if (!showSetSlider.value) {
        setData.value = [];
        currentRow.value = null;
        mode.value = 'form';
      }
    });
    async function handleSetVariable(row) {
      showSetSlider.value = true;
      currentRow.value = row;
      setSliderLoading.value = true;
      let data = { results: [] };
      if (row.scope === 'cluster') {
        data = await getClusterVariable({
          $variableID: row.id,
        });
      } else {
        data = await getNamespaceVariable({
          $variableID: row.id,
        });
      }
      setData.value = data.results || [];
      setSliderLoading.value = false;
      reset();
    }
    async function handleSave() {
      setSliderLoading.value = true;
      let result = false;
      if (currentRow.value?.scope === 'cluster') {
        result = await handleUpdateClusterVariable({
          $variableID: currentRow.value?.id,
          data: setData.value.map(item => ({ clusterID: item.clusterID, value: item.value })),
        });
      } else {
        result = await handleUpdateNamespaceVariable({
          $variableID: currentRow.value?.id,
          data: setData.value.map(item => ({
            clusterID: item.clusterID,
            value: item.value,
            namespace: item.namespace,
          })),
        });
      }
      setSliderLoading.value = false;
      if (result) {
        $bkMessage({
          theme: 'success',
          message: $i18n.t('设置成功'),
        });
        getVariableDefinitions(params.value);
        showSetSlider.value = false;
      }
    }
    function handleJsonDataChange(content) {
      try {
        setData.value = JSON.parse(content);
      } catch (error) {
        console.log(error);
      }
    }

    onMounted(() => {
      getVariableDefinitions(params.value);
    });
    return {
      mode,
      setData,
      setSliderLoading,
      fileLoading,
      exampleData,
      showSideslider,
      showSetSlider,
      selections,
      currentRow,
      dialogLoading,
      scopeList,
      formRef,
      formRules,
      formData,
      isLoading,
      searchKey,
      scope,
      variableList,
      pagination,
      showCreateOrEdit,
      handlePageChange,
      handlePageLimitChange,
      selectable,
      handleAddVariable,
      handleConfirmDialog,
      handleEdit,
      handleDelete,
      handleSelectionChange,
      handleBatchDelete,
      handleFileChange,
      handleSetVariable,
      handleSave,
      handleJsonDataChange,
      setChanged,
      handleBeforeClose,
    };
  },
});
</script>
<style lang="postcss" scoped>
.mw320 {
  min-width: 320px;
}
.mw200 {
  min-width: 200px;
  background-color: #fff;
}
.file-input {
  position: absolute;
  width: 100%;
  height: 100%;
  left: 0;
  opacity: 0;
}
.code-wrapper {
  height: calc(100vh - 192px);
}

.slider-wrapper {
  max-height: calc(100vh - 60px);
  overflow: auto;
}

.switch-mode-btn {
  line-height: 1;
  font-weight: normal;
  margin-right: 20px;
}

>>> .slider-wrapper-table .bk-table-body-wrapper {
  max-height: calc(100vh - 200px);
  overflow-y: auto;
}
</style>
