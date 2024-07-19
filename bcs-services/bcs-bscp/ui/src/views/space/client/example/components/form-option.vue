<template>
  <div class="headline">{{ $t('示例参数') }}</div>
  <bk-form class="form-example-wrap" :model="formData" :rules="rules" form-type="vertical" ref="formRef">
    <bk-form-item property="clientKey" required>
      <template #label>
        {{ $t('客户端密钥') }}
        <info
          class="icon-info"
          v-bk-tooltips="{
            content: $t('用于客户端拉取配置时身份验证，下拉列表只会展示关联过此服务且状态为启用的密钥'),
            placement: 'top',
          }" />
      </template>
      <KeySelect ref="keySelectorRef" @current-key="setCredential" />
    </bk-form-item>
    <bk-form-item v-if="props.directoryShow" property="tempDir" :required="props.directoryShow">
      <template #label>
        {{ $t('临时目录') }}
        <info
          class="icon-info"
          v-bk-tooltips="{
            content: $t('用于客户端拉取文件型配置后的临时存储目录'),
            placement: 'top',
          }" />
      </template>
      <bk-input v-model="formData.tempDir" :placeholder="$t('请输入')" clearable />
    </bk-form-item>
    <bk-form-item>
      <!-- 添加标签 -->
      <AddLabel ref="addLabelRef" :label-name="labelName" @send-label="(obj) => (formData.labelArr = obj)" />
    </bk-form-item>
    <bk-form-item v-if="p2pShow">
      <!-- p2p网络加速 -->
      <p2p-acceleration
        ref="p2pAccelerationRef"
        @send-cluster="
          ({ clusterSwitch, clusterInfo }) => {
            formData.clusterSwitch = clusterSwitch;
            formData.clusterInfo = clusterInfo;
          }
        " />
    </bk-form-item>
  </bk-form>
</template>

<script lang="ts" setup>
  import { onMounted, ref, watch } from 'vue';
  import KeySelect from './key-selector.vue';
  import { Info } from 'bkui-vue/lib/icon';
  import AddLabel from './add-label.vue';
  import p2pAcceleration from './p2p-acceleration.vue';
  import { IExampleFormData } from '../../../../../../types/client';
  import { useI18n } from 'vue-i18n';
  import { cloneDeep } from 'lodash';

  const props = defineProps({
    directoryShow: {
      type: Boolean,
      default: true,
    },
    labelName: {
      type: String,
      default: '标签',
    },
    p2pShow: {
      type: Boolean,
      default: false,
    },
  });

  const emits = defineEmits(['update-option-data']);

  const { t } = useI18n();
  const sysDirectories: string[] = ['/bin', '/boot', '/dev', '/lib', '/lib64', '/proc', '/run', '/sbin', '/sys'];

  const addLabelRef = ref();
  const keySelectorRef = ref();
  const p2pAccelerationRef = ref();
  const formRef = ref();
  const formData = ref<IExampleFormData>({
    clientKey: '', // 客户端密钥
    privacyCredential: '', // 脱敏的密钥
    tempDir: '/data/bscp', // 临时目录
    labelArr: [], // 添加的标签
    clusterSwitch: false, // 集群开关
    clusterInfo: {
      name: '', // 集群名称
      value: '', // 集群id
    },
  });

  const rules = {
    clientKey: [
      {
        required: true,
        message: t('请先选择客户端密钥，替换下方示例代码后，再尝试复制示例'),
        validator: (value: string) => value.length,
      },
    ],
    tempDir: [
      {
        required: props.directoryShow,
        message: t('请输入路径地址，替换下方示例代码后，再尝试复制示例'),
        validator: (value: string) => value.length,
        trigger: 'change',
      },
      {
        required: props.directoryShow,
        message: t('禁止使用系统目录'),
        validator: (value: string) => !sysDirectories.some((dir) => value === dir || value.startsWith(`${dir}/`)),
        trigger: 'change',
      },
      {
        required: props.directoryShow,
        validator: (value: string) => {
          // 必须为绝对路径, 且不能以/结尾
          if (!value.startsWith('/') || value.endsWith('/')) {
            return false;
          }
          const parts = value.split('/').slice(1);
          let isValid = true;
          // 文件路径校验
          parts.some((part) => {
            if (part.startsWith('.') || !/^[\u4e00-\u9fa5A-Za-z0-9.\-_#%,@^+=\\[\]{}]+$/.test(part)) {
              isValid = false;
              return true;
            }
            return false;
          });
          return isValid;
        },
        trigger: 'change',
        message: t('无效的路径,路径不符合Unix文件路径格式规范'),
      },
    ],
  };

  watch(formData.value, () => {
    sendAll();
  });

  onMounted(() => {
    sendAll();
  });

  const handleValidate = () => {
    // label验证，数组长度为空时返回true
    const labelValid = addLabelRef.value.isAllValid();
    // p2p网络加速验证，目前只有Sidecar使用，根据有无使用决定验证情况
    const p2pValid = props.p2pShow ? p2pAccelerationRef.value.isValid() : true;
    // 密钥验证
    const keyValid = keySelectorRef.value.validateCredential();
    const isAllValid = [labelValid, p2pValid, keyValid].includes(false);
    if (isAllValid) {
      formRef.value.validate();
      return Promise.reject();
    }
    return formRef.value.validate();
  };
  const setCredential = (key: string, privacyKey: string) => {
    formData.value.clientKey = key;
    formData.value.privacyCredential = privacyKey;
    formRef.value.validate();
  };
  const sendAll = () => {
    const filterFormData = cloneDeep(formData.value);
    // 过滤发送数据
    if (!props.directoryShow) {
      delete filterFormData.tempDir;
    }
    if (!props.p2pShow) {
      delete filterFormData.clusterSwitch;
      delete filterFormData.clusterInfo;
    }
    console.log(filterFormData);
    emits('update-option-data', filterFormData);
  };

  defineExpose({
    handleValidate,
  });
</script>

<style scoped lang="scss">
  .headline {
    font-size: 14px;
    font-weight: 700;
    line-height: 22px;
    color: #63656e;
  }
  .form-example-wrap {
    margin-top: 16px;
    width: 537px;
    :deep(.bk-form-label) {
      font-size: 12px;
      & > span {
        position: relative;
      }
    }
  }
  .icon-info {
    position: absolute;
    right: -33px;
    top: 50%;
    transform: translateY(-50%);
    font-size: 14px;
    color: #979ba5;
    cursor: pointer;
  }
  // :deep(.is-error) {
  //   &.add-label-item .bk-input {
  //     border-color: #c4c6cc;
  //   }
  // }
</style>
