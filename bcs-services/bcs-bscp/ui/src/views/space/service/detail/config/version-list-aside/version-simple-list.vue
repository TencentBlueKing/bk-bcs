<template>
  <section class="version-container">
    <div class="service-selector-wrapper">
      <ServiceSelector :value="props.appId" />
    </div>
    <bk-loading :loading="versionListLoading">
      <div class="version-search-wrapper">
        <SearchInput v-model="searchStr" class="config-search-input" :placeholder="t('版本名称')" />
      </div>
      <section class="versions-wrapper">
        <section v-if="!searchStr" class="unnamed-version">
          <section
            :class="['version-item', { active: versionData.id === 0 }]"
            @click="handleSelectVersion(unNamedVersion)">
            <i class="bk-bscp-icon icon-edit-small edit-icon" />
            <div class="version-name">{{ t('未命名版本') }}</div>
          </section>
          <div class="divider"></div>
        </section>
        <section
          v-for="version in versionsInView"
          :key="version.id"
          :class="['version-item', { active: versionData.id === version.id }]"
          @click="handleSelectVersion(version)">
          <div :class="['dot', version.status.publish_status]"></div>
          <bk-overflow-title class="version-name" type="tips">
            {{ version.spec.name }}
          </bk-overflow-title>
          <div
            v-if="version.status.fully_released"
            :class="['all-tag', { 'full-release': version.status.fully_release }]"
            v-bk-tooltips="{
              content: version.status.fully_release ? t('当前线上全量版本') : t('历史全量上线过的版本'),
            }">
            All
          </div>
          <Ellipsis class="action-more-icon" @mouseenter="handlePopShow(version, $event)" @mouseleave="handlePopHide" />
        </section>
        <TableEmpty v-if="searchStr && versionsInView.length === 0" :is-search-empty="true" @clear="searchStr = ''" />
      </section>
    </bk-loading>
    <VersionDiff v-model:show="showDiffPanel" :current-version="currentOperatingVersion" />
    <VersionOperateConfirmDialog
      v-model:show="showOperateConfirmDialog"
      :title="t('确认废弃该版本')"
      :tips="t('此操作不会删除版本，如需找回或彻底删除请去版本详情的废弃版本列表操作')"
      :confirm-fn="handleDeprecateVersion"
      :version="currentOperatingVersion" />
    <div
      class="action-list"
      ref="popover"
      v-show="popShow"
      @mouseenter="handlePopContentMouseEnter"
      @mouseleave="handlePopContentMouseLeave">
      <div class="action-item" @click="handleDiffDialogShow(selectedVersion!)">{{ t('版本对比') }}</div>
      <div
        v-bk-tooltips="{
          disabled: selectedVersion?.status.publish_status === 'not_released',
          placement: 'bottom',
          content: t('只支持未上线版本'),
        }"
        :class="['action-item', { disabled: selectedVersion?.status.publish_status !== 'not_released' }]"
        @click="handleDeprecateDialogShow(selectedVersion!)">
        {{ t('版本废弃') }}
      </div>
    </div>
  </section>
</template>
<script setup lang="ts">
  import { ref, onMounted, computed, watch } from 'vue';
  import { useRoute, useRouter } from 'vue-router';
  import { storeToRefs } from 'pinia';
  import { useI18n } from 'vue-i18n';
  import { Message } from 'bkui-vue';
  import { Ellipsis } from 'bkui-vue/lib/icon';
  import useConfigStore from '../../../../../../store/config';
  import { getConfigVersionList, deprecateVersion } from '../../../../../../api/config';
  import { GET_UNNAMED_VERSION_DATA } from '../../../../../../constants/config';
  import { IConfigVersion } from '../../../../../../../types/config';
  import ServiceSelector from '../../components/service-selector.vue';
  import SearchInput from '../../../../../../components/search-input.vue';
  import TableEmpty from '../../../../../../components/table/table-empty.vue';
  import VersionDiff from '../../config/components/version-diff/index.vue';
  import VersionOperateConfirmDialog from './version-operate-confirm-dialog.vue';

  const configStore = useConfigStore();
  const { versionData, refreshVersionListFlag, publishedVersionId } = storeToRefs(configStore);

  const route = useRoute();
  const router = useRouter();
  const { t } = useI18n();

  const props = defineProps<{
    bkBizId: string;
    appId: number;
  }>();

  const unNamedVersion: IConfigVersion = GET_UNNAMED_VERSION_DATA();
  const versionListLoading = ref(false);
  const versionList = ref<IConfigVersion[]>([]);
  const searchStr = ref('');
  const showDiffPanel = ref(false);
  const currentOperatingVersion = ref();
  const showOperateConfirmDialog = ref(false);
  const selectedVersion = ref<IConfigVersion>();
  const popShow = ref(false);
  const popover = ref<HTMLInputElement | null>(null);
  const popHideTimerId = ref(0);
  const isMouseenter = ref(false);

  const versionsInView = computed(() => {
    if (searchStr.value === '') {
      return versionList.value.slice(1);
    }
    return versionList.value.filter((item) => {
      const isNameMatched = item.spec.name.toLowerCase().includes(searchStr.value.toLocaleLowerCase());
      return item.id > 0 && isNameMatched;
    });
  });

  // 监听刷新版本列表标识，处理新增版本场景，默认选中新增的版本
  watch(refreshVersionListFlag, async (val) => {
    if (val) {
      await getVersionList();
      let versionDetail;
      // 判断当前是生成版本还是上线版本
      if (publishedVersionId.value) {
        versionDetail = versionList.value.find((item) => item.id === publishedVersionId.value);
        publishedVersionId.value = 0;
      } else {
        versionDetail = versionList.value[1];
      }
      if (versionDetail) {
        versionData.value = versionDetail;
        refreshVersionListFlag.value = false;
        // 默认选中新增的版本时，路由参数versionId需要更新
        router.push({ name: route.name as string, params: { versionId: versionDetail.id } });
      }
    }
  });

  watch(
    () => props.appId,
    () => {
      getVersionList();
    },
  );

  onMounted(async () => {
    init();
  });

  const init = async () => {
    await getVersionList();
    if (route.params.versionId) {
      const version = versionList.value.find((item) => item.id === Number(route.params.versionId));
      if (version) {
        versionData.value = version;
      }
    }
  };

  const getVersionList = async () => {
    try {
      versionListLoading.value = true;
      const params = {
        // 未命名版本不在实际的版本列表里，需要特殊处理
        start: 0,
        all: true,
      };
      const res = await getConfigVersionList(props.bkBizId, props.appId, params);
      versionList.value = [unNamedVersion, ...res.data.details];
      const index = versionList.value.findIndex((version: IConfigVersion) =>
        version.status.released_groups.some((group) => group.id === 0),
      );
      if (index > -1) versionList.value[index].status.fully_release = true;
    } catch (e) {
      console.error(e);
    } finally {
      versionListLoading.value = false;
    }
  };

  const handleSelectVersion = (version: IConfigVersion) => {
    configStore.$patch((state) => {
      state.allExistConfigCount = 0;
      state.conflictFileCount = 0;
    });
    versionData.value = version;
    const params: { spaceId: string; appId: number; versionId?: number } = {
      spaceId: props.bkBizId,
      appId: props.appId,
    };
    if (version.id !== 0) {
      params.versionId = version.id;
    }
    router.push({ name: route.name as string, params });
  };

  const handleDiffDialogShow = (version: IConfigVersion) => {
    currentOperatingVersion.value = version;
    showDiffPanel.value = true;
  };

  const handleDeprecateDialogShow = (version: IConfigVersion) => {
    if (version.status.publish_status !== 'not_released') {
      return;
    }
    currentOperatingVersion.value = version;
    showOperateConfirmDialog.value = true;
  };

  const handleDeprecateVersion = () =>
    new Promise(() => {
      const id = currentOperatingVersion.value.id;
      deprecateVersion(props.bkBizId, props.appId, id).then(() => {
        showOperateConfirmDialog.value = false;
        Message({
          theme: 'success',
          message: '版本废弃成功',
        });
        if (id !== versionData.value.id) {
          return;
        }

        const versions = versionsInView.value.filter((item) => item.id > 0);
        const index = versions.findIndex((item) => item.id === id);

        if (versions.length === 1) {
          handleSelectVersion(unNamedVersion);
        } else if (index === versions.length - 1) {
          handleSelectVersion(versions[index - 1]);
        } else {
          handleSelectVersion(versions[index + 1]);
        }

        versionList.value = versionList.value.filter((item) => item.id !== id);
      });
    });

  const handlePopShow = (version: IConfigVersion, event: any) => {
    selectedVersion.value = version;
    const element = event.target;
    const rect = element.getBoundingClientRect();
    const distanceToBottom = window.innerHeight - rect.bottom;
    if (distanceToBottom < 70) {
      popover.value!.style.top = `${rect.top - 140}px`;
    } else {
      popover.value!.style.top = `${rect.top - 20}px`;
    }
    popHideTimerId.value && clearTimeout(popHideTimerId.value);
    popShow.value = true;
  };

  const handlePopHide = () => {
    popHideTimerId.value = window.setTimeout(() => {
      popShow.value = false;
    }, 300);
  };

  const handlePopContentMouseEnter = () => {
    if (popHideTimerId.value) {
      isMouseenter.value = true;
      clearTimeout(popHideTimerId.value);
      popHideTimerId.value = 0;
    }
  };
  const handlePopContentMouseLeave = () => {
    if (isMouseenter.value) {
      handlePopHide();
      isMouseenter.value = false;
    }
  };
</script>

<style lang="scss" scoped>
  .version-container {
    height: 100%;
  }
  .service-selector-wrapper {
    padding: 10px 8px 9px;
    width: 280px;
    border-bottom: 1px solid #eaebf0;
  }
  .bk-nested-loading {
    height: calc(100% - 52px);
  }
  .version-search-wrapper {
    padding: 8px 16px;
  }
  .versions-wrapper {
    position: relative;
    height: calc(100% - 48px);
    overflow: auto;
  }
  .version-steps {
    padding: 16px 0;
    overflow: auto;
  }
  .unnamed-version {
    .divider {
      margin: 8px 24px;
      border-bottom: 1px solid #dcdee5;
    }
  }
  .version-item {
    position: relative;
    padding: 0 40px 0 48px;
    cursor: pointer;
    &.active {
      background: #e1ecff;
    }
    &:hover {
      background: #e1ecff;
    }
    .edit-icon {
      position: absolute;
      top: 10px;
      left: 24px;
      font-size: 22px;
      color: #979ba5;
    }
    .dot {
      position: absolute;
      left: 28px;
      top: 16px;
      width: 8px;
      height: 8px;
      border-radius: 50%;
      border: 1px solid #c4c6cc;
      background: #f0f1f5;
      &.not_released {
        border: 1px solid #ff9c01;
        background: #ffe8c3;
      }
      &.full_released,
      &.partial_released {
        border: 1px solid #3fc06d;
        background: #e5f6ea;
      }
    }
    .all-tag {
      position: absolute;
      right: 53px;
      top: 0;
      transform: translateY(50%);
      width: 31px;
      height: 22px;
      font-size: 12px;
      background: #fafbfd;
      border: 1px solid #dcdee5;
      border-radius: 2px;
      color: #63656e;
      text-align: center;
      line-height: 22px;
      &.full-release {
        background: #e4faf0;
        border: 1px solid #14a5684d;
        border-radius: 2px;
        color: #14a568;
      }
    }
  }
  .version-name {
    height: 42px;
    width: 140px;
    line-height: 42px;
    font-size: 12px;
    color: #313238;
    text-align: left;
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
  }
  .action-more-icon {
    position: absolute;
    top: 10px;
    right: 10px;
    transform: rotate(90deg);
    width: 22px;
    height: 22px;
    color: #979ba5;
    border-radius: 50%;
    cursor: pointer;
    &:hover {
      background: rgba(99, 101, 110, 0.1);
      color: #3a84ff;
    }
  }
  .list-pagination {
    margin-top: 16px;
  }
  .action-list {
    position: absolute;
    right: 25px;
    padding: 4px 0;
    border: 1px solid #dcdee5;
    box-shadow: 0 2px 6px 0 #0000001a;
    background-color: #fff;
    border-radius: 4px;
    .action-item {
      padding: 0 12px;
      height: 32px;
      line-height: 32px;
      color: #63656e;
      font-size: 12px;
      cursor: pointer;
      &:hover {
        background: #f5f7fa;
      }
      &.disabled {
        color: #dcdee5;
        cursor: not-allowed;
      }
    }
  }
</style>
