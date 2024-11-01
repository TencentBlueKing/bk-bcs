<template>
  <section class="cmd-tool-wrap">
    <form-option
      ref="fileOptionRef"
      :directory-show="basicInfo!.serviceType.value === 'file'"
      :associate-config-show="basicInfo!.serviceType.value === 'file'"
      :dual-system-support="true"
      :config-show="true"
      :config-label="basicInfo?.serviceType.value === 'file' ? '配置文件名' : '配置项名称'"
      :selected-key-data="props.selectedKeyData"
      @update-option-data="getOptionData"
      @selected-key-data="emits('selected-key-data', $event)" />
    <div class="preview-container">
      <p class="headline">{{ $t('配置指引与示例预览') }}</p>
      <div class="guide-wrap">
        <div class="guide-content" v-for="(item, index) in cmdContent" :key="item.value + index">
          <p class="guide-text">{{ `${index + 1}. ${item.title}` }}</p>
          <p class="guide-text guide-text--copy" v-if="item.value" @click="copyText(item.value, index)">
            {{ item.value }}
            <copy-shape class="icon-copy" />
          </p>
          <template v-else>
            <bk-button theme="primary" class="copy-btn" @click="copyExample">
              {{ $t(optionData.systemType === 'Windows' ? '复制配置' : '复制命令') }}
            </bk-button>
            <code-preview
              :class="[
                'preview-component',
                { 'preview-component--kvcmd': basicInfo!.serviceType.value === 'kv' },
                { 'rules-height': optionData.rules.length },
                { 'windows-path-height': optionData.systemType === 'Windows' },
              ]"
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
  import { ref, Ref, computed, inject, onMounted, nextTick } from 'vue';
  import { copyToClipBoard } from '../../../../../../utils/index';
  import { IVariableEditParams } from '../../../../../../../types/variable';
  import { newICredentialItem } from '../../../../../../../types/client';
  import BkMessage from 'bkui-vue/lib/message';
  import FormOption from '../form-option.vue';
  import CodePreview from '../code-preview.vue';
  import { CopyShape } from 'bkui-vue/lib/icon';
  import { useI18n } from 'vue-i18n';
  import { useRoute } from 'vue-router';

  const props = defineProps<{
    contentScrollTop: Function;
    kvName: string;
    selectedKeyData: newICredentialItem['spec'] | null;
  }>();

  const emits = defineEmits(['selected-key-data']);

  const { t } = useI18n();
  const route = useRoute();
  const basicInfo = inject<{ serviceName: Ref<string>; serviceType: Ref<string> }>('basicInfo');

  const fileOptionRef = ref();
  const bkBizId = ref(String(route.params.spaceId));
  const codeVal = ref(''); // 存储yaml字符原始值
  const replaceVal = ref('');
  const copyReplaceVal = ref(''); // 渲染的值，用于复制未脱敏密钥的yaml数据
  const variables = ref<IVariableEditParams[]>();
  // fileOption组件传递过来的数据汇总
  const optionData = ref({
    clientKey: '',
    privacyCredential: '',
    labelArr: [],
    tempDir: '',
    rules: [],
    systemType: 'Unix',
    configName: '', // 配置项
  });

  const fileText = computed(() => [
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
      value: `./bscp pull -a ${basicInfo!.serviceName.value} -c ./bscp.yaml`,
    },
    {
      title: t('获取服务下所有配置文件列表'),
      value: `./bscp get file -a ${basicInfo!.serviceName.value} -c ./bscp.yaml`,
    },
    {
      title: t(
        '下载配置文件时，保留目录层级，并将其保存到指定目录下，例如：将 {path} 文件下载到 /tmp 目录时，文件保存在 /tmp{path}',
        { path: optionData.value.configName },
      ),
      value: `./bscp get file ${optionData.value.configName} -a ${basicInfo!.serviceName.value} -c ./bscp.yaml -d /tmp`,
    },
    {
      title: t(
        '下载配置文件时，不保留目录层级，并将其保存到指定目录下，例如：将 {path} 文件下载到 /tmp 目录时，文件保存在 /tmp/{name}',
        {
          path: optionData.value.configName,
          name: optionData.value.configName.substring(optionData.value.configName.lastIndexOf('/') + 1),
        },
      ),
      value: `./bscp get file ${optionData.value.configName} -a ${basicInfo!.serviceName.value} -c ./bscp.yaml -d /tmp --ignore-dir`,
    },
  ]);

  const windowsFileText = computed(() => [
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
      title: t('为命令行工具创建所需的配置文件bscp.yaml'),
      value: '',
    },
    {
      title: t('获取业务下的服务列表'),
      value: '.\\bscp.exe get app -c .\\bscp.yaml',
    },
    {
      title: t('拉取服务下所有配置文件'),
      value: `.\\bscp.exe pull -a ${basicInfo!.serviceName.value} -c .\\bscp.yaml`,
    },
    {
      title: t('获取服务下所有配置文件列表'),
      value: `.\\bscp.exe get file -a ${basicInfo!.serviceName.value} -c .\\bscp.yaml`,
    },
    {
      title: t(
        '下载配置文件时，保留目录层级，并将其保存到指定目录下，例如：将 {path} 文件下载到当前目录时，文件保存在 .{windowsPath}',
        { path: optionData.value.configName, windowsPath: optionData.value.configName.replaceAll('/', '\\') },
      ),
      value: `.\\bscp.exe get file ${optionData.value.configName} -a ${basicInfo!.serviceName.value} -c .\\bscp.yaml -d .\\`,
    },
    {
      title: t(
        '下载配置文件时，不保留目录层级，并将其保存到指定目录下，例如：将 {path} 文件下载到当前目录时，文件保存在 .\\{windowsName}',
        {
          path: optionData.value.configName,
          windowsName: optionData.value.configName
            .substring(optionData.value.configName.lastIndexOf('/') + 1)
            .replaceAll('/', '\\'),
        },
      ),
      value: `.\\bscp.exe get file ${optionData.value.configName} -a ${basicInfo!.serviceName.value} -c .\\bscp.yaml -d .\\ --ignore-dir`,
    },
  ]);

  const kvText = computed(() => [
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
      value: `./bscp get kv -a ${basicInfo!.serviceName.value} -c ./bscp.yaml`,
    },
    {
      title: t('获取指定服务下指定配置项列表，支持多个配置项，配置项之间用空格分隔'),
      value: `./bscp get kv -a ${basicInfo!.serviceName.value} ${optionData.value.configName} -c ./bscp.yaml`,
    },
    {
      title: t('获取指定服务下指定配置项值，只支持单个配置项值获取'),
      value: `./bscp get kv -a ${basicInfo!.serviceName.value} ${optionData.value.configName} -c ./bscp.yaml -o value`,
    },
    {
      title: t(
        '获取指定服务下指定配置项元数据，支持多个配置项元数据获取，没有指定配置项，获取服务下所有配置项的元数据',
      ),
      value: `./bscp get kv -a ${basicInfo!.serviceName.value} ${optionData.value.configName} -c ./bscp.yaml -o json`,
    },
  ]);

  const windowsKvText = computed(() => [
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
      title: t('为命令行工具创建所需的配置文件bscp.yaml'),
      value: '',
    },
    {
      title: t('获取业务下的服务列表'),
      value: '.\\bscp.exe get app -c .\\bscp.yaml',
    },
    {
      title: t('获取指定服务下所有配置项列表'),
      value: `.\\bscp.exe get kv -a ${basicInfo!.serviceName.value} -c .\\bscp.yaml`,
    },
    {
      title: t('获取指定服务下指定配置项列表，支持多个配置项，配置项之间用空格分隔'),
      value: `.\\bscp.exe get kv -a ${basicInfo!.serviceName.value} ${optionData.value.configName} -c .\\bscp.yaml`,
    },
    {
      title: t('获取指定服务下指定配置项值，只支持单个配置项值获取'),
      value: `.\\bscp.exe get kv -a ${basicInfo!.serviceName.value} ${optionData.value.configName} -c .\\bscp.yaml -o value`,
    },
    {
      title: t(
        '获取指定服务下指定配置项元数据，支持多个配置项元数据获取，没有指定配置项，获取服务下所有配置项的元数据',
      ),
      value: `.\\bscp.exe get kv -a ${basicInfo!.serviceName.value} ${optionData.value.configName} -c .\\bscp.yaml -o json`,
    },
  ]);

  const cmdContent = computed(() => {
    if (basicInfo!.serviceType.value === 'file') {
      if (optionData.value.systemType === 'Windows') {
        return windowsFileText.value; // 文件型-windows路径提示文案
      }
      return fileText.value; // 文件型-unix路径提示文案
    }
    if (optionData.value.systemType === 'Windows') {
      return windowsKvText.value; // 键值型-windows路径提示文案
    }
    return kvText.value;
  });

  onMounted(async () => {
    const newKvData = await changeKvData(basicInfo!.serviceType.value);
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
    updateString = updateString.replace('{{ .Bk_Bscp_Variable_ServiceName }}', basicInfo!.serviceName.value);
    updateString = updateString.replaceAll('{{ .Bk_Bscp_Variable_FEED_ADDR }}', (window as any).GRPC_ADDR);
    // 文件配置筛选规则动态增/删
    if (optionData.value.rules?.length) {
      const rulesPart = `
  # 当客户端无需拉取配置服务中的全量配置文件时，指定相应的通配符，可仅拉取客户端所需的文件，支持多个通配符
  config_matches: {{ .Bk_Bscp_Variable_Rules_Value }}`;
      updateString = updateString.replaceAll('{{ .Bk_Bscp_Variable_Rules }}', rulesPart);
    } else {
      updateString = updateString.replaceAll('{{ .Bk_Bscp_Variable_Rules }}', 'delete');
    }
    // 临时目录为windows路径时去除首尾两行
    if (optionData.value.systemType === 'Windows') {
      updateString = updateString.replaceAll('cat << EOF > ./bscp.yaml', 'delete').replaceAll('EOF', 'delete');
    }
    // 去除 动态插入的值为空的情况下产生的空白行
    replaceVal.value = updateString.replaceAll(/(delete\r?\n|\r?\ndelete)/g, '');
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
      {
        name: 'Bk_Bscp_Variable_Rules_Value',
        type: '',
        default_val: `[${optionData.value.rules.map((rule) => `"${rule}"`).join(',')}]`,
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
      props.contentScrollTop();
      console.log(error);
    }
  };
  const copyText = (copyVal: string, configIndex: number) => {
    // 未选择配置项时不能复制内容 file：5、6项，kv：4、5、6项涉及配置项复制
    const serviceType = basicInfo!.serviceType.value || 'file';
    const validConfigIndices: Record<string, number[]> = {
      file: [5, 6],
      kv: [4, 5, 6],
    };
    if (validConfigIndices[serviceType].includes(configIndex) && !optionData.value.configName) {
      BkMessage({
        theme: 'error',
        message: t(
          `请先选择${serviceType === 'file' ? '配置文件名' : '配置项名称'}，替换下方示例代码后，再尝试复制示例`,
        ),
      });
      return;
    }
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
  const changeKvData = (serviceType = 'file') => {
    return serviceType === 'file'
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
    height: 336px;
    padding: 16px 10px;
    background-color: #f5f7fa;
    &.rules-height {
      height: 394px;
    }
    &.windows-path-height {
      height: 298px;
    }
    &.rules-height.windows-path-height {
      height: 356px;
    }
    &--kvcmd {
      height: 279px;
      // 键值型cmd没有规则筛选
      &.windows-path-height {
        height: 242px;
      }
    }
  }
</style>
