<template>
  <div class="template-version-manage-page">
    <div class="page-header">
      <ArrowsLeft class="arrow-icon" @click="goToTemplateListPage" />
      <div v-if="templateDetail" class="title-name">
        {{t('版本管理')}} - {{ templateDetail.spec.name }}
        <span class="path">{{ templateDetail.spec.path }}</span>
      </div>
    </div>
    <div class="operation-area">
      <bk-button theme="primary" @click="openSelectVersionDialog">
        <Plus class="button-icon" />
        {{t('新建版本')}}
      </bk-button>
      <SearchInput v-model="searchStr" :placeholder="t('版本号/版本说明/更新人')" :width="320" @search="refreshList()" />
    </div>
    <div class="version-content-area">
      <VersionFullTable
        v-if="!versionDetailModeData.open"
        :space-id="spaceId"
        :template-space-id="templateSpaceId"
        :template-id="templateId"
        :list="versionList"
        :bound-by-apps-count-loading="boundByAppsCountLoading"
        :bound-by-apps-count-list="boundByAppsCountList"
        :pagination="pagination"
        @deleted="handleVersionDeleted"
        @select="handleOpenDetailTable($event, 'view')"
      />
      <VersionDetailTable
        v-else
        :space-id="spaceId"
        :template-space-id="templateSpaceId"
        :template-id="templateId"
        :list="versionList"
        :pagination="pagination"
        :type="versionDetailModeData.type"
        :version-id="versionDetailModeData.id"
        @created="handleCreatedVersion"
        @close="versionDetailModeData.open = false"
        @select="handleOpenDetailTable($event, 'view')"
      />
    </div>
    <bk-dialog
      :title="t('新建版本')"
      width="480"
      dialog-type="operation"
      :is-show="selectVersionDialog.open"
      :is-loading="allVersionListLoading"
      @confirm="handleSelectVersionConfirm"
      @closed="selectVersionDialog.open = false"
    >
      <bk-form ref="selectVersionFormRef" form-type="vertical" :model="{ id: selectVersionDialog.id }">
        <bk-form-item :label="t('选择载入版本')" required property="id">
          <bk-select v-model="selectVersionDialog.id" :clearable="false" :filterable="true" :input-search="false">
            <bk-option v-for="item in allVersionList" v-overflow-title :key="item.id" :id="item.id" :label="item.name">
            </bk-option>
          </bk-select>
        </bk-form-item>
      </bk-form>
    </bk-dialog>
  </div>
</template>
<script lang="ts" setup>
import { ref, computed, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { storeToRefs } from 'pinia';
import { useRoute, useRouter } from 'vue-router';
import { ArrowsLeft, Plus } from 'bkui-vue/lib/icon';
import useGlobalStore from '../../../../store/global';
import useTemplateStore from '../../../../store/template';
import { ITemplateConfigItem, ITemplateVersionItem } from '../../../../../types/template';
import { IPagination, ICommonQuery } from '../../../../../types/index';
import {
  getTemplatesDetailByIds,
  getTemplateVersionList,
  getCountsByTemplateVersionIds,
} from '../../../../api/template';
import SearchInput from '../../../../components/search-input.vue';
import VersionFullTable from './version-full-table.vue';
import VersionDetailTable from './version-detail/version-detail-table.vue';

const { t } = useI18n();
const getRouteId = (id: string) => {
  if (id && typeof Number(id) === 'number') {
    return Number(id);
  }
  return 0;
};

const { spaceId } = storeToRefs(useGlobalStore());
const { versionListPageShouldOpenEdit, versionListPageShouldOpenView } = storeToRefs(useTemplateStore());
const route = useRoute();
const router = useRouter();
const templateDetailLoading = ref(false);
const templateDetail = ref<ITemplateConfigItem>();
const versionListLoading = ref(false);
const versionList = ref<ITemplateVersionItem[]>([]);
const allVersionListLoading = ref(false);
const allVersionList = ref<{ id: number; name: string }[]>([]); // 全量版本列表，选择载入版本使用
const boundByAppsCountLoading = ref(false);
const boundByAppsCountList = ref([]);
const searchStr = ref('');
const selectVersionFormRef = ref();
const selectVersionDialog = ref<{ open: boolean; id: number | string }>({
  open: false,
  id: '',
});
const versionDetailModeData = ref<{ open: boolean; type: string; id: number }>({
  open: false,
  type: 'create',
  id: 0,
});
const pagination = ref<IPagination>({
  count: 0,
  current: 1,
  limit: 10,
});

const templateSpaceId = computed(() => getRouteId(route.params.templateSpaceId as string));
const packageId = computed(() => route.params.packageId);
const templateId = computed(() => getRouteId(route.params.templateId as string));

onMounted(async () => {
  getTemplateDetail();
  await getVersionList();
  if (versionListPageShouldOpenView.value) {
    versionListPageShouldOpenView.value = false;
    handleOpenDetailTable(versionList.value[0].id, 'view');
  }
  if (versionListPageShouldOpenEdit.value) {
    versionListPageShouldOpenEdit.value = false;
    handleOpenDetailTable(versionList.value[0].id, 'create');
  }
});

const getTemplateDetail = async () => {
  templateDetailLoading.value = true;
  const res = await getTemplatesDetailByIds(spaceId.value, [templateId.value]);
  if (res.details.length > 0) {
    templateDetail.value = res.details[0];
  }
  templateDetailLoading.value = false;
};

const getVersionList = async () => {
  versionListLoading.value = true;
  const params: ICommonQuery = {
    start: (pagination.value.current - 1) * pagination.value.limit,
    limit: pagination.value.limit,
  };
  if (searchStr.value) {
    params.search_value = searchStr.value;
    params.search_fields = 'revision_name,revision_memo,creator';
  }
  const res = await getTemplateVersionList(spaceId.value, templateSpaceId.value, templateId.value, params);
  versionList.value = res.details;
  pagination.value.count = res.count;
  versionListLoading.value = false;
  const ids = versionList.value.map(item => item.id);
  boundByAppsCountList.value = [];
  if (ids.length > 0) {
    loadBoundByAppsList(ids);
  }
};

const getAllVersionList = async () => {
  allVersionListLoading.value = true;
  const res = await getTemplateVersionList(spaceId.value, templateSpaceId.value, templateId.value, {
    start: 0,
    all: true,
  });
  allVersionList.value = res.details.map((item: ITemplateVersionItem) => {
    const { id, spec } = item;
    const name = spec.revision_memo ? `${spec.revision_name}(${spec.revision_memo})` : spec.revision_name;
    return { id, name };
  });
  allVersionListLoading.value = false;
};

const loadBoundByAppsList = async (ids: number[]) => {
  boundByAppsCountLoading.value = true;
  const res = await getCountsByTemplateVersionIds(spaceId.value, templateSpaceId.value, templateId.value, ids);
  boundByAppsCountList.value = res.details;
  boundByAppsCountLoading.value = false;
};

const goToTemplateListPage = () => {
  router.push({
    name: 'templates-list',
    params: {
      templateSpaceId: templateSpaceId.value,
      packageId: packageId.value,
    },
  });
};

const openSelectVersionDialog = async () => {
  selectVersionDialog.value.open = true;
  await getAllVersionList();
  selectVersionDialog.value.id = allVersionList.value.length > 0 ? allVersionList.value[0].id : '';
};

const handleVersionMenuSelect = (id: number) => {
  if (id === 0) {
    handleOpenDetailTable(0, 'create');
  } else {
    handleOpenDetailTable(id, 'view');
  }
};

const handleSelectVersionConfirm = async () => {
  await selectVersionFormRef.value.validate();
  handleOpenDetailTable(selectVersionDialog.value.id as number, 'create');
  selectVersionDialog.value.open = false;
};

const handleOpenDetailTable = (id: number, type: string) => {
  versionDetailModeData.value = {
    open: true,
    type,
    id,
  };
};

const handleVersionDeleted = () => {
  if (versionList.value.length === 1 && pagination.value.current > 1) {
    pagination.value.current -= 1;
  }
  getVersionList();
};

// 新建版本成功后，重新拉取版本列表，并选中最新版本
const handleCreatedVersion = async (id: number) => {
  pagination.value.current = 1;
  await getVersionList();
  handleVersionMenuSelect(id);
};

const refreshList = (current = 1) => {
  pagination.value.current = current;
  getVersionList();
};
</script>
<style lang="scss" scoped>
.template-version-manage-page {
  height: 100%;
  background: #f5f7fa;
}
.page-header {
  display: flex;
  align-items: center;
  padding: 0 24px;
  height: 52px;
  background: #ffffff;
  box-shadow: 0 3px 4px 0 #0000000a;
  .arrow-icon {
    font-size: 24px;
    color: #3a84ff;
    cursor: pointer;
  }
  .title-name {
    padding: 14px 0;
    font-size: 16px;
    line-height: 24px;
    color: #313238;
    .path {
      margin-left: 12px;
      color: #979ba5;
      font-size: 12px;
    }
  }
}
.operation-area {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 24px;
  padding: 0 24px;
  .button-icon {
    font-size: 18px;
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
.version-content-area {
  padding: 16px 24px 24px;
  height: calc(100% - 110px);
  overflow: hidden;
}
</style>
