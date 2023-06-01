<script setup lang="ts">
  import { ref, onMounted } from 'vue'
  import { useRouter } from 'vue-router'
  import { storeToRefs } from 'pinia'
  import { Share } from 'bkui-vue/lib/icon'
  import { IScriptItem, IScriptVersion } from '../../../../../../../../../types/script'
  import { useGlobalStore } from '../../../../../../../../store/global'
  import { getScriptList, getScriptVersionList } from '../../../../../../../../api/script'
  import ScriptEditor from '../../../../../../scripts/components/script-editor.vue'

  const { spaceId } = storeToRefs(useGlobalStore())

  const router = useRouter()

  const content = ref('')
  const scriptsLoading = ref(false)
  const scriptsData = ref<IScriptItem[]>([])
  const scriptVersionsLoading = ref(false)
  const scriptVersions = ref<IScriptVersion[]>([])

  onMounted(() => {
    getScripts()
  })

  // 获取脚本列表
  const getScripts = async () => {
    scriptsLoading.value = true
    const params = {
      start: 0,
      all: true
    }
    const res = await getScriptList(spaceId.value, params)
    scriptsData.value = res.detail
    scriptsLoading.value = false
  }

  // 获取脚本版本列表
  const getScriptVersions = async (id: number) => {
    scriptVersionsLoading.value = true
    const res = await getScriptVersionList(spaceId.value, id, { start: 0 })
    scriptVersions.value = res.detail
    scriptVersionsLoading.value = false
  }

  const goToScripts = () => {
    const { href } = router.resolve({ name: 'scripts-management', params: { spaceId: spaceId.value } })
    window.open(href, '__blank')
  }
</script>
<template>
  <bk-form form-type="vertical">
    <bk-form-item label="脚本引用" :required="true">
      <div class="script-cite">
        <bk-select class="script-selector">
          <bk-option
            v-for="option in scriptsData"
            :key="option.id"
            :value="option.id"
            :label="option.spec.name">
          </bk-option>
        </bk-select>
        <bk-select class="version-selector">
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
      <ScriptEditor v-model="content" />
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