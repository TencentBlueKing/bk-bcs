<!-- eslint-disable @typescript-eslint/semi -->
<template>
  <div>
    <bcs-navigation
      navigation-type="top-bottom"
      :need-menu="needMenu"
      :default-open="openSideMenu"
      :hover-enter-delay="300"
      @toggle-click="handleToggleClickNav">
      <template #side-header>
        <span class="title-icon"><img src="@/images/bcs.svg" class="w-[28px] h-[28px]"></span>
        <span
          class="title-desc cursor-pointer"
          @click="handleGoHome">
          {{ $INTERNAL ? $t('TKEx-IEG 容器平台') : $t('蓝鲸容器管理平台') }}
        </span>
      </template>
      <template #header>
        <!-- 顶部菜单 -->
        <ol class="flex text-[14px] text-[#96a2b9]">
          <li
            v-for="(item, index) in menus"
            :class="[
              'mr-[40px] hover:text-[#d3d9e4] cursor-pointer',
              { 'text-[#d3d9e4]': activeNav.id === item.id }
            ]"
            :key="index"
            @click="handleChangeMenu(item)">
            {{ item.title }}
          </li>
        </ol>
        <!-- 项目选载 -->
        <ProjectSelector class="ml-auto w-[240px] mr-[18px]"></ProjectSelector>
        <!-- 语言切换 -->
        <PopoverSelector class="mr-[8px]">
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
              <li class="bcs-dropdown-item" @click="handleGotoHelp">{{ $t('产品文档') }}</li>
              <li class="bcs-dropdown-item" @click="handleShowSystemLog">{{ $t('版本日志') }}</li>
              <li class="bcs-dropdown-item" @click="handleShowFeatures">{{ $t('功能特性') }}</li>
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
              <li class="bcs-dropdown-item" @click="handleGotoUserToken">{{ $t('个人密钥') }}</li>
              <li class="bcs-dropdown-item" @click="handleLogout">{{ $t('退出登录') }}</li>
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
      :title="$t('产品功能特性')"
      :show-footer="false"
      width="480">
      <BcsMd :code="releaseData.feature.content" />
    </bcs-dialog>
  </div>
</template>
<script lang="ts">
import { defineComponent, ref, computed, toRef, reactive, onMounted } from 'vue';
import SystemLog from '@/views/app/log.vue';
import BcsMd from '@/components/bcs-md/index.vue';
import ProjectSelector from '@/views/app/project-selector.vue';
import PopoverSelector from '../../components/popover-selector.vue';
import $store from '@/store';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import menusData, { IMenu } from './menus';
import { releaseNode } from '@/api/modules/project';
import { setCookie } from '@/common/util';

export default defineComponent({
  name: 'NewNavigation',
  components: {
    SystemLog,
    BcsMd,
    ProjectSelector,
    PopoverSelector,
  },
  setup() {
    const menus = ref<IMenu[]>(menusData);
    const langs = ref([
      {
        icon: 'bk-icon icon-english',
        name: 'English',
        id: 'en-US',
      },
      {
        icon: 'bk-icon icon-chinese',
        name: '中文',
        id: 'zh-CN',
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
    // path上有项目Code且项目开启了容器服务时才显示左侧菜单
    const needMenu = computed(() => {
      const { projectCode } = route.value.params;
      return !!projectCode && route.value.fullPath.indexOf(projectCode) > -1 && !!curProject.value.kind;
    });

    // 当前导航
    const findCurrentNav = (menus: IMenu[]) => menus.find((item) => {
      if (item.children?.length) return findCurrentNav(item.children);

      return item.route === route.value.name || item.id === route.value.meta?.menuId;
    });

    // 切换菜单
    const handleChangeMenu = (item: IMenu) => {
      const name = item.route || item.children?.[0]?.route || '404';
      if (route.value.name === name) return;

      $router.push({
        name,
        params: {
          projectCode: $store.getters.curProjectCode,
          clusterId: $store.getters.curClusterId,
        },
      });
    };

    // 左侧菜单折叠和收起
    const openSideMenu = computed(() => $store.state.openSideMenu);
    const handleToggleClickNav = (value) => {
      $store.commit('updateOpenSideMenu', !!value);
    };
    // 首页
    const handleGoHome = () => {
      $router.push({ name: 'home' });
    };

    // 切换语言
    const handleChangeLang = (item) => {
      $i18n.locale = item.id;
      setCookie('blueking_language', item.id);
      window.location.reload();
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
      window.location.href = `${window.LOGIN_FULL}?c_url=${window.location}`;
    };

    // release信息
    const releaseData = ref({
      changelog: [],
      feature: { content: '' },
    });
    onMounted(async () => {
      releaseData.value = await releaseNode().catch(() => ({ changelog: [], feature: {} }));
    });

    return {
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
      handleGoHome,
      handleChangeLang,
      handleGotoHelp,
      handleShowSystemLog,
      handleShowFeatures,
      handleGotoUserToken,
      handleLogout,
      handleToggleClickNav,
      handleChangeMenu,
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
