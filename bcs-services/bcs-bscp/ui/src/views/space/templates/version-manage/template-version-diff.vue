<script setup lang="ts">
  import { onMounted, ref, watch } from 'vue'
  import { ITemplateVersionItem } from '../../../../../types/template'
  import { IConfigDiffDetail } from '../../../../../types/config'
  import { getTemplateVersionsDetailByIds, getTemplateVersionList } from '../../../../api/template'
  import Diff from '../../../../components/diff/index.vue'

  const props = defineProps<{
    show: boolean;
    spaceId: string;
    templateSpaceId: number;
    crtVersion: { id: number; versionId: number; name: string; };
  }>()

  const emits = defineEmits(['update:show'])

  const selectedVersion = ref()
  const versionList = ref<ITemplateVersionItem[]>([])
  const versionListLoading = ref(false)
  const configDiffData = ref<IConfigDiffDetail>({
    id: props.crtVersion.id,
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

  watch(() => props.show, async val => {
    if (val) {
      getVersionList()
      const detail = await getTemplateVersionDetail(props.crtVersion.versionId)
      const { spec, content_spec } = detail
      configDiffData.value.file_type = spec.file_type
      configDiffData.value.base.signature = content_spec.signature
      configDiffData.value.base.byte_size = content_spec.byte_size
    }
  })

  // 获取版本列表
  const getVersionList = async() => {
    versionListLoading.value = true
    const params = {
      start: 0,
      all: true
    }
    const res = await getTemplateVersionList(props.spaceId, props.templateSpaceId, props.crtVersion.id, params)
    versionList.value = res.details.filter((item: ITemplateVersionItem) => item.id !== props.crtVersion.versionId)
    versionListLoading.value = false
  }

  const getTemplateVersionDetail = async(versionId: number) => {
    return getTemplateVersionsDetailByIds(props.spaceId, [versionId] ).then(res => res.details[0])
  }

  const handleSelectVersion = (id: number) => {
    const version = versionList.value.find(item => item.id === id)
    if (version) {
      // baseContent.value = version.spec.content
    }
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
    <div class="diff-content-area">
      <diff :app-id="props.crtVersion.id" :config="configDiffData">
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
                <div class="version-name">{{ props.crtVersion.name }}</div>
              </div>
            </slot>
        </template>
      </diff>
    </div>
    <div class="actions-btn">
      <bk-button @click="handleClose">关闭</bk-button>
    </div>
  </bk-sideslider>
</template>
<style lang="scss" scoped>
  .diff-content-area {
    height: calc(100vh - 100px);
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
  .actions-btn {
    padding: 8px 24px;
    background: #fafbfd;
    box-shadow: 0 -1px 0 0 #dcdee5;
    .bk-button {
      min-width: 88px;
    }
  }
</style>
