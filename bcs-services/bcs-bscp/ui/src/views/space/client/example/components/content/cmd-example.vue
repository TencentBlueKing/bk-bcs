<template>
  <section class="cmd-tool-wrap">
    <form-option ref="fileOptionRef" :directory-show="props.kvName !== 'kv-cmd'" @update-option-data="getOptionData" />
    <div class="preview-container">
      <p class="headline">{{ $t('配置指引与示例预览') }}</p>
      <div class="guide-wrap">
        <div class="guide-content" v-for="(item, index) in cmdContent" :key="item.title">
          <p class="guide-text">{{ `${index + 1}. ${item.title}` }}</p>
          <p class="guide-text guide-text--copy" v-if="item.value" @click="copyText(item.value)">
            {{ item.value }}
            <copy-shape class="icon-copy" />
          </p>
          <template v-else>
            <bk-button theme="primary" class="copy-btn" @click="copyExample">{{ $t('复制命令') }}</bk-button>
            <code-preview
              :class="['preview-component', { 'preview-component--kvcmd': props.kvName === 'kv-cmd' }]"
              :code-val="replaceVal"
              :variables="variables"
              :language="kvName"
              @change="(val: string) => (copyReplaceVal = val)" />
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
  </section>
</template>

<script lang="ts" setup>
  import { computed, provide, ref, onMounted, inject, Ref, nextTick } from 'vue';
  import { copyToClipBoard } from '../../../../../../utils/index';
  import { IVariableEditParams } from '../../../../../../../types/variable';
  import BkMessage from 'bkui-vue/lib/message';
  import FormOption from '../form-option.vue';
  import CodePreview from '../code-preview.vue';
  import { CopyShape } from 'bkui-vue/lib/icon';
  import { useI18n } from 'vue-i18n';
  import { useRoute } from 'vue-router';

  const props = defineProps<{ contentScrollTop: Function; kvName: string }>();

  const { t } = useI18n();
  const route = useRoute();
  const serviceName = inject<Ref<string>>('serviceName');
  const fileText = [
    {
      title: t('下载二进制命令行'),
      value: 'go install github.com/TencentBlueKing/bscp-go/cmd/bscp@latest',
      tips: {
        title: t('如果没有安装 GO，可以通过浏览器手动下载，建议下载最新版本'),
        value: t('下载地址：'),
        url: 'https://github.com/TencentBlueKing/bscp-go/releases/',
      },
    },
    {
      title: t('为命令行工具创建所需的配置文件bscp.yaml，请复制以下命令并在与该命令行工具相同的目录下执行'),
      value: '',
    },
    {
      title: t('获取业务下的服务列表'),
      value: './bscp get app -c ./bscp.yaml',
    },
    {
      title: t('拉取服务下所有配置文件'),
      value: `./bscp pull -a ${serviceName!.value} -c ./bscp.yaml`,
    },
    {
      title: t('获取服务下所有配置文件列表'),
      value: `./bscp get file -a ${serviceName!.value} -c ./bscp.yaml`,
    },
    {
      title: t(
        '下载配置文件时，保留目录层级，并将其保存到指定目录下，例如：将 /etc/nginx.conf 文件下载到 /tmp 目录时，文件保存在 /tmp/etc/nginx.conf',
      ),
      value: `./bscp get file /etc/nginx.conf -a ${serviceName!.value} -c ./bscp.yaml -d /tmp`,
    },
    {
      title: t(
        '下载配置文件时，不保留目录层级，并将其保存到指定目录下，例如：将 /etc/nginx.conf 文件下载到 /tmp 目录时，文件保存在 /tmp/nginx.conf',
      ),
      value: `./bscp get file /etc/nginx.conf -a ${serviceName!.value} -c ./bscp.yaml -d /tmp --ignore-dir`,
    },
  ];

  const kvText = [
    {
      title: t('下载二进制命令行'),
      value: 'go install github.com/TencentBlueKing/bscp-go/cmd/bscp@latest',
      tips: {
        title: t('如果没有安装 GO，可以通过浏览器手动下载，建议下载最新版本'),
        value: t('下载地址：'),
        url: 'https://github.com/TencentBlueKing/bscp-go/releases/',
      },
    },
    {
      title: t('为命令行工具创建所需的配置文件bscp.yaml，请复制以下命令并在与该命令行工具相同的目录下执行'),
      value: '',
    },
    {
      title: t('获取业务下的服务列表'),
      value: './bscp get app -c ./bscp.yaml',
    },
    {
      title: t('获取指定服务下所有配置项列表'),
      value: `./bscp get kv -a ${serviceName!.value} -c ./bscp.yaml`,
    },
    {
      title: t('获取指定服务下指定配置项列表，多个配置项'),
      value: `./bscp get kv -a ${serviceName!.value} key1 key2 -c ./bscp.yaml`,
    },
    {
      title: t('获取指定服务下指定配置项值，只支持单个配置项值获取'),
      value: `./bscp get kv -a ${serviceName!.value} key1  -c ./bscp.yaml -o value`,
    },
    {
      title: t(
        '获取指定服务下指定配置项元数据，支持多个配置项元数据获取，没有指定配置项，获取服务下所有配置项的元数据',
      ),
      value: `./bscp get kv -a ${serviceName!.value} key1 key2  -c ./bscp.yaml -o json`,
    },
  ];

  const fileOptionRef = ref();
  const bkBizId = ref(String(route.params.spaceId));
  const codeVal = ref(''); // 存储yaml字符原始值
  const replaceVal = ref('');
  const copyReplaceVal = ref(''); // 渲染的值，用于复制未脱敏密钥的yaml数据
  const variables = ref<IVariableEditParams[]>();
  const formError = ref<number>();
  provide('formError', formError);
  // fileOption组件传递过来的数据汇总
  const optionData = ref({
    clientKey: '',
    privacyCredential: '',
    labelArr: [],
    tempDir: '',
  });

  const cmdContent = computed(() => {
    return props.kvName === 'file-cmd' ? fileText : kvText;
  });

  onMounted(async () => {
    const newKvData = await changeKvData(props.kvName);
    codeVal.value = newKvData.default;
    replaceVal.value = newKvData.default;
    updateReplaceVal();
  });

  // 监听传来的数据
  const getOptionData = (data: any) => {
    const labelArr = data.labelArr.length ? data.labelArr.join(', ') : '';
    optionData.value = {
      ...data,
      labelArr,
    };
    updateVariables(); // 表单数据更新，配置需要同时更新
    replaceVal.value = codeVal.value; // 数据重置
    nextTick(() => {
      // 等待monaco渲染完成(高亮)再改固定值
      updateReplaceVal();
    });
  };
  const updateReplaceVal = () => {
    let updateString = replaceVal.value;
    updateString = updateString.replace('{{ .Bk_Bscp_Variable_BkBizId }}', bkBizId.value);
    updateString = updateString.replace('{{ .Bk_Bscp_Variable_ServiceName }}', serviceName!.value);
    replaceVal.value = updateString.replaceAll('{{ .Bk_Bscp_Variable_FEED_ADDR }}', (window as any).FEED_ADDR);
  };
  const updateVariables = () => {
    variables.value = [
      {
        name: 'Bk_Bscp_Variable_Leabels',
        type: '',
        default_val: `{${optionData.value.labelArr}}`,
        memo: '',
      },
      {
        name: 'Bk_Bscp_Variable_ClientKey',
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
  };
  // 复制示例
  const copyExample = async () => {
    try {
      await fileOptionRef.value.handleValidate();
      // 复制示例使用未脱敏的密钥
      const reg = /'(.{1}|.{3})\*{3}(.{1}|.{3})'/g;
      const copyVal = copyReplaceVal.value.replaceAll(reg, `'${optionData.value.clientKey}'`);
      copyToClipBoard(copyVal);
      BkMessage({
        theme: 'success',
        message: t('复制成功'),
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
  // 命令行file与kv的数据模板切换
  /**
   *
   * @param serviceType 数据模板名称
   */
  const changeKvData = (serviceType = 'file-cmd') => {
    return serviceType === 'file-cmd'
      ? import('/src/assets/example-data/file-cmd.yaml?raw')
      : import('/src/assets/example-data/kv-cmd.yaml?raw');
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
    height: 334px;
    padding: 16px 0 0;
    background-color: #f5f7fa;
    &--kvcmd {
      height: 276px;
    }
  }
</style>
