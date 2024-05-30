<template>
  <div class="package-list-comp">
    <div class="search-wrapper">
      <div class="search-input">
        <SearchInput v-model="searchStr" :placeholder="t('搜索模板套餐')" @search="handleSearch" />
      </div>
      <div class="create-btn" v-bk-tooltips="t('新建模板套餐')" @click="isCreatePackageDialogShow = true">
        <Plus />
      </div>
    </div>
    <div v-if="menuList.length > 0" class="package-list">
      <PackageItem
        v-for="pkg in menuList"
        :key="pkg.id"
        :pkg="pkg"
        :current-pkg="currentPkg"
        @select="handleSelect"
        @open-slider="handlePkgAction" />
    </div>
    <div v-else class="exception-notice">
      <TableEmpty :is-search-empty="isSearchEmpty" @clear="clearSearch"></TableEmpty>
    </div>
    <div class="other-package-list">
      <PackageItem :pkg="MenuItemOfallConfigList" :current-pkg="currentPkg" @select="handleSelect">
        <template #icon>
          <i class="bk-bscp-icon icon-app-store all-config-icon"></i>
        </template>
      </PackageItem>
      <PackageItem
        v-if="countOfTemplatesForNoSpecifiedPackage"
        :pkg="menuItemOfNoSpecifiedPackage"
        :current-pkg="currentPkg"
        @select="handleSelect">
        <template #icon>
          <i class="bk-bscp-icon icon-empty empty-config-icon"></i>
        </template>
      </PackageItem>
    </div>
  </div>
  <PackageCreate
    v-model:show="isCreatePackageDialogShow"
    :template-space-id="currentTemplateSpace"
    @created="handlePkgCreated" />
  <PackageEdit
    v-model:show="editingPkgData.open"
    :template-space-id="currentTemplateSpace"
    :pkg="editingPkgData.data as ITemplatePackageItem"
    @edited="getList" />
  <PackageClone
    v-model:show="cloningPkgData.open"
    :template-space-id="currentTemplateSpace"
    :pkg="cloningPkgData.data as ITemplatePackageItem"
    @created="getList" />
  <PackageDelete
    v-model:show="deletingPkgData.open"
    :template-space-id="currentTemplateSpace"
    :pkg="deletingPkgData.data as ITemplatePackageItem"
    @deleted="handlePkgDeleted" />
</template>
<script lang="ts" setup>
  import { computed, onMounted, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { useRouter } from 'vue-router';
  import { storeToRefs } from 'pinia';
  import { Plus } from 'bkui-vue/lib/icon';
  import useGlobalStore from '../../../../../store/global';
  import useTemplateStore from '../../../../../store/template';
  import {
    getTemplatePackageList,
    getTemplatesWithNoSpecifiedPackage,
    getTemplatesBySpaceId,
  } from '../../../../../api/template';
  import { ITemplatePackageItem, IPackageMenuItem } from '../../../../../../types/template';
  import SearchInput from '../../../../../components/search-input.vue';
  import PackageItem from './item.vue';
  import PackageCreate from './package-create.vue';
  import PackageEdit from './package-edit.vue';
  import PackageClone from './package-clone.vue';
  import PackageDelete from './package-delete.vue';
  import TableEmpty from '../../../../../components/table/table-empty.vue';

  const router = useRouter();
  const { spaceId } = storeToRefs(useGlobalStore());
  const { t } = useI18n();
  const templateStore = useTemplateStore();
  const {
    currentTemplateSpace,
    currentPkg,
    CountOfAllTemplatesInSpace,
    countOfTemplatesForNoSpecifiedPackage,
    needRefreshMenuFlag,
  } = storeToRefs(templateStore);

  const loading = ref(false);
  const packages = ref<ITemplatePackageItem[]>([]); // 全部套餐列表
  const menuList = ref<IPackageMenuItem[]>([]); // 展示到菜单栏的套餐列表，处理搜索场景
  const searchStr = ref('');
  const isSearchEmpty = ref(false);
  const isCreatePackageDialogShow = ref(false);
  const editingPkgData = ref<{ open: boolean; data: ITemplatePackageItem | undefined }>({
    open: false,
    data: undefined,
  });
  const cloningPkgData = ref<{ open: boolean; data: ITemplatePackageItem | undefined }>({
    open: false,
    data: undefined,
  });
  const deletingPkgData = ref<{ open: boolean; data: ITemplatePackageItem | undefined }>({
    open: false,
    data: undefined,
  });

  const MenuItemOfallConfigList = computed(() => ({
    id: 'all',
    name: t('全部配置文件'),
    count: CountOfAllTemplatesInSpace.value,
  }));

  const menuItemOfNoSpecifiedPackage = computed(() => ({
    id: 'no_specified',
    name: t('未指定套餐'),
    count: countOfTemplatesForNoSpecifiedPackage.value,
  }));

  watch(
    () => currentTemplateSpace.value,
    async (val) => {
      searchStr.value = '';
      if (val) {
        getMenuInitData();
      }
    },
  );

  watch(
    () => needRefreshMenuFlag.value,
    async (val) => {
      if (val) {
        await Promise.all([getCountOfAllTemplatesInSpace(), getCountOfTemplatesForNoSpecifiedPackage(), getList()]);
        templateStore.$patch((state) => {
          state.needRefreshMenuFlag = false;
        });
      }
    },
  );

  onMounted(() => {
    if (currentTemplateSpace.value) {
      getMenuInitData();
    }
  });

  const getMenuInitData = async () => {
    getCountOfAllTemplatesInSpace();
    getCountOfTemplatesForNoSpecifiedPackage();
    await getList();
    if (!currentPkg.value) {
      if (packages.value.length > 0) {
        const { id } = packages.value[0];
        setCurrentPackage(id);
        updateRouter(id);
      }
    } else {
      setCurrentPackage(currentPkg.value);
    }
  };

  // 获取全部配置文件数量
  const getCountOfAllTemplatesInSpace = async () => {
    const params = {
      start: 0,
      limit: 1,
    };
    const res = await getTemplatesBySpaceId(spaceId.value, currentTemplateSpace.value, params);
    templateStore.$patch((state) => {
      state.CountOfAllTemplatesInSpace = res.count;
    });
  };

  // 获取未指定套餐配置文件数量
  const getCountOfTemplatesForNoSpecifiedPackage = async () => {
    const params = {
      start: 0,
      limit: 1,
    };
    const res = await getTemplatesWithNoSpecifiedPackage(spaceId.value, currentTemplateSpace.value, params);
    templateStore.$patch((state) => {
      state.countOfTemplatesForNoSpecifiedPackage = res.count;
    });
  };

  // 获取套餐列表
  const getList = async () => {
    loading.value = true;
    const params = {
      start: 0,
      all: true,
    };
    const res = await getTemplatePackageList(spaceId.value, currentTemplateSpace.value, params);
    packages.value = res.details;
    menuList.value = res.details.map((item: ITemplatePackageItem) => {
      const { id, spec } = item;
      return { id, name: spec.name, count: item.spec.template_ids.length };
    });
    templateStore.$patch((state) => {
      state.packageList = res.details;
    });
    loading.value = false;
  };

  const handleSearch = () => {
    let result: ITemplatePackageItem[] = [];
    isSearchEmpty.value = searchStr.value !== '';
    if (searchStr.value) {
      result = packages.value.filter((item) => item.spec.name.toLowerCase().includes(searchStr.value.toLowerCase()));
    } else {
      result = packages.value.slice();
    }
    menuList.value = result.map((item) => {
      const { id, spec } = item;
      return { id, name: spec.name, count: item.spec.template_ids.length };
    });
  };

  const handlePkgAction = (id: number, type: string) => {
    const pkg = packages.value.find((item) => item.id === id);
    if (pkg) {
      if (type === 'edit') {
        editingPkgData.value = {
          open: true,
          data: { ...pkg },
        };
      } else if (type === 'clone') {
        cloningPkgData.value = {
          open: true,
          data: { ...pkg },
        };
      } else if (type === 'delete') {
        deletingPkgData.value = {
          open: true,
          data: { ...pkg },
        };
      }
    }
  };

  const handleSelect = (id: number | string) => {
    setCurrentPackage(id);
    updateRouter(id);
  };

  const handlePkgCreated = (id: number) => {
    updateRouter(id);
    setCurrentPackage(id);
    getList();
  };

  const handlePkgDeleted = (id: number) => {
    if (id === currentPkg.value) {
      templateStore.$patch((state) => {
        state.currentPkg = '';
      });
      getMenuInitData();
    } else {
      getList();
    }
  };

  const setCurrentPackage = (id: number | string) => {
    templateStore.$patch((state) => {
      state.currentPkg = id;
    });
  };

  const updateRouter = (id: number | string) => {
    router.push({ name: 'templates-list', params: { templateSpaceId: currentTemplateSpace.value, packageId: id } });
  };

  const clearSearch = () => {
    searchStr.value = '';
    handleSearch();
  };
</script>
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
      padding: 16px 0 8px;
      max-height: calc(100% - 104px);
      overflow: auto;
    }
    .exception-notice {
      padding: 20px 0 40px;
      :deep(.bk-exception-footer) {
        font-size: 12px;
        color: #63656e;
      }
    }
    .other-package-list {
      padding-top: 8px;
      border-top: 1px solid #dcdee5;
    }
    .empty-config-icon {
      transform-origin: 0 50%;
      transform: scale(0.7);
    }
  }
</style>
