<template>
  <section class="kv-example-template">
    <form-option @option-data="getOptionData" ref="fileOptionRef" :contents-show="false" />
    <div class="preview-container">
      <div class="kv-handle-content">
        <span class="preview-label">示例预览</span>
        <div class="changeMethod">
          <div
            class="tab-wrap"
            :class="['tab-wrap', { 'is-active': activeTab === index }]"
            v-for="(item, index) in tabArr"
            :key="item.name"
            @click="handleTab(index)">
            {{ item.name }}
          </div>
        </div>
        <bk-button @click="copyExample" theme="primary" class="copy-btn">复制示例</bk-button>
        <bk-alert class="alert-tips-wrap" v-show="topTipShow" theme="info">
          <div class="alert-tips">
            <p>{{ topTip }}</p>
            <close-line class="close-line" @click="topTipShow = false" />
          </div>
        </bk-alert>
      </div>
      <code-preview class="preview-component" :code-val="replaceVal" ref="codePreviewRef" />
    </div>
  </section>
</template>

<script lang="ts" setup>
  import { ref, provide, computed, onMounted, watch } from 'vue';
  import { copyToClipBoard } from '../../../../../../utils/index';
  import { CloseLine } from 'bkui-vue/lib/icon';
  import BkMessage from 'bkui-vue/lib/message';
  import FormOption from '../form-option.vue';
  import codePreview from '../code-preview.vue';
  import { useI18n } from 'vue-i18n';
  import { useRoute } from 'vue-router';
  const props = defineProps<{
    kvName: string;
  }>();
  const codePreviewRef = ref();
  const { t } = useI18n();
  const route = useRoute();
  const fileOptionRef = ref();
  const bkBizId = ref(String(route.params.spaceId));
  const codeVal = ref('');
  const formError = ref<number>();
  provide('formError', formError);
  const tabArr = [
    {
      name: 'Get方法',
      topTip: 'Get 方法：用于一次性拉取最新的配置信息，适用于需要获取并更新配置的场景。',
    },
    {
      name: 'Watch方法',
      topTip:
        'Watch 方法：通过建立长连接，实时监听配置版本的变更，当新版本的配置发布时，将自动调用回调方法处理新的配置信息，适用于需要实时响应配置变更的场景。',
    },
  ];
  const activeTab = ref(0); // 激活tab索引
  // 代码预览上方提示
  const topTip = computed(() => {
    return tabArr[activeTab.value].topTip;
  });
  const topTipShow = ref(true);
  // fileOption组件传递过来的数据汇总
  const optionData = ref({
    clientKey: '',
    privacyCredential: '',
    labelArr: [],
  });
  const getOptionData = (data: any) => {
    optionData.value = data;
  };
  // 修改后的预览数据
  const replaceVal = computed(() => {
    const labelArr = optionData.value.labelArr.length ? JSON.stringify(optionData.value.labelArr.join(', ')) : '';
    let updateString = codeVal.value.replace('动态替换bkBizId', bkBizId.value);
    updateString = updateString.replaceAll('动态替换labels', labelArr);
    updateString = updateString.replaceAll('动态替换clientKey', optionData.value.privacyCredential);
    return updateString;
  });
  onMounted(() => {
    handleTab();
  });
  // 复制示例
  const copyExample = async () => {
    try {
      await fileOptionRef.value.formRef.validate();
      // 复制示例使用未脱敏的密钥
      const reg = /"(.{1}|.{3})\*{3}(.{1}|.{3})"/g;
      const copyVal = replaceVal.value.replaceAll(reg, `"${optionData.value.clientKey}"`);
      copyToClipBoard(copyVal);
      BkMessage({
        theme: 'success',
        message: t('示例已复制'),
      });
    } catch (error) {
      // 通知密钥选择组件校验状态
      formError.value = new Date().getTime();
      console.log(error);
    }
  };
  // 切换tab
  const handleTab = async (index = 0) => {
    // scrollTo.value = new Date().getTime(); // 通知code-preview，滚动条到最顶
    codePreviewRef.value.scrollTo();
    activeTab.value = index;
    const newKvData = await changeKvData(props.kvName, index);
    codeVal.value = newKvData.default;
  };
  // 键值型数据模板切换
  /**
   *
   * @param kvName 数据模板名称
   * @param methods 方法，0: get，1: watch
   */
  const changeKvData = (kvName = 'python', methods = 0) => {
    switch (kvName) {
      case 'python':
        return !methods
          ? import('/src/assets/exampleData/kv-python-get.yaml?raw')
          : import('/src/assets/exampleData/kv-python-watch.yaml?raw');
      case 'go':
        return !methods
          ? import('/src/assets/exampleData/kv-go-get.yaml?raw')
          : import('/src/assets/exampleData/kv-go-watch.yaml?raw');
      case 'java':
        return !methods
          ? import('/src/assets/exampleData/kv-python-get.yaml?raw')
          : import('/src/assets/exampleData/kv-python-watch.yaml?raw');
      case 'c++':
        return !methods
          ? import('/src/assets/exampleData/kv-python-get.yaml?raw')
          : import('/src/assets/exampleData/kv-python-watch.yaml?raw');
      case 'kv-cmd':
        return !methods
          ? import('/src/assets/exampleData/kv-python-get.yaml?raw')
          : import('/src/assets/exampleData/kv-python-watch.yaml?raw');
      default:
        return '';
    }
  };
  watch(
    () => props.kvName,
    () => {
      handleTab();
    },
  );
</script>

<style scoped lang="scss">
  .kv-example-template {
    display: flex;
    flex-direction: column;
    height: 100%;
    .changeMethod {
      margin: 0 16px;
      padding: 4px;
      display: inline-flex;
      background: #f0f1f5;
      border-radius: 2px;
    }
    .tab-wrap {
      padding: 0 12px;
      min-width: 72px;
      line-height: 24px;
      font-size: 12px;
      color: #63656e;
      cursor: pointer;
      transition: 0.3s;
      &.is-active {
        color: #3a84ff;
        background-color: #fff;
      }
    }
  }
  .preview-container {
    margin-top: 32px;
    padding: 8px 0;
    display: flex;
    flex-direction: column;
    flex: 1;
    height: 100%;
    border-top: 1px solid #dcdee5;
    // overflow: hidden;
  }
  .kv-handle-content {
    flex-shrink: 0;
  }
  .preview-component {
    margin-top: 8px;
    padding: 16px 0;
    flex: 1;
    height: 100%;
    background-color: #f5f7fa;
  }
  .alert-tips-wrap {
    margin-top: 8px;
    .close-line {
      margin-left: auto;
      cursor: pointer;
    }
  }
  .alert-tips {
    display: flex;
    > p {
      margin: 0;
      line-height: 20px;
    }
  }
</style>
