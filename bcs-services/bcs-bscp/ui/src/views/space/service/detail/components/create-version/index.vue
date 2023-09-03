<script setup lang="ts">
  import { ref, watch, onMounted } from 'vue'
  import { storeToRefs } from 'pinia'
  import { Message } from 'bkui-vue'
  import { useConfigStore } from '../../../../../../store/config'
  import { IConfigListQueryParams, IConfigVersion } from '../../../../../../../types/config'
  import { IVariableEditParams } from '../../../../../../../types/variable';
  import { getConfigList } from '../../../../../../api/config'
  import CreateVersionSlider from './create-version-slider.vue'
  import VersionDiff from '../../config/components/version-diff/index.vue'

  const props = defineProps<{
    bkBizId: string,
    appId: number
  }>()

  const emits = defineEmits(['confirm'])

  const { versionData } = storeToRefs(useConfigStore())

  const isVersionSliderShow = ref(false)
  const isDiffSliderShow = ref(false)
  const configCount = ref(0)
  const variableList = ref<IVariableEditParams[]>([])
  const loading = ref(false)
  const createPending = ref(false)
  const createSliderRef = ref()

  watch(() => versionData.value.id, val => {
    if (val === 0) {
      getconfigsCountData()
    }
  })

  onMounted(() => {
    if (versionData.value.id === 0) {
      getconfigsCountData()
    }
  })

  // 调用列表接口查询配置项数量
  const getconfigsCountData = async () => {
    loading.value = true
    try {
      const params: IConfigListQueryParams = {
        start: 0,
        limit: 1
      }
      const res = await getConfigList(props.bkBizId, props.appId, params)
      configCount.value = res.count
    } catch (e) {
      console.error(e)
    } finally {
      loading.value = false
    }
  }

  const handleVersionSliderOpen = () => {
    isVersionSliderShow.value = true
  }

  const handleDiffSliderOpen = (variables: IVariableEditParams[]) => {
    isDiffSliderShow.value = true
    variableList.value = variables
  }

  // 触发生成版本确认操作
  const triggerCreate = async() => {
    try {
      createPending.value = true
      await createSliderRef.value.confirm()
    } catch (e) {
      console.log(e)
    } finally {
      createPending.value = false
    }
  }

  const handleCreated = (versionData: IConfigVersion, isPublish: boolean) => {
    isDiffSliderShow.value = false
    isVersionSliderShow.value = false
    emits('confirm', versionData, isPublish)
    Message({ theme: 'success', message: '新版本已生成' })
  }

</script>
<template>
  <bk-button
    v-if="versionData.id === 0"
    class="trigger-button"
    theme="primary"
    :loading="loading"
    :disabled="configCount === 0"
    @click="handleVersionSliderOpen">
    生成版本
  </bk-button>
  <CreateVersionSlider
    v-model:show="isVersionSliderShow"
    ref="createSliderRef"
    :bk-biz-id="props.bkBizId"
    :app-id="props.appId"
    :is-diff-slider-show="isDiffSliderShow"
    @open-diff="handleDiffSliderOpen"
    @created="handleCreated" />
    <VersionDiff
      v-model:show="isDiffSliderShow"
      :current-version="versionData">
      <template #footerActions>
        <bk-button
          class="create-version-btn"
          theme="primary"
          :loading="createPending"
          @click="triggerCreate">
          生成版本
        </bk-button>
        <bk-button @click="isDiffSliderShow = false">关闭</bk-button>
      </template>
    </VersionDiff>
</template>
<style lang="scss" scoped>
  .trigger-button {
    margin-left: 8px;
  }
  .create-version-btn {
    margin-right: 8px;
  }
</style>
