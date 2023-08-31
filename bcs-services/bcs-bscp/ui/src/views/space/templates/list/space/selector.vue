<script lang="ts" setup>
  import { ref, computed, onMounted, watch } from 'vue';
  import { useRouter } from 'vue-router'
  import { storeToRefs } from 'pinia'
  import { DownShape, Del } from 'bkui-vue/lib/icon'
  import { InfoBox, Message } from 'bkui-vue/lib'
  import { useGlobalStore } from '../../../../../store/global'
  import { useTemplateStore } from '../../../../../store/template'
  import { ITemplateSpaceItem } from '../../../../../../types/template'
  import { ICommonQuery } from '../../../../../../types/index'
  import { getTemplateSpaceList, deleteTemplateSpace, getTemplatesBySpaceId } from '../../../../../api/template'

  import Create from './create.vue'
  import Edit from './edit.vue'

  const router = useRouter()
  const { spaceId } = storeToRefs(useGlobalStore())
  const templateStore = useTemplateStore()
  const { currentTemplateSpace, templateSpaceList } = storeToRefs(templateStore)

  const loading = ref(false)
  const spaceList = ref<ITemplateSpaceItem[]>([])
  const selectorOpen = ref(false)
  const selectorRef = ref()
  const isShowCreateDialog = ref(false)
  const templatesLoading = ref(false)
  const editingData = ref({
    open: false,
    data: { id: 0, name: '', memo: '' }
  })

  const spaceName = computed(() => {
    if (templateSpaceDetail.value.name === 'default_space') {
      return '默认空间'
    }
    return templateSpaceDetail.value.name
  })

  const templateSpaceDetail = computed(() => {
    const item = templateSpaceList.value.find(item => item.id === currentTemplateSpace.value)
    if (item) {
      const { name, memo } = item.spec
      return { name, memo }
    }
    return { name: '', memo: '' }
  })

  watch(() => spaceId.value, () => {
    initData()
  })

  onMounted(() => {
    initData()
  })

  const initData = async() => {
    await loadList()
    if (!currentTemplateSpace.value) {
      const spaceId = spaceList.value[0].id
      if (spaceId) { // url中没有模版空间id，且空间列表不为空时，默认选中第一个空间
        setTemplateSpace(spaceId)
        updateRouter(spaceId)
      }
    } else {
      setTemplateSpace(currentTemplateSpace.value)
    }
  }

  const loadList = async () => {
    loading.value = true
    const params: ICommonQuery = {
      start: 0,
      limit: 1000
      // all: true
    }
    const res = await getTemplateSpaceList(spaceId.value, params)
    spaceList.value = res.details
    templateStore.$patch((state) => {
      state.templateSpaceList = spaceList.value
    })
    loading.value = false
  }

  const handleCreateOpen = () => {
    isShowCreateDialog.value = true
    selectorRef.value.hidePopover()
  }

  const handleEditOpen = (space: ITemplateSpaceItem) => {
    console.log('edit open: ', space)
    const { id, spec } = space
    editingData.value = {
      open: true,
      data: {
        id,
        name: spec.name,
        memo: spec.memo
      }
    }
    selectorRef.value.hidePopover()
  }

  const handleDelete = async(space: ITemplateSpaceItem) => {
    templatesLoading.value = true
    const params = {
      start: 0,
      limit: 1,
      // all: true
    }
    const res = await getTemplatesBySpaceId(spaceId.value, space.id, params)
    if (res.count > 0) {
      InfoBox({
        title: `未能删除【${space.spec.name}】`,
        subTitle: '请先确认删除此空间下所有配置项',
        infoType: 'warning',
        dialogType: 'confirm',
        confirmText: '我知道了',
      } as any)
    } else {
      InfoBox({
        title: `确认删除【${space.spec.name}】`,
        extCls: 'delete-space-infobox',
        onConfirm: async () => {
          await deleteTemplateSpace(spaceId.value, space.id)
          if (space.id === currentTemplateSpace.value) {
            templateStore.$patch(state => {
              state.currentTemplateSpace = ''
            })
            initData()
          } else {
            loadList()
          }
          Message({
            theme: 'success',
            message: '删除成功'
          })
        },
      } as any)
    }

    selectorRef.value.hidePopover()
  }

  const handleSelect = (id: number) => {
    setTemplateSpace(id)
    setCurrentPackage()
    updateRouter(id)
  }

  const setTemplateSpace = (id: number) => {
    templateStore.$patch((state) => {
      state.currentTemplateSpace = id
    })
  }

  const setCurrentPackage = () => {
    templateStore.$patch((state) => {
      state.currentPkg = ''
    })
  }

  const updateRouter = (id: number) => {
    router.push({ name: 'templates-list', params: { templateSpaceId: id } })
  }

</script>
<template>
  <div class="space-selector">
    <bk-select
      ref="selectorRef"
      search-placeholder="搜索空间"
      filterable
      :input-search="false"
      :model-value="currentTemplateSpace"
      :popover-options="{ theme: 'light bk-select-popover template-space-selector-popover' }"
      @toggle="selectorOpen = $event"
      @change="handleSelect">
      <template #trigger>
        <div class="select-trigger">
          <h5 class="space-name" :title="spaceName">{{ spaceName }}</h5>
          <div class="space-desc">{{ templateSpaceDetail.memo }}</div>
          <DownShape :class="['triangle-icon', { up: selectorOpen }]" />
        </div>
      </template>
      <bk-option v-for="item in spaceList" :key="item.id" :value="item.id">
        <div class="space-option-item">
          <div class="name-text">{{ item.spec.name }}</div>
          <div class="actions">
            <i class="bk-bscp-icon icon-edit-small" @click.stop="handleEditOpen(item)"></i>
            <Del v-if="templateSpaceDetail.name !== 'default_space'" class="delete-icon" @click.stop="handleDelete(item)" />
          </div>
        </div>
      </bk-option>
      <template #extension>
        <div class="create-space-extension" @click="handleCreateOpen">
          <i class="bk-bscp-icon icon-add"></i>
          创建空间
        </div>
      </template>
    </bk-select>
  </div>
  <Create v-model:show="isShowCreateDialog" @created="loadList" />
  <Edit v-model:show="editingData.open" :data="editingData.data" @edited="loadList" />
</template>
<style lang="scss" scoped>
  .select-trigger {
    position: relative;
    margin: 0 16px;
    padding: 8px;
    height: 58px;
    background: #f5f7fa;
    border-radius: 2px;
    cursor: pointer;
    .space-name {
      margin: 0;
      width: calc(100% - 20px);
      font-size: 14px;
      color: #313238;
      line-height: 22px;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
    .space-desc {
      width: calc(100% - 20px);
      font-size: 12px;
      color: #979ba5;
      line-height: 20px;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
    .triangle-icon {
      position: absolute;
      top: 50%;
      right: 9px;
      font-size: 12px;
      color: #979ba5;
      transform: translateY(-50%);
      transition: transform .3s cubic-bezier(.4,0,.2,1);
      &.up {
        transform: translateY(-50%) rotate(-180deg);
      }
    }
  }
  .space-option-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 12px;
    width: 100%;
    &:hover {
      .actions {
        display: flex;
      }
    }
    .name-text {
      width: calc(100% - 34px);
      white-space: nowrap;
      text-overflow: ellipsis;
      overflow: hidden;
    }
    .actions {
      display: none;
      align-items: center;
      width: 34px;
      .icon-edit-small {
        margin-right: 4px;
        font-size: 18px;
        color: #979ba5;
        &:hover {
          color: #3a84ff;
        }
      }
      .delete-icon {
        font-size: 12px;
        color: #979ba5;
        &:hover {
          color: #3a84ff;
        }
      }
    }
  }
  .create-space-extension {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 100%;
    font-size: 12px;
    color: #63656e;
    cursor: pointer;
    &:hover {
      color: #3a84ff;
      .icon-add {
        color: #3a84ff;
      }
    }
    .icon-add {
      margin-right: 5px;
      font-size: 14px;
      color: #979ba5;
    }
  }
</style>
<style lang="scss">
  .template-space-selector-popover.bk-popover.bk-pop2-content.bk-select-popover .bk-select-content-wrapper .bk-select-option {
    padding: 0;
  }
</style>
