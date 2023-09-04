<script lang="ts" setup>
  import { ref, computed, watch } from 'vue'
  import { storeToRefs } from 'pinia'
  import { Message } from 'bkui-vue'
  import { useGlobalStore } from '../../../../../../../../store/global'
  import { useTemplateStore } from '../../../../../../../../store/template'
  import { ITemplatePackageItem, ITemplateConfigItem } from '../../../../../../../../../types/template';
  import useModalCloseConfirmation from '../../../../../../../../utils/hooks/use-modal-close-confirmation'
  import { PACKAGE_MENU_OTHER_TYPE_MAP } from '../../../../../../../../constants/template'
  import { addTemplateToPackage, getTemplatesBySpaceId, getTemplatePackageList } from '../../../../../../../../api/template'
  import PackageTable from './package-table.vue'
  import SearchInput from '../../../../../../../../components/search-input.vue'
import search from 'bkui-vue/lib/icon/search'

  interface IPackageTableGroup {
    id: number|string;
    name: string;
    configs: ITemplateConfigItem[];
  }

  const { spaceId } = storeToRefs(useGlobalStore())
  const { currentTemplateSpace, currentPkg } = storeToRefs(useTemplateStore())

  const props = defineProps<{
    show: boolean;
  }>()

  const emits = defineEmits(['update:show', 'added'])

  const isShow = ref(false)
  const isFormChange = ref(false)
  const loading = ref(false)
  const packageGroups = ref<IPackageTableGroup[]>([]) // 所有套餐配置项数据
  const packageGroupsOnShow = ref<IPackageTableGroup[]>([]) // 实际展示的数据，处理搜索的场景
  const pending = ref(false)
  const searchStr = ref('')
  const openedPkgTable = ref<number|string>('')
  const selectedConfigs = ref<{ id: number; name: string; }[]>([])

  watch(() => props.show, val => {
    isShow.value = val
    if (val) {
      openedPkgTable.value = ''
      selectedConfigs.value = []
      isFormChange.value = false
      getGroupConfigs()
    }
  })

  // 加载全部配置项
  const getGroupConfigs = async () => {
    loading.value = true
    const params = {
      start: 0,
      all: true
    }
    const [packagesRes, configsRes] = await Promise.all([
      getTemplatePackageList(spaceId.value, currentTemplateSpace.value, params),
      getTemplatesBySpaceId(spaceId.value, currentTemplateSpace.value, params)
    ])
    // 第一个分组默认为“全部配置项”
    const packages: IPackageTableGroup[] = [{
      id: 0,
      name: '全部配置项',
      configs: configsRes.details
    }]
    packagesRes.details.filter((pkg: ITemplatePackageItem) => pkg.id !== currentPkg.value).forEach((pkg: ITemplatePackageItem) => {
      const { name, template_ids } = pkg.spec
      const pkgGroup: IPackageTableGroup = {
        id: pkg.id,
        name,
        configs: []
      }
      template_ids.forEach(id => {
        const config = configsRes.details.find((item: ITemplateConfigItem) => item.id === id)
        if (config) {
          pkgGroup.configs.push(config)
        }
      })
      packages.push(pkgGroup)
    })
    packageGroups.value = packages.slice()
    packageGroupsOnShow.value = packages.slice()
    loading.value = false
  }

  const handleSearch = () => {
    if (searchStr.value) {
      const list: IPackageTableGroup[] = []
      packageGroups.value.forEach(pkg => {
        const matchedConfigs = pkg.configs.filter(config => {
          const { name, path, memo } = config.spec
          const lowerSearchStr = searchStr.value.toLocaleLowerCase()
          return name.toLocaleLowerCase().includes(lowerSearchStr)
            || path.toLocaleLowerCase().includes(lowerSearchStr)
            || memo.toLocaleLowerCase().includes(lowerSearchStr)
        })
        if (matchedConfigs.length > 0) {
          const { id, name } = pkg
          list.push({ id, name, configs: matchedConfigs })
        }
      })
      packageGroupsOnShow.value = list
    } else {
      packageGroupsOnShow.value = packageGroups.value.slice()
    }
  }

  const handleToggleOpenTable = (id: string|number) => {
    openedPkgTable.value = openedPkgTable.value === id ? '' : id
  }

  const handleDeleteConfig = (id: number) => {
    const index = selectedConfigs.value.findIndex(item => item.id === id)
    if (index > -1) {
      selectedConfigs.value.splice(index, 1)
    }
  }

  const handleAddConfigs = async() => {
    try {
      pending.value = true
      const configIds = selectedConfigs.value.map(item => item.id)
      await addTemplateToPackage(spaceId.value, currentTemplateSpace.value, configIds, [<number>currentPkg.value])
      emits('added')
      close()
      Message({
        theme: 'success',
        message: '添加配置项成功'
      })
    } catch (e) {
      console.log(e)
    } finally {
      pending.value = false
    }
  }

  const handleBeforeClose = async() => {
    if (isFormChange.value) {
      const result = await useModalCloseConfirmation()
      return result
    }
    return true
  }

  const close = () => {
    emits('update:show', false)
  }

</script>
<template>
  <bk-sideslider
    title="从已有配置项添加"
    :width="640"
    :is-show="isShow"
    :before-close="handleBeforeClose"
    @closed="close">
    <div v-bkloading="{ loading }" class="slider-content-container">
      <div class="package-configs-pick">
        <div class="search-wrapper">
          <SearchInput v-model="searchStr" placeholder="配置项名称/路径/描述" @search="handleSearch" />
        </div>
        <div class="package-tables">
          <PackageTable
            v-for="pkg in packageGroupsOnShow"
            v-model:selected-configs="selectedConfigs"
            :key="pkg.id"
            :pkg="pkg"
            :open="openedPkgTable === pkg.id"
            :config-list="pkg.configs"
            @change="isFormChange = true"
            @toggleOpen="handleToggleOpenTable" />
        </div>
      </div>
      <div class="selected-panel">
        <h5 class="title-text">已选 <span class="num">{{ selectedConfigs.length }}</span> 个配置项</h5>
        <div class="selected-list">
          <div v-for="config in selectedConfigs" class="config-item" :key="config.id">
            <div class="name" :title="config.name">{{ config.name }}</div>
            <i class="bk-bscp-icon icon-reduce delete-icon" @click="handleDeleteConfig(config.id)" />
          </div>
          <p v-if="selectedConfigs.length === 0" class="empty-tips">请先从左侧选择配置项</p>
        </div>
      </div>
    </div>
    <div class="action-btns">
      <bk-button
        theme="primary"
        :loading="pending"
        :disabled="loading || selectedConfigs.length === 0"
        @click="handleAddConfigs">
        添加
      </bk-button>
      <bk-button @click="close">取消</bk-button>
    </div>
  </bk-sideslider>
</template>
<style lang="scss" scoped>
  .slider-content-container {
    display: flex;
    align-items: flex-start;
    height: calc(100vh - 101px);
    overflow: auto;
  }
  .search-wrapper {
    padding: 0 16px 0 24px;
    .search-input-icon {
      padding-right: 10px;
      color: #979ba5;
      background: #ffffff;
      font-size: 16px;
    }
  }
  .package-configs-pick {
    padding: 20px 0;
    width: 440px;
    height: 100%;
    .package-tables {
      padding: 16px 16px 0 24px;
      height: calc(100% - 32px);
      overflow: auto;
      .package-config-table:not(:last-of-type) {
        margin-bottom: 16px;
      }
    }
  }
  .selected-panel {
    padding: 20px 24px 20px 16px;
    width: 200px;
    height: 100%;
    background: #f5f7fa;
    .title-text {
      margin: 0;
      line-height: 16px;
      font-size: 12px;
      font-weight: normal;
      color: #63656e;
      .num {
        color: #3a84ff;
        font-weight: 700;
      }
    }
    .selected-list {
      padding-top: 16px;
      height: calc(100% - 16px);
      overflow: auto;
      .config-item {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 0 9px 0 12px;
        height: 32px;
        font-size: 12px;
        color: #63656e;
        background: #ffffff;
        border-radius: 2px;
        &:not(:last-of-type) {
          margin-bottom: 4px;
        }
        .name {
          text-overflow: ellipsis;
          overflow: hidden;
          white-space: nowrap;
        }
        .delete-icon {
          margin-left: 4px;
          font-size: 12px;
          cursor: pointer;
          &:hover {
            color: #3a84ff;
          }
        }
      }
      .empty-tips {
        margin: 56px 0 0;
        font-size: 12px;
        color: #979ba5;
        text-align: center;
      }
    }
  }
  .action-btns {
    border-top: 1px solid #dcdee5;
    padding: 8px 24px;
    .bk-button {
      margin-right: 8px;
      min-width: 88px;
    }
  }
</style>
