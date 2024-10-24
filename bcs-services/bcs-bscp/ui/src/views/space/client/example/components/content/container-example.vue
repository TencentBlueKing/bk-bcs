<template>
  <section class="example-wrap">
    <form-option
      ref="fileOptionRef"
      :p2p-show="true"
      :associate-config-show="true"
      :selected-key-data="props.selectedKeyData"
      @update-option-data="getOptionData"
      @selected-key-data="emits('selected-key-data', $event)" />
    <div class="preview-container">
      <span class="preview-label">{{ $t('示例预览') }}</span>
      <bk-button theme="primary" class="copy-btn" @click="copyExample">{{ $t('复制示例') }}</bk-button>
      <code-preview
        class="preview-component"
        :style="{ height: optionData.clusterSwitch ? '1990px' : '1496px' }"
        :code-val="replaceVal"
        :variables="variables"
        language="yaml"
        @change="(val: string) => (copyReplaceVal = val)" />
    </div>
  </section>
</template>

<script lang="ts" setup>
  import { ref, inject, Ref, nextTick } from 'vue';
  import { copyToClipBoard } from '../../../../../../utils/index';
  import { IVariableEditParams } from '../../../../../../../types/variable';
  import BkMessage from 'bkui-vue/lib/message';
  import FormOption from '../form-option.vue';
  import CodePreview from '../code-preview.vue';
  import { IExampleFormData, newICredentialItem } from '../../../../../../../types/client';
  import { useI18n } from 'vue-i18n';
  import { useRoute } from 'vue-router';
  import yamlString from '/src/assets/example-data/file-container.yaml?raw';

  const props = defineProps<{ contentScrollTop: Function; selectedKeyData: newICredentialItem['spec'] | null }>();

  const emits = defineEmits(['selected-key-data']);

  const { t } = useI18n();
  const route = useRoute();

  const fileOptionRef = ref();
  // fileOption组件传递过来的数据汇总
  const optionData = ref<IExampleFormData>({
    clientKey: '',
    privacyCredential: '',
    labelArr: [],
    tempDir: '',
    clusterSwitch: false,
    clusterInfo: '',
    rules: [],
  });
  const bkBizId = ref(String(route.params.spaceId));
  const replaceVal = ref('');
  const copyReplaceVal = ref(''); // 渲染的值，用于复制未脱敏密钥的yaml数据
  const basicInfo = inject<{ serviceName: Ref<string>; serviceType: Ref<string> }>('basicInfo');
  const variables = ref<IVariableEditParams[]>();

  const getOptionData = (data: any) => {
    // 标签展示方式加工
    const labelArr = data.labelArr.length ? data.labelArr.join(', ') : '';
    optionData.value = {
      ...data,
      labelArr,
    };
    replaceVal.value = yamlString; // 获取/重置传递的数据
    updateVariables(); // 表单数据更新，配置需要同时更新
    nextTick(() => {
      // 等待monaco渲染完成(高亮)再改固定值
      updateReplaceVal();
    });
  };
  const updateReplaceVal = () => {
    let updateString = replaceVal.value;
    updateString = updateString.replaceAll('{{ .Bk_Bscp_Variable_BkBizId }}', bkBizId.value);
    updateString = updateString.replaceAll('{{ .Bk_Bscp_Variable_ServiceName }}', basicInfo!.serviceName.value);
    updateString = updateString.replaceAll('{{ .Bk_Bscp_Variable_FEED_ADDR }}', (window as any).GRPC_ADDR);
    // p2p网络加速打开后动态插入内容
    if (optionData.value.clusterSwitch) {
      const p2pPart1 = `
            # 是否启用P2P网络加速
            - name: enable_p2p_download
              value: 'true'
            # 以下几个环境变量在启用 p2p 文件加速时为必填项
            # BCS集群ID
            - name: cluster_id
              value: {{ .Bk_Bscp_Variable_Cluster_Value }}
            # BSCP容器名称
            - name: container_name
              value: bscp-init
            - name: pod_id
              valueFrom:
                fieldRef:
                  apiVersion: v1
                  fieldPath: metadata.uid`;
      const p2pPart2 = `
            - name: enable_p2p_download
              value: 'true'
            - name: cluster_id
              value: {{ .Bk_Bscp_Variable_Cluster_Value }}
            - name: container_name
              value: bscp-sidecar
            - name: pod_id
              valueFrom:
                fieldRef:
                  apiVersion: v1
                  fieldPath: metadata.uid`;
      updateString = updateString.replaceAll('{{ .Bk_Bscp_Variable_p2p_part1 }}', p2pPart1);
      updateString = updateString.replaceAll('{{ .Bk_Bscp_Variable_p2p_part2 }}', p2pPart2);
    } else {
      updateString = updateString.replaceAll('{{ .Bk_Bscp_Variable_p2p_part1 }}', '');
      updateString = updateString.replaceAll('{{ .Bk_Bscp_Variable_p2p_part2 }}', '');
    }

    // 文件配置筛选规则动态增/删
    if (optionData.value.rules?.length) {
      const rulesPart1 = `
      # 当客户端无需拉取配置服务中的全量配置文件时，指定相应的通配符，可仅拉取客户端所需的文件，支持多个通配符
            - name: config_matches
              value: {{ .Bk_Bscp_Variable_Rules_Value }}`;
      const rulesPart2 = `
            - name: config_matches
              value: {{ .Bk_Bscp_Variable_Rules_Value }}`;
      updateString = updateString.replaceAll('{{ .Bk_Bscp_Variable_Rules1 }}', rulesPart1);
      updateString = updateString.replaceAll('{{ .Bk_Bscp_Variable_Rules2 }}', rulesPart2);
    } else {
      updateString = updateString.replaceAll('{{ .Bk_Bscp_Variable_Rules1 }}', '');
      updateString = updateString.replaceAll('{{ .Bk_Bscp_Variable_Rules2 }}', '');
    }
    // 去除 动态插入的值为空的情况下产生的空白行
    replaceVal.value = updateString.replaceAll(/\r?\n\s+\r?\n/g, '\n');
  };
  // 高亮配置
  const updateVariables = () => {
    variables.value = [
      {
        name: 'Bk_Bscp_Variable_Leabels',
        type: '',
        default_val: `'{${optionData.value.labelArr}}'`,
        memo: '',
      },
      {
        name: 'Bk_Bscp_Variable_ClientKey',
        type: '',
        default_val: `'${optionData.value.privacyCredential}'`,
        memo: '',
      },
      {
        name: 'Bk_Bscp_Variable_TempDir',
        type: '',
        default_val: `'${optionData.value.tempDir}'`,
        memo: '',
      },
      {
        name: 'Bk_Bscp_Variable_Cluster_Value',
        type: '',
        default_val: `${optionData.value.clusterInfo}`,
        memo: '',
      },
      {
        name: 'Bk_Bscp_Variable_Rules_Value',
        type: '',
        default_val: `'${optionData.value.rules}'`,
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
        message: t('示例已复制'),
      });
    } catch (error) {
      props.contentScrollTop();
      console.log(error);
    }
  };
</script>

<style scoped lang="scss">
  .example-wrap {
    .preview-component {
      margin-top: 16px;
      padding: 16px 8px;
      background-color: #f5f7fa;
    }
  }
  .preview-container {
    margin-top: 32px;
    padding: 16px 0 0;
    border-top: 1px solid #dcdee5;
    overflow: hidden;
  }

  .preview-label {
    font-weight: 700;
    font-size: 14px;
    letter-spacing: 0;
    line-height: 22px;
    color: #63656e;
  }
  .copy-btn {
    margin-left: 16px;
  }
</style>
