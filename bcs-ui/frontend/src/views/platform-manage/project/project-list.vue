<template>
  <BcsContent :title="$t('nav.platformProject')" hide-back>
    <Row>
      <template #left>
        <bcs-input
          :placeholder="$t('generic.placeholder.searchName')"
          class="min-w-[360px]"
          right-icon="bk-icon icon-search"
          clearable
          v-model="searchValue">
        </bcs-input>
      </template>
    </Row>
    <bcs-table
      class="mt-[20px]"
      :pagination="pagination"
      :data="projectList"
      v-bkloading="{ isLoading: loading }"
      @page-change="pageChange"
      @page-limit-change="pageSizeChange">
      <bcs-table-column :label="$t('generic.label.name')">
        <template #default="{ row }">
          <bcs-button
            text
            @click="handleShowProjectDetail(row)">
            <span class="bcs-ellipsis">
              {{row.name}}
            </span>
          </bcs-button>
        </template>
      </bcs-table-column>
      <bcs-table-column
        :label="$t('projects.project.ID')"
        prop="projectID"
        show-overflow-tooltip>
      </bcs-table-column>
      <bcs-table-column
        :label="$t('generic.label.description')"
        prop="description"
        show-overflow-tooltip>
      </bcs-table-column>
      <bcs-table-column
        :label="$t('projects.project.business')"
        prop="businessName"
        show-overflow-tooltip>
        <template #default="{ row }">
          {{ row.businessName || '--' }}
        </template>
      </bcs-table-column>
      <bcs-table-column
        :label="$t('generic.label.createdBy')"
        prop="creator"
        show-overflow-tooltip>
      </bcs-table-column>
      <bcs-table-column
        :label="$t('generic.label.managers')"
        prop="managers"
        show-overflow-tooltip>
      </bcs-table-column>
      <bcs-table-column :label="$t('generic.label.status')" prop="isOffline">
        <template #default="{ row }">
          <StatusIcon
            :status="row.isOffline ? 'offline' : 'online'"
            :status-color-map="statusColorMap"
            :status-text-map="statusTextMap">
            {{statusTextMap[row.isOffline ? 'offline' : 'online']}}
          </StatusIcon>
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('generic.label.action')" width="120">
        <template #default="{ row }">
          <bcs-button
            text
            @click="openLink(row)">{{ $t('projects.project.goto') }}</bcs-button>
        </template>
      </bcs-table-column>
      <template #empty>
        <BcsEmptyTableStatus :type="searchValue ? 'search-empty' : 'empty'" @clear="handleClearSearchData" />
      </template>
    </bcs-table>
    <!-- 项目详情侧滑 -->
    <bcs-sideslider
      :width="950"
      quick-close
      :is-show.sync="showProjectDetail">
      <template #header>
        <div class="relative">
          <span class="text-[16px]">{{ curProject?.name || '--' }}</span>
          <div class="absolute inset-0 flex justify-center items-center">
            <div class="bk-button-group">
              <bcs-button
                size="small"
                :class="{ 'is-selected': detailMode === 'code' }"
                @click="detailMode = 'code'">
                {{ $t('generic.label.codeMode') }}
              </bcs-button>
              <bcs-button
                size="small"
                :class="{ 'is-selected': detailMode === 'form' }"
                @click="detailMode = 'form'">
                {{ $t('dashboard.label.formMode') }}
              </bcs-button>
            </div>
          </div>
        </div>
      </template>
      <template #content>
        <div :class="['h-full', { 'p-[20px]': detailMode === 'form' }]">
          <bk-form
            v-if="detailMode === 'form'"
            class="bcs-small-form grid grid-cols-2 gap-[16px] px-[20px]"
            v-bkloading="{ isLoading: detailLoading }"
            form-type="vertical">
            <bk-form-item :label="$t('generic.label.name')">
              {{ curProject?.name || '--' }}
            </bk-form-item>
            <bk-form-item :label="$t('projects.project.ID')">
              {{ curProject?.projectID || '--' }}
            </bk-form-item>
            <bk-form-item :label="$t('generic.label.description')">
              <EditFormItem
                :maxlength="100"
                :value="curProject?.description || ''"
                :clearable="true"
                type="textarea"
                @save="handleUpdateDescription" />
            </bk-form-item>
            <bk-form-item :label="$t('projects.project.business')">
              <EditField
                :has-btn="true"
                :value="curProject?.businessName"
                :edit-mode="businessIDEditMode"
                @update:editMode="businessIDEditMode = $event"
                @save="handleUpdateBusinessID">
                <bcs-select
                  class="w-full"
                  v-model="businessID"
                  :clearable="false"
                  searchable
                  id-key="bk_biz_id"
                  display-key="bk_biz_name">
                  <bcs-option
                    v-for="option in businessList"
                    :key="option.bk_biz_id"
                    :id="option.bk_biz_id"
                    :name="option.bk_biz_name"
                  ></bcs-option>
                </bcs-select>
              </EditField>
            </bk-form-item>
            <bk-form-item :label="$t('projects.project.businessID')">
              {{ curProject?.businessID || '--' }}
            </bk-form-item>
            <bk-form-item :label="$t('generic.label.createdBy1')">
              {{ curProject?.creator || '--' }}
            </bk-form-item>
            <bk-form-item :label="$t('generic.label.managers')">
              <EditField
                :has-btn="true"
                :value="curProject?.managers"
                :edit-mode="managersEditMode"
                @update:editMode="managersEditMode = $event"
                @save="handleUpdateManagers">
                <BkUserSelector
                  v-model="curManagers"
                  class="w-full"
                  :api="userSelectorAPI"
                  ref="inputRef" />
              </EditField>
            </bk-form-item>
            <bk-form-item :label="$t('generic.label.status')">
              <StatusIcon
                :status="curProject?.isOffline ? 'offline' : 'online'"
                :status-color-map="statusColorMap"
                :status-text-map="statusTextMap">
                {{statusTextMap[curProject?.isOffline ? 'offline' : 'online']}}
              </StatusIcon>
            </bk-form-item>
          </bk-form>
          <div v-else class="h-full">
            <CodeEditor
              class="h-[calc(100vh-104px)]"
              multi-document
              :value="content"
              :title="$t('projects.project.info')"
              :no-validate="false"
              v-bkloading="{ isLoading: false }"
              ref="codeEditorRef" />
            <div class="h-[52px] bg-[#2E2E2E] px-[16px] py-[10px]">
              <bcs-button
                theme="primary"
                @click="handleUpdateProject">
                {{ $t('generic.button.save') }}
              </bcs-button>
            </div>
          </div>
        </div>
      </template>
    </bcs-sideslider>
  </BcsContent>
</template>
<script lang="ts" setup>
import { onMounted, ref, watch } from 'vue';

import BkUserSelector from '@blueking/user-selector';

import { BusinessService, ProjectService } from '@/api/modules/platform-manage';
import $bkMessage from '@/common/bkmagic';
import EditField from '@/components/edit-field.vue';
import BcsContent from '@/components/layout/Content.vue';
import Row from '@/components/layout/Row.vue';
import CodeEditor from '@/components/monaco-editor/ai-editor.vue';
import StatusIcon from '@/components/status-icon';
import useDebouncedRef from '@/composables/use-debounce';
import $i18n from '@/i18n/i18n-setup';
import EditFormItem from '@/views/cluster-manage/components/edit-form-item.vue';

// 搜索
const searchValue = useDebouncedRef<string>('', 300);

// 项目列表
const loading = ref(false);
const projectList = ref<IPlatformProject[]>([]);
const pagination = ref({
  count: 0,
  current: 1,
  limit: 10,
});

// 状态映射
const statusColorMap = ref({
  online: 'green',
  offline: 'red',
});
const statusTextMap = ref({
  online: $i18n.t('generic.label.enabled'),
  offline: $i18n.t('generic.label.disabled'),
});

// 项目详情
const showProjectDetail = ref(false);
const curProject = ref<Partial<IPlatformProject>>({});
const detailLoading = ref(false);
const detailMode = ref<'form' | 'code'>('form');
const content = ref('');
const codeEditorRef = ref<InstanceType<typeof CodeEditor>>();

// 业务列表
const businessList = ref<IPlatformBusiness[]>([]);
const businessLoading = ref(false);

// 获取项目列表
const handleGetProjectList = async (defaultLoading = true) => {
  loading.value = defaultLoading;
  const res = await ProjectService.ListProject({
    limit: pagination.value.limit,
    offset: pagination.value.current - 1,
    searchName: searchValue.value,
  }).catch(() => ({ results: [], total: 0 }));
  projectList.value = res?.results || [];
  pagination.value.count = res?.total || 0;
  loading.value = false;
};

// 更新业务ID
const businessIDEditMode = ref(false);
const businessID = ref<number | string>();
const handleUpdateBusinessID = async () => {
  if (!curProject.value.projectID) return;
  const result = await ProjectService.UpdateProject({
    $projectId: curProject.value.projectID,
    businessID: String(businessID.value),
  }).then(() => true)
    .catch(() => false);

  if (result) {
    $bkMessage({
      theme: 'success',
      message: $i18n.t('generic.msg.success.save'),
    });
    curProject.value.businessID = String(businessID.value);
    curProject.value.businessName = businessList.value.find(item => item.bk_biz_id === Number(businessID.value))?.bk_biz_name || '';
    businessIDEditMode.value = false;
    handleGetProjectList();
  }
};

// 获取业务列表
const handleGetBusinessList = async () => {
  businessLoading.value = true;
  const res = await BusinessService.ListBusiness().catch(() => [])  ;
  businessList.value = res || [];
  businessLoading.value = false;
};

// 更新管理员
const userSelectorAPI = `${window.BK_USER_HOST}/api/c/compapi/v2/usermanage/fs_list_users/?app_code=bk-magicbox&page_size=100&page=1&callback=USER_LIST_CALLBACK_0`;
const managersEditMode = ref(false);
const curManagers = ref<string[]>([]);
const handleUpdateManagers = async () => {
  if (!curProject.value.projectID) return;
  const result = await ProjectService.UpdateProjectManager({
    $projectId: curProject.value.projectID,
    managers: curManagers.value.join(','),
  }).then(() => true)
    .catch(() => false);

  if (result) {
    $bkMessage({
      theme: 'success',
      message: $i18n.t('generic.msg.success.save'),
    });
  }
  curProject.value.managers = curManagers.value.join(',');
  managersEditMode.value = false;
  handleGetProjectList();
};
const handleUpdateProject = async () => {
  if (!curProject.value.projectID) return;
  try {
    const editorContent = codeEditorRef.value?.getData();
    const obj = JSON.parse(editorContent || '');
    Object.assign(curProject.value, obj);
  } catch (error) {
    console.warn('json 解析失败', error);
    return;
  }
  const result = await ProjectService.UpdateProject({
    $projectId: curProject.value.projectID,
    ...curProject.value,
  }).then(() => true)
    .catch(() => false);

  if (result) {
    $bkMessage({
      theme: 'success',
      message: $i18n.t('generic.msg.success.save'),
    });
  }
  handleGetProjectList();
};

// 分页
const pageChange = (page: number) => {
  pagination.value.current = page;
  handleGetProjectList();
};

const pageSizeChange = (size: number) => {
  pagination.value.current = 1;
  pagination.value.limit = size;
  handleGetProjectList();
};

// 搜索
watch(searchValue, () => {
  pagination.value.current = 1;
  handleGetProjectList();
});

// 不能编辑的字段
const uneditableFields = ref(['createTime', 'updateTime', 'creator', 'updater', 'businessName', 'isOffline']);
// 显示项目详情
const handleShowProjectDetail = async (row: IPlatformProject) => {
  curProject.value = await ProjectService.GetProject({
    $projectId: row.projectID,
  }).catch(() => ({}));
  businessID.value = curProject.value.businessID ? Number(curProject.value.businessID) : '';
  curManagers.value = (curProject.value?.managers || '').split(',').filter(item => item);
  const temp = JSON.parse(JSON.stringify(curProject.value));

  // 删除不能编辑的字段
  uneditableFields.value.forEach((item) => {
    delete temp[item];
  });

  content.value = JSON.stringify(temp, null, 2);
  showProjectDetail.value = true;
};
// 打开链接
const openLink = (row: IPlatformProject) => {
  row.link && window.open(row.link);
};

// 编辑描述
const handleUpdateDescription = async (value: string) => {
  if (!curProject.value.projectID) return;
  const result = await ProjectService.UpdateProject({
    $projectId: curProject.value.projectID,
    description: value,
  }).then(() => true)
    .catch(() => false);

  if (result) {
    $bkMessage({
      theme: 'success',
      message: '保存成功',
    });
  }
  // 更新当前项目信息
  curProject.value.description = value;
  // 刷新列表
  handleGetProjectList();
};

// 清空搜索
const handleClearSearchData = () => {
  searchValue.value = '';
  pagination.value.current = 1;
  handleGetProjectList();
};

onMounted(() => {
  handleGetProjectList();
  handleGetBusinessList();
});
</script>
<style lang="postcss" scoped>
:deep(.bk-form-content) {
  margin-top: 6px;
}
:deep(.bk-sideslider-content) {
  height: calc(-52px + 100vh);
}
</style>
