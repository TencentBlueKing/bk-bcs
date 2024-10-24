<template>
  <section class="kv-example-template">
    <form-option
      ref="fileOptionRef"
      :directory-show="false"
      :config-show="props.kvName === 'http' || (props.kvName === 'python' && activeTab === 0)"
      :config-label="basicInfo?.serviceType.value === 'file' ? '配置文件名' : '配置项名称'"
      :selected-key-data="props.selectedKeyData"
      @update-option-data="(data) => getOptionData(data)"
      @selected-key-data="emits('selected-key-data', $event)" />
    <div class="preview-container">
      <div class="kv-handle-content">
        <span class="preview-label">{{ $t('示例预览') }}</span>
        <div class="change-method">
          <div
            :class="['tab-wrap', { 'is-active': activeTab === index }]"
            v-for="(item, index) in tabArr"
            :key="item"
            @click="handleTab(index)">
            {{ item }}
          </div>
        </div>
        <bk-button theme="primary" class="copy-btn" @click="copyExample">{{ $t('复制示例') }}</bk-button>
        <bk-alert class="alert-tips-wrap" v-show="topTipShow && kvConfig.topTip" theme="info">
          <div class="alert-tips">
            <span v-html="kvConfig.topTip"></span>
            <close-line class="close-line" @click="topTipShow = false" />
          </div>
        </bk-alert>
      </div>
      <code-preview
        class="preview-component"
        :style="{ height: `${kvConfig.codePreviewHeight}px` }"
        :code-val="replaceVal"
        :variables="variables"
        :language="codeLanguage"
        @change="(val: string) => (copyReplaceVal = val)" />
    </div>
  </section>
</template>

<script lang="ts" setup>
  import { ref, Ref, inject, computed, watch, nextTick } from 'vue';
  import { copyToClipBoard } from '../../../../../../utils/index';
  import { CloseLine } from 'bkui-vue/lib/icon';
  import { IVariableEditParams } from '../../../../../../../types/variable';
  import { newICredentialItem } from '../../../../../../../types/client';
  import BkMessage from 'bkui-vue/lib/message';
  import FormOption from '../form-option.vue';
  import codePreview from '../code-preview.vue';
  import { useI18n } from 'vue-i18n';
  import { useRoute } from 'vue-router';

  const props = defineProps<{
    kvName: string;
    selectedKeyData: newICredentialItem['spec'] | null;
  }>();

  const emits = defineEmits(['selected-key-data']);

  const basicInfo = inject<{ serviceName: Ref<string>; serviceType: Ref<string> }>('basicInfo');
  const { t } = useI18n();
  const route = useRoute();

  const tabArr = ref([t('Get方法'), t('Watch方法')]);
  const fileOptionRef = ref();
  const bkBizId = ref(String(route.params.spaceId));
  const codeVal = ref(''); // 存储yaml字符原始值
  const replaceVal = ref(''); // 替换后的值
  const copyReplaceVal = ref(''); // 渲染的值，用于复制未脱敏密钥的yaml数据
  const variables = ref<IVariableEditParams[]>();
  const activeTab = ref(0); // 激活tab索引
  const topTipShow = ref(true);
  // fileOption组件传递过来的数据汇总
  const optionData = ref({
    clientKey: '',
    privacyCredential: '',
    labelArr: [],
    labelArrType: '', // 展示格式
    configName: '', // 配置项
  });

  // 代码预览上方提示框
  const kvConfig = computed(() => {
    // @ts-ignore
    // eslint-disable-next-line
    const url = (typeof BSCP_CONFIG !== 'undefined' && BSCP_CONFIG.python_sdk_dependency_doc) || '';
    switch (props.kvName) {
      case 'python':
        // get
        if (!activeTab.value) {
          return {
            topTip: `${t('用于主动获取配置项值的场景，此方法不会监听服务器端的配置更改，有关Python SDK的部署环境和依赖组件，请参阅白皮书中的')} <a href="${url}" target="_blank">${t('BSCP Python SDK依赖说明')}</a>`,
            codePreviewHeight: 336,
          };
        }
        // watch
        return {
          topTip: `${t('通过建立长连接，实时监听配置版本的变更，当新版本的配置发布时，将自动调用回调方法处理新的配置信息，适用于需要实时响应配置变更的场景，有关Python SDK的部署环境和依赖组件，请参阅白皮书中的')} <a href="${url}" target="_blank">${t('BSCP Python SDK依赖说明')}</a>`,
          codePreviewHeight: 640,
        };
      case 'go':
        if (!activeTab.value) {
          return {
            topTip: t('Get方法：用于一次性拉取最新的配置信息，适用于需要获取并更新配置的场景。'),
            codePreviewHeight: 1002,
          };
        }
        return {
          topTip: t(
            'Watch方法：通过建立长连接，实时监听配置版本的变更，当新版本的配置发布时，将自动调用回调方法处理新的配置信息，适用于需要实时响应配置变更的场景。',
          ),
          codePreviewHeight: 1250,
        };
      case 'java':
        if (!activeTab.value) {
          return {
            topTip: t('Get方法：用于一次性拉取最新的配置信息，适用于需要获取并更新配置的场景。'),
            codePreviewHeight: 1156,
          };
        }
        return {
          topTip: t(
            'Watch方法：通过建立长连接，实时监听配置版本的变更，当新版本的配置发布时，将自动调用回调方法处理新的配置信息，适用于需要实时响应配置变更的场景。',
          ),
          codePreviewHeight: 1172,
        };
      case 'cpp':
        if (!activeTab.value) {
          return {
            topTip: t('Get方法：用于一次性拉取最新的配置信息，适用于需要获取并更新配置的场景。'),
            codePreviewHeight: 1324,
          };
        }
        return {
          topTip: t(
            'Watch方法：通过建立长连接，实时监听配置版本的变更，当新版本的配置发布时，将自动调用回调方法处理新的配置信息，适用于需要实时响应配置变更的场景。',
          ),
          codePreviewHeight: 1990,
        };
      case 'http':
        // shell
        if (!activeTab.value) {
          return {
            topTip: '',
            codePreviewHeight: basicInfo?.serviceType.value === 'file' ? 470 : 604,
          };
        }
        // python
        return {
          topTip: '',
          codePreviewHeight: basicInfo?.serviceType.value === 'file' ? 982 : 850,
        };
      default:
        return {
          topTip: '',
          codePreviewHeight: 0,
        };
    }
  });
  // http分别使用shell和python高亮格式，其他保持原有
  const codeLanguage = computed(() => {
    if (props.kvName === 'http') {
      return tabArr.value[activeTab.value].toLocaleLowerCase();
    }
    return props.kvName;
  });

  watch(
    () => props.kvName,
    (newV) => {
      tabArr.value = newV === 'http' ? ['Shell', 'Python'] : [t('Get方法'), t('Watch方法')];
      codeVal.value = '';
      nextTick(() => handleTab());
    },
    { immediate: true },
  );

  const getOptionData = (data: any) => {
    // labels展示方式加工，并替换数据
    let labelArrType = '';
    switch (props.kvName) {
      case 'java':
        if (data.labelArr.length) {
          labelArrType = data.labelArr
            .map((item: string) => {
              const [key, value] = item.split(':');
              return `labels.put(${key}, ${value});`;
            })
            .join('');
        }
        break;
      case 'cpp':
        labelArrType = data.labelArr.length
          ? `{${data.labelArr.map((item: string) => `{${item.split(':').join(', ')}}`).join(', ')}}`
          : '{}';
        break;
      case 'http':
        labelArrType = data.labelArr.length ? `{${data.labelArr.join(', ')}}` : '{}';
        if (!activeTab.value) {
          // 文件型的shell需要添加转义符
          labelArrType =
            basicInfo?.serviceType.value === 'file'
              ? `'\\${labelArrType.slice(0, labelArrType.length - 1)}\\${labelArrType.slice(labelArrType.length - 1, labelArrType.length)}'`
                .replaceAll(' ', '',)
              : `'${labelArrType}'`;
        }
        break;
      default:
        labelArrType = data.labelArr.length ? `{${data.labelArr.join(', ')}}` : '{}';
        break;
    }
    optionData.value = {
      ...data,
      labelArrType,
    };
    replaceVal.value = codeVal.value; // 数据重置
    updateVariables(); // 表单数据更新，配置需要同时更新
    nextTick(() => {
      // 等待monaco渲染完成(高亮)再改固定值
      updateReplaceVal();
    });
  };
  const updateReplaceVal = () => {
    let updateString = replaceVal.value;
    let feedAddrVal = (window as any).GRPC_ADDR;
    if (props.kvName === 'http') {
      // http的host特殊处理
      feedAddrVal = (window as any).HTTP_ADDR;
    }
    updateString = updateString.replace('{{ .Bk_Bscp_Variable_BkBizId }}', bkBizId.value);
    updateString = updateString.replace('{{ .Bk_Bscp_Variable_ServiceName }}', basicInfo!.serviceName.value);
    replaceVal.value = updateString.replaceAll('{{ .Bk_Bscp_Variable_FEED_ADDR }}', feedAddrVal);
  };
  const updateVariables = () => {
    variables.value = [
      {
        name: 'Bk_Bscp_Variable_Leabels',
        type: '',
        default_val: optionData.value.labelArrType,
        memo: '',
      },
      {
        name: 'Bk_Bscp_Variable_ClientKey',
        type: '',
        default_val: `"${optionData.value.privacyCredential}"`,
        memo: '',
      },
      // {
      //   name: 'Bk_Bscp_Variable_Python_Key',
      //   type: '',
      //   default_val: '{{ YOUR_KEY }}',
      //   memo: '',
      // },
      {
        name: 'Bk_Bscp_Variable_KeyName',
        type: '',
        default_val: `"${optionData.value.configName}"`,
        memo: '',
      },
    ];
  };

  // 复制示例
  const copyExample = async () => {
    try {
      await fileOptionRef.value.handleValidate();
      // 复制示例使用未脱敏的密钥
      const reg = /"(.{1}|.{3})\*{3}(.{1}|.{3})"/g;
      let copyVal = copyReplaceVal.value.replaceAll(reg, `"${optionData.value.clientKey}"`);
      let tempStr = '';
      // 键值型示例复制时，内容开头插入注释信息(http除外)；插入文案除python以外，其他都一样
      if (props.kvName === 'python') {
        // watch
        tempStr = `'''\n${t('通过建立长连接，实时监听配置版本的变更，当新版本的配置发布时，将自动调用回调方法处理新的配置信息，适用于需要实时响应配置变更的场景\n有关Python SDK的部署环境和依赖组件，请参阅白皮书中的 [BSCP Python SDK依赖说明]')}\n(https://bk.tencent.com/docs/markdown/ZH/BSCP/1.29/UserGuide/Function/python_sdk_dependency.md)\n'''\n`;
        if (!activeTab.value) {
          // get
          tempStr = `'''\n${t('用于主动获取配置项值的场景，此方法不会监听服务器端的配置更改\n有关Python SDK的部署环境和依赖组件，请参阅白皮书中的 [BSCP Python SDK依赖说明]')}\n(https://bk.tencent.com/docs/markdown/ZH/BSCP/1.29/UserGuide/Function/python_sdk_dependency.md)\n'''\n`;
        }
      } else if (props.kvName !== 'http') {
        // watch
        tempStr = `// ${t('Watch方法：通过建立长连接，实时监听配置版本的变更，当新版本的配置发布时，将自动调用回调方法处理新的配置信息，适用于需要实时响应配置变更的场景。')}\n`;
        if (!activeTab.value) {
          // get
          tempStr = `// ${t('Get方法：用于一次性拉取最新的配置信息，适用于需要获取并更新配置的场景。')}\n`;
        }
      }
      copyVal = `${tempStr}${tempStr ? '\n' : ''}${copyVal}`;
      copyToClipBoard(copyVal);
      BkMessage({
        theme: 'success',
        message: t('示例已复制'),
      });
    } catch (error) {
      console.log(error);
    }
  };
  // 切换tab
  const handleTab = async (index = 0) => {
    if (index === activeTab.value && codeVal.value) return;
    activeTab.value = index;
    const newKvData = await changeKvData(props.kvName, index);
    codeVal.value = newKvData.default;
    replaceVal.value = newKvData.default;
    getOptionData(optionData.value);
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
          ? import('/src/assets/example-data/kv-python-get.yaml?raw')
          : import('/src/assets/example-data/kv-python-watch.yaml?raw');
      case 'go':
        return !methods
          ? import('/src/assets/example-data/kv-go-get.yaml?raw')
          : import('/src/assets/example-data/kv-go-watch.yaml?raw');
      case 'java':
        return !methods
          ? import('/src/assets/example-data/kv-java-get.yaml?raw')
          : import('/src/assets/example-data/kv-java-watch.yaml?raw');
      case 'cpp':
        return !methods
          ? import('/src/assets/example-data/kv-c++-get.yaml?raw')
          : import('/src/assets/example-data/kv-c++-watch.yaml?raw');
      case 'http':
        // http独有的file型
        if (basicInfo?.serviceType.value === 'file') {
          return !methods
            ? import('/src/assets/example-data/file-http-shell.yaml?raw')
            : import('/src/assets/example-data/file-http-python.yaml?raw');
        }
        return !methods
          ? import('/src/assets/example-data/kv-http-shell.yaml?raw')
          : import('/src/assets/example-data/kv-http-python.yaml?raw');
      default:
        return '';
    }
  };
</script>

<style scoped lang="scss">
  .kv-example-template {
    display: flex;
    flex-direction: column;
    height: 100%;
    .change-method {
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
      text-align: center;
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
    margin-top: 8px;
    padding: 8px 0 0;
    border-top: 1px solid #dcdee5;
  }
  .preview-label {
    font-weight: 700;
    font-size: 14px;
    letter-spacing: 0;
    line-height: 22px;
    color: #63656e;
  }
  .preview-component {
    margin-top: 8px;
    padding: 16px 10px;
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
