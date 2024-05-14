<template>
  <DetailLayout :name="`${t('版本管理')} - ${scriptDetail.spec.name}`" :show-footer="false" @close="handleClose">
    <template #content>
      <bk-loading :loading="initLoading" class="script-version-manage">
        <div class="operation-area">
          <CreateVersion
            :script-id="scriptId"
            :disabled="createBtnDisabled"
            :creatable="!unPublishVersion"
            @create="handleCreateVersionClick"
            @edit="handleEditVersionClick" />
          <bk-input
            v-model.trim="searchStr"
            class="search-input"
            :placeholder="t('版本号/版本说明/更新人')"
            :clearable="true"
            @enter="refreshList"
            @clear="refreshList"
            @change="handleSearchInputChange">
            <template #suffix>
              <Search class="search-input-icon" />
            </template>
          </bk-input>
        </div>
        <div :class="['version-data-container', { 'script-panel-open': versionEditData.panelOpen }]">
          <div class="table-data-area">
            <VersionListFullTable
              v-if="!versionEditData.panelOpen"
              :script-id="scriptId"
              :list="versionList"
              :pagination="pagination"
              :is-search-empty="isSearchEmpty"
              @view="handleViewVersionClick"
              @page-change="refreshList"
              @page-limit-change="handlePageLimitChange"
              @clear-str="clearStr">
              <template #operations="{ data }">
                <div v-if="data.hook_revision" class="action-btns">
                  <bk-button
                    v-if="['not_deployed', 'shutdown'].includes(data.hook_revision.spec.state)"
                    text
                    theme="primary"
                    @click="handlePublishClick(data.hook_revision)">
                    {{ t('上线') }}
                  </bk-button>
                  <bk-button
                    v-if="data.hook_revision.spec.state === 'not_deployed'"
                    text
                    theme="primary"
                    @click="handleEditVersionClick">
                    {{ t('编辑') }}
                  </bk-button>
                  <bk-button text theme="primary" @click="handleVersionDiff(data.hook_revision)">
                    {{ t('版本对比') }}
                  </bk-button>
                  <bk-button
                    v-if="data.hook_revision.spec.state !== 'not_deployed'"
                    text
                    theme="primary"
                    :disabled="!!unPublishVersion"
                    v-bk-tooltips="{ content: '当前已有「未上线」版本', disabled: !unPublishVersion }"
                    @click="handleCreateVersionClick(data.hook_revision.spec.content)">
                    {{ t('复制并新建') }}
                  </bk-button>
                  <bk-button
                    v-if="data.hook_revision.spec.state === 'not_deployed'"
                    text
                    theme="primary"
                    :disabled="pagination.count <= 1"
                    @click="handleDelClick(data.hook_revision)">
                    {{ t('删除') }}
                  </bk-button>
                </div>
              </template>
            </VersionListFullTable>
            <template v-else>
              <bk-button class="back-table-btn" text theme="primary" @click="versionEditData.panelOpen = false">
                {{ t('展开列表') }}
                <AngleDoubleRightLine class="arrow-icon" />
              </bk-button>
              <VersionListSimpleTable
                :version-id="versionEditData.form.id"
                :list="versionList"
                :pagination="pagination"
                @select="handleViewVersionClick"
                @page-change="refreshList" />
            </template>
          </div>
          <div v-if="versionEditData.panelOpen" class="script-edit-area">
            <VersionEdit
              :type="scriptDetail.spec.type"
              :version-data="versionEditData.form"
              :script-id="scriptId"
              :editable="versionEditData.editable"
              :hook-revision="versionEditData.hook_revision"
              :has-unpublish-version="!!unPublishVersion"
              @submitted="handleVersionEditSubmitted"
              @publish="handlePublishClick"
              @edit="handleEditVersionClick"
              @copy-and-create="handleCreateVersionClick"
              @delete="handleDelClick"
              @close="versionEditData.panelOpen = false" />
          </div>
        </div>
      </bk-loading>
    </template>
  </DetailLayout>
  <ScriptVersionDiff
    v-if="showVersionDiff"
    v-model:show="showVersionDiff"
    :type="scriptDetail.spec.type"
    :crt-version="crtVersion as IScriptVersion"
    :space-id="spaceId"
    :script-id="scriptId" />
  <DeleteConfirmDialog
    v-model:isShow="isDeleteScriptVersionDialogShow"
    :title="t('确认删除该脚本版本？')"
    @confirm="handleDeleteScriptVersionConfirm">
    <div style="margin-bottom: 8px">
      {{ t('脚本名称') }}:
      <span style="color: #313238; font-weight: 600">
        {{ deleteScriptVersionItem?.spec.name }}
      </span>
    </div>
    <div>{{ t('一旦删除，该操作将无法撤销，请谨慎操作') }}</div>
  </DeleteConfirmDialog>
</template>

<script setup lang="ts">
  import { ref, onMounted } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { useRouter, useRoute } from 'vue-router';
  import { InfoBox, Message } from 'bkui-vue';
  import { Search, AngleDoubleRightLine } from 'bkui-vue/lib/icon';
  import { storeToRefs } from 'pinia';
  import dayjs from 'dayjs';
  import useGlobalStore from '../../../../store/global';
  import useScriptStore from '../../../../store/script';
  import { IScriptVersion, IScriptVersionListItem, IScriptVersionForm } from '../../../../../types/script';
  import {
    getScriptDetail,
    getScriptVersionList,
    deleteScriptVersion,
    publishVersion,
    getScriptVersionDetail,
  } from '../../../../api/script';
  import DetailLayout from '../components/detail-layout.vue';
  import VersionListFullTable from './version-list-full-table.vue';
  import VersionListSimpleTable from './version-list-simple-table.vue';
  import CreateVersion from './create-version.vue';
  import VersionEdit from './version-edit.vue';
  import ScriptVersionDiff from './script-version-diff.vue';
  import DeleteConfirmDialog from '../../../../components/delete-confirm-dialog.vue';

  const { spaceId } = storeToRefs(useGlobalStore());
  const { versionListPageShouldOpenEdit, versionListPageShouldOpenView } = storeToRefs(useScriptStore());
  const router = useRouter();
  const route = useRoute();
  const { t } = useI18n();

  const scriptId = ref(Number(route.params.scriptId));
  const initLoading = ref(false);
  const detailLoading = ref(true);
  const scriptDetail = ref({ spec: { name: '', type: '' }, not_release_id: 0 });
  const versionLoading = ref(true);
  const versionList = ref<IScriptVersionListItem[]>([]);
  const unPublishVersion = ref<IScriptVersion | null>(null); // 未发布版本
  const createBtnDisabled = ref(false);
  const showVersionDiff = ref(false);
  const crtVersion = ref<IScriptVersion | null>(null);
  const isSearchEmpty = ref(false);
  const isDeleteScriptVersionDialogShow = ref(false);
  const deleteScriptVersionItem = ref<IScriptVersion>();
  const versionEditData = ref<{
    panelOpen: boolean;
    editable: boolean;
    hook_revision: IScriptVersion | null;
    form: IScriptVersionForm;
  }>({
    panelOpen: false,
    editable: true,
    hook_revision: null,
    form: {
      // 版本编辑、新建、查看数据
      id: 0,
      name: '',
      memo: '',
      content: '',
    },
  });
  const searchStr = ref('');
  const pagination = ref({
    current: 1,
    count: 0,
    limit: 10,
  });

  onMounted(async () => {
    initLoading.value = true;
    await getVersionList();
    await getScriptDetailData();
    if (scriptDetail.value.not_release_id) {
      unPublishVersion.value = await getScriptVersionDetail(
        spaceId.value,
        scriptId.value,
        scriptDetail.value.not_release_id,
      );
    }
    initLoading.value = false;
    if (versionListPageShouldOpenEdit.value) {
      versionListPageShouldOpenEdit.value = false;
      if (scriptDetail.value.not_release_id) {
        handleEditVersionClick();
      } else {
        handleCreateVersionClick(versionList.value[0].hook_revision.spec.content);
      }
    }
    if (versionListPageShouldOpenView.value) {
      versionListPageShouldOpenView.value = false;
      handleViewVersionClick(versionList.value[0]);
    }
  });

  // 获取脚本详情
  const getScriptDetailData = async () => {
    detailLoading.value = true;
    const res = await getScriptDetail(spaceId.value, scriptId.value);
    const { name, type, releases } = res.spec;
    scriptDetail.value = { spec: { name, type }, not_release_id: releases.not_release_id };
    detailLoading.value = false;
  };

  // 获取版本列表
  const getVersionList = async () => {
    versionLoading.value = true;
    const params: { start: number; limit: number; searchKey?: string } = {
      start: (pagination.value.current - 1) * pagination.value.limit,
      limit: pagination.value.limit,
    };
    if (searchStr.value) {
      params.searchKey = searchStr.value;
    }
    const res = await getScriptVersionList(spaceId.value, scriptId.value, params);
    versionList.value = res.details;
    pagination.value.count = res.count;
    versionLoading.value = false;
  };

  // 点击新建版本
  const handleCreateVersionClick = (content: string) => {
    versionEditData.value = {
      panelOpen: true,
      editable: true,
      hook_revision: null,
      form: {
        id: 0,
        name: `v${dayjs().format('YYYYMMDDHHmmss')}`,
        memo: '',
        content,
      },
    };
  };

  // 编辑未上线版本
  const handleEditVersionClick = () => {
    if (unPublishVersion.value) {
      const { name, memo, content } = unPublishVersion.value.spec;
      versionEditData.value = {
        panelOpen: true,
        editable: true,
        hook_revision: unPublishVersion.value,
        form: {
          id: unPublishVersion.value?.id,
          name,
          memo,
          content,
        },
      };
    }
  };

  // 查看版本
  const handleViewVersionClick = (version: IScriptVersionListItem) => {
    const { name, memo, content } = version.hook_revision.spec;
    versionEditData.value = {
      panelOpen: true,
      editable: false,
      hook_revision: version.hook_revision,
      form: {
        id: version.hook_revision.id,
        name,
        memo,
        content,
      },
    };
  };

  // 上线版本
  const handlePublishClick = (version: IScriptVersion) => {
    InfoBox({
      title: t('确定上线此版本？'),
      subTitle: t('上线后，之前的线上版本将被置为「已下线」,若要使该版本在现网中生效，需要重新发布引用此脚本的服务'),
      'ext-cls': 'info-box-style',
      // infoType: 'warning',
      confirmText: t('确定'),
      cancelText: t('取消'),
      onConfirm: async () => {
        await publishVersion(spaceId.value, scriptId.value, version.id);
        unPublishVersion.value = null;
        versionEditData.value.panelOpen = false;
        getVersionList();
        Message({
          theme: 'success',
          message: t('上线版本成功'),
        });
      },
    } as any);
  };

  // 删除版本
  const handleDelClick = (version: IScriptVersion) => {
    isDeleteScriptVersionDialogShow.value = true;
    deleteScriptVersionItem.value = version;
  };

  const handleDeleteScriptVersionConfirm = async () => {
    await deleteScriptVersion(spaceId.value, scriptId.value, deleteScriptVersionItem.value!.id);
    if (versionList.value.length === 1 && pagination.value.current > 1) {
      pagination.value.current = pagination.value.current - 1;
    }
    unPublishVersion.value = null;
    versionEditData.value.panelOpen = false;
    isDeleteScriptVersionDialogShow.value = false;
    getVersionList();
    Message({
      theme: 'success',
      message: t('删除版本成功'),
    });
  };

  // 新建、编辑脚本后回调
  const handleVersionEditSubmitted = async (
    data: { id: number; name: string; memo: string; content: string },
    type: string,
  ) => {
    versionEditData.value.form = { ...data };
    refreshList();
    // 如果是创建新版本，则需要更新未发布版本数据
    if (type === 'create') {
      createBtnDisabled.value = true;
      scriptDetail.value.not_release_id = data.id;
      unPublishVersion.value = await getScriptVersionDetail(
        spaceId.value,
        scriptId.value,
        scriptDetail.value.not_release_id,
      );
      createBtnDisabled.value = false;
      handleEditVersionClick();
    } else {
      // 如果是编辑旧版本，则直接修改版本数据
      const { memo, content, name } = data;
      unPublishVersion.value!.spec.memo = memo;
      unPublishVersion.value!.spec.content = content;
      unPublishVersion.value!.spec.name = name;
    }
    versionEditData.value.editable = false;
  };

  // 版本对比
  const handleVersionDiff = (version: IScriptVersion) => {
    crtVersion.value = version;
    showVersionDiff.value = true;
  };

  const handleSearchInputChange = (val: string) => {
    if (!val) {
      refreshList();
    }
  };

  const refreshList = () => {
    isSearchEmpty.value = searchStr.value !== '';
    pagination.value.current = 1;
    getVersionList();
  };

  const handlePageLimitChange = (val: number) => {
    pagination.value.limit = val;
    refreshList();
  };

  const handleClose = () => {
    router.push({ name: 'script-list', params: { spaceId: spaceId.value } });
  };

  const clearStr = () => {
    searchStr.value = '';
    refreshList();
  };
</script>
<style lang="scss" scoped>
  .script-version-manage {
    padding: 24px;
    height: 100%;
    background: #f5f7fa;
  }
  .operation-area {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 16px;
    .search-input {
      width: 320px;
    }
    .search-input-icon {
      padding-right: 10px;
      color: #979ba5;
      background: #ffffff;
    }
  }
  .action-btns {
    .bk-button {
      margin-right: 8px;
    }
  }
  .version-data-container {
    height: calc(100% - 48px);
    border-radius: 2px;
    &.script-panel-open {
      display: flex;
      align-items: flex-start;
      background: #ffffff;
      .table-data-area {
        width: 216px;
        border: 1px solid #dcdee5;
      }
    }
    .table-data-area {
      position: relative;
      height: 100%;
      .back-table-btn {
        position: absolute;
        top: 16px;
        right: 10px;
        font-size: 12px;
        z-index: 10;
        .arrow-icon {
          margin-left: 4px;
        }
      }
    }
    .script-edit-area {
      width: calc(100% - 216px);
      height: 100%;
    }
  }
</style>
