<!-- eslint-disable @typescript-eslint/semi -->
<template>
  <div>
    <bcs-navigation
      navigation-type="top-bottom"
      :need-menu="needMenu"
      :default-open="openSideMenu"
      :hover-enter-delay="300"
      @toggle-click="handleToggleClickNav"
      @hover="handleInitSliderListHeight"
      @leave="handleInitSliderListHeight">
      <template #side-header>
        <span class="title-icon"><img :src="appLogo" class="w-[28px] h-[28px]"></span>
        <span
          class="title-desc cursor-pointer"
          @click="handleGoHome">
          {{ appName }}
        </span>
      </template>
      <template #header>
        <!-- 顶部菜单 -->
        <ol class="flex flex-1 w-[0] text-[14px] text-[#96a2b9] overflow-hidden" ref="navRef">
          <li
            v-for="(item, index) in menus"
            :class="[
              'mr-[40px] hover:text-[#d3d9e4] cursor-pointer whitespace-nowrap',
              {
                'text-[#d3d9e4]': activeNav.id === item.id,
                'opacity-0': (index >= breakIndex) && (breakIndex > -1)
              }
            ]"
            :key="index"
            ref="navItemRefs"
            @click="handleChangeMenu(item)">
            <a
              :class="[
                'text-[#96a2b9] hover:text-[#d3d9e4]',
                {
                  'text-[#d3d9e4]': activeNav.id === item.id,
                }
              ]"
              :href="resolveMenuLink(item)"
              @click.prevent>
              {{ item.title }}
            </a>
          </li>
        </ol>
        <!-- 折叠的菜单 -->
        <PopoverSelector class="w-[24px] relative right-[24px]" v-show="breakIndex > -1">
          <span class="text-[#fff] text-[18px] h-[52px] cursor-pointer">
            <i class="bk-icon icon-ellipsis relative top-[-1px]"></i>
          </span>
          <template #content>
            <ul>
              <li
                v-for="item in hiddenMenus"
                :key="item.id"
                :class="['bcs-dropdown-item', { active: activeNav.id === item.id }]"
                @click="handleChangeMenu(item)">
                {{ item.title }}
              </li>
            </ul>
          </template>
        </PopoverSelector>
        <!-- 返回旧版 -->
        <span
          :class="[
            'flex items-center h-[32px] px-[12px] cursor-pointer',
            'text-[#96A2B9] text-[12px] mr-[24px] hover:bg-[linear-gradient(270deg,#253047,#263247)] rounded-sm'
          ]"
          v-if="showPreVersionBtn"
          v-bk-trace.click="{
            module: 'app',
            operation: 'backToLegacy',
            desc: '返回旧版',
            username: $store.state.user.username,
            projectCode: $store.getters.curProjectCode,
          }"
          @click="backToPreVersion">
          {{ $t('bcs.preVersion') }}
        </span>
        <!-- 项目切换 -->
        <ProjectSelector class="ml-auto w-[240px] mr-[18px]"></ProjectSelector>
        <!-- 语言切换 -->
        <PopoverSelector class="mr-[8px]" ref="langRef">
          <span class="header-icon text-[18px]">
            <i :class="curLang.icon"></i>
          </span>
          <template #content>
            <ul>
              <li
                v-for="(item, index) in langs"
                :key="index"
                :class="['bcs-dropdown-item', { active: curLang.id === item.id }]"
                @click="handleChangeLang(item)">
                <i :class="['text-[18px] mr5', item.icon]"></i>
                {{item.name}}
              </li>
            </ul>
          </template>
        </PopoverSelector>
        <!-- 帮助文档 -->
        <PopoverSelector class="mr-[8px]">
          <span id="siteHelp" class="header-icon !text-[16px]">
            <i class="bcs-icon bcs-icon-help-document-fill"></i>
          </span>
          <template #content>
            <ul>
              <li class="bcs-dropdown-item" @click="handleGotoHelp">{{ $t('blueking.docs') }}</li>
              <li class="bcs-dropdown-item" @click="handleShowSystemLog">{{ $t('blueking.releaseNotes') }}</li>
              <li class="bcs-dropdown-item" @click="handleShowFeatures">{{ $t('blueking.features') }}</li>
            </ul>
          </template>
        </PopoverSelector>
        <!-- 用户设置 -->
        <PopoverSelector class="ml-[4px]">
          <span class="flex items-center text-[#96A2B9] hover:text-[#d3d9e4]">
            <span class="text-[14px]">{{user.username}}</span>
            <i class="ml-[4px] text-[12px] bk-icon icon-down-shape"></i>
          </span>
          <template #content>
            <ul>
              <li class="bcs-dropdown-item" @click="handleGotoUserToken">{{ $t('blueking.apiToken') }}</li>
              <li class="bcs-dropdown-item" @click="handleLogout">{{ $t('blueking.signOut') }}</li>
            </ul>
          </template>
        </PopoverSelector>
      </template>
      <!-- 左侧菜单 -->
      <template #menu>
        <slot name="sideMenu"></slot>
      </template>
      <!-- 视图 -->
      <template #default>
        <slot></slot>
      </template>
    </bcs-navigation>
    <!-- 系统日志 -->
    <SystemLog v-model="showSystemLog" :list="releaseData.changelog" />
    <!-- 产品特性 -->
    <bcs-dialog
      v-model="showFeatures"
      :title="$t('blueking.features1')"
      :show-footer="false"
      width="480">
      <BcsMd :code="releaseData.feature.content" />
    </bcs-dialog>
  </div>
</template>
<script lang="ts">
import jsonp from 'jsonp';
import { computed, defineComponent, onBeforeUnmount, onMounted, reactive, ref, toRef } from 'vue';

import PopoverSelector from '../../components/popover-selector.vue';

import useMenu, { IMenu } from './use-menu';

import { releaseNote } from '@/api/modules/project';
import { setCookie } from '@/common/util';
import BcsMd from '@/components/bcs-md/index.vue';
import useCalcHeight from '@/composables/use-calc-height';
import usePlatform from '@/composables/use-platform';
import $i18n from '@/i18n/i18n-setup';
import logoSvg from '@/images/rtx.svg';
import $router from '@/router';
import $store from '@/store';
import SystemLog from '@/views/app/log.vue';
import ProjectSelector from '@/views/app/project-selector.vue';

export default defineComponent({
  name: 'NewNavigation',
  components: {
    SystemLog,
    BcsMd,
    ProjectSelector,
    PopoverSelector,
  },
  setup() {
    const { init } = useCalcHeight([
      {
        prop: 'height',
        el: '.nav-slider-list',
        calc: ['#bcs-notice-com', '.bk-navigation-header', '.nav-slider-footer'],
      },
    ]);
    // 修复nav悬浮时高度被组件内部覆盖问题
    const handleInitSliderListHeight = () => {
      setTimeout(() => {
        init();
      });
    };

    const { menusData: menus } = useMenu();
    const { config } = usePlatform();
    const appLogo = computed(() => config.appLogo || logoSvg);
    const appName = computed(() => config.i18n.name);
    const langs = ref([
      {
        icon: 'bk-icon icon-english',
        name: 'English',
        id: 'en-US',
        locale: 'en',
      },
      {
        icon: 'bk-icon icon-chinese',
        name: '中文',
        id: 'zh-CN', // bcs 前端语言标识
        locale: 'zh-cn', // cookie标识
      },
    ]);
    const curLang = computed(() => langs.value.find(item => item.id === $i18n.locale) || { id: 'zh-CN', icon: 'bk-icon icon-chinese' });
    const showSystemLog = ref(false);
    const showFeatures = ref(false);
    const user = computed(() => $store.state.user);
    const curProject = computed(() => $store.state.curProject);
    const activeNav = computed(() => findCurrentNav(menus.value) || {});
    // 当前路由
    const route = computed(() => toRef(reactive($router), 'currentRoute').value);
    const hideMenu = computed(() => !!route.value.meta?.hideMenu || !!route.value.matched?.[0]?.meta?.hideMenu);
    // 是否显示左侧菜单
    const needMenu = computed(() => {
      const { projectCode } = route.value.params;
      return !!projectCode
        && route.value.fullPath.indexOf(projectCode) > -1 // 1.跟项目无关界面
        && (!!curProject.value.kind && !!curProject.value.businessID && curProject.value.businessID !== '0')// 2. 当前项目未开启容器服务
        && !['404', 'token'].includes(route.value.name || '')// 404 和 token特殊界面
        && !hideMenu.value;
    });

    // 导航自适应
    const breakIndex = ref(-1);
    const hiddenMenus = computed(() => {
      if (breakIndex.value === -1) return [];
      return menus.value.slice(breakIndex.value, menus.value.length);
    });
    const navRef = ref();
    const navItemRefs = ref<any[]>([]);
    const resizeObserver = new ResizeObserver(() => {
      window.requestAnimationFrame(() => {
        const navWrapperWidth = navRef.value?.clientWidth;
        let tmpWidth = 24;// 最小宽度
        const index = navItemRefs.value?.findIndex((item) => {
          tmpWidth += (item.clientWidth + 40); // 40: margin-right: 40px
          return tmpWidth >= navWrapperWidth;
        });
        breakIndex.value = index;
      });
    });

    // 当前导航
    const findCurrentNav = (menus: IMenu[]) => menus.find((item) => {
      if (item.route === route.value.name || item.id === route.value.meta?.menuId) return true;

      if (item.children?.length) return findCurrentNav(item.children);

      return false;
    });

    // 菜单link
    const resolveMenuLink = (item: IMenu) => {
      const name = item.route || item.children?.[0]?.route || '404';
      const { href } = $router.resolve({
        name,
        params: {
          projectCode: $store.getters.curProjectCode,
          clusterId: $store.getters.curClusterId,
        },
      });
      return href;
    };
    // 切换菜单
    const handleChangeMenu = (item: IMenu) => {
      const name = item.route || item.children?.[0]?.route || '404';
      if (route.value.name === name) return;

      // 更新当前一级导航信息
      $store.commit('updateCurNav', item);
      $router.push({
        name,
        params: {
          projectCode: $store.getters.curProjectCode,
          // 资源视图集群视图时切换路由不能丢失集群ID，其余菜单默认不给
          clusterId: item.id === 'CLUSTERRESOURCE' ? $router.currentRoute?.params?.clusterId : '',
        },
      }).catch(err => console.warn(err));
    };

    // 左侧菜单折叠和收起
    const openSideMenu = computed(() => $store.state.openSideMenu);
    const handleToggleClickNav = (value) => {
      $store.commit('updateOpenSideMenu', !!value);
      handleInitSliderListHeight();
    };
    // 首页
    const handleGoHome = () => {
      if ($router.currentRoute?.name === 'dashboardWorkloadDeployments') return;
      $router.push({
        name: 'dashboardWorkloadDeployments',
        params: {
          clusterId: $router.currentRoute?.params?.clusterId,
        },
      });
    };

    // 切换语言
    const langRef = ref();
    const handleChangeLang = async (item) => {
      // $i18n.locale = item.id;// 后面 $router.go(0) 会重新加载界面，这里会导致一瞬间被切换了，然后界面再刷新
      langRef.value?.hide();
      // 修改cookie
      setCookie('blueking_language', item.locale, window.BK_DOMAIN);
      // 修改用户管理语言
      jsonp(
        `${window.BK_USER_HOST}/api/c/compapi/v2/usermanage/fe_update_user_language/?language=${item.locale}`,
        { param: 'callback', timeout: 100 },
        () => {
          $router.go(0);
        },
      );
    };
    // 帮助文档
    const handleGotoHelp  = () => {
      window.open(window.BCS_CONFIG?.help);
    };
    // 版本日志
    const handleShowSystemLog = () => {
      showSystemLog.value = true;
    };
    // 版本特性
    const handleShowFeatures = () => {
      showFeatures.value = true;
    };
    // 跳转用户token
    const handleGotoUserToken = () => {
      if (route.value.name === 'token') return;
      $router.push({
        name: 'token',
      });
    };
    // 注销登录态
    const handleLogout = () => {
      // 注销登录只注销当前登录态，清除bk_token，不做登录弹窗
      window.location.href = `${window.LOGIN_FULL}?is_from_logout=1&c_url=${encodeURIComponent(window.location.href)}`;
    };

    // release信息
    const releaseData = ref({
      changelog: [],
      feature: { content: '' },
    });

    // 返回旧版
    const showPreVersionBtn = ref(!!window.BCS_CONFIG.backToLegacyButtonUrl);
    const backToPreVersion = () => {
      window.open(window.BCS_CONFIG.backToLegacyButtonUrl);
    };

    onMounted(async () => {
      navRef.value && resizeObserver.observe(navRef.value);
      releaseData.value = await releaseNote().catch(() => ({ changelog: [], feature: {} }));
    });

    onBeforeUnmount(() => {
      navRef.value && resizeObserver.unobserve(navRef.value);
    });

    onBeforeUnmount(() => {
      navRef.value && resizeObserver.unobserve(navRef.value);
    });

    return {
      appLogo,
      appName,
      langRef,
      navRef,
      navItemRefs,
      breakIndex,
      hiddenMenus,
      activeNav,
      needMenu,
      curProject,
      openSideMenu,
      releaseData,
      menus,
      langs,
      curLang,
      showSystemLog,
      showFeatures,
      user,
      showPreVersionBtn,
      handleGoHome,
      handleChangeLang,
      handleGotoHelp,
      handleShowSystemLog,
      handleShowFeatures,
      handleGotoUserToken,
      handleLogout,
      handleToggleClickNav,
      handleChangeMenu,
      resolveMenuLink,
      backToPreVersion,
      handleInitSliderListHeight,
    };
  },
});

</script>
<style lang="postcss" scoped>
.header-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  color: #768197;
  cursor: pointer;
  width: 32px;
  height: 32px;
  &:hover {
    background: linear-gradient(270deg,#253047,#263247);
    border-radius: 100%;
    color: #d3d9e4;
  }
}
>>> .container-content {
  padding: 0!important;
}
</style>
