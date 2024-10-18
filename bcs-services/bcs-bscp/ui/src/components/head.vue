<template>
  <div class="header">
    <div class="head-left">
      <div class="title-wrap" @click="router.push({ name: 'service-all', params: { spaceId } })">
        <span class="logo">
          <img :src="appGlobalConfig.appLogo || logo" alt="BSCP" />
        </span>
        <span class="head-title"> {{ appGlobalConfig.i18n.name }} </span>
      </div>
      <div class="head-routes">
        <div v-for="nav in navList" :key="nav.id" :class="['nav-item', { actived: isFirstNavActived(nav.module) }]">
          <div v-if="nav.children">
            <div :class="['firstNav-item', { actived: isFirstNavActived(nav.module) }]">{{ nav.name }}</div>
            <div class="secondNav-list">
              <div
                v-for="secondNav in nav.children"
                :key="secondNav.id"
                :class="['secondNav-item', { actived: isSecondNavActived(secondNav.module) }]">
                <a @click.stop="handleNavClick(secondNav.id)">
                  {{ secondNav.name }}
                </a>
              </div>
            </div>
          </div>
          <a v-else @click.stop="handleNavClick(nav.id)">
            {{ nav.name }}
          </a>
        </div>
      </div>
    </div>
    <div class="head-right">
      <bk-select
        class="space-selector"
        id-key="space_id"
        display-key="space_name"
        enable-virtual-render
        :model-value="spaceId"
        :popover-options="{ theme: 'light bk-select-popover space-selector-popover' }"
        :list="optionList"
        :filterable="true"
        :clearable="false"
        :input-search="false"
        :remote-method="handleSpaceSearch"
        @change="handleSelectSpace">
        <template #trigger>
          <div class="space-name">
            <input readonly :value="crtSpaceText" />
            <AngleDown class="arrow-icon" />
          </div>
        </template>
        <template #extension>
          <div class="create-operation" @click="handleToCMDB">
            <plus />
            <div class="content">{{ t('新建业务') }}</div>
          </div>
        </template>
        <template #virtualScrollRender="{ item }">
          <div
            v-cursor="{ active: !item.permission }"
            :class="['biz-option-item', { 'no-perm': !item.permission }]"
            v-bk-tooltips="{
              content: `${t('业务名')}: ${item.space_name}\n${t('业务')}ID: ${item.space_id}`,
              placement: 'left',
            }">
            <div class="name-wrapper">
              <span class="text">{{ item.space_name }}</span>
              <span class="id">({{ item.space_id }})</span>
            </div>
            <span class="tag">{{ locale === 'zh-cn' ? item.space_type_name : item.space_en_name }}</span>
          </div>
        </template>
      </bk-select>
      <bk-popover ext-cls="login-out-popover" trigger="hover" placement="bottom-center" theme="light" :arrow="false">
        <div class="international">
          <span :class="['bk-bscp-icon', locale === 'zh-cn' ? 'icon-lang-cn' : 'icon-lang-en']"></span>
        </div>
        <template #content>
          <div class="international-item" @click="switchLanguage('zh-cn')">
            <span class="bk-bscp-icon icon-lang-cn"></span> 中文
          </div>
          <div class="international-item" @click="switchLanguage('en')">
            <span class="bk-bscp-icon icon-lang-en"></span> English
          </div>
        </template>
      </bk-popover>
      <bk-dropdown trigger="hover" ext-cls="dropdown" :is-show="isShowDropdown" @hide="isShowDropdown = false">
        <bk-button text :class="['dropdown-trigger', isShowDropdown ? 'active' : '']">
          <help-document-fill width="16" height="16" :fill="isShowDropdown ? '#fff' : '#96a2b9'" />
        </bk-button>
        <template #content>
          <bk-dropdown-menu ext-cls="dropdown-menu">
            <bk-dropdown-item
              v-for="item in dropdownList"
              :key="item.title"
              ext-cls="dropdown-item"
              @click="item.click">
              {{ item.title }}
            </bk-dropdown-item>
          </bk-dropdown-menu>
        </template>
      </bk-dropdown>
      <bk-popover ext-cls="login-out-popover" trigger="click" placement="bottom-center" theme="light" :arrow="false">
        <div class="username-wrapper">
          <span class="text">{{ userInfo.username }}</span>
          <DownShape class="arrow-icon" />
        </div>
        <template #content>
          <div class="login-out-btn" @click="handleLoginOut">{{ t('退出登录') }}</div>
        </template>
      </bk-popover>
    </div>
  </div>
  <version-log :log-list="logList" v-model:is-show="isShowVersionLog"></version-log>
  <features :detail="featuresContent" v-model:is-show="isShowFeatures"></features>
</template>

<script setup lang="ts">
  import { ref, computed, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { useRoute, useRouter, RouteRecordName } from 'vue-router';
  import { storeToRefs } from 'pinia';
  import { AngleDown, HelpDocumentFill, DownShape, Plus } from 'bkui-vue/lib/icon';
  import useGlobalStore from '../store/global';
  import useUserStore from '../store/user';
  import useTemplateStore from '../store/template';
  import { ISpaceDetail } from '../../types/index';
  import { loginOut } from '../api/index';
  import logo from '../assets/logo.svg';
  import type { IVersionLogItem } from '../../types/version-log';
  import VersionLog from './version-log.vue';
  import features from './features-dialog.vue';
  import MarkdownIt from 'markdown-it';
  import { setCookie } from '../utils';

  const route = useRoute();
  const router = useRouter();
  const { t, locale } = useI18n();
  const {
    bscpVersion,
    spaceId,
    spaceList,
    spaceFeatureFlags,
    showPermApplyPage,
    showApplyPermDialog,
    permissionQuery,
    appGlobalConfig,
  } = storeToRefs(useGlobalStore());
  const { userInfo } = storeToRefs(useUserStore());
  const templateStore = useTemplateStore();
  const md = new MarkdownIt({
    html: true,
    linkify: true,
    typographer: true,
  });
  const navList = computed(() => [
    { id: 'service-all', module: 'service', name: t('服务管理') },
    { id: 'groups-management', module: 'groups', name: t('分组管理') },
    { id: 'script-list', module: 'scripts', name: t('脚本管理') },
    {
      id: 'templates-and-variables',
      module: 'templates-and-variables',
      name: t('模板与变量'),
      children: [
        { id: 'templates-list', module: 'templates', name: t('模板管理') },
        { id: 'variables-management', module: 'variables', name: t('变量管理') },
      ],
    },

    {
      id: 'client-manage',
      module: 'client',
      name: t('客户端管理'),
      children: [
        { id: 'client-statistics', module: 'client-statistics', name: t('客户端统计') },
        { id: 'client-search', module: 'client-search', name: t('客户端查询') },
        { id: 'credentials-management', module: 'credentials', name: t('客户端密钥') },
        { id: 'configuration-example', module: 'example', name: t('配置示例') },
      ],
    },
    { id: 'records-all', module: 'records', name: t('操作记录') },
  ]);

  const optionList = ref<ISpaceDetail[]>([]);

  const crtSpaceText = computed(() => {
    const space = spaceList.value.find((item) => item.space_id === spaceId.value);
    if (space) {
      return `${space.space_name}(${spaceId.value})`;
    }
    return '';
  });

  watch(
    spaceList,
    (val) => {
      optionList.value = val.slice();
    },
    {
      immediate: true,
    },
  );

  const isFirstNavActived = (name: string) => {
    const nav = navList.value.find((item) => item.module === name);
    if (nav!.children) {
      return (
        spaceFeatureFlags.value.BIZ_VIEW &&
        !showPermApplyPage.value &&
        nav!.children.find((item) => item.module === route.meta.navModule)
      );
    }
    return spaceFeatureFlags.value.BIZ_VIEW && !showPermApplyPage.value && route.meta.navModule === name;
  };

  const isSecondNavActived = (secondNavName: string) => {
    return spaceFeatureFlags.value.BIZ_VIEW && !showPermApplyPage.value && route.meta.navModule === secondNavName;
  };

  const handleNavClick = (navId: string) => {
    if (['service-all', 'client-statistics', 'client-search', 'configuration-example'].includes(navId)) {
      const lastAccessedServiceDetail = localStorage.getItem('lastAccessedServiceDetail');
      if (lastAccessedServiceDetail) {
        const detail = JSON.parse(lastAccessedServiceDetail);
        if (detail.spaceId === spaceId.value) {
          router.push({
            name: navId === 'service-all' && !showPermApplyPage.value ? 'service-config' : (navId as RouteRecordName),
            params: { spaceId: detail.spaceId, appId: detail.appId },
          });
          return;
        }
      }
    }
    router.push({ name: navId, params: { spaceId: spaceId.value || 0 } });
  };

  const handleSpaceSearch = (searchStr: string) => {
    if (searchStr) {
      optionList.value = spaceList.value.filter((item) => {
        const spaceName = item.space_name.toLowerCase();
        return spaceName.includes(searchStr.toLowerCase()) || String(item.space_id).includes(searchStr);
      });
    } else {
      optionList.value = spaceList.value.slice();
    }
  };

  const handleSelectSpace = async (id: string) => {
    const space = spaceList.value.find((item: ISpaceDetail) => item.space_id === id);
    if (space) {
      if (!space.permission) {
        permissionQuery.value = {
          resources: [
            {
              biz_id: id,
              basic: {
                type: 'biz',
                action: 'find_business_resource',
                resource_id: id,
              },
            },
          ],
        };

        showApplyPermDialog.value = true;
        return;
      }
      templateStore.$patch((state) => {
        state.templateSpaceList = [];
        state.currentTemplateSpace = 0;
        state.currentPkg = '';
      });
      const nav = navList.value.find((item) => item.module === route.meta.navModule);
      if (nav) {
        router.push({ name: nav.id, params: { spaceId: id } });
      } else {
        router.push({ name: 'service-mine', params: { spaceId: id } });
      }
    }
  };

  // 下拉菜单
  const dropdownList = computed(() => [
    {
      title: t('产品文档'),
      click() {
        // @ts-ignore
        // eslint-disable-next-line
        window.open(BSCP_CONFIG.help);
      },
    },
    {
      title: t('版本日志'),
      click() {
        isShowVersionLog.value = true;
      },
    },
    {
      title: t('功能特性'),
      click() {
        isShowFeatures.value = true;
      },
    },
    {
      title: t('问题反馈'),
      click() {
        window.open('https://bk.tencent.com/s-mart/community/question');
      },
    },
  ]);

  const isShowDropdown = ref(false);

  // 版本日志
  const logList = ref<IVersionLogItem[]>([]);
  const isShowVersionLog = ref(false);
  // @ts-ignore
  // const modules = import.meta.glob('../../../docs/changelog/zh_CN/*.md', {
  //   as: 'raw',
  //   eager: true,
  // });
  // const modules = import.meta.glob('../../../docs/changelog/en_US/*.md', {
  //   as: 'raw',
  //   eager: true,
  // });

  const logModules = computed(() => {
    if (locale.value === 'zh-cn') {
      // @ts-ignore
      return import.meta.glob('../../../docs/changelog/zh_CN/*.md', {
        as: 'raw',
        eager: true,
      });
    }
    // @ts-ignore
    return import.meta.glob('../../../docs/changelog/en_US/*.md', {
      as: 'raw',
      eager: true,
    });
  });
  Object.keys(logModules.value).forEach((path) => {
    const separator = locale.value === 'zh-cn' ? 'CN/' : 'US/';
    logList.value!.push({
      title: path.split(separator)[1].split('_')[0],
      date: path.split(separator)[1].split('_')[1].slice(0, -3),
      // @ts-ignore
      detail: md.render(logModules.value[path]),
    });
  });

  if (logList.value.length > 0) {
    bscpVersion.value = logList.value[0].title;
  }

  // 功能特性
  const featuresContent = ref('');
  const isShowFeatures = ref(false);
  // @ts-ignore
  const featuresModule = computed(() => {
    if (locale.value === 'zh-cn') {
      // @ts-ignore
      return import.meta.glob('../../../docs/features/features.md', {
        as: 'raw',
        eager: true,
      });
    }
    // @ts-ignore
    return import.meta.glob('../../../docs/features/features_en.md', {
      as: 'raw',
      eager: true,
    });
  });
  Object.keys(featuresModule.value).forEach((path) => {
    featuresContent.value = md.render(featuresModule.value[path]);
  });
  const handleLoginOut = () => {
    loginOut();
  };
  const handleToCMDB = () => {
    // @ts-ignore
    window.open(`${BK_CC_HOST}/#/resource/business`); // eslint-disable-line no-undef
  };

  // 切换语言
  const switchLanguage = (language: string) => {
    const domain = window.location.hostname.replace(/^[^.]+(.*)$/, '$1');
    setCookie('blueking_language', language, domain);
    locale.value = language;
    location.reload();
  };
</script>

<style lang="scss" scoped>
  .header {
    height: 52px;
    background: #182132;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 20px;
    .head-left {
      display: flex;
      align-items: center;
      .logo {
        display: inline-flex;
        width: 30px;
        height: 30px;
      }
      .head-routes {
        display: flex;
        padding-left: 90px;
        font-size: 14px;
        .nav-item {
          position: relative;
          display: flex;
          align-items: center;
          height: 52px;
          padding: 0 16px;
          font-size: 14px;
          color: #96a2b9;
          cursor: pointer;
          a {
            color: #96a2b9;
          }
          &:hover {
            color: #c2cee5;
            a {
              color: #c2cee5;
            }
            .secondNav-list {
              display: block;
            }
          }
          &.actived {
            color: #ffffff;
            a {
              color: #ffffff;
            }
          }
          .firstNav-item {
            height: 100%;
            display: flex;
            align-items: center;
            cursor: default;
          }
          .secondNav-list {
            display: none;
            position: absolute;
            top: 52px;
            left: 0;
            z-index: 9999;
            background: #182132;
            border-radius: 0 0 2px 2px;
            padding: 4px 1px;
            .secondNav-item {
              min-width: 102px;
              height: 40px;
              line-height: 40px;
              padding: 0 16px;
              font-size: 14px;
              white-space: nowrap;
              cursor: pointer;
              a {
                color: #96a2b9;
              }
              &:hover {
                color: #c2cee5;
                a {
                  color: #c2cee5;
                }
              }
              &.actived {
                background: #2f3746;
                a {
                  color: #fff;
                }
              }
            }
          }
        }
      }

      .head-title {
        display: inline-flex;
        padding-left: 20px;
        font-weight: Bold;
        font-size: 18px;
        color: #96a2b9;
      }
    }

    .head-right {
      display: flex;
      align-items: center;
      justify-self: flex-end;
      justify-content: space-between;
      font-size: 14px;
      color: #979ba5;
    }
  }
  .space-selector {
    margin-right: 24px;
    width: 240px;
    &.popover-show {
      .space-name .arrow-icon {
        transform: rotate(-180deg);
      }
    }
    .space-name {
      position: relative;
      input {
        padding: 0 24px 0 10px;
        width: 100%;
        line-height: 32px;
        font-size: 12px;
        border: none;
        outline: none;
        background: #303d55;
        border-radius: 2px;
        color: #d3d9e4;
        cursor: pointer;
      }
      .arrow-icon {
        position: absolute;
        top: 0;
        right: 4px;
        height: 100%;
        font-size: 20px;
        color: #979ba5;
        transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
      }
    }
  }
  .biz-option-item {
    position: relative;
    padding: 0 12px;
    width: 100%;
    &.no-perm {
      background-color: #fafafa !important;
      color: #cccccc !important;
    }
    .name-wrapper {
      padding-right: 30px;
      display: flex;
      align-items: center;
      .text {
        flex: 0 1 auto;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }
      .id {
        flex: 0 0 auto;
        margin-left: 4px;
        color: #979ba5;
      }
    }
    .tag {
      position: absolute;
      top: 8px;
      right: 4px;
      padding: 2px;
      font-size: 12px;
      line-height: 1;
      color: #cccccc;
      border: 1px solid #cccccc;
      border-radius: 2px;
      transform: scale(0.8);
    }
  }
  .username-wrapper {
    display: flex;
    align-items: center;
    font-size: 12px;
    color: #c4c4cc;
    cursor: pointer;
    &:hover {
      color: #3a84ff;
    }
    .arrow-icon {
      font-size: 14px;
      margin-left: 4px;
    }
  }
  .login-out-btn {
    padding: 0 16px;
    height: 32px;
    line-height: 32px;
    font-size: 12px;
    cursor: pointer;
    &:hover {
      background-color: #eaf3ff;
      color: #3a84ff;
    }
  }
  .create-operation {
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    height: 33px;
    padding: 0 12px;
    width: 100%;
    color: #2b353e;
    span {
      font-size: 16px;
      margin-right: 3px;
    }
  }
  .title-wrap {
    display: flex;
    align-items: center;
    cursor: pointer;
  }
  .international {
    width: 28px;
    height: 28px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 20px;
    cursor: pointer;
    &:hover {
      background-color: rgba(150, 162, 185, 0.3);
      span {
        color: #ffffff;
      }
    }
  }
  .international-item {
    height: 32px;
    line-height: 32px;
    padding: 0 16px;
    cursor: pointer;
    span {
      font-size: 20px;
    }
    &:hover {
      background-color: #eaf3ff;
      color: #3a84ff;
    }
  }
</style>
<style lang="scss">
  .space-selector-popover .bk-select-option {
    padding: 0 !important;
  }
  .dropdown {
    display: flex;
    align-items: center;
    justify-content: center;
    margin-right: 10px;
    width: 40px;
    height: 40px;
    .dropdown-trigger {
      width: 28px;
      height: 28px;
      border-radius: 50%;
      &:hover {
        background-color: rgba($color: #96a2b9, $alpha: 0.3);
        .bk-icon {
          fill: #fff !important;
        }
      }
    }
    .active {
      background-color: rgba($color: #96a2b9, $alpha: 0.3);
    }
  }
  .dropdown-menu .dropdown-item:hover {
    background-color: #f0f1f5;
    color: #3a84ff;
  }
  .version-dialog {
    .bk-dialog-header {
      display: none;
    }
    .bk-modal-content {
      padding: 0 !important;
    }
  }
  .login-out-popover.bk-popover.bk-pop2-content {
    padding: 4px 0;
  }
</style>
