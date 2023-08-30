<script lang="ts" setup>
  import { ref, watch, onMounted } from 'vue'
  import { useRoute } from 'vue-router'
  import { storeToRefs } from 'pinia'
  import { useServiceStore } from '../../../../../../../../store/service'
  import { IConfigVersion, IConfigListQueryParams, IConfigDiffDetail, IConfigDetail } from '../../../../../../../../../types/config'
  import { IFileConfigContentSummary } from '../../../../../../../../../types/config';
  import { getConfigList, getConfigItemDetail, getConfigContent } from '../../../../../../../../api/config'
  import { byteUnitConverse } from '../../../../../../../../utils'
  import MenuList from './menu-list.vue'

  const props = defineProps<{
    currentVersionId: number;
    baseVersionId: number;
    currentConfigId?: string|number;
  }>()

  const emits = defineEmits(['selected'])

  const route = useRoute()
  const bkBizId = ref(String(route.params.spaceId))
  const { appData } = storeToRefs(useServiceStore())

  const diffCount = ref(0)
  const selected = ref()
  const currentList = ref<IConfigDetail[]>([])
  const baseList = ref<IConfigDetail[]>([])
  // 汇总的配置项列表，包含未修改、增加、删除、修改的所有配置项
  const aggregatedList = ref<IConfigDiffDetail[]>([])

  // 基准版本变化，更新选中对比项
  watch(() => props.baseVersionId, async() => {
    let id = selected.value
    baseList.value = await getConfigsOfVersion(props.baseVersionId)
    calcDiff()
    if (typeof selected.value === 'number') {
      if (aggregatedList.value.length > 0 && !aggregatedList.value.find(item => item.id === id)) {
        id = aggregatedList.value[0].id
      }
      selectConfig(id)
    }
  })
  
  watch(() => props.currentConfigId, (val) => {
    selected.value = val
  }, {
    immediate: true
  })

  onMounted(async() => {
    await getAllConfigList()
    calcDiff()
    if (props.currentVersionId && aggregatedList.value.length > 0) {
      selectConfig(aggregatedList.value[0].id)
    }
  })

  // 获取某一版本下配置项列表
  // 不传releaseId表示当前版本
  const getConfigsOfVersion = async (releaseId?: number) => {
    const listQueryParams: IConfigListQueryParams = {
      start: 0,
      all: true
    }
    const listDetailQueryParams: { release_id?: number } = {}

    if (releaseId) {
      listQueryParams.release_id = releaseId
      listDetailQueryParams.release_id = releaseId
    }

    const res = await getConfigList(bkBizId.value, <number>appData.value.id, listQueryParams)

    // 遍历配置项列表，拿到每个配置项的signature
    return Promise.all(res.details.map((item: IConfigVersion) => getConfigItemDetail(bkBizId.value, item.id, <number>appData.value.id, listDetailQueryParams)))
  }

  // 获取当前版本和基准版本的所有配置项列表
  const getAllConfigList = async () => {
    if (props.currentVersionId) {
      currentList.value = await getConfigsOfVersion(props.currentVersionId)
    }
    if (props.baseVersionId) {
      baseList.value = await getConfigsOfVersion(props.baseVersionId)
    }
  }

  // 计算配置被修改、被删除、新增的差异
  const calcDiff = () => {
    diffCount.value = 0
    const list: IConfigDiffDetail[]= []
    currentList.value.forEach(currentItem => {
        const { config_item } = currentItem
        const baseItem = baseList.value.find(item => config_item.id === item.config_item.id)
        if (baseItem) {
            const diffConfig = {
                id: config_item.id,
                name: config_item.spec.name,
                type: '',
                current: currentItem.content.signature,
                base: baseItem.content.signature
            }
            if (currentItem.content.signature !== baseItem.content.signature) {
              diffCount.value++
              diffConfig.type = props.baseVersionId ? 'modify' : ''
            }
            list.push(diffConfig)
        } else { // 在基准版本中被删除
            diffCount.value++
            list.push({
                id: config_item.id,
                name: config_item.spec.name,
                type: props.baseVersionId ? 'add' : '',
                current: currentItem.content.signature,
                base: ''
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
                type: props.baseVersionId ? 'delete' : '',
                current: '',
                base: baseItem.content.signature,
            })
        }
    })
    aggregatedList.value = list
  }

  // 选择对比配置项
  const selectConfig = async (id: number) => {
    const config = aggregatedList.value.find(item => item.id === id)
    if (config) {
      selected.value = config.id
      const data = await getConfigDiffDetail(config)
      emits('selected', selected.value, data)
    }
  }

  const getConfigDiffDetail = async (config: IConfigDiffDetail) => {
    const currentConfig = currentList.value.find(item => item.config_item.id === config.id)
    const baseConfig = baseList.value.find(item => item.config_item.id === config.id)
    const contentType = (currentConfig?.config_item.spec.file_type || baseConfig?.config_item.spec.file_type) === 'binary' ? 'file' : 'text'
    let currentConfigContent: string|IFileConfigContentSummary = ''
    let baseConfigContent: string|IFileConfigContentSummary = ''

    if (currentConfig) {
      currentConfigContent = await loadConfigContent(currentConfig)
    }

    if (baseConfig) {
      baseConfigContent = await loadConfigContent(baseConfig)
    }

    return {
      contentType,
      base: {
        content: baseConfigContent
      },
      current: {
        content: currentConfigContent
      }
    }
  }
  // 加载配置内容详情
  const loadConfigContent = async(config: IConfigDetail) => {
    if (!config.content.signature) {
      return ''
    }
    if (config.config_item.spec.file_type === 'binary') {
      return {
        id: config.config_item.id,
        name: config.config_item.spec.name,
        signature: config.content.signature,
        update_at: config.config_item.revision?.update_at,
        size: byteUnitConverse(Number(config.content.byte_size))
      }
    }
    const configContent = await getConfigContent(bkBizId.value, <number>appData.value.id, config.content.signature)
    return String(configContent)
  }

</script>
<template>
  <div class="configs-menu">
    <MenuList title="配置项" :value="selected" :list="aggregatedList" @selected="selectConfig" />
  </div>
</template>
<style lang="scss" scoped></style>
