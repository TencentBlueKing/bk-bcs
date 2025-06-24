<template>
  <div
    :class="[
      !openSideMenu ? 'w-[60px]' : '',
      'overflow-x-hidden'
    ]"
    @mouseenter="handleMenuMouseEnter"
    @mouseleave="handleMenuMouseLeave">
    <div
      :class="[
        'flex flex-col h-full transition-all',
        !openSideMenu && !isHover ? 'w-[60px]' : 'w-[260px]',
      ]">
      <!-- 菜单搜索 -->
      <div class="flex-[0_0_auto] flex items-center justify-center px-[16px] pt-[16px] pb-[8px]">
        <bk-input
          right-icon="bk-icon icon-search"
          clearable
          :placeholder="$t('view.placeholder.searchResourceKind')"
          v-model.trim="searchName"
          v-if="isMenuOpen">
        </bk-input>
        <i
          :class="[
            'flex items-center justify-center',
            'h-[32px] w-[32px] text-[16px]',
            'bk-icon bk-icon icon-search'
          ]"
          v-else>
        </i>
      </div>
      <div v-show="openSideMenu" class="flex-[0_0_auto] flex items-center justify-between px-[16px] pt-[16px] pb-[8px]">
        <bk-checkbox
          :true-value="true"
          :false-value="false"
          v-model="hideEmptyCountResourceMenu"
          @change="handleToggleResourceMenuShow">
          <span class="text-[12px]">{{ $t('view.button.onlyData') }}</span>
        </bk-checkbox>
        <span
          :class="[
            'flex items-center text-[#3A84FF] text-[12px]',
            isLoading ? 'cursor-not-allowed' : 'cursor-pointer'
          ]"
          @click.stop="refreshMenu">
          <LoadingIcon v-if="isLoading"></LoadingIcon>
          <i v-else class="bcs-icon bcs-icon-reset mr-[5px] p-[2px]"></i>
          <span>{{ $t('generic.log.button.refresh') }}</span>
        </span>
      </div>
      <!-- 菜单 -->
      <div class="side-menu-wrapper flex-1 overflow-y-auto">
        <bcs-navigation-menu
          item-hover-bg-color="#EAEBF0"
          item-active-bg-color="#E1ECFF"
          item-hover-color="#63656E"
          item-active-color="#3A84FF"
          sub-menu-open-bg-color="#F5F7FA"
          item-active-icon-color="#3A84FF"
          :unique-opened="false"
          :default-active="activeMenu.id"
          :before-nav-change="handleBeforeNavChange"
          :key="menuKey"
          class="resource-view-menu">
          <bcs-navigation-menu-item
            v-for="menu in clusterResourceMenus"
            :key="menu.id"
            :id="menu.id"
            :icon="['bcs-icon', menu.icon]"
            :has-child="menu.children && !!menu.children.length"
            :disabled="disabledMenuIDs.includes(menu.id)"
            ref="menuItemsRef"
            @click="handleChangeMenu(menu)">
            <a
              :class="[
                'flex items-center justify-between no-underline',
                activeMenu.id === menu.id ? 'text-[#3a84ff]' : 'text-[#63656e]'
              ]"
              :href="menu.children && !!menu.children.length ? 'javascript:void(0)' : resolveMenuLink(menu)">
              <span class="flex items-center">
                <span :title="menu.title" class="flex-1 bcs-ellipsis" v-bk-overflow-tips>{{ menu.title }}</span>
                <bcs-tag theme="danger" v-if="menu.tag">{{ menu.tag }}</bcs-tag>
              </span>
              <span
                :class="[
                  'flex items-center justify-center',
                  'px-[8px] h-[16px] bg-[#F0F1F5] rounded-sm',
                  countMap[menu.meta.kind] > 0 ? 'text-[#979BA5]' : 'text-[#C4C6CC]',
                  activeMenu.id === menu.id ? '!bg-[#A3C5FD] !text-[#fff]' : ''
                ]"
                v-if="menu.meta && menu.meta.kind && (menu.meta.kind in countMap)">
                {{ countMap[menu.meta.kind] || 0 }}
              </span>
            </a>
            <template #child>
              <bcs-navigation-menu-item
                v-for="child in menu.children"
                :key="child.id"
                :id="child.id"
                :icon="child.icon ? ['bcs-icon', child.icon] : undefined"
                :disabled="disabledMenuIDs.includes(child.id)"
                @click="handleChangeMenu(child)">
                <a
                  :class="[
                    'flex items-center justify-between no-underline',
                    activeMenu.id === child.id ? 'text-[#3a84ff]' : 'text-[#63656e]'
                  ]"
                  :href="resolveMenuLink(child)">
                  <span class="flex items-center">
                    <span :title="child.title" class="flex-1 bcs-ellipsis" v-bk-overflow-tips>{{ child.title }}</span>
                    <bcs-tag theme="danger" v-if="child.tag" class="mr-[6px]">{{ child.tag }}</bcs-tag>
                  </span>
                  <span
                    :class="[
                      'flex items-center justify-center',
                      'px-[8px] h-[16px] bg-[#F0F1F5] rounded-sm',
                      countMap[child.meta.kind] > 0 ? 'text-[#979BA5]' : 'text-[#C4C6CC]',
                      activeMenu.id === child.id ? '!bg-[#A3C5FD] !text-[#fff]' : ''
                    ]"
                    v-if="child.meta && child.meta.kind && (child.meta.kind in countMap)">
                    {{ countMap[child.meta.kind] || 0 }}
                  </span>
                </a>
              </bcs-navigation-menu-item>
            </template>
          </bcs-navigation-menu-item>
          <!-- 更多资源 -->
          <div class="w-full px-[15px] py-[10px]" v-if="!isEmptyCrdData && !searchName">
            <bcs-divider color="#A3C5FD">
              <LoadingIcon v-if="isLoading">
                <span class="text-[12px] text-[#3a84ff]">{{ $t('generic.status.loading') }}</span>
              </LoadingIcon>
              <bcs-button
                text
                theme="primary"
                size="small"
                class="!px-[0px]"
                v-else
                @click="toggleMoreResource">
                <i
                  :class="[
                    'bk-icon',
                    'flex items-center justify-center',
                    'w-[18px] h-[18px] text-[18px]',
                    showMore ? 'icon-angle-double-up' : 'icon-angle-double-down'
                  ]">
                </i>
                <span v-if="isMenuOpen">{{ $t('view.button.moreResource') }}</span>
              </bcs-button>
            </bcs-divider>
          </div>
          <template v-if="!isLoading">
            <bcs-navigation-menu-item
              v-for="apiVersion in Object.keys(searchCRData)"
              :key="apiVersion"
              :id="apiVersion"
              :icon="['bcs-icon', 'bcs-icon-crd-3']"
              has-child
              v-show="showMore"
              ref="menuItemsRef">
              <span :title="apiVersion" class="flex-1 bcs-ellipsis" v-bk-overflow-tips>{{ apiVersion }}</span>
              <template #child>
                <bcs-navigation-menu-item
                  v-for="(item, index) in searchCRData[apiVersion]"
                  :key="`${index}-${item.kind}`"
                  :id="item.kind"
                  @click="handleChangeCRD(item)">
                  <a
                    :class="[
                      'flex items-center justify-between no-underline',
                      activeMenu.id === item.kind ? 'text-[#3a84ff]' : 'text-[#63656e]'
                    ]"
                    :href="resolveCRDLink(item)">
                    <span :title="item.kind" v-bk-overflow-tips>{{ item.kind }}</span>
                  </a>
                </bcs-navigation-menu-item>
              </template>
            </bcs-navigation-menu-item>
          </template>
        </bcs-navigation-menu>
        <div
          class="h-[100%] flex items-center justify-center"
          v-if="!clusterResourceMenus.length && !Object.keys(searchCRData).length">
          <bcs-exception
            type="empty"
            scene="part"
            class="!w-[260px]" />
        </div>
      </div>
      <!-- 菜单折叠和收起 -->
      <div class="flex-[0_0_auto] flex items-center h-[56px] text-[16px] text-[#C4C6CC] pl-[14px]">
        <i
          :class="[
            'bk-icon icon-expand-line',
            'flex items-center justify-center w-[32px] h-[32px] cursor-pointer transition-all',
            openSideMenu ? 'bcs-rotate' : ''
          ]"
          @click="expandMenu">
        </i>
      </div>
    </div>
  </div>
</template>
<script lang="ts" setup>
import { cloneDeep, isEqual } from 'lodash';
import { computed, onBeforeMount, onBeforeUnmount, reactive, ref, set, toRef, watch } from 'vue';

import useTableData from '../common/use-table-data';

import useViewConfig from './use-view-config';

import { bus } from '@/common/bus';
import LoadingIcon from '@/components/loading-icon.vue';
import $router from '@/router';
import $store from '@/store';
import useMenu, { IMenu } from '@/views/app/use-menu';

// 悬浮时展开菜单
const isHover = ref(false);
const handleMenuMouseEnter = () => {
  // 暂时隐藏到悬浮展开的交互
  // setTimeout(() => {
  //   isHover.value = true;
  // }, 300);
};
const handleMenuMouseLeave = () => {
  isHover.value = false;
};

const searchName = ref('');
// 搜索时展开所有二级菜单
const menuItemsRef = ref();
watch(searchName, () => {
  if (searchName.value) {
    setTimeout(() => {
      menuItemsRef.value?.forEach((item) => {
        item.handleOpen();
      });
    });
  }
});

const { menus, disabledMenuIDs, flatLeafMenus } = useMenu();
// 只显示有数据的资源
const hideEmptyCountResourceMenu = ref(localStorage.getItem('__BCS_RESOURCE_VIEW_MENU_HIDE__') === 'true');
// 解决菜单渲染问题
const randomKey = ref('');
const menuKey = computed(() => `${searchName.value}-${hideEmptyCountResourceMenu.value}-${randomKey.value}`);
// 切换菜单显示和隐藏
function handleToggleResourceMenuShow(v: boolean) {
  localStorage.setItem('__BCS_RESOURCE_VIEW_MENU_HIDE__', `${v}`);
}
// 刷新菜单
function refreshMenu() {
  if (isLoading.value) return;
  handleGetCustomResourceDefinition();
  handleGetMultiClusterResourcesCount();
}
// 左侧菜单
const activeMenu = ref<Partial<IMenu>>({});
// 所有叶子菜单项
const leafMenus = computed(() => flatLeafMenus(menus.value));
// 一级菜单
const clusterResourceMenus = computed<IMenu[]>(() => {
  const data = menus.value.find(item => item.id === 'CLUSTERRESOURCE')?.children || [];
  return data.reduce<IMenu[]>((pre, item) => {
    if (item.children?.length) {
      const newChildrenList = item.children
        .filter(menu => filterMenu(menu));
      if (newChildrenList.length) {
        pre.push({
          ...item,
          children: newChildrenList,
        });
      }
    } else if (filterMenu(item)) {
      pre.push(item);
    }
    return pre;
  }, []);
});
// 过滤菜单
function filterMenu(menu: IMenu) {
  const searchValue = searchName.value.toLocaleLowerCase();
  if (hideEmptyCountResourceMenu.value) {
    return menu?.title?.toLocaleLowerCase()?.includes(searchValue) && countMap.value[menu.meta?.kind] !== 0;
  }
  return menu?.title?.toLocaleLowerCase()?.includes(searchValue);
}

// 当前路由
const route = computed(() => toRef(reactive($router), 'currentRoute').value);
// 设置当前菜单ID
watch(
  [
    () => route.value,
  ],
  () => {
    // 路由上配置了菜单ID或者路由名称与当前子菜单项路由名称一致
    const menu = leafMenus.value
      .find(item => item.route === route.value.name || item.id === route.value.meta?.menuId);

    if (menu) {
      activeMenu.value = menu || {};
      // 更新当前一级导航信息
      $store.commit('updateCurNav', activeMenu.value);
    } else if (route.value?.query?.kind) {
      activeMenu.value = { id: route.value?.query?.kind as string };
    } else {
      console.warn(`current route ${route.value.name} has no matched menuId`);
    }
  },
  { immediate: true },
);

// 切换菜单
const handleBeforeNavChange = () => false;
const resolveMenuLink = (item: IMenu) => {
  const { href } = $router.resolve({
    name: item.route || item.children?.[0]?.route || '404',
    params: {
      projectCode: $store.getters.curProjectCode,
    },
  });
  return href;
};
const handleChangeMenu = (item: IMenu) => {
  if (route.value.name === item.route) return;

  $store.commit('updateCrdData', {});
  const queryData = cloneDeep({
    ...route.value.query,
    viewID: dashboardViewID.value,
  });
  delete queryData.crd;
  delete queryData.kind;
  delete queryData.scope;
  !dashboardViewID.value && (delete queryData.viewID);
  $router.push({
    name: item.route || item.children?.[0]?.route || '404',
    params: {
      projectCode: $store.getters.curProjectCode,
      clusterId: route.value.params?.clusterId,
    },
    query: {
      ...queryData,
    },
  });
};

// 切换自定义资源
const resolveCRDLink = (item) => {
  const routeData = {
    name: 'dashboardCustomObjects',
    params: {
      projectCode: $store.getters.curProjectCode,
      clusterId: route.value.params?.clusterId, // 保留之前的集群ID
    },
    query: {
      crd: item.name,
      kind: item.kind,
      scope: item.scope,
    },
  };
  const { href } = $router.resolve(routeData);
  return href;
};
const handleChangeCRD = async (item) => {
  const queryData = cloneDeep(route.value.query);
  queryData.crd = `${item.resource}.${item.group}`;
  queryData.kind = item.kind;
  queryData.scope = item.namespaced ? 'Namespaced' : 'Cluster';
  queryData.group = item.group;
  queryData.version = item.version;
  queryData.resource = item.resource;
  queryData.namespaced = item.namespaced;
  // 自定义资源
  $store.commit('updateCrdData', queryData);
  const routeData = {
    name: 'dashboardCustomObjects',
    params: {
      projectCode: $store.getters.curProjectCode,
      clusterId: route.value.params?.clusterId, // 保留之前的集群ID
    },
    query: queryData,
  };
  const { resolved } = await $router.resolve(routeData);
  if (resolved?.fullPath === route.value.fullPath) return;

  $router.push(routeData);
};

// 菜单折叠和收起
const openSideMenu = computed(() => $store.state.openSideMenu);
const isMenuOpen = computed(() => isHover.value || openSideMenu.value);
const expandMenu = () => {
  $store.commit('updateOpenSideMenu', !openSideMenu.value);
  isHover.value = false;
};

// 更多资源
const {
  getMultiClusterResourcesCount,
  getMultiClusterAPIResources,
} = useTableData();
const { curViewData, dashboardViewID } = useViewConfig();
const isLoading = ref(false);
const crdData = ref<Record<string, any[]>>({});
const searchCRData = computed(() => {
  // 资源名称搜索
  const searchValue = searchName.value.toLocaleLowerCase();
  return Object.keys(crdData.value).reduce<Record<string, any[]>>((pre, key) => {
    const data = crdData.value[key]?.filter(item => item.kind?.toLocaleLowerCase().includes(searchValue));
    if (data?.length) {
      pre[key] = data;
    }
    return pre;
  }, {});
});
const isEmptyCrdData = computed(() => !Object.keys(crdData.value).length);
const showMoreResource = ref(true);
const showMore = computed(() => showMoreResource.value || searchName.value);
const toggleMoreResource = () => {
  showMoreResource.value = !showMoreResource.value;
};
const tkexCRDList = computed(() => flatLeafMenus(menus.value.find(item => item.id === 'CLUSTERRESOURCE')?.children || []));
const handleGetCustomResourceDefinition = async () => {
  if (!curViewData.value) return;

  isLoading.value = true;
  const res = await getMultiClusterAPIResources(route.value.params?.clusterId || '', false);
  isLoading.value = false;
  crdData.value = Object.keys(res.resources || {}).reduce((pre, key) => {
    const item = res.resources?.[key] || [];

    const list: any[] = [];
    for (const resource of item) {
      if (!tkexCRDList.value.find(v => v.meta?.kind === resource.kind)) { // tkex自定义资源有单独UI展示
        list.push(resource);
      }
    }

    pre[key] = list;
    return pre;
  }, {});
};
// 资源统计
const countMap = ref<Record<string, number>>({});
const handleGetMultiClusterResourcesCount = async () => {
  countMap.value = {};// 重置数据
  if (!curViewData.value) return;

  countMap.value = await getMultiClusterResourcesCount({
    ...curViewData.value,
  });
};
watch(countMap, (newValue, oldValue) => {
  // hack 修复count数量获取后菜单没有更新的异常
  if (!isEqual(newValue, oldValue) && hideEmptyCountResourceMenu.value) {
    randomKey.value = `${Math.random() * 10}`;
  }
});

watch(curViewData, (newValue, oldValue) => {
  if (!curViewData.value || isEqual(newValue, oldValue)) return;
  handleGetCustomResourceDefinition();
  handleGetMultiClusterResourcesCount();
}, { deep: true });

onBeforeMount(() => {
  bus.$on('set-resource-count', (kind: string, count: number) => {
    if (count === undefined) return;

    set(countMap.value, kind, count);
  });
  if (curViewData.value?.clusterNamespaces?.some(item => item?.clusterID === '-')) {
    return;
  }
  handleGetCustomResourceDefinition();
  handleGetMultiClusterResourcesCount();
});

onBeforeUnmount(() => {
  bus.$off('set-resource-count');
});

</script>
<!-- 覆盖导航默认样式 -->
<style lang="postcss">
.nav-slider {
  .navigation-sbmenu {
    margin-bottom: 2px;
  }
  .nav-slider-list {
    padding: 6px 0 4px 0!important;
  }
  .navigation-menu-item:hover:not(.is-disabled) {
    background-color: #EAEBF0;
  }
  .footer-icon {
    color: #C4C6CC !important;
  }
}
</style>
<style scoped lang="postcss">
>>> .resource-view-menu {
  .navigation-menu-item-name {
    width: 100%;
  }
}
.side-menu-wrapper {
  &::-webkit-scrollbar {
    display: none;
  }
  .navigation-menu:first-child {
    margin-top: 6px;
  }
}
</style>
