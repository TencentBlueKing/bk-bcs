<script lang="ts" setup>
  import { ref, onMounted, computed, watch } from 'vue'
  import { useRoute } from 'vue-router'
  import { IVersionHook } from '../../../../../../../../../types/config'
  import { getScriptVersionDetail } from '../../../../../../../../api/script'
  import { getDiffType } from '../../../../../../../../utils/index'
  import MenuList from './menu-list.vue'

  const route = useRoute()
  const bkBizId = String(route.params.spaceId)

  const props = defineProps<{
    baseVersionId: number;
    currentHook: IVersionHook;
    baseHook: IVersionHook;
    value: string|number|undefined;
  }>()

  const emits = defineEmits(['selected'])

  const scriptDetailList = ref([
    {
      id: 'pre',
      name: '前置脚本',
      type: '',
      current: {
        language: '',
        content: ''
      },
      base: {
        language: '',
        content: ''
      }
    },
    {
      id: 'post',
      name: '后置脚本',
      type: '',
      current: {
        language: '',
        content: ''
      },
      base: {
        language: '',
        content: ''
      }
    }
  ])
  const selected = ref()

  watch(() => props.baseHook, async(val) => {
    await updateDiff(val, 'base')
    if (typeof selected === 'string') {
      selectScript(selected)
    }
  })

  watch(() => props.value, (val) => {
    selected.value = val
  }, {
    immediate: true
  })

  onMounted(() => {
    updateDiff(props.currentHook, 'current')
  })

  // 获取脚本内容
  const getScriptDetail = (id: number, versionId: number) => {
    if (id) {
      return getScriptVersionDetail(bkBizId, id, versionId).then(res => {
        const { content, type } = res.spec
        return { content, language: type }
      })
    }
    return { language: '', content: '' }
  }
  

  // 计算前置脚本或后置脚本差异
  const updateDiff = async(hook: IVersionHook, type: 'current'|'base') => {
    const { pre_hook_id, pre_hook_release_id, post_hook_id, post_hook_release_id } = hook
    const [preHook, postHook] = await Promise.all([
      getScriptDetail(pre_hook_id, pre_hook_release_id),
      getScriptDetail(post_hook_id, post_hook_release_id)
    ])
    scriptDetailList.value[0][type] = preHook
    scriptDetailList.value[1][type] = postHook
    // 选择基准版本后才计算变更状态
    if (props.baseVersionId) {
      scriptDetailList.value[0].type = getDiffType(scriptDetailList.value[0].base.content, scriptDetailList.value[0].current.content)
      scriptDetailList.value[1].type = getDiffType(scriptDetailList.value[1].base.content, scriptDetailList.value[1].current.content)
    }
  }

  const selectScript = (id: string) => {
    const script = id === 'pre' ? scriptDetailList.value[0] : scriptDetailList.value[1]
    const { base, current } = script
    const diffData = { contentType: 'text', base, current }
    selected.value = id
    emits('selected', selected.value, diffData)
  }

</script>
<template>
  <div class="scripts-menu">
    <MenuList title="初始化脚本" :value="selected" :list="scriptDetailList" @selected="selectScript" />
  </div>
</template>
<style lang="scss" scoped></style>
