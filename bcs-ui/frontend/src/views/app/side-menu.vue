<template>
  <bcs-navigation-menu
    item-hover-bg-color="#EAEBF0"
    item-active-bg-color="#E1ECFF"
    item-hover-color="#63656E"
    item-active-color="#3A84FF"
    sub-menu-open-bg-color="#F5F7FA"
    item-active-icon-color="#3A84FF"
    :unique-opened="false"
    :default-active="activeMenu.id"
    :before-nav-change="handleBeforeNavChange">
    <bcs-navigation-menu-item
      v-for="menu in activeNav.children"
      :key="menu.id"
      :id="menu.id"
      :icon="['bcs-icon', menu.icon]"
      :has-child="menu.children && !!menu.children.length"
      :disabled="disabledMenuIDs.includes(menu.id)"
      @click="handleChangeMenu(menu)">
      <span :title="menu.title">{{ menu.title }}</span>
      <bcs-tag theme="danger" v-if="menu.tag">{{ menu.tag }}</bcs-tag>
      <template #child>
        <bcs-navigation-menu-item
          v-for="child in menu.children"
          :key="child.id"
          :id="child.id"
          :icon="['bcs-icon', child.icon]"
          :disabled="disabledMenuIDs.includes(child.id)"
          @click="handleChangeMenu(child)">
          <span :title="child.title">{{ child.title }}</span>
          <bcs-tag theme="danger" v-if="child.tag">{{ child.tag }}</bcs-tag>
        </bcs-navigation-menu-item>
      </template>
    </bcs-navigation-menu-item>
  </bcs-navigation-menu>
</template>
<script lang="ts">
import { computed, defineComponent, reactive, ref, toRef, watch } from 'vue';

import useMenu, { IMenu } from './use-menu';

import { useProject } from '@/composables/use-app';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import $store from '@/store';

export default defineComponent({
  name: 'SideMenu',
  setup() {
    const { menus, disabledMenuIDs, flatLeafMenus } = useMenu();
    // 左侧菜单
    const activeMenu = ref<Partial<IMenu>>({});
    // 一级菜单
    const activeNav = ref<Partial<IMenu>>({});
    // 所有叶子菜单项
    const leafMenus = computed(() => flatLeafMenus(menus.value));
    // 当前路由
    const route = computed(() => toRef(reactive($router), 'currentRoute').value);

    // 设置当前菜单ID
    watch(
      [
        () => route.value,
        () => $i18n.locale,
      ],
      () => {
        // 路由上配置了菜单ID或者路由名称与当前子菜单项路由名称一致
        const menu = leafMenus.value
          .find(item => item.route === route.value.name || item.id === route.value.meta?.menuId);

        if (!menu) {
          console.warn(`current route ${route.value.name} has no matched menuId`);
        } else {
          activeMenu.value = menu || {};
          activeNav.value = menu?.root || {};
          $store.commit('updateCurSideMenu', activeMenu.value);
        }
      },
      { immediate: true },
    );

    // 切换菜单
    const { projectCode } = useProject();
    const handleBeforeNavChange = () => false;
    const handleChangeMenu = (item: IMenu) => {
      if (route.value.name === item.route) return;

      if (item.id === 'MONITOR') {
        window.open(`${window.BKMONITOR_HOST}/?space_uid=bkci__${projectCode.value}#/k8s`);
      } else {
        $router.push({
          name: item.route || item.children?.[0]?.route || '404',
          params: {
            projectCode: $store.getters.curProjectCode,
          },
        });
      }
    };

    return {
      activeMenu,
      activeNav,
      disabledMenuIDs,
      handleChangeMenu,
      handleBeforeNavChange,
    };
  },
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
