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
        label="项目ID"
        prop="projectID"
        show-overflow-tooltip>
      </bcs-table-column>
      <bcs-table-column
        :label="$t('generic.label.description')"
        prop="description"
        show-overflow-tooltip>
      </bcs-table-column>
      <bcs-table-column
        label="所属业务"
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
        label="管理者"
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
            @click="handleShowProjectDetail(row)">{{ $t('前往项目') }}</bcs-button>
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
      :title="curProject?.name || '--'"
      :is-show.sync="showProjectDetail">
      <template #content>
        <div class="p-[20px]">
          <bk-form
            class="bcs-small-form grid grid-cols-2 gap-[16px] px-[20px]"
            v-bkloading="{ isLoading: detailLoading }"
            form-type="vertical">
            <bk-form-item :label="$t('generic.label.name')">
              {{ curProject?.name || '--' }}
            </bk-form-item>
            <bk-form-item label="项目ID">
              {{ curProject?.projectID || '--' }}
            </bk-form-item>
            <bk-form-item :label="$t('generic.label.description')">
              <EditField
                :maxlength="100"
                :value="curProject?.description || ''"
                :clearable="true"
                :edit-mode="descriptionEditMode"
                type="textarea"
                @update:editMode="descriptionEditMode = $event"
                @blur="handleUpdateDescription" />
            </bk-form-item>
            <bk-form-item label="所属业务">
              <EditField
                :value="curProject?.businessName"
                :edit-mode="businessIDEditMode"
                @update:editMode="businessIDEditMode = $event">
                <bcs-select
                  class="w-full"
                  :value="Number(curProject?.businessID) || ''"
                  :clearable="false"
                  searchable
                  id-key="bk_biz_id"
                  display-key="bk_biz_name"
                  @change="handleUpdateBusinessID">
                  <bcs-option
                    v-for="option in businessList"
                    :key="option.bk_biz_id"
                    :id="option.bk_biz_id"
                    :name="option.bk_biz_name"
                  ></bcs-option>
                </bcs-select>
              </EditField>
            </bk-form-item>
            <bk-form-item label="创建者">
              {{ curProject?.creator || '--' }}
            </bk-form-item>
            <bk-form-item label="管理者">
              {{ curProject?.managers || '--' }}
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
        </div>
      </template>
    </bcs-sideslider>
  </BcsContent>
</template>
<script lang="ts" setup>
import { onMounted, ref, watch } from 'vue';

import { BusinessService, ProjectService } from '@/api/modules/platform-manage';
import $bkMessage from '@/common/bkmagic';
import EditField from '@/components/edit-field.vue';
import BcsContent from '@/components/layout/Content.vue';
import Row from '@/components/layout/Row.vue';
import StatusIcon from '@/components/status-icon';
import useDebouncedRef from '@/composables/use-debounce';

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
  online: '已启用',
  offline: '已停用',
});

// 项目详情
const showProjectDetail = ref(false);
const curProject = ref<Partial<IPlatformProject>>({});
const detailLoading = ref(false);

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
const handleUpdateBusinessID = async (value: number) => {
  if (!curProject.value.projectID) return;
  const result = await ProjectService.UpdateProject({
    $projectId: curProject.value.projectID,
    businessID: String(value),
  }).then(() => true)
    .catch(() => false);

  if (result) {
    $bkMessage({
      theme: 'success',
      message: '保存成功',
    });
  }
  curProject.value.businessID = String(value);
  curProject.value.businessName = businessList.value.find(item => item.bk_biz_id === Number(value))?.bk_biz_name || '';
  businessIDEditMode.value = false;
  handleGetProjectList();
};

// 获取业务列表
const handleGetBusinessList = async () => {
  businessLoading.value = true;
  const res = await BusinessService.ListBusiness().catch(() => [])  ;
  businessList.value = res || [];
  businessLoading.value = false;
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

// 显示项目详情
const handleShowProjectDetail = async (row: IPlatformProject) => {
  curProject.value = row;
  showProjectDetail.value = true;
};

// 编辑描述
const descriptionEditMode = ref(false);
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
  descriptionEditMode.value = false;
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
/deep/ .bk-form-content {
  margin-top: 6px;
}
</style>
