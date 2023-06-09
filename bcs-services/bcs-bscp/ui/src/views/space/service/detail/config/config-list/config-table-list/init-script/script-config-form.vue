<script setup lang="ts">
  import { ref, onMounted, nextTick } from 'vue'
  import { useRouter } from 'vue-router'
  import { storeToRefs } from 'pinia'
  import { Share } from 'bkui-vue/lib/icon'
  import { IScriptItem, IScriptVersion } from '../../../../../../../../../types/script'
  import { useGlobalStore } from '../../../../../../../../store/global'
  import { getScriptList, getScriptVersionList, getScriptVersionDetail } from '../../../../../../../../api/script'
  import ScriptEditor from '../../../../../../scripts/components/script-editor.vue'

  const { spaceId } = storeToRefs(useGlobalStore())

  const router = useRouter()

  const props = defineProps<{
    scriptId: number|string;
    versionId: number|string;
  }>()

  const emits = defineEmits(['change'])

  const content = ref('')
  const contentLoading = ref(false)
  const scriptsLoading = ref(false)
  const scriptsData = ref<IScriptItem[]>([])
  const scriptVersionsLoading = ref(false)
  const scriptVersions = ref<IScriptVersion[]>([])
  const formRef = ref()

  onMounted(() => {
    getScripts()
    if (props.scriptId) {
      getScriptVersions()
      if (props.versionId) {
        getScriptDetail()
      }
    }
  })

  // 获取脚本列表
  const getScripts = async () => {
    scriptsLoading.value = true
    const params = {
      start: 0,
      limit: 10,
      all: true
    }
    const res = await getScriptList(spaceId.value, params)
    scriptsData.value = res.details
    scriptsLoading.value = false
  }

  // 获取脚本版本列表
  const getScriptVersions = async () => {
    scriptVersionsLoading.value = true
    const res = await getScriptVersionList(spaceId.value, <number>props.scriptId, { start: 0, limit: 200 })
    scriptVersions.value = res.details
    scriptVersionsLoading.value = false
  }

  const getScriptDetail = async () => {
    contentLoading.value = true
    const res = await getScriptVersionDetail(spaceId.value, <number>props.scriptId, <number>props.versionId)
    content.value = res.spec.content
    contentLoading.value = false
  }

  const handleScriptChange = (id: number) => {
    handleDataChange({ scriptId: id, versionId: '' })
    content.value = ''
    nextTick(() => {
      getScriptVersions()
    })
  }

  const handleVersionChange = (id: number) => {
    handleDataChange({ scriptId: props.scriptId, versionId: id })
    nextTick(() => {
      getScriptDetail()
    })
  }

  const goToScripts = () => {
    const { href } = router.resolve({ name: 'scripts-management', params: { spaceId: spaceId.value } })
    window.open(href, '__blank')
  }

  const handleDataChange = (data: { scriptId: number|string; versionId: number|string; }) => {
    emits('change', data)
  }
</script>
<template>
  <bk-form ref="formRef" form-type="vertical">
    <bk-form-item label="脚本引用" :required="true">
      <div class="script-cite">
        <bk-select
          class="script-selector"
          :model-value="props.scriptId"
          :clearable="false"
          :loading="scriptsLoading"
          @change="handleScriptChange">
          <bk-option
            v-for="option in scriptsData"
            :key="option.id"
            :value="option.id"
            :label="option.spec.name">
          </bk-option>
        </bk-select>
        <bk-select
          class="version-selector"
          :model-value="props.versionId"
          :clearable="false"
          :loading="scriptVersionsLoading"
          @change="handleVersionChange">
          <bk-option
            v-for="option in scriptVersions"
            :key="option.id"
            :value="option.id"
            :label="option.spec.name">
          </bk-option>
        </bk-select>
        <Share class="link-icon" @click="goToScripts" />
      </div>
    </bk-form-item>
    <bk-form-item label="脚本内容">
      <ScriptEditor v-model="content" :editable="false" />
    </bk-form-item>
  </bk-form>
</template>
<style lang="scss" scoped>
  .script-cite {
    display: flex;
    align-items: center;
    .script-selector {
      margin-right: 8px;
      width: 240px;
    }
    .version-selector {
      margin-right: 8px;
      width: 120px;
    }
    .link-icon {
      font-size: 12px;
      color: #3a84ff;
      cursor: pointer;
    }
  }
  .script-content {
    height: 600px;
  }
</style>