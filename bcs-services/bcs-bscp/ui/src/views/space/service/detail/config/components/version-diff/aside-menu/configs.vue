<script lang="ts" setup>
  import { ref, computed, watch, onMounted } from 'vue'
  import { useRoute } from 'vue-router'
  import { storeToRefs } from 'pinia'
  import { Search, RightShape } from 'bkui-vue/lib/icon'
  import { useServiceStore } from '../../../../../../../../store/service'
  import { ICommonQuery } from '../../../../../../../../../types/index'
  import { IConfigItem, IConfigListQueryParams, IBoundTemplateDetail } from '../../../../../../../../../types/config'
  import { IFileConfigContentSummary } from '../../../../../../../../../types/config';
  import { getConfigList, getConfigItemDetail, getConfigContent, getBoundTemplates, getBoundTemplatesByAppVersion } from '../../../../../../../../api/config'
  import { byteUnitConverse } from '../../../../../../../../utils'
  import SearchInput from '../../../../../../../../components/search-input.vue'

  interface IConfigMenuItem {
    type: string;
    id: number;
    name: string;
    file_type: string;
    file_state: string;
    update_at: string;
    byte_size: string;
    signature: string;
    template_space_id: number;
    template_space_name: string;
    template_set_id: number;
    template_set_name: string;
    template_revision_id: number;
  }

  interface IConfigDiffItem extends IConfigMenuItem {
    diff_type: string;
    current: string;
    base: string;
  }

  interface IConfigsGroupData {
    id: number;
    name: string;
    expand: boolean;
    configs: IConfigDiffItem[];
  }

  const props = defineProps<{
    currentVersionId: number;
    baseVersionId: number|undefined;
    currentConfigId?: string|number;
  }>()

  const emits = defineEmits(['selected'])

  const route = useRoute()
  const bkBizId = ref(String(route.params.spaceId))
  const { appData } = storeToRefs(useServiceStore())

  const diffCount = ref(0)
  const selected = ref<IConfigDiffItem>()
  const currentList = ref<IConfigMenuItem[]>([])
  const baseList = ref<IConfigMenuItem[]>([])
  // 汇总的配置项列表，包含未修改、增加、删除、修改的所有配置项
  const aggregatedList = ref<IConfigsGroupData[]>([])
  const searchStr = ref('')

  // 是否实际选择了对比的基准版本，为了区分的未命名版本id为0的情况
  const isBaseVersionExist = computed(() => {
    return typeof props.baseVersionId === 'number'
  })

  // 基准版本变化，更新选中对比项
  watch(() => props.baseVersionId, async() => {
    let id = selected.value
    baseList.value = await getConfigsOfVersion(props.baseVersionId)
    aggregatedList.value = calcDiff()
    // if (typeof selected.value === 'number') {
    //   if (aggregatedList.value.length > 0 && !aggregatedList.value.find(item => item.id === id)) {
    //     id = aggregatedList.value[0].id
    //   }
    //   selectConfig(id)
    // }
  })

  // 当前版本默认选中的配置项
  watch(() => props.currentConfigId, (val) => {
    if (val) {
      // selected.value = val
    }
  }, {
    immediate: true
  })

  onMounted(async() => {
    await getAllConfigList()
    aggregatedList.value = calcDiff()
    // if (aggregatedList.value.length > 0) {
    //   selectConfig(aggregatedList.value[0].id)
    // }
  })

  // 判断版本是否为未命名版本
  const isUnNamedVersion = (id: number) => {
    return id === 0
  }

  // 获取某一版本下配置项和模板列表
  const getConfigsOfVersion = async (releaseId: number|undefined) => {
    if (typeof releaseId !== 'number') {
      return []
    }

    const [commonConfigList, templateList] = await Promise.all([
      getCommonConfigList(releaseId),
      getBoundTemplateList(releaseId)
    ])

    return commonConfigList.concat(templateList)
  }

  // 获取非模板配置项列表
  const getCommonConfigList = async(id: number) => {
    const params: IConfigListQueryParams = {
      start: 0,
      all: true
    }
    const configsDetailQueryParams: { release_id?: number } = {}

    if (!isUnNamedVersion(id)) {
      params.release_id = id
      configsDetailQueryParams.release_id = id
    }

    const configsRes = await getConfigList(bkBizId.value, <number>appData.value.id, params)
    // 未命名版本中包含被删除的配置项，需要过滤掉
    const configs: IConfigItem[] = configsRes.details.filter((item: IConfigItem) => item.file_state !== 'DELETE')

    // 遍历配置项列表，拿到每个配置项的signature
    const configsDetailRes =  await Promise.all(configs.map(item => getConfigItemDetail(bkBizId.value, item.id, <number>appData.value.id, configsDetailQueryParams)))

    return configs.map((config, index) => {
      const { id, spec, revision, file_state } = config
      const { name, file_type } = spec
      const { byte_size, signature } = configsDetailRes[index].content
      return {
        type: 'config',
        id,
        name,
        file_type,
        file_state,
        update_at: revision.update_at,
        byte_size,
        signature,
        template_revision_id: 0,
        template_space_id: 0,
        template_space_name: '',
        template_set_id: 0,
        template_set_name: ''
      }
    })
  }

  // 获取模板配置项列表
  const getBoundTemplateList = async(id: number) => {
      const params: ICommonQuery = {
        start: 0,
        all: true
      }
      let res
      if (isUnNamedVersion(id)) {
        res = await getBoundTemplates(bkBizId.value, <number>appData.value.id, params)
      } else {
        res = await getBoundTemplatesByAppVersion(bkBizId.value, <number>appData.value.id, id)
      }
      return res.details.filter((template: IBoundTemplateDetail) => template.file_state !== 'DELETE').map((template: IBoundTemplateDetail) => {
        const {
            template_id, name, file_type, file_state, byte_size, signature, template_revision_id,
            template_space_id, template_space_name, template_set_id, template_set_name
        } = template
        return {
          type: 'template',
          id: template_id,
          name,
          file_type,
          file_state,
          update_at: '',
          byte_size,
          signature,
          template_revision_id: template_revision_id,
          template_space_id,
          template_space_name,
          template_set_id,
          template_set_name
        }
      })
  }

  // 获取当前版本和基准版本的所有配置项列表
  const getAllConfigList = async () => {
    currentList.value = await getConfigsOfVersion(props.currentVersionId)
    baseList.value = await getConfigsOfVersion(props.baseVersionId)
  }

  // 计算配置被修改、被删除、新增的差异
  const calcDiff = () => {
    diffCount.value = 0
    const list: IConfigDiffItem[]= []
    currentList.value.forEach(currentItem => {
      const baseItem = baseList.value.find(config => {
        return config.id === currentItem.id && config.template_revision_id === currentItem.template_revision_id
      })
      if (baseItem) {
          const diffConfig = {
              ...currentItem,
              diff_type: '',
              current: currentItem.signature,
              base: baseItem.signature
          }
          if (diffConfig.current !== diffConfig.base) {
            diffCount.value++
            diffConfig.diff_type = isBaseVersionExist.value ? 'modify' : ''
          }
          list.push(diffConfig)
      } else { // 在基准版本中被删除
          diffCount.value++
          list.push({
              ...currentItem,
              diff_type: isBaseVersionExist.value ? 'add' : '',
              current: currentItem.signature,
              base: ''
          })
      }
    })
    // 基准版本中的新增项
    baseList.value.forEach(baseItem => {
        const currentItem = currentList.value.find(config => {
          return config.id === baseItem.id && config.template_revision_id === baseItem.template_revision_id
        })
        if (!currentItem) {
            diffCount.value++
            list.push({
                ...baseItem,
                diff_type: isBaseVersionExist.value ? 'delete' : '',
                current: '',
                base: baseItem.signature
            })
        }
    })
    return groupTplsByPkg(list)
  }

  // 将模板按套餐分组
  const groupTplsByPkg = (configs: IConfigDiffItem[]) => {
    const groups: IConfigsGroupData[] = []
    configs.forEach(config => {
      const { template_space_name, template_set_id, template_set_name } = config
      const group = groups.find(item => item.id === template_set_id)
      if (group) {
        group.configs.push(config)
      } else {
        groups.push({
          id: template_set_id,
          name: template_set_id === 0 ? '非配置项分组' : `${template_space_name} - ${template_set_name}`,
          expand: template_set_id === 0,
          configs: [config]
        })
      }
    })
    return groups
  }

  // 选择对比配置项后，加载配置项详情，组装对比数据
  const handleSelectItem = async (id: number, versionId: number) => {
    let config: IConfigDiffItem | undefined
    aggregatedList.value.some(group => {
      group.configs.some(configItem => {
        if (configItem.id === id && configItem.template_revision_id === versionId) {
          config = configItem
          return true
        }
      })
    })
    if (config) {
      selected.value = config
      const data = await getConfigDiffDetail(config)
      emits('selected', selected.value, data)
    }
  }

  const getConfigDiffDetail = async (config: IConfigDiffItem) => {
    const currentConfig = currentList.value.find(item => item.id === config.id && item.template_revision_id === config.template_revision_id)
    const baseConfig = baseList.value.find(item => item.id === config.id && item.template_revision_id === config.template_revision_id)
    const contentType = (currentConfig?.file_type || baseConfig?.file_type) === 'binary' ? 'file' : 'text'
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
  const loadConfigContent = async(config: IConfigMenuItem) => {
    if (!config.signature) {
      return ''
    }
    if (config.file_type === 'binary') {
      return {
        id: config.id,
        name: config.name,
        signature: config.signature,
        update_at: config.update_at,
        size: byteUnitConverse(Number(config.byte_size))
      }
    }
    const configContent = await getConfigContent(bkBizId.value, <number>appData.value.id, config.signature)
    return String(configContent)
  }

</script>
<template>
  <div class="configs-menu">
    <div class="title-area">
      <div class="title">配置项</div>
      <div class="title-extend">
        <bk-checkbox class="view-diff-checkbox">只查看差异项({{ diffCount }})</bk-checkbox>
        <div class="search-trigger">
          <Search />
        </div>
      </div>
    </div>
    <div class="search-wrapper">
      <SearchInput v-model="searchStr" />
    </div>
    <div class="groups-wrapper">
      <div v-for="group in aggregatedList" class="config-group-item" :key="group.id">
        <div :class="['group-header', { expand: group.expand }]" @click="group.expand = !group.expand">
          <RightShape class="arrow-icon" />
          <span v-overflow-title class="name">{{ group.name }}</span>
        </div>
        <div v-if="group.expand" class="config-list">
          <div
            v-for="config in group.configs"
            :key="config.id"
            :class="['config-item', { actived: config.id === selected?.id && config.template_revision_id === selected.template_revision_id }]"
            @click="handleSelectItem(config.id, config.template_revision_id)">
            <i v-if="config.diff_type" :class="['status-icon', config.diff_type]"></i>
            {{ config.name }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
<style lang="scss" scoped>
  .configs-menu {
    background: #fafbfd;
  }
  .title-area {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 12px 8px 24px;
    .title {
      font-size: 14px;
      color: #313238;
      font-weight: 700;
    }
    .title-extend {
      display: flex;
      align-items: center;
      .view-diff-checkbox {
        padding-right: 8px;
        border-right: 1px solid #dcdee5;
        :deep(.bk-checkbox-label) {
          font-size: 12px;
        }
      }
    }
    .search-trigger {
      display: flex;
      align-items: center;
      justify-content: center;
      margin-left: 8px;
      width: 20px;
      height: 20px;
      font-size: 12px;
      color: #63656e;
      background: #edeff1;
      border-radius: 2px;
      cursor: pointer;
      &:hover {
        background: #e1ecff;
        color: #3a84ff;
      }
    }
  }
  .search-wrapper {
    padding: 0 12px 8px;
  }
  .groups-wrapper {

  }
  .config-group-item {
    .group-header {
      display: flex;
      align-items: center;
      padding: 8px 12px;
      line-height: 20px;
      font-size: 12px;
      color: #313238;
      cursor: pointer;
      &.expand {
        .arrow-icon {
          transform: rotate(90deg);
          color: #3a84ff;
        }
      }
    }
    .arrow-icon {
      margin-right: 8px;
      font-size: 14px;
      color: #c4c6cc;
      transition: transform .2s ease-in-out;
    }
    .config-list {
      margin-bottom: 8px;
      .config-item {
        position: relative;
        padding: 0 12px 0 24px;
        height: 40px;
        line-height: 40px;
        font-size: 12px;
        color: #63656e;
        border-bottom: 1px solid #dcdee5;
        cursor: pointer;
        &:hover {
          background: #e1ecff;
        }
        &.actived {
          background: #e1ecff;
          color: #3a84ff;
        }
        .status-icon {
          position: absolute;
          top: 18px;
          left: 10px;
          width: 4px;
          height: 4px;
          border-radius: 50%;
          &.add {
            background: #3a84ff;
          }
          &.delete {
            background: #ea3536;
          }
          &.modify {
            background: #fe9c00;
          }
        }
      }
    }
  }
</style>
