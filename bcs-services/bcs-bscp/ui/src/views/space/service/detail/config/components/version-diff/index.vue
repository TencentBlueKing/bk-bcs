<script setup lang="ts">
  import { ref, computed, watch } from 'vue'
  import { useRoute } from 'vue-router'
  import { IConfigVersion } from '../../../../../../../../types/config'
  import { IDiffDetail } from '../../../../../../../../types/service'
  import { getConfigVersionList } from '../../../../../../../api/config'
  import AsideMenu from './aside-menu/index.vue'
  import Diff from '../../../../../../../components/diff/index.vue'

  const props = defineProps<{
    show: boolean;
    showPublishBtn?: boolean; // 是否显示发布按钮
    currentVersion: IConfigVersion; // 当前版本
    currentConfigId?: number; // 配置项id
    baseVersionId?: number; // 默认选中的基准版本id
  }>()

  const emits = defineEmits(['update:show', 'publish'])

  const route = useRoute()
  const bkBizId = String(route.params.spaceId)
  const appId = Number(route.params.appId)
  const versionList = ref<IConfigVersion[]>([])
  const versionListLoading = ref(false)
  const selectedVersion = ref()
  const selectedDiff = ref<IDiffDetail>({
    contentType: 'text',
    current: {
      language: '',
      content: ''
    },
    base: {
      language: '',
      content: ''
    }
  })

  const loading = computed(() => {
    return versionListLoading.value
  })

  watch(() => props.show, async(val) => {
    if (val) {
      getVersionList()
    }
  })

  watch(() => props.baseVersionId, (val) => {
    selectedVersion.value = val
  })

  // 获取所有对比基准版本
  const getVersionList = async() => {
    try {
      versionListLoading.value = true
      const res = await getConfigVersionList(bkBizId, appId, { start: 0, all: true })
      versionList.value = res.data.details.filter((item: IConfigVersion) => item.id !== props.currentVersion.id)
    } catch (e) {
      console.error(e)
    } finally {
      versionListLoading.value = false
    }
  }

  // 选择对比基准版本
  const handleSelectVersion = async(val: number) => {
    selectedVersion.value = val
  }

  const handleSelect = (data: IDiffDetail) => {
    selectedDiff.value = data
  }

  const handleClose = () => {
    selectedVersion.value = undefined
    versionList.value = []
    emits('update:show', false)
  }

</script>
<template>
  <bk-sideslider
    :is-show="props.show"
    title="版本对比"
    :width="1200"
    @closed="handleClose">
    <bk-loading class="loading-wrapper" :loading="loading">
      <div class="version-diff-content">
          <AsideMenu
            :base-version-id="selectedVersion"
            :current-version-id="currentVersion.id"
            :current-config-id="props.currentConfigId"
            @selected="handleSelect" />
          <div :class="['diff-content-area', { light: selectedDiff.contentType === 'file' }]">
            <diff :diff="selectedDiff" :loading="false">
              <template #leftHead>
                  <slot name="baseHead">
                    <div class="diff-panel-head">
                      <div class="version-tag base-version">对比版本</div>
                      <bk-select
                        :model-value="selectedVersion"
                        style="width: 320px;"
                        :loading="versionListLoading"
                        :clearable="false"
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
        <bk-button v-if="showPublishBtn" class="publish-btn" theme="primary" @click="emits('publish')">上线版本</bk-button>
        <bk-button @click="handleClose">关闭</bk-button>
      </div>
    </template>
  </bk-sideslider>
</template>
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