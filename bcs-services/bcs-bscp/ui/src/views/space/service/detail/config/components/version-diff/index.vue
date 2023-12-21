<template>
  <bk-sideslider
    :is-show="props.show"
    title="版本对比"
    ext-cls="config-version-diff-slider"
    :width="1200"
    @closed="handleClose"
  >
    <bk-loading class="loading-wrapper" :loading="loading">
      <div v-if="!loading" class="version-diff-content">
        <AsideMenu
          :base-version-id="selectedBaseVersion"
          :current-version-id="currentVersion.id"
          :un-named-version-variables="props.unNamedVersionVariables"
          :selected-config="props.selectedConfig"
          :selected-config-kv="props.selectedConfigKv"
          @selected="handleSelectDiffItem"
        />
        <div :class="['diff-content-area', { light: diffDetailData.contentType === 'file' }]">
          <diff :diff="diffDetailData" :id="appId" :loading="false">
            <template #leftHead>
              <slot name="baseHead">
                <div class="diff-panel-head">
                  <div class="version-tag base-version">{{showPublishBtn ? '线上版本' : '对比版本'}}</div>
                  <bk-select
                    :model-value="selectedBaseVersion"
                    style="width: 320px"
                    :loading="versionListLoading"
                    :clearable="false"
                    no-data-text="暂无数据"
                    @change="handleSelectVersion"
                  >
                    <bk-option
                      v-for="version in versionList"
                      :key="version.id"
                      :label="version.spec.name"
                      :value="version.id"
                    >
                    </bk-option>
                  </bk-select>
                </div>
              </slot>
            </template>
            <template #rightHead>
              <slot name="currentHead">
                <div class="diff-panel-head">
                  <div class="version-tag">当前版本</div>
                  <div class="version-name">{{ props.currentVersion.spec.name }}</div>
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
          <bk-button v-if="showPublishBtn" class="publish-btn" theme="primary" @click="emits('publish')"
            >上线版本</bk-button
          >
          <bk-button @click="handleClose">关闭</bk-button>
        </slot>
      </div>
    </template>
  </bk-sideslider>
</template>
<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { useRoute } from 'vue-router';
import { IConfigVersion, IConfigDiffSelected } from '../../../../../../../../types/config';
import { IDiffDetail } from '../../../../../../../../types/service';
import { IVariableEditParams } from '../../../../../../../../types/variable';
import { getConfigVersionList } from '../../../../../../../api/config';
import AsideMenu from './aside-menu/index.vue';
import Diff from '../../../../../../../components/diff/index.vue';

const props = defineProps<{
  show: boolean;
  showPublishBtn?: boolean; // 是否显示发布按钮
  currentVersion: IConfigVersion; // 当前版本详情
  unNamedVersionVariables?: IVariableEditParams[];
  baseVersionId?: number; // 默认选中的基准版本id
  selectedConfig?: IConfigDiffSelected; // 默认选中的配置文件id
  versionDiffList?: IConfigVersion[];
  selectedConfigKv?: number // 默认选中的文件id
}>();

const emits = defineEmits(['update:show', 'publish']);

const route = useRoute();
const bkBizId = String(route.params.spaceId);
const appId = Number(route.params.appId);
const versionList = ref<IConfigVersion[]>([]);
const versionListLoading = ref(false);
const selectedBaseVersion = ref(); // 基准版本ID
const diffDetailData = ref<IDiffDetail>({
  // 差异详情数据
  contentType: 'text',
  current: {
    language: '',
    content: '',
  },
  base: {
    language: '',
    content: '',
  },
});

const loading = computed(() => versionListLoading.value);


watch(
  () => props.show,
  async (val) => {
    if (val) {
      getVersionList();
      if (props.baseVersionId) {
        selectedBaseVersion.value = props.baseVersionId;
      }
    }
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
    const res = await getConfigVersionList(bkBizId, appId, { start: 0, all: true });
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
};

const handleClose = () => {
  selectedBaseVersion.value = undefined;
  versionList.value = [];
  emits('update:show', false);
};
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
  &:not(.light) {
    .diff-panel-head {
      background: #313238;
      :deep(.bk-select) {
        .bk-input {
          border-color: #63656e;
        }
        .bk-input--text {
          color: #b6b6b6;
          background: #313238;
        }
      }
    }
    :deep(.right-panel) {
      border-color: #1d1d1d;
    }
  }
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
