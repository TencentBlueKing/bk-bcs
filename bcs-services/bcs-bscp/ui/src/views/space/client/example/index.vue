<template>
  <section class="configuration-example-page">
    <div class="example-aside">
      <!-- 选择服务 -->
      <service-selector class="sel-service" @select-service="selectService" />
      <!-- 示例列表 -->
      <div class="type-wrap" v-show="serviceName && serviceType">
        <bk-menu :active-key="renderComponent" @update:active-key="changeTypeItem">
          <bk-menu-item :need-icon="false" v-for="item in navList" :key="item.val"> {{ item.name }} </bk-menu-item>
        </bk-menu>
      </div>
    </div>
    <!-- 右侧区域 -->
    <div class="example-main" ref="exampleMainRef">
      <bk-alert v-show="(serviceType === 'file' || renderComponent === 'shell') && topTip" theme="info">
        <div class="alert-tips">
          <p>{{ topTip }}</p>
        </div>
      </bk-alert>
      <div class="content-wrap">
        <bk-loading style="height: 100%" :loading="loading">
          <component
            :is="currentTemplate"
            :kv-name="renderComponent"
            :content-scroll-top="contentScrollTop"
            :selected-key-data="selectedClientKey"
            :key="renderComponent"
            @selected-key-data="selectedClientKey = $event" />
        </bk-loading>
      </div>
    </div>
  </section>
</template>

<script lang="ts" setup>
  import { computed, ref, nextTick, provide } from 'vue';
  import ServiceSelector from './components/service-selector.vue';
  import { useI18n } from 'vue-i18n';
  import ContainerExample from './components/content/container-example.vue';
  import NodeManaExample from './components/content/node-mana-example.vue';
  import CmdExample from './components/content/cmd-example.vue';
  import KvExample from './components/content/kv-example.vue';
  import Exception from '../components/exception.vue';

  const { t } = useI18n();
  const fileTypeArr = [
    { name: t('Sidecar容器'), val: 'sidecar' },
    { name: t('节点管理插件'), val: 'node' },
    { name: t('HTTP(S)接口调用'), val: 'http' }, // 文件型也有http(s)接口，页面结构和键值型一样，但脚本内容、部分文案不一样
    { name: t('命令行工具'), val: 'shell' },
  ];
  const kvTypeArr = [
    { name: 'Python SDK', val: 'python' },
    { name: 'Go SDK', val: 'go' },
    { name: 'Java SDK', val: 'java' },
    { name: 'C++ SDK', val: 'cpp' },
    { name: t('HTTP(S)接口调用'), val: 'http' },
    { name: t('命令行工具'), val: 'shell' },
  ];

  const selectedClientKey = ref(); // 记忆选择的客户端密钥信息,用于切换不同示例时默认选中密钥
  const exampleMainRef = ref();
  const renderComponent = ref(''); // 渲染的示例组件
  const serviceName = ref('');
  const serviceType = ref('');
  const topTip = ref('');
  const loading = ref(true);
  provide('basicInfo', { serviceName, serviceType });

  const navList = computed(() => (serviceType.value === 'file' ? fileTypeArr : kvTypeArr));
  // 展示的示例组件与顶部提示语
  const currentTemplate = computed(() => {
    if (serviceType.value && !loading.value) {
      switch (renderComponent.value) {
        case 'sidecar':
          topTip.value = t('Sidecar 容器客户端用于容器化应用程序拉取文件型配置场景。');
          return ContainerExample;
        case 'node':
          topTip.value = t('节点管理插件客户端用于非容器化应用程序 (传统主机) 拉取文件型配置场景。');
          return NodeManaExample;
        case 'shell':
          topTip.value = t(
            '命令行工具通常用于在脚本 (如 Bash、Python 等) 中手动拉取应用程序配置，同时支持文件型和键值型配置的获取。',
          );
          return CmdExample;
        default:
          // 键值类型模板都一样
          return KvExample;
      }
    }
    // 无数据模板
    return Exception;
  });

  // 服务切换
  const selectService = (serviceTypeVal: string, serviceNameVal: string) => {
    // 重置已选择的密钥信息
    selectedClientKey.value = null;
    if (!serviceTypeVal || !serviceNameVal) {
      return (loading.value = false);
    }
    if (serviceName.value !== serviceNameVal || serviceType.value !== serviceTypeVal) {
      loading.value = true;
      serviceName.value = serviceNameVal;
      serviceType.value = serviceTypeVal;
    }
    changeTypeItem(navList.value[0].val);
  };
  // 服务的子类型切换
  const changeTypeItem = (data: string) => {
    renderComponent.value = data;
    nextTick(() => {
      loading.value = false;
    });
  };
  // 返回顶部
  const contentScrollTop = () => {
    if (exampleMainRef.value.scrollTop > 64) {
      exampleMainRef.value.scrollTo({ top: 0, behavior: 'smooth' });
    }
  };
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
    border-right: 1px solid #dcdee5;
    background-color: #fff;
  }
  .example-main {
    flex: 1;
    height: 100%;
    overflow-y: auto;
    :deep(.bk-alert-wraper) {
      align-items: center;
    }
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
    box-shadow: 0 2px 4px 0 #1919290d;
    overflow: auto;
    background-color: #fff;
    flex: 1;
    min-height: 1px;
  }
</style>
