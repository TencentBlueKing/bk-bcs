<script lang="ts" setup>
  import { ref, computed, onMounted } from 'vue';
  import { storeToRefs } from 'pinia'
  import { DownShape, Del } from 'bkui-vue/lib/icon'
  import { InfoBox, Message } from 'bkui-vue/lib'
  import { useGlobalStore } from '../../../../../../store/global'
  import { useTemplateStore } from '../../../../../../store/template'
  import { ITemplateSpaceItem } from '../../../../../../../types/template'
  import { ICommonQuery } from '../../../../../../../types/index'
  import { getTemplateSpaceList, deleteTemplateSpace } from '../../../../../../api/template'

  import Create from './create.vue'
  import Edit from './edit.vue'

  const { spaceId } = storeToRefs(useGlobalStore())
  const templateStore = useTemplateStore()
  const { currentTemplateSpace, templateSpaceList } = storeToRefs(templateStore)

  const loading = ref(false)
  const spaceList = ref<ITemplateSpaceItem[]>([])
  const selectorOpen = ref(false)
  const isShowCreateDialog = ref(false)
  const editingData = ref({
    open: false,
    data: { id: 0, name: '', memo: '' }
  })
  const deletingData = ref({
    open: false,
    data: { id: 0, name: '', memo: '' }
  })

  const templateSpaceDetail = computed(() => {
    const item = templateSpaceList.value.find(item => item.id === currentTemplateSpace.value)
    if (item) {
      const { name, memo } = item.spec
      return { name, memo }
    }
    return { name: '', memo: '' }
  })

  onMounted(async () => {
    await loadList()
    templateStore.$patch((state) => {
      state.currentTemplateSpace = spaceList.value[0]?.id
    })
  })

  const loadList = async () => {
    loading.value = true
    const params: ICommonQuery = {
      start: 0,
      limit: 100
      // all: true
    }
    const res = await getTemplateSpaceList(spaceId.value, params)
    spaceList.value = res.details
    templateStore.$patch((state) => {
      state.templateSpaceList = spaceList.value
    })
    loading.value = false
  }

  const handleSelectSpace = (val: number) => {
    templateStore.$patch((state) => {
      state.currentTemplateSpace = val
    })
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
  }

  const handleDelete = (space: ITemplateSpaceItem) => {
    console.log('delete open: ', space)
    InfoBox({
      title: `确认删除【${space.spec.name}】`,
      extCls: 'delete-space-infobox',
      onConfirm: async () => {
        await deleteTemplateSpace(spaceId.value, space.id)
        loadList()
        Message({
          theme: 'success',
          message: '删除成功'
        })
      },
    } as any)
    // InfoBox({
    //   title: `未能删除【${space.spec.name}】`,
    //   subTitle: '请先确认删除此空间下所有配置项',
    //   infoType: 'warning',
    //   dialogType: 'confirm',
    //   confirmText: '我知道了',
    // } as any)
  }

</script>
<template>
  <div class="space-selector">
    <bk-select
      search-placeholder="搜索空间"
      filterable
      :input-search="false"
      :model-value="currentTemplateSpace"
      :popover-options="{ theme: 'light bk-select-popover template-space-selector-popover' }"
      @toggle="selectorOpen = $event"
      @change="handleSelectSpace">
      <template #trigger>
        <div class="select-trigger">
          <h5 class="space-name" :title="templateSpaceDetail.name">{{ templateSpaceDetail.name }}</h5>
          <div class="space-desc">{{ templateSpaceDetail.memo }}</div>
          <DownShape :class="['triangle-icon', { up: selectorOpen }]" />
        </div>
      </template>
      <bk-option v-for="item in spaceList" :key="item.id" :value="item.id">
        <div class="space-option-item">
          <div class="name-text">{{ item.spec.name }}</div>
          <div class="actions">
            <i class="bk-bscp-icon icon-edit-small" @click.stop="handleEditOpen(item)"></i>
            <Del class="delete-icon" @click.stop="handleDelete(item)" />
          </div>
        </div>
      </bk-option>
      <template #extension>
        <div class="create-space-extension" @click="isShowCreateDialog = true">
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
