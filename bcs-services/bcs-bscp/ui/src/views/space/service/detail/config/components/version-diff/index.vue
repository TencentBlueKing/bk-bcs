<script setup lang="ts">
  import { ref, watch } from 'vue'
  import { useRoute } from 'vue-router'
  import { PlayShape } from 'bkui-vue/lib/icon'
  import { IConfigVersion, IConfigListQueryParams, IConfigDiffDetail, IConfigDetail, IConfigItem } from '../../../../../../../../types/config'
  import { getConfigVersionList, getConfigList, getConfigItemDetail } from '../../../../../../../api/config'
  import Diff from '../../../../../../../components/diff/index.vue'

  const props = defineProps<{
    show: boolean;
    showPublishBtn?: boolean;
    currentVersion: IConfigVersion;
    currentConfig?: number;
    baseVersionId?: number; // 默认选中的基准版本id
  }>()

  const emits = defineEmits(['update:show', 'publish'])

  const route = useRoute()
  const bkBizId = String(route.params.spaceId)
  const appId = Number(route.params.appId)
  const versionList = ref<IConfigVersion[]>([])
  const versionListLoading = ref(false)
  const selectedVersion = ref(0)
  // 当前版本配置列表
  const currentConfigLoading = ref(true)
  const currentList = ref<IConfigDetail[]>([])
  // 基准版本配置列表
  const baseList = ref<IConfigDetail[]>([])
  const baseConfigLoading = ref(false)
  // 汇总的配置项列表，包含未修改、增加、删除、修改的所有配置项
  const aggregatedList = ref<IConfigDiffDetail[]>([])
  const diffCount = ref(0)
  const selectedConfig = ref<IConfigDiffDetail>({
    id: 0,
    name: '',
    file_type: '',
    current: {
      signature: '',
      byte_size: '',
      update_at: ''
    },
    base: {
      signature: '',
      byte_size: '',
      update_at: ''
    }
  })

  watch(() => props.show, async(val) => {
    if (val) {
      currentConfigLoading.value = true
      if (props.baseVersionId) {
        selectedVersion.value = props.baseVersionId
      }
      getVersionList()
      const list = await getConfigsForVersion(props.currentVersion.id)
      currentList.value = await getConfigDetails(list, props.currentVersion.id)
      calcDiff()
      setSelectedConfig(props.currentConfig)
      currentConfigLoading.value = false
    }
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

  // 获取某个版本下配置项列表
  const getConfigsForVersion = async (releaseId?: number) => {
    const params: IConfigListQueryParams = {
      start: 0,
      limit: 200 // @todo 分页条数待确认
    }
    if (releaseId) {
      params.release_id = releaseId
    }

    const res = await getConfigList(bkBizId, appId, params)
    return res.details
  }

  // 获取配置项详情，主要为了拿signature
  const getConfigDetails = (list: IConfigItem[], version?: number) => {
    const params: { release_id?: number } = {}
    if (version) {
      params.release_id = version
    }
    return Promise.all(list.map(item => getConfigItemDetail(bkBizId, item.id, appId, params)))
  }

  // 选择对比基准版本
  const handleSelectVersion = async(val: number) => {
    baseConfigLoading.value = true
    selectedVersion.value = val
    const list = await getConfigsForVersion(val)
    baseList.value = await getConfigDetails(list, val)
    calcDiff()
    setSelectedConfig(selectedConfig.value.id)
    baseConfigLoading.value = false
  }

  // 计算配置被修改、被删除、新增的差异
  const calcDiff = () => {
    diffCount.value = 0
    const list: IConfigDiffDetail[]= []
    currentList.value.forEach(currentItem => {
        const { config_item } = currentItem
        const baseItem = baseList.value.find(item => config_item.id === item.config_item.id)
        // 在基准版本中存在
        if (baseItem) {
            const diffConfig = {
                id: config_item.id,
                name: config_item.spec.name,
                file_type: config_item.spec.file_type,
                current: {
                    ...currentItem.content,
                    update_at: config_item.revision?.update_at
                },
                base: {
                    ...baseItem.content,
                    update_at: baseItem.config_item.revision?.update_at
                }
            }
            if (currentItem.content.signature !== baseItem.content.signature) {
                diffCount.value++
              }
              list.push(diffConfig)
        } else { // 在基准版本中被删除
            diffCount.value++
            list.push({
                id: config_item.id,
                name: config_item.spec.name,
                file_type: config_item.spec.file_type,
                current: {
                    ...currentItem.content,
                    update_at: config_item.revision?.update_at
                },
                base: {
                    signature: '',
                    byte_size: '',
                    update_at: ''
                }
            })
        }
    })
    // 基准版本中的新增项
    baseList.value.forEach(baseItem => {
        const { config_item: base_config_item } = baseItem
        const currentItem = currentList.value.find(item => base_config_item.id === item.config_item.id)
        if (!currentItem) {
            diffCount.value++
            list.push({
                id: base_config_item.id,
                name: base_config_item.spec.name,
                file_type: base_config_item.spec.file_type,
                current: {
                    signature: '',
                    byte_size: '',
                    update_at: ''
                },
                base: {
                    ...baseItem.content,
                    update_at: ''
                }
            })
        }
    })
    aggregatedList.value = list
  }

  // 设置选中配置文件的数据
  const setSelectedConfig = (id: number|undefined) => {
    if (id) {
      const config = aggregatedList.value.find(item => item.id === id)
      if (config) {
        selectedConfig.value = config
        return
      }
    }
    selectedConfig.value = aggregatedList.value[0]
  }

  const handleClose = () => {
    selectedVersion.value = 0
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
    <bk-loading class="loading-wrapper" :loading="currentConfigLoading">
      <div class="version-diff-content">
          <aside class="config-list-side">
            <div class="title-area">
              <span class="title">配置项</span>
              <span>共 <span class="count">{{ diffCount }}</span> 项配置有差异</span>
            </div>
            <ul class="configs-wrapper">
              <li
                v-for="config in aggregatedList"
                :key="config.id"
                :class="{ active: selectedConfig.id === config.id }"
                @click="selectedConfig = config">
                <div class="name">{{ config.name }}</div>
                <PlayShape v-if="selectedConfig.id === config.id" class="arrow-icon" />
              </li>
            </ul>
          </aside>
          <div class="diff-content-area">
            <diff v-if="!!selectedConfig.id" :app-id="appId" :config="selectedConfig">
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
  .config-list-side {
    width: 264px;
    height: 100%;
    .title-area {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 0 24px;
      height: 49px;
      color: #979ba5;
      font-size: 12px;
      border-bottom: 1px solid #dcdee5;
      .title {
        font-size: 14px;
        font-weight: 700;
        color: #63656e;
      }
      .count {
        color: #313238;
      }
    }
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
    background: #313238;
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
