<script setup lang="ts">
  import { ref, computed, onMounted } from 'vue'
  import { storeToRefs } from 'pinia'
  import { useGlobalStore } from '../../../../../../../../store/global'
  import { IConfigInitScript } from '../../../../../../../../../types/script'
  import { getConfigInitScript } from '../../../../../../../../api/script'
  import ScriptConfigForm from './script-config-form.vue'

  const { spaceId } = storeToRefs(useGlobalStore())

  const props = defineProps<{
    appId: number;
  }>()

  const show = ref(false)
  const scriptData = ref<IConfigInitScript>({
    pre_hook_id: 0,
  	pre_hook_release_id: '',
  	post_hook_id: 0,
  	post_hook_release_id: ''
  })
  const loading = ref(false)
  const currentType = ref('post')

  const TYPES = [
    { id: 'pre', name: '前置脚本' },
    { id: 'post', name: '后置脚本' }
  ]

  const tips = computed(() => {
    return currentType.value === 'post' ? '后置脚本：下载配置文件后执行脚本，一般用于二进制重新 reload 配置文件。' : '前置脚本：下载配置文件前执行脚本，一般用于文件备份等前置准备工作。'
  })

  const scriptId = computed(() => {
    return currentType.value === 'post' ? scriptData.value.post_hook_id : scriptData.value.pre_hook_id
  })

  const versionId = computed(() => {
    return currentType.value === 'post' ? scriptData.value.post_hook_release_id : scriptData.value.pre_hook_release_id
  })

  onMounted(() => {
    // getData()
  })

  const getData = async () => {
    loading.value = true
    const res = await getConfigInitScript(spaceId.value, props.appId)
    console.log(res)
    loading.value = false
  }

  const handleOpenSlider = () => {
    show.value = true
  }

  const handleFormChange = (data: { scriptId: number; versionId: number }) => {
    const { scriptId, versionId } = data
    if (currentType.value === 'post') {
      scriptData.value.post_hook_id = scriptId
      scriptData.value.post_hook_release_id = versionId
    } else {
      scriptData.value.pre_hook_id = scriptId
      scriptData.value.pre_hook_release_id = versionId
    }
  }
  
  const handleClose = () => {
    show.value = false
  }
</script>
<template>
  <bk-button @click="handleOpenSlider">初始化脚本</bk-button>
  <bk-sideslider
    title="初始化脚本"
    width="1200"
    :is-show="show"
    @closed="handleClose">
    <div class="header-wrapper">
      <div class="script-types">
        <span
          v-for="item in TYPES"
          :key="item.id"
          :class="['type-text', { actived: item.id === currentType }]"
          @click="currentType = item.id">
          {{ item.name }}
        </span>
      </div>
    </div>
    <div class="script-config">
      <bk-alert theme="info" :title="tips"></bk-alert>
      <ScriptConfigForm :key="currentType" :scriptId="scriptId" :versionId="versionId" @change="handleFormChange" />
    </div>
    <div class="actions-btns">
      <bk-button theme="primary">保存</bk-button>
      <bk-button>取消</bk-button>
    </div>
  </bk-sideslider>
</template>
<style lang="scss" scoped>
  .header-wrapper {
    .script-types {
      display: flex;
      align-items: center;
      padding: 6px 24px 0;
      background: #f0f1f5;
    }
    .type-text {
      padding: 10px 24px;
      line-height: 22px;
      color: #63656e;
      font-size: 14px;
      cursor: pointer;
      &.actived {
        background: #ffffff;
        color: #3a84ff;
      }
      &:hover {
        color: #3a84ff;
      }
    }
  }
  .script-config {
    padding: 8px 24px 24px;
    height: calc(100vh - 149px);
    overflow: auto;
    .bk-alert {
      margin: 8px 0 20px;
    }
  }
  .actions-btns {
    border-top: 1px solid #dcdee5;
    padding: 8px 24px;
    .bk-button {
      margin-right: 8px;
      min-width: 88px;
    }
  }
</style>