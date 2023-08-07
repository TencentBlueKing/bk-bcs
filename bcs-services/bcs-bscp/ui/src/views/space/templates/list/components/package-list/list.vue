<script lang="ts" setup>
  import { ref, watch } from 'vue';
  import { storeToRefs } from 'pinia'
  import { Plus, Search } from 'bkui-vue/lib/icon'
  import { useGlobalStore } from '../../../../../../store/global'
  import { useTemplateStore } from '../../../../../../store/template'
  import { getTemplatePackageList } from '../../../../../../api/template'
  import { ITemplatePackageItem } from '../../../../../../../types/template'
  import PackageItem from './item.vue'
  import PackageCreate from './package-create.vue'
  import PackageEdit from './package-edit.vue';
  import PackageClone from './package-clone.vue';
  import PackageDelete from './package-delete.vue';

  const { spaceId } = storeToRefs(useGlobalStore())
  const templateStore = useTemplateStore()
  const { currentTemplateSpace, currentPkg } = storeToRefs(templateStore)

  const loading = ref(false)
  const list = ref<ITemplatePackageItem[]>([])
  const isCreatePackageDialogShow = ref(false)
  const editingPkgData = ref<{ open: boolean; data: ITemplatePackageItem|undefined }>({
    open: false,
    data: undefined
  })
  const cloningPkgData = ref<{ open: boolean; data: ITemplatePackageItem|undefined }>({
    open: false,
    data: undefined
  })
  const deletingPkgData = ref<{ open: boolean; data: ITemplatePackageItem|undefined }>({
    open: false,
    data: undefined
  })

  watch([() => spaceId.value, () => currentTemplateSpace.value], async([newSpaceId, newTemplateSpace]) => {
    console.log(newSpaceId, newTemplateSpace)
    await getList()
    templateStore.$patch((state) => {
      state.currentPkg = state.packageList?.[0].id
    })
  })

  const getList = async () => {
    loading.value = true
    const params = {
      start: 0,
      limit: 1000,
      // all: true
    }
    const res = await getTemplatePackageList(spaceId.value, currentTemplateSpace.value, params)
    list.value = res.details
    templateStore.$patch((state) => {
      state.packageList = res.details
    })
    loading.value = false
  }

  const handleEdit = (pkg: ITemplatePackageItem) => {
    editingPkgData.value = {
      open: true,
      data: { ...pkg }
    }
  }

  const handleClone = (pkg: ITemplatePackageItem) => {
    cloningPkgData.value = {
      open: true,
      data: { ...pkg }
    }
  }

  const handleDelete = (pkg: ITemplatePackageItem) => {
    console.log(pkg)
    deletingPkgData.value = {
      open: true,
      data: { ...pkg }
    }
  }

</script>
<template>
  <div class="package-list-comp">
    <div class="search-wrapper">
      <div class="create-btn" v-bk-tooltips="'新建模板套餐'" @click="isCreatePackageDialogShow = true">
        <Plus />
      </div>
      <div class="search-input">
        <bk-input placeholder="搜索模板套餐">
          <template #suffix>
            <Search class="search-icon" />
          </template>
        </bk-input>
      </div>
    </div>
    <div class="package-list">
      <PackageItem
        v-for="pkg in list"
        :key="pkg.id"
        :pkg="pkg"
        :current-pkg="currentPkg"
        :name="pkg.spec.name"
        :count="pkg.spec.template_ids.length"
        @edit="handleEdit"
        @clone="handleClone"
        @delete="handleDelete" />
    </div>
    <div class="other-package-list">
      <PackageItem  name="全部配置项" :count="30" :current-pkg="currentPkg">
        <template #icon>
          <i class="bk-bscp-icon icon-app-store all-config-icon"></i>
        </template>
      </PackageItem>
      <PackageItem name="未指定套餐" :count="15" :current-pkg="currentPkg">
        <template #icon>
          <i class="bk-bscp-icon icon-empty empty-config-icon"></i>
        </template>
      </PackageItem>
    </div>
  </div>
  <PackageCreate
    v-model:show="isCreatePackageDialogShow"
    :template-space-id="currentTemplateSpace"
    @created="getList" />
  <PackageEdit
    v-model:show="editingPkgData.open"
    :template-space-id="currentTemplateSpace"
    :pkg="(editingPkgData.data as ITemplatePackageItem)"
    @edited="getList" />
  <PackageClone
    v-model:show="cloningPkgData.open"
    :template-space-id="currentTemplateSpace"
    :pkg="(cloningPkgData.data as ITemplatePackageItem)"
    @created="getList" />
  <PackageDelete
    v-model:show="deletingPkgData.open"
    :template-space-id="currentTemplateSpace"
    :pkg="(deletingPkgData.data as ITemplatePackageItem)"
    @deleted="getList" />
</template>
<style lang="scss" scoped>
  .package-list-comp {
    padding-top: 12px;
    height: calc(100% - 58px);
    .search-wrapper {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 0 16px;
    }
    .create-btn {
      display: inline-flex;
      align-items: center;
      justify-content: center;
      margin-right: 8px;
      width: 32px;
      height: 32px;
      font-size: 24px;
      color: #c4c6cc;
      border: 1px solid #c4c6cc;
      border-radius: 2px;
      cursor: pointer;
      &:hover {
        color: #3a84ff;
        border-color: #3a84ff;
      }
    }
    .search-input {
      width: calc(100% - 40px);
    }
    .search-icon {
      margin-right: 10px;
      color: #979ba5;
    }
    .package-list {
      padding-top: 16px;
      height: calc(100% - 104px);
    }
    .other-package-list {
      padding-top: 8px;
      border-top: 1px solid #dcdee5;
    }
    // .all-config-icon {
    //   transform-origin: 0 50%;
    //   transform: scale(0.8);
    // }
    .empty-config-icon {
      transform-origin: 0 50%;
      transform: scale(0.7);
    }
  }
</style>
