<template>
  <div :class="['package-content-detail', { 'with-apps-panel': isPackage && isAppsPanelOpen }]">
    <div class="detail-container">
      <div class="header-wrapper">
        <div v-if="headerInfo.isPublic" class="tag">{{ t('公开') }}</div>
        <h4 class="package-name">{{ headerInfo.name }}</h4>
        <p class="package-desc" :title="headerInfo.memo" v-if="headerInfo.primary_name !== '默认套餐'">
          {{ headerInfo.memo }}
        </p>
      </div>
      <div class="table-list-wrapper">
        <component v-if="tableComp" :is="tableComp" />
      </div>
    </div>
    <div v-if="isPackage" class="app-list-panel">
      <div class="panel-switch-trigger" @click="isAppsPanelOpen = !isAppsPanelOpen">
        <div :class="{ 'en-text': locale === 'en' }">{{ t('使用模板') }}</div>
        <AngleDoubleRightLine v-if="isAppsPanelOpen" class="arrow-icon" />
        <AngleDoubleLeftLine v-else class="arrow-icon" />
      </div>
      <UsePackageApps :tpl-counts="templateCounts" @toggle-open="isAppsPanelOpen = $event" />
    </div>
  </div>
</template>
<script lang="ts" setup>
  import { ref, computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { storeToRefs } from 'pinia';
  import { AngleDoubleRightLine, AngleDoubleLeftLine } from 'bkui-vue/lib/icon';
  import useTemplateStore from '../../../../../store/template';
  import PACKAGE_MENU_OTHER_TYPE_MAP from '../../../../../constants/template';
  import UsePackageApps from './use-package-apps.vue';
  import ConfigInPackageTable from './tables/config-in-package.vue';
  import ConfigInAllTable from './tables/config-in-all.vue';
  import ConfigWithoutPackageTable from './tables/config-without-package.vue';

  const { packageList, currentPkg } = storeToRefs(useTemplateStore());
  const { t, locale } = useI18n();

  const isAppsPanelOpen = ref(true);
  const templateCounts = ref(0);

  // 当前选中的套餐菜单项是否为实际的套餐
  const isPackage = computed(() => typeof currentPkg.value === 'number');

  const tableComp = computed(() => {
    if (isPackage.value) {
      return ConfigInPackageTable;
    }
    if (currentPkg.value === 'all') {
      return ConfigInAllTable;
    }
    if (currentPkg.value === 'no_specified') {
      return ConfigWithoutPackageTable;
    }
    return null;
  });

  const headerInfo = computed(() => {
    let name = '';
    let memo = '';
    let isPublic = false;
    let primary_name = '';
    if (isPackage.value) {
      const pkgDetail = packageList.value.find((item) => item.id === currentPkg.value);
      if (pkgDetail) {
        name = pkgDetail.spec.name;
        memo = pkgDetail.spec.memo;
        isPublic = pkgDetail.spec.public;
        primary_name = pkgDetail.spec.name;
        templateCounts.value = pkgDetail.spec.template_ids.length;
      }
    } else {
      name = PACKAGE_MENU_OTHER_TYPE_MAP[currentPkg.value as keyof typeof PACKAGE_MENU_OTHER_TYPE_MAP];
      templateCounts.value = 0;
    }
    return { name, memo, isPublic, primary_name };
  });
</script>
<style lang="scss" scoped>
  .package-content-detail {
    position: relative;
    height: 100%;
    &.with-apps-panel {
      .detail-container {
        width: calc(100% - 240px);
      }
      .use-package-apps {
        display: block;
      }
    }
  }
  .detail-container {
    padding: 18px 24px;
    .header-wrapper {
      display: flex;
      align-items: center;
      .tag {
        flex-shrink: 0;
        margin-right: 8px;
        padding: 0 8px;
        height: 22px;
        line-height: 22px;
        font-size: 12px;
        color: #4193e5;
        background: #dae9fd;
        border-radius: 2px;
        text-align: center;
      }
      .package-name {
        flex-shrink: 0;
        margin: 0;
        line-height: 22px;
        font-size: 14px;
        font-weight: 700;
        color: #63656e;
      }
      .package-desc {
        margin: 0 0 0 16px;
        line-height: 20px;
        color: #979ba5;
        font-size: 12px;
        overflow: hidden;
        white-space: nowrap;
        text-overflow: ellipsis;
      }
    }
    .operation-wrapper {
      display: flex;
      align-items: center;
      justify-content: space-between;
      .create-config-btn {
        margin-right: 8px;
      }
      .config-search {
        width: 320px;
        background-color: #ffffff;
        .search-icon {
          margin-right: 10px;
          color: #979ba5;
        }
      }
    }
    .table-list-wrapper {
      margin-top: 16px;
    }
  }
  .app-list-panel {
    position: absolute;
    top: 0;
    right: 0;
    bottom: 0;
    .panel-switch-trigger {
      position: absolute;
      top: 16px;
      left: -20px;
      padding: 8px 0;
      width: 20px;
      line-height: 16px;
      color: #ffffff;
      font-size: 12px;
      text-align: center;
      background: #3a84ff;
      border-radius: 4px 0 0 4px;
      cursor: pointer;
      .arrow-icon {
        margin-top: 4px;
        font-size: 12px;
      }
      .en-text {
        writing-mode: vertical-lr;
      }
    }
    .use-package-apps {
      display: none;
      box-shadow: -1px 0 0 0 #dcdee5;
    }
  }
</style>
