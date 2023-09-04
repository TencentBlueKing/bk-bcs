<script lang="ts" setup>
  import { ref, computed, watch } from 'vue'
  import { storeToRefs } from 'pinia'
  import { RightShape } from 'bkui-vue/lib/icon';
  import { useGlobalStore } from '../../../../../../../../store/global'
  import { useTemplateStore } from '../../../../../../../../store/template'
  import { ITemplateConfigItem } from '../../../../../../../../../types/template';
  import { getTemplatesByPackageId, getTemplatesBySpaceId } from '../../../../../../../../api/template';

  const { spaceId } = storeToRefs(useGlobalStore())
  const { currentTemplateSpace } = storeToRefs(useTemplateStore())

  const props = defineProps<{
    pkg: { id: number|string; name: string; };
    open: boolean;
    selectedConfigs: { id: number; name: string; }[]
  }>()

  const emits = defineEmits(['toggleOpen', 'update:selectedConfigs'])

  const loading = ref(false)
  const configList = ref<ITemplateConfigItem[]>([])
  const page = ref(1)

  const isAllSelected = computed(() => {
    return configList.value.length > 0 && configList.value.every(item => isConfigSelected(item.id))
  })

  const isIndeterminate = computed(() => {
    return configList.value.length > 0 && props.selectedConfigs.length > 0 && !isAllSelected.value
  })

  watch(() => props.open, val => {
    if (val) {
      page.value = 1
      getConfigList()
    }
  })

  const getConfigList = async () => {
    loading.value = true
    let res
    const params = {
      start: (page.value - 1) * 10,
      all: true
    }
    if (typeof props.pkg.id === 'number') {
      res = await getTemplatesByPackageId(spaceId.value, currentTemplateSpace.value, props.pkg.id, params)
    } else {
      res = await getTemplatesBySpaceId(spaceId.value, currentTemplateSpace.value, params)
    }
    configList.value = res.details
    loading.value = false
  }

  const isConfigSelected = (id: number) => {
    return props.selectedConfigs.findIndex(item => item.id === id) > -1
  }

  const handleAllSelectionChange = (checked: boolean) => {
    const configs = props.selectedConfigs.slice()
    if (checked) {
      configList.value.forEach(config => {
        if(!configs.find(item => item.id === config.id)) {
          const { id, spec } = config
          configs.push({ id, name: spec.name })
        }
      })
    } else {
      configList.value.forEach(config => {
        const index = configs.findIndex(item => item.id === config.id)
        if(index > -1) {
          configs.splice(index, 1)
        }
      })
    }
    emits('update:selectedConfigs', configs)
  }

  const handleConfigSelectionChange = (checked: boolean, config: ITemplateConfigItem) => {
    const configs = props.selectedConfigs.slice()
    if (checked) {
      if(!configs.find(item => item.id === config.id)) {
        const { id, spec } = config
        configs.push({ id, name: spec.name })
      }
    } else {
      const index = configs.findIndex(item => item.id === config.id)
      if(index > -1) {
        configs.splice(index, 1)
      }
    }
    emits('update:selectedConfigs', configs)
  }

</script>
<template>
  <div :class="['package-config-table', {'table-open': props.open }]">
    <div class="head-area" @click="emits('toggleOpen', props.pkg.id)">
      <RightShape class="triangle-icon" />
      <div class="title">{{ props.pkg.name }}</div>
    </div>
    <div v-show="props.open" v-bkloading="{ loading }" class="config-table-wrapper">
      <table class="config-table">
        <thead>
          <tr>
            <th class="th-cell name">
              <div class="name-info">
                <bk-checkbox
                :model-value="isAllSelected"
                :indeterminate="isIndeterminate"
                @change="handleAllSelectionChange" />
                <div class="name-text">配置项名称</div>
              </div>
            </th>
            <th class="th-cell path">配置项路径</th>
            <th class="th-cell memo">配置项描述</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="config in configList" :key="config.id">
            <td class="td-cell name">
              <div class="cell name-info">
                <bk-checkbox
                  :model-value="isConfigSelected(config.id)"
                  @change="handleConfigSelectionChange($event, config)" />
                <div class="name-text">{{ config.spec.name }}</div>
              </div>
            </td>
            <td class="td-cell name">
              <div class="cell">
                {{ config.spec.path }}
              </div>
            </td>
            <td class="td-cell name">
              <div class="cell">
                {{ config.spec.memo || '--' }}
              </div>
            </td>
          </tr>
          <tr v-if="configList.length === 0">
            <td  class="td-cell" :colspan="3">
              <bk-exception class="empty-tips" type="empty" scene="part">暂无配置项</bk-exception>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
<style lang="scss" scoped>
  .package-config-table.table-open {
    .triangle-icon {
      transform: rotate(90deg);
    }
  }
  .head-area {
    display: flex;
    align-items: center;
    padding: 0 8px;
    height: 28px;
    background: #eaebf0;
    cursor: pointer;
    .triangle-icon {
      margin-right: 8px;
      font-size: 12px;
      color: #979ba5;
      transition: transform .3s cubic-bezier(.4,0,.2,1);;
    }
    .title {
      font-size: 12px;
      font-weight: 700;
      color: #63656e;
    }
  }
  .config-table-wrapper {
    position: relative;
    max-height: 60vh;
    overflow: auto;
  }
  .config-table {
    width: 100%;
    border: 1px solid #dcdee5;
    border-top: none;
    table-layout: fixed;
    border-collapse: collapse;
    thead {
      position: sticky;
      top: 0;
      z-index: 1;
    }
    .th-cell {
      padding: 0 16px;
      height: 42px;
      color: #313238;
      font-size: 12px;
      font-weight: normal;
      text-align: left;
      background: #fafbfd;
      border-bottom: 1px solid #dcdee5;
    }
    .td-cell {
      padding: 0 16px;
      text-align: left;
      border-bottom: 1px solid #dcdee5;
    }
    .cell {
      height: 42px;
      line-height: 42px;
      color: #63656e;
      font-size: 12px;
      white-space: nowrap;
      text-overflow: ellipsis;
      overflow: hidden;
    }
    .name-info {
      display: flex;
      align-items: center;
      height: 100%;
      .name-text {
        margin-left: 8px;
        white-space: nowrap;
        text-overflow: ellipsis;
        overflow: hidden;
      }
    }
    .empty-tips {
      margin-bottom: 20px;
    }
  }
</style>
