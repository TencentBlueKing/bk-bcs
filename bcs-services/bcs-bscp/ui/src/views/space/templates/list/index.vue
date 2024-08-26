<template>
  <div class="templates-page">
    <bk-alert class="template-tips" theme="info">
      <div class="tips-wrapper">
        <div class="message">
          {{ t('配置模板用于同一业务下服务间配置文件复用，可以在创建服务配置时引用配置模板。') }}
        </div>
        <!-- <bk-button text theme="primary">go template</bk-button> -->
      </div>
    </bk-alert>
    <bk-resize-layout class="main-content-container" :min="240" :initial-divide="240" :max="480">
      <template #aside>
        <div class="side-menu-area">
          <space-selector></space-selector>
          <package-menu></package-menu>
        </div>
      </template>
      <template #main>
        <div class="package-detail-area">
          <package-detail></package-detail>
        </div>
      </template>
    </bk-resize-layout>
  </div>
</template>
<script lang="ts" setup>
  import { useRoute } from 'vue-router';
  import { useI18n } from 'vue-i18n';
  import useTemplateStore from '../../../../store/template';
  import SpaceSelector from './space/selector.vue';
  import PackageMenu from './package-menu/menu.vue';
  import PackageDetail from './package-detail/index.vue';
  import { onMounted } from 'vue';

  const route = useRoute();
  const templateStore = useTemplateStore();
  const { t } = useI18n();

  onMounted(() => {
    const { templateSpaceId, packageId } = route.params;
    templateStore.$patch((state) => {
      if (templateSpaceId) {
        state.currentTemplateSpace = Number(templateSpaceId);
      }
      if (packageId) {
        state.currentPkg = /\d+/.test(packageId as string) ? Number(packageId) : (packageId as string);
      }
    });
  });
</script>
<style lang="scss" scoped>
  .templates-page {
    height: 100%;
  }
  .template-tips {
    height: 38px;
  }
  .tips-wrapper {
    display: flex;
    align-items: center;
    .message {
      line-height: 20px;
    }
    .bk-button {
      margin-left: 8px;
    }
  }
  .main-content-container {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    height: calc(100% - 38px);
    border: none;
  }
  .side-menu-area {
    padding: 16px 0;
    height: 100%;
    background: #ffffff;
  }
  .package-detail-area {
    height: 100%;
    background: #f5f7fa;
  }
</style>

<style>
  .main-content-container > .bk-resize-layout-aside {
    height: 100%;
  }
  .main-content-container > .bk-resize-layout-main {
    height: 100%;
    overflow: auto;
  }
</style>
