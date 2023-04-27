<script setup lang="ts">
  import { ref, computed, watch } from 'vue'
  import Diff from '../../../../../components/diff/index.vue'
  import { IConfigItem, IConfigDetail, IConfigVersion, IConfigListQueryParams } from '../../../../../../types/config'
  import { FilterOp, RuleOp } from '../../../../../types'
  import { getConfigVersionList, getConfigList, getConfigItemDetail } from '../../../../../api/config'

  const emits = defineEmits(['update:show'])

  const props = defineProps<{
    show: boolean,
    bkBizId: string,
    appId: number,
    versionName: string,
    releaseId: number,
    config: IConfigItem
  }>()

  const diffConfig = ref({
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
  const versionListLoading = ref(true)
  const versionList = ref<IConfigVersion[]>([])
  const selectedVersion = ref<number>()
  const baseConfigLoading = ref(false)
  const filter = {
    op: FilterOp.AND,
    rules: [{
      field: "deprecated",
      op: RuleOp.eq,
      value: false
    }]
  }
  const page = {
    count: false,
    start: 0,
    limit: 200 // @todo 分页条数待确认
  }

  const title = computed(() => {
    return `配置项对比 - ${props.config?.spec?.name}`
  })
  
  watch(() => props.show, async(val) => {
    if (val) {
      getVersionList()
      const data = await getConfigDetail(props.config.id, props.releaseId)
      const { config_item, content } = data
      diffConfig.value = {
        id: props.config.id,
        name: props.config.spec.name,
        file_type: props.config.spec.file_type,
        current: {
          signature: content.signature,
          byte_size: content.byte_size,
          update_at: config_item.revision?.update_at
        },
        base: {
          signature: '',
          byte_size: '',
          update_at: ''
        }
      }
    }
  })

  // 获取所有版本
  const getVersionList = async() => {
    try {
      versionListLoading.value = true
      const res = await getConfigVersionList(props.bkBizId, props.appId, { start: 0, all: true })
      versionList.value = res.data.details.filter((item: IConfigVersion) => item.id !== props.releaseId)
    } catch (e) {
      console.error(e)
    } finally {
      versionListLoading.value = false
    }
  }

  // 获取某个版本下配置项列表
  const getConfigsForVersion = async () => {
    baseConfigLoading.value = true
    try {
      const params: IConfigListQueryParams = {
          release_id: selectedVersion.value,
          start: 0,
          limit: 200 // @todo 分页条数待确认
      }

      const res = await getConfigList(props.bkBizId, props.appId, params)
      const baseConfig = res.details.find((item: IConfigItem) => item.id === props.config.id)
      if (baseConfig) {
        const detailData = await getConfigDetail(props.config.id, <number>selectedVersion.value)
        const { config_item, content } = detailData
        diffConfig.value.base = {
          signature: content.signature,
          byte_size: content.byte_size,
          update_at: config_item.revision?.update_at
        }
      }
    } catch (e) {
      console.error(e)
    } finally {
      baseConfigLoading.value = false
    }
  }

  // 获取配置项详情
  const getConfigDetail = async (id: number, version: number) => {
    const params: { release_id?: number } = {}
    if (version) {
      params.release_id = version
    }

    return getConfigItemDetail(props.bkBizId, id, props.appId, params)
  }

  const handleSelectVersion = (val: number) => {
    selectedVersion.value = val
    getConfigsForVersion()
  }

  const handleClose = () => {
    emits('update:show', false)
  }

</script>
<template>
  <bk-dialog
    :title="title"
    ext-cls="version-compare-dialog"
    :width="1200"
    :is-show="props.show"
    :esc-close="false"
    :quick-close="false"
    @closed="handleClose">
      <div class="diff-content-wrapper">
        <bk-loading style="height: 100%;" :loading="baseConfigLoading">
          <diff v-if="!baseConfigLoading" :panelName="props.versionName" :config="diffConfig">
            <template #leftHead>
              <div class="version-selector">
                对比版本：
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
            </template>
          </diff>
        </bk-loading>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <bk-button @click="handleClose">关闭</bk-button>
        </div>
      </template>
  </bk-dialog>
</template>
<style lang="scss" scoped>
  .diff-content-wrapper {
    margin-bottom: 20px;
    height: 580px;
    border: 1px solid #dcdee5;
    border-left: none;
  }
  .version-selector {
      display: flex;
      align-items: center;
      height: 100%;
      padding: 0 24px;
      font-size: 12px;
  }
</style>