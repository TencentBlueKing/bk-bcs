<script lang="ts" setup>
  import { ref, computed, watch } from 'vue'
  import { storeToRefs } from 'pinia'
  import { useGlobalStore } from '../../../../../../../../store/global'
  import { useTemplateStore } from '../../../../../../../../store/template'
  import { Search } from 'bkui-vue/lib/icon'
  import useModalCloseConfirmation from '../../../../../../../../utils/hooks/use-modal-close-confirmation'
  import { PACKAGE_MENU_OTHER_TYPE_MAP } from '../../../../../../../../constants/template'
  import { updateTemplatePackage } from '../../../../../../../../api/template'
  import PackageTable from './package-table.vue'
import { Message } from 'bkui-vue'

  const { spaceId } = storeToRefs(useGlobalStore())
  const { currentTemplateSpace, currentPkg, packageList } = storeToRefs(useTemplateStore())

  const props = defineProps<{
    show: boolean;
  }>()

  const emits = defineEmits(['update:show', 'added'])

  const isShow = ref(false)
  const isFormChange = ref(false)
  const pending = ref(false)
  const searchStr = ref('')
  const openedPkgTable = ref<number|string>('')
  const selectedConfigs = ref<{ id: number; name: string; }[]>([])

  const allPackages = computed(() => {
    const packages: { id: number|string; name: string; }[] = [{ id: 'all', name: PACKAGE_MENU_OTHER_TYPE_MAP['all'] }]
    packageList.value.forEach(item => {
      const { id, spec } = item
      packages.push({ id, name: spec.name })
    })
    return packages
  })

  watch(() => props.show, val => {
    isShow.value = val
    isFormChange.value = false
    selectedConfigs.value = []
  })

  const handleToggleOpenTable = (id: string|number) => {
    openedPkgTable.value = openedPkgTable.value === id ? '' : id
  }

  const handleDeleteConfig = (id: number) => {
    console.log('delete config: ', id)
  }

  const handleSelectConfig = (ids: number[]) => {
    isFormChange.value = true
    console.log('select config: ', ids)
  }

  const handleAddConfigs = async() => {
    const pkg = packageList.value.find(item => item.id === currentPkg.value)
    if (!pkg) return

    try {
      pending.value = true
      const { name, memo, public: isPublic, bound_apps } = pkg.spec
      const params = {
        name,
        memo,
        template_ids: selectedConfigs.value.map(item => item.id),
        bound_apps,
        public: isPublic
      }
      await updateTemplatePackage(spaceId.value, currentTemplateSpace.value, <number>currentPkg.value, params)
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
    <div class="slider-content-container">
      <div class="package-configs-pick">
        <div class="search-wrapper">
          <bk-input
            v-model="searchStr"
            class="search-input"
            placeholder="配置项名称/路径/描述"
            :clearable="true">
              <template #suffix>
                <Search class="search-input-icon" />
              </template>
          </bk-input>
        </div>
        <div class="package-tables">
          <PackageTable
            v-for="pkg in allPackages"
            v-model:selected-configs="selectedConfigs"
            :key="pkg.id"
            :pkg="pkg"
            :open="openedPkgTable === pkg.id"
            @toggleOpen="handleToggleOpenTable" />
        </div>
      </div>
      <div class="selected-panel">
        <h5 class="title-text">已选 <span class="num">0</span> 个配置项</h5>
        <div class="selected-list">
          <div v-for="config in selectedConfigs" class="config-item" :key="config.id">
            <div class="name" :title="config.name">{{ config.name }}</div>
            <i class="bk-bscp-icon icon-reduce delete-icon" @click="handleDeleteConfig(config.id)" />
          </div>
          <div class="config-item">
            <div class="name">nginxtestnginxtestnginxtestnginxtestnginxtestnginxtest.bin</div>
            <i class="bk-bscp-icon icon-reduce delete-icon"></i>
          </div>
          <div class="config-item">
            <div class="name">config.bin</div>
            <i class="bk-bscp-icon icon-reduce delete-icon"></i>
          </div>
          <p v-if="selectedConfigs.length === 0" class="empty-tips">请先从左侧选择配置项</p>
        </div>
      </div>
    </div>
    <div class="action-btns">
      <bk-button
        theme="primary"
        :loading="pending"
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
