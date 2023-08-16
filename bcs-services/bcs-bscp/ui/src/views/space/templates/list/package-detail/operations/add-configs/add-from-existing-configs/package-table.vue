<script lang="ts" setup>
  import { ref, watch } from 'vue'
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
      limit: 10
    }
    if (typeof props.pkg.id === 'number') {
      res = await getTemplatesByPackageId(spaceId.value, currentTemplateSpace.value, props.pkg.id, params)
    } else {
      res = await getTemplatesBySpaceId(spaceId.value, currentTemplateSpace.value, params)
    }
    configList.value = res.details
  }

  const handleSelectionChange = ({ checked, isAll, row}: { checked: boolean; isAll: boolean; row: ITemplateConfigItem }) => {
    const configs = props.selectedConfigs.slice()
    if (isAll) {
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
    } else {
      if (checked) {
        if(!configs.find(item => item.id === row.id)) {
          const { id, spec } = row
          configs.push({ id, name: spec.name })
        }
      } else {
        const index = configs.findIndex(item => item.id === row.id)
        if(index > -1) {
          configs.splice(index, 1)
        }
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
    <bk-table v-show="props.open" :data="configList" @selection-change="handleSelectionChange">
      <bk-table-column type="selection" :min-width="20" :width="30" />
      <bk-table-column label="配置项名称" prop="spec.name"></bk-table-column>
      <bk-table-column label="配置项路径" prop="spec.path"></bk-table-column>
      <bk-table-column label="配置项描述" prop="spec.memo"></bk-table-column>
    </bk-table>
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
</style>
