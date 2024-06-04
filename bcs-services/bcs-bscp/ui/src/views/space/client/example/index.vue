<template>
  <section class="configuration-example-page">
    <div class="example-aside">
      <!-- 选择服务 -->
      <Selector class="sel-service" @change-service="changeService" />
      <!-- 示例列表 -->
      <div class="type-wrap">
        <bk-menu v-if="serviceType === 'file'" :active-key="fileTypeArr[0].val" @update:active-key="changeTypeItem">
          <bk-menu-item :need-icon="false" v-for="item in fileTypeArr" :key="item.val"> {{ item.name }} </bk-menu-item>
        </bk-menu>
        <bk-menu v-if="serviceType === 'kv'" :active-key="kvTypeArr[0].val" @update:active-key="changeTypeItem">
          <bk-menu-item :need-icon="false" v-for="item in kvTypeArr" :key="item.val"> {{ item.name }} </bk-menu-item>
        </bk-menu>
      </div>
    </div>
    <!-- 右侧区域 -->
    <div class="example-main">
      <bk-alert v-show="serviceType === 'file' && topTip" theme="info">
        <div class="alert-tips">
          <p>{{ topTip }}</p>
        </div>
      </bk-alert>
      <div class="content-wrap">
        <component :is="currentTemplate" />
      </div>
    </div>
  </section>
</template>

<script lang="ts" setup>
  import { computed, watch, ref } from 'vue';
  import Selector from './components/service-selector.vue';
  import BkMenu from 'bkui-vue/lib/menu';
  import { useI18n } from 'vue-i18n';
  import ContainerExample from './components/content/container-example.vue';
  import NodeManaExample from './components/content/node-mana-example.vue';
  import CmdExample from './components/content/cmd-example.vue';
  import KvExample from './components/content/kv-example.vue';
  const { t } = useI18n();
  const fileTypeArr = ref([
    { name: t('Sidecar容器'), val: 'sidecar' },
    { name: t('节点管理插件'), val: 'node' },
    { name: t('命令行工具'), val: 'file-cmd' },
  ]);
  const kvTypeArr = ref([
    { name: 'Python SDK', val: 'python' },
    { name: 'Go SDK', val: 'go' },
    { name: 'Java SDK', val: 'java' },
    { name: 'C++ SDK', val: 'c++' },
    { name: t('命令行工具'), val: 'kv-cmd' },
  ]);
  const renderComponent = ref(''); // 渲染的示例组件
  const serviceType = ref(''); // 配置类型：file/kv
  const topTip = ref('');
  // 服务切换
  const changeService = (data: string) => {
    serviceType.value = data;
    // console.log('切换服务', data);
  };
  // 服务的子类型切换
  const changeTypeItem = (data: string) => {
    // console.log('切换item', data);
    renderComponent.value = data;
  };
  // 展示的示例组件与顶部提示语
  const currentTemplate = computed(() => {
    if (serviceType.value && serviceType.value === 'file') {
      switch (renderComponent.value) {
        case 'sidecar':
          topTip.value = t('Sidecar 容器客户端用于容器化应用程序拉取文件型配置场景。');
          return ContainerExample;
        case 'node':
          topTip.value = t('节点管理插件客户端用于非容器化应用程序 (传统主机) 拉取文件型配置场景。');
          return NodeManaExample;
        case 'file-cmd':
          topTip.value = t(
            '命令行工具通常用于在脚本 (如 Bash、Python 等) 中手动拉取应用程序配置，同时支持文件型和键值型配置的获取。',
          );
          return CmdExample;
        default:
          return '';
      }
    } else {
      // 键值类型模板都一样
      return KvExample;
    }
  });
  // 服务类型变化后，重新选择渲染的示例模板
  watch(serviceType, (newV) => {
    if (newV === 'file') {
      changeTypeItem(fileTypeArr.value[0].val);
    } else {
      changeTypeItem(kvTypeArr.value[0].val);
    }
  });
</script>

<style scoped lang="scss">
  .configuration-example-page {
    display: flex;
    justify-content: flex-start;
    align-items: flex-start;
    width: 100%;
    height: 100%;
    background: #f5f7fa;
  }
  .example-aside {
    display: flex;
    flex-direction: column;
    justify-content: flex-start;
    align-items: center;
    flex-shrink: 0;
    width: 240px;
    height: 100%;
    border-right: 1px solid #f5f7fa;
    background-color: #fff;
  }
  .example-main {
    flex: 1;
    height: 100%;
  }
  .alert-tips {
    display: flex;
    > p {
      margin: 0;
      line-height: 20px;
    }
  }
  .sel-service {
    flex-shrink: 0;
    padding: 10px 8px;
    width: 239px;
    border-bottom: 1px solid #f0f1f5;
  }
  .type-wrap {
    margin-top: 12px;
    width: 100%;
    flex: 1;
    overflow-y: auto;
  }
  .bk-menu {
    width: 239px;
    background: #fff;
    .bk-menu-item {
      padding: 0 22px;
      margin: 0;
      color: #63656e;
      &.is-active {
        color: #3a84ff;
        background: #e1ecff;
        &:hover {
          color: #3a84ff;
        }
      }
      &:hover {
        color: #63656e;
      }
    }
  }
  .content-wrap {
    margin: 24px;
    padding: 24px;
    min-height: 600px;
    box-shadow: 0 2px 4px 0 #1919290d;
    background-color: #fff;
  }
</style>
