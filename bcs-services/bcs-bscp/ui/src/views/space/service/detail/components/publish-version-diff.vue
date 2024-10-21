<template>
  <VersionDiff
    ref="diffRef"
    :btn-loading="props.btnLoading"
    :show="props.show"
    :current-version="props.currentVersion"
    :base-version-id="baseVersionId"
    :show-publish-btn="true"
    :version-diff-list="props.versionList"
    :is-approval-mode="props.isApprovalMode"
    @publish="emits('publish')"
    @update:show="emits('close')">
    <template #baseHead>
      <div class="panel-header">
        <div class="version-tag base">{{ $t('线上版本') }}</div>
        <div class="base-version-tab">
          <bk-tab :active="activeTab" type="unborder-card" @change="handleVersionChange">
            <bk-tab-panel v-for="version in props.versionList" :key="version.id" :name="version.id">
              <template #label>
                <div class="version-tab-label">
                  <span class="name">{{ version.spec.name }}</span>
                  <ReleasedGroupViewer
                    placement="bottom-start"
                    :bk-biz-id="props.bkBizId"
                    :app-id="props.appId"
                    :groups="version.status.released_groups">
                    <i class="bk-bscp-icon icon-resources-fill view-detail-icon" />
                  </ReleasedGroupViewer>
                </div>
              </template>
            </bk-tab-panel>
          </bk-tab>
        </div>
      </div>
    </template>
    <template #currentHead>
      <div class="panel-header">
        <div class="version-tag current">{{ $t('待上线版本') }}</div>
        <div class="version-title">
          <span class="text">{{ props.currentVersion.spec.name }}</span>
          <ReleasedGroupViewer
            placement="bottom-start"
            :bk-biz-id="props.bkBizId"
            :app-id="props.appId"
            :groups="props.currentVersionGroups"
            :is-pending="true">
            <i class="bk-bscp-icon icon-resources-fill"></i>
          </ReleasedGroupViewer>
        </div>
      </div>
    </template>
  </VersionDiff>
</template>
<script setup lang="ts">
  import { ref, watch } from 'vue';
  import { IConfigVersion, IReleasedGroup } from '../../../../../../types/config';
  import VersionDiff from '../config/components/version-diff/index.vue';
  import ReleasedGroupViewer from '../config/components/released-group-viewer.vue';

  const props = defineProps<{
    bkBizId: string;
    appId: number;
    show: boolean;
    currentVersion: IConfigVersion; // 当前版本详情信息
    baseVersionId: number; // 对比目标版本id
    versionList: IConfigVersion[]; // 对比版本列表
    currentVersionGroups: IReleasedGroup[]; // 当前版本上线分组实例
    isApprovalMode?: boolean; // 是否审批模式(操作记录-去审批-拒绝)
    btnLoading?: boolean;
  }>();

  const emits = defineEmits(['publish', 'close']);

  const activeTab = ref(props.baseVersionId);
  const diffRef = ref();

  watch(
    () => props.baseVersionId,
    (val) => {
      activeTab.value = val;
    },
  );

  const handleVersionChange = (val: number) => {
    activeTab.value = val;
    if (diffRef.value) {
      diffRef.value.handleSelectVersion(val);
    }
  };
</script>
<style lang="scss" scoped>
  .panel-header {
    display: flex;
    align-items: center;
    padding: 0 16px;
    height: 100%;
    background: transparent;
  }
  .version-tag {
    flex-shrink: 0;
    margin-right: 8px;
    padding: 0 10px;
    height: 22px;
    line-height: 22px;
    font-size: 12px;
    color: #14a568;
    background: #e4faf0;
    border-radius: 2px;
    &.base {
      color: #3a84ff;
      background: #edf4ff;
    }
  }
  .base-version-tab {
    flex: 1;
    overflow: hidden;
    :deep(.bk-tab-header) {
      border-bottom: none;
    }
    :deep(.bk-tab-content) {
      display: none;
    }
  }
  .version-tab-label {
    position: relative;
    .view-detail-icon {
      position: absolute;
      right: -20px;
      top: 17px;
      font-size: 16px;
      color: #979ba5;
      &:hover {
        color: #3a84ff;
        cursor: pointer;
      }
    }
  }
  .version-title {
    flex: 1;
    display: flex;
    align-items: center;
    padding-right: 20px;
    color: #b6b6b6;
    font-size: 12px;
    overflow: hidden;
    .bk-bscp-icon {
      margin-left: 3px;
      font-size: 16px;
      color: #979ba5;
      &:hover {
        color: #3a84ff;
        cursor: pointer;
      }
    }
  }
</style>
