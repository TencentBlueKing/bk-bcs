<template>
  <div class="cmd-tool-wrap">
    <form-option @option-data="getOptionData" ref="fileOptionRef" />
    <div class="preview-container">
      <p class="headline">配置指引与示例预览</p>
      <div class="guide-wrap">
        <div class="guide-content" v-for="(item, index) in guideText" :key="item.title">
          <p class="guide-text">{{ `${index + 1}. ${item.title}` }}</p>
          <p class="guide-text guide-text--copy" v-if="item.value" @click="copyText(item.value)">
            {{ item.value }}
            <copy-shape class="icon-copy" />
          </p>
          <template v-else>
            <bk-button @click="copyExample" theme="primary" class="copy-btn">复制示例</bk-button>
            <code-preview
              class="preview-component"
              :code-val="replaceVal"
              :variables="variables"
              @change="(val) => (copyReplaceVal = val)" />
          </template>
          <template v-if="item.tips">
            <p class="guide-text guide-text--margin">{{ item.tips.title }}</p>
            <p class="guide-text">
              {{ item.tips.value }}：<span class="guide-text-url" @click="linkUrl(item.tips.url)">{{
                item.tips.url
              }}</span>
            </p>
          </template>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { computed, provide, ref, onMounted, watch } from 'vue';
  import { copyToClipBoard } from '../../../../../../utils/index';
  import BkMessage from 'bkui-vue/lib/message';
  import FormOption from '../form-option.vue';
  import CodePreview from '../code-preview.vue';
  import { CopyShape } from 'bkui-vue/lib/icon';
  import { useI18n } from 'vue-i18n';
  import { useRoute } from 'vue-router';
  import yamlString from '/src/assets/exampleData/file-cmd.yaml?raw';

  const props = defineProps<{ contentScrollTop: Function }>();

  const { t } = useI18n();
  const route = useRoute();
  const guideText = [
    {
      title: '下载二进制命令行',
      value: 'go install github.com/TencentBlueKing/bscp-go/cmd/bscp@latest',
      tips: {
        title: '如果没有安装 GO，可以通过浏览器手动下载，建议下载最新版本',
        value: '下载地址：',
        url: 'https://github.com/TencentBlueKing/bscp-go/releases/',
      },
    },
    {
      title: '创建命令配置文件，配置文件为 YAML 格式',
      value: '',
    },
    {
      title: '获取业务下的服务列表',
      value: './bscp get app -c ./bscp.yaml',
    },
    {
      title: '拉取服务下所有配置文件',
      value: './bscp pull -a alkaid-test-file -c ./bscp.yaml',
    },
    {
      title: '获取服务下所有配置文件列表',
      value: './bscp get file -a alkaid-test-file -c ./bscp.yaml',
    },
    {
      title: '下载指定配置文件到指定目录，例如指定文件为 /etc/nginx/nginx.conf，下载文件到 /root/config 目录',
      value: './bscp get file /etc/nginx/nginx.conf  -a alkaid-test-file -c ./bscp.yaml  -d /root/config',
    },
  ];

  const fileOptionRef = ref();
  const bkBizId = ref(String(route.params.spaceId));
  const replaceVal = ref('');
  const copyReplaceVal = ref(''); // 渲染的值，用于复制未脱敏密钥的yaml数据
  const formError = ref<number>();
  provide('formError', formError);
  // fileOption组件传递过来的数据汇总
  const optionData = ref({
    clientKey: '',
    privacyCredential: '',
    labelArr: [],
    tempDir: '',
  });

  const variables = computed(() => {
    replaceVal.value = '';
    return [
      {
        name: 'Bk_Bscp_VariableLeabels',
        type: '',
        default_val: `'{${optionData.value.labelArr}}'`,
        memo: '',
      },
      {
        name: 'Bk_Bscp_VariableClientKey',
        type: '',
        default_val: `'${optionData.value.privacyCredential}'`,
        memo: '',
      },
      {
        name: 'Bk_Bscp_VariableTempDir',
        type: '',
        default_val: `'${optionData.value.tempDir}'`,
        memo: '',
      },
    ];
  });

  watch(
    () => replaceVal.value,
    () => {
      // 初始化值，variables对应配置生效
      let updateString = yamlString;
      updateString = updateString.replace('{{ .Bk_Bscp_VariableBkBizId }}', bkBizId.value);
      replaceVal.value = updateString.replace('{{ .Bk_Bscp_VariableFEED_ADDR }}', (window as any).FEED_ADDR);
    },
  );

  onMounted(() => {
    replaceVal.value = yamlString;
  });

  // 监听传来的数据
  const getOptionData = (data: any) => {
    optionData.value = data;
    optionData.value = computed(() => {
      // 标签展示方式加工
      const labelArr = data.labelArr.length ? data.labelArr.join(', ') : '';
      return {
        ...data,
        labelArr,
      };
    }).value;
  };
  // 复制示例
  const copyExample = async () => {
    try {
      await fileOptionRef.value.formRef.validate();
      // 复制示例使用未脱敏的密钥
      const reg = /'(.{1}|.{3})\*{3}(.{1}|.{3})'/g;
      const copyVal = copyReplaceVal.value.replaceAll(reg, `'${optionData.value.clientKey}'`);
      copyToClipBoard(copyVal);
      BkMessage({
        theme: 'success',
        message: t('示例已复制'),
      });
    } catch (error) {
      // 通知密钥选择组件校验状态
      formError.value = new Date().getTime();
      props.contentScrollTop();

      console.log(error);
    }
  };
  const copyText = (copyVal: string) => {
    copyToClipBoard(copyVal);
    BkMessage({
      theme: 'success',
      message: t('复制成功'),
    });
  };
  // 跳转链接
  const linkUrl = (url: string) => {
    window.open(url, '__blank');
  };
</script>

<style scoped lang="scss">
  .cmd-tool-wrap {
    .guide-content {
      margin-top: 24px;
      &:first-child {
        margin-top: 19px;
      }
    }
    .guide-text {
      margin: 0;
      font-size: 12px;
      line-height: 20px;
      color: #63656e;
      &--margin {
        margin-top: 10px;
      }
      &--copy {
        padding: 0 6px;
        margin-left: -4px;
        display: inline-block;
        vertical-align: middle;
        line-height: 24px;
        cursor: pointer;
        &:hover {
          background-color: #f5f7fa;
          .icon-copy {
            visibility: visible;
          }
        }
      }
      &-url {
        cursor: pointer;
        color: #3a84ff;
      }
    }
    .copy-btn {
      margin: 8px 0;
    }
  }
  .preview-container {
    margin-top: 32px;
    padding: 8px 0;
    flex: 1;
    height: 100%;
    border-top: 1px solid #dcdee5;
    .headline {
      margin: 0;
      font-size: 14px;
      font-weight: 700;
      line-height: 22px;
      color: #63656e;
    }
    .icon-copy {
      margin-left: 11px;
      font-size: 12px;
      color: #3a84ff;
      vertical-align: middle;
      visibility: hidden;
    }
  }
  .preview-component {
    height: 292px;
    padding: 16px 0 0;
    background-color: #f5f7fa;
  }
</style>
