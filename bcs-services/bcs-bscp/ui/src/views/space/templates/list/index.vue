<script lang="ts" setup>
  import SpaceSelector from './components/space/selector.vue';
  import PackageList from './components/package-list/list.vue';
  import PackageDetail from './components/package-detail/index.vue';
</script>
<template>
  <div class="templates-page">
    <bk-alert class="template-tips" theme="info">
      <div class="tips-wrapper">
        <div class="message">配置模板用于统一业务下服务间配置文件复用，可以在创建服务配置时引用配置模板。模板变量使用方法请参考</div>
        <bk-button text theme="primary">go template</bk-button>
      </div>
    </bk-alert>
    <div class="main-content-container">
      <div class="side-menu-area">
        <space-selector></space-selector>
        <package-list></package-list>
      </div>
      <div class="package-detail-area">
        <package-detail></package-detail>
      </div>
    </div>
  </div>
</template>
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
  }
  .side-menu-area {
    padding: 16px 0;
    width: 240px;
    height: 100%;
    background: #ffffff;
    box-shadow: 1px 0 0 0 #dcdee5;
  }
  .package-detail-area {
    width: calc(100% - 240px);
    padding: 18px 24px;
    height: 100%;
    background: #f5f7fa;
  }
</style>
