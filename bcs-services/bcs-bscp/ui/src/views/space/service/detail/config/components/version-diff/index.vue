<template>
  <bk-sideslider
    :is-show="props.show"
    :title="t('版本对比')"
    ext-cls="config-version-diff-slider"
    :width="1200"
    @closed="handleClose">
    <bk-loading class="loading-wrapper" :loading="loading">
      <div v-if="!loading" class="version-diff-content">
        <AsideMenu
          :base-version-id="selectedBaseVersion"
          :current-version-id="currentVersion.id"
          :un-named-version-variables="props.unNamedVersionVariables"
          :selected-config="props.selectedConfig"
          :selected-kv-config-id="selectedKV"
          :is-publish="props.showPublishBtn"
          @selected="handleSelectDiffItem"
          @render="publishBtnLoading = $event" />
        <div class="diff-content-area">
          <diff :diff="diffDetailData" :id="appId" :selected-kv-config-id="selectedKV" :loading="false">
            <template #leftHead>
              <slot name="baseHead">
                <div class="diff-panel-head">
                  <div class="version-tag base-version">{{ showPublishBtn ? t('线上版本') : t('对比版本') }}</div>
                  <bk-select
                    :model-value="selectedBaseVersion"
                    :style="{ width: locale === 'zh-cn' ? '320px' : '300px' }"
                    :loading="versionListLoading"
                    :clearable="false"
                    :no-data-text="t('暂无数据')"
                    :placeholder="t('请选择')"
                    @change="handleSelectVersion">
                    <bk-option
                      v-for="version in versionList"
                      :key="version.id"
                      :label="version.spec.name"
                      :value="version.id">
                    </bk-option>
                  </bk-select>
                </div>
              </slot>
            </template>
            <template #rightHead>
              <slot name="currentHead">
                <div class="diff-panel-head">
                  <div class="version-tag">{{ showPublishBtn ? t('待上线版本') : t('当前版本') }}</div>
                  <bk-overflow-title class="version-name" type="tips">
                    {{ props.currentVersion.spec.name }}
                  </bk-overflow-title>
                </div>
              </slot>
            </template>
          </diff>
        </div>
      </div>
    </bk-loading>
    <template #footer>
      <div class="actions-btns">
        <slot name="footerActions">
          <bk-button
            v-if="showPublishBtn"
            :loading="publishBtnLoading || props.btnLoading"
            :disabled="publishBtnLoading || props.btnLoading"
            class="publish-btn"
            theme="primary"
            @click="emits('publish')">
            {{ isApprovalMode ? t('通过') : t('上线版本') }}
          </bk-button>
          <bk-button v-if="isApprovalMode" :loading="publishBtnLoading || props.btnLoading" @click="emits('reject')">
            {{ t('驳回') }}
          </bk-button>
          <bk-button v-else @click="handleClose">{{ t('关闭') }}</bk-button>
        </slot>
      </div>
    </template>
  </bk-sideslider>
</template>
<script setup lang="ts">
  import { ref, computed, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { useRoute } from 'vue-router';
  import { IConfigVersion, IConfigDiffSelected } from '../../../../../../../../types/config';
  import { IDiffDetail } from '../../../../../../../../types/service';
  import { IVariableEditParams } from '../../../../../../../../types/variable';
  import { getConfigVersionList } from '../../../../../../../api/config';
  import AsideMenu from './aside-menu/index.vue';
  import Diff from '../../../../../../../components/diff/index.vue';

  const getDefaultDiffData = (): IDiffDetail => ({
    // 差异详情数据
    id: 0,
    contentType: 'text',
    is_secret: false,
    secret_hidden: false,
    current: {
      language: '',
      content: '',
    },
    base: {
      language: '',
      content: '',
    },
  });

  const { t, locale } = useI18n();
  const props = defineProps<{
    show: boolean;
    showPublishBtn?: boolean; // 是否显示发布按钮
    currentVersion: IConfigVersion; // 当前版本详情
    unNamedVersionVariables?: IVariableEditParams[];
    baseVersionId?: number; // 默认选中的基准版本id
    selectedConfig?: IConfigDiffSelected; // 默认选中的配置文件
    versionDiffList?: IConfigVersion[];
    selectedKvConfigId?: number; // 选中的kv类型配置id
    isApprovalMode?: boolean; // 是否审批模式(操作记录-去审批-拒绝)
    btnLoading?: boolean;
  }>();

  const emits = defineEmits(['update:show', 'publish', 'reject']);

  const route = useRoute();
  const bkBizId = ref(String(route.params.spaceId));
  const appId = ref(Number(route.params.appId));
  const versionList = ref<IConfigVersion[]>([]);
  const versionListLoading = ref(false);
  const selectedBaseVersion = ref(); // 基准版本ID
  const selectedKV = ref(props.selectedKvConfigId);
  const diffDetailData = ref<IDiffDetail>(getDefaultDiffData());

  const loading = computed(() => versionListLoading.value);
  const publishBtnLoading = ref(true);

  watch(
    () => props.show,
    async (val) => {
      publishBtnLoading.value = true;
      if (val) {
        await getVersionList();
        if (props.baseVersionId) {
          selectedBaseVersion.value = props.baseVersionId;
        } else if (versionList.value.length > 0) {
          selectedBaseVersion.value = versionList.value[0].id;
        }
      }
    },
  );

  watch(
    () => route.params.appId,
    (val) => {
      if (val) {
        appId.value = Number(val);
      }
    },
  );

  watch(
    () => props.selectedKvConfigId,
    (val) => {
      selectedKV.value = val;
    },
  );

  // 获取所有对比基准版本
  const getVersionList = async () => {
    try {
      versionListLoading.value = true;
      if (props.versionDiffList) {
        versionList.value = props.versionDiffList;
        return;
      }
      const res = await getConfigVersionList(bkBizId.value, appId.value, { start: 0, all: true });
      versionList.value = res.data.details.filter((item: IConfigVersion) => item.id !== props.currentVersion.id);
    } catch (e) {
      console.error(e);
    } finally {
      versionListLoading.value = false;
    }
  };

  // 选择对比基准版本
  const handleSelectVersion = async (val: number) => {
    selectedBaseVersion.value = val;
  };

  // 选中对比对象，配置或者脚本
  const handleSelectDiffItem = (data: IDiffDetail) => {
    diffDetailData.value = data;
    if (data.contentType === 'singleLineKV') {
      selectedKV.value = data.id as number;
    }
  };

  const handleClose = () => {
    selectedBaseVersion.value = undefined;
    versionList.value = [];
    selectedKV.value = 0;
    diffDetailData.value = getDefaultDiffData();
    emits('update:show', false);
  };

  defineExpose({
    handleSelectVersion,
  });
</script>
<style lang="scss" scoped>
  .loading-wrapper {
    height: calc(100vh - 106px);
  }
  .version-diff-content {
    display: flex;
    align-items: center;
    height: 100%;
  }
  .configs-wrapper {
    height: calc(100% - 49px);
    overflow: auto;
    & > li {
      display: flex;
      align-items: center;
      justify-content: space-between;
      position: relative;
      padding: 0 24px;
      height: 41px;
      color: #313238;
      border-bottom: 1px solid #dcdee5;
      cursor: pointer;
      &:hover {
        background: #e1ecff;
        color: #3a84ff;
      }
      &.active {
        background: #e1ecff;
        color: #3a84ff;
      }
      .name {
        width: calc(100% - 24px);
        line-height: 16px;
        font-size: 12px;
        white-space: nowrap;
        text-overflow: ellipsis;
        overflow: hidden;
      }
      .arrow-icon {
        position: absolute;
        top: 50%;
        right: 5px;
        transform: translateY(-60%);
        font-size: 12px;
        color: #3a84ff;
      }
    }
  }
  .diff-content-area {
    width: calc(100% - 264px);
    height: 100%;
  }
  .diff-panel-head {
    display: flex;
    align-items: center;
    padding: 0 16px;
    width: 100%;
    height: 100%;
    font-size: 12px;
    color: #b6b6b6;
    .version-tag {
      margin-right: 8px;
      padding: 0 10px;
      height: 22px;
      line-height: 22px;
      font-size: 12px;
      color: #14a568;
      background: #e4faf0;
      border-radius: 2px;
      &.base-version {
        color: #3a84ff;
        background: #edf4ff;
      }
    }
    .version-name {
      max-width: 300px;
    }
  }
  .actions-btns {
    padding: 0 24px;
    .bk-button {
      min-width: 88px;
    }
    .publish-btn {
      margin-right: 8px;
    }
  }
</style>
<style lang="scss">
  .config-version-diff-slider {
    .bk-modal-body {
      transform: none;
    }
  }
</style>
