<template>
  <bk-form ref="formRef" form-type="vertical" :model="localVal" :rules="rules">
    <bk-form-item :label="t('配置文件名')" property="fileAP" :required="true">
      <bk-input
        v-model="localVal.fileAP"
        :placeholder="t('请输入配置文件的完整路径和文件名，例如：/etc/nginx/nginx.conf')"
        :disabled="isEdit"
        @input="handleFileAPInput" />
    </bk-form-item>
    <bk-form-item :label="t('配置文件描述')" property="memo">
      <bk-input
        v-model="localVal.memo"
        type="textarea"
        :maxlength="200"
        :placeholder="t('请输入')"
        :resize="true"
        @input="change" />
    </bk-form-item>
    <bk-form-item :label="t('配置文件格式')">
      <bk-radio-group v-model="localVal.file_type" :required="true" @change="change">
        <bk-radio v-for="typeItem in CONFIG_FILE_TYPE" :key="typeItem.id" :label="typeItem.id" :disabled="isEdit">{{
          typeItem.name
        }}</bk-radio>
      </bk-radio-group>
    </bk-form-item>
    <div class="user-settings">
      <bk-form-item :label="t('文件权限')" property="privilege" required>
        <div class="perm-input">
          <bk-popover
            ext-cls="privilege-tips-wrap"
            theme="light"
            trigger="manual"
            placement="top"
            :is-show="showPrivilegeErrorTips">
            <bk-input
              v-model="privilegeInputVal"
              type="number"
              :placeholder="t('请输入三位权限数字')"
              @blur="handlePrivilegeInputBlur" />
            <template #content>
              <div>{{ t('只能输入三位 0~7 数字') }}</div>
              <div class="privilege-tips-btn-area">
                <bk-button text theme="primary" @click="showPrivilegeErrorTips = false">{{ t('我知道了') }}</bk-button>
              </div>
            </template>
          </bk-popover>
          <bk-popover ext-cls="privilege-select-popover" theme="light" trigger="click" placement="bottom">
            <div :class="['perm-panel-trigger']">
              <i class="bk-bscp-icon icon-configuration-line"></i>
            </div>
            <template #content>
              <div class="privilege-select-panel">
                <div v-for="(item, index) in PRIVILEGE_GROUPS" class="group-item" :key="index" :label="item">
                  <div class="header">{{ item }}</div>
                  <div class="checkbox-area">
                    <bk-checkbox-group
                      class="group-checkboxs"
                      :model-value="privilegeGroupsValue[index]"
                      @change="handleSelectPrivilege(index, $event)">
                      <bk-checkbox size="small" :label="4" :disabled="index === 0">
                        {{ t('读') }}
                      </bk-checkbox>
                      <bk-checkbox size="small" :label="2">{{ t('写') }}</bk-checkbox>
                      <bk-checkbox size="small" :label="1">{{ t('执行') }}</bk-checkbox>
                    </bk-checkbox-group>
                  </div>
                </div>
              </div>
            </template>
          </bk-popover>
        </div>
      </bk-form-item>
      <bk-form-item :label="t('用户')" property="user" :required="true">
        <bk-input v-model="localVal.user" :placeholder="t('请输入')" @input="change"></bk-input>
      </bk-form-item>
      <bk-form-item :label="t('用户组')" :placeholder="t('请输入')" property="user_group" :required="true">
        <bk-input v-model="localVal.user_group" @input="change"></bk-input>
      </bk-form-item>
    </div>
    <bk-form-item v-if="localVal.file_type === 'binary'" :label="t('配置内容')" :required="true">
      <bk-upload
        class="config-uploader"
        url=""
        theme="button"
        :tip="t('文件大小{size}M以内', { size: props.fileSizeLimit })"
        :size="100000"
        :multiple="false"
        :custom-request="handleFileUpload">
      </bk-upload>
      <bk-loading
        v-if="uploadFile"
        mode="spin"
        theme="primary"
        :opacity="0.6"
        size="mini"
        :title="t('文件下载中，请稍后')"
        :loading="fileDownloading"
        class="file-down-loading">
        <div :class="['file-wrapper', { 'upload-fail': uploadFile.status === 'fail' }]" @click="handleDownloadFile">
          <TextFill class="file-icon" />
          <div class="file-content">
            <div class="name" :title="uploadFile?.file.name">{{ uploadFile?.file.name }}</div>
            <div v-if="uploadFile.status === 'checking'" class="check-status">
              <Spinner class="spinner-icon" /> {{ $t('文件上传准备中，请稍候…') }}
            </div>
            <div v-else-if="uploadProgress.status === 'uploading'">
              <bk-progress
                :percent="uploadProgress.percent"
                :theme="uploadFile.status === 'fail' ? 'danger' : 'primary'"
                size="small"
                :show-text="false" />
            </div>
            <div v-else :class="[uploadFile.status === 'success' ? 'success-text' : 'error-text', 'status-icon-area']">
              <Done v-if="uploadFile.status === 'success'" class="success-icon" />
              <Error v-if="uploadFile.status === 'fail'" class="error-icon" />
              <span :class="[uploadFile.status === 'success' ? 'success-text' : 'error-text']">
                {{ uploadFile.status === 'success' ? t('上传成功') : `${t('上传失败')} ${uploadFile.errorMessage}` }}
                <span v-if="uploadFile.status === 'success' && uploadFile.isExist">
                  {{ $t('( 后台已存在此文件，上传快速完成 )') }}
                </span>
              </span>
            </div>
          </div>
          <span class="size">
            ({{ byteUnitConverse(uploadFile.file.size) }})
            <span v-if="uploadProgress.status === 'uploading'">{{ `${uploadProgress.percent}%` }}</span>
          </span>
        </div>
      </bk-loading>
    </bk-form-item>
    <bk-form-item v-else>
      <template #label>
        <div class="config-content-label">
          <span>{{ t('配置内容') }}</span>
          <info v-bk-tooltips="{ content: t('tips.createConfig'), placement: 'top' }" fill="#3a84ff" />
        </div>
      </template>
      <ConfigContentEditor
        :content="stringContent"
        :editable="true"
        :variables="props.variables"
        :size-limit="props.fileSizeLimit"
        @change="handleStringContentChange" />
    </bk-form-item>
  </bk-form>
</template>
<script setup lang="ts">
  import { ref, computed, watch, onMounted } from 'vue';
  import { useI18n } from 'vue-i18n';
  import SHA256 from 'crypto-js/sha256';
  import WordArray from 'crypto-js/lib-typedarrays';
  import CryptoJS from 'crypto-js';
  import { TextFill, Done, Info, Error, Spinner } from 'bkui-vue/lib/icon';
  import BkMessage from 'bkui-vue/lib/message';
  import { cloneDeep } from 'lodash';
  import { IConfigEditParams, IFileConfigContentSummary } from '../../../../../../../../types/config';
  import { IVariableEditParams } from '../../../../../../../../types/variable';
  import {
    updateConfigContent,
    downloadConfigContent,
    getConfigUploadFileIsExist,
  } from '../../../../../../../api/config';
  import {
    downloadTemplateContent,
    updateTemplateContent,
    getTemplateUploadFileIsExist,
  } from '../../../../../../../api/template';
  import { stringLengthInBytes, byteUnitConverse } from '../../../../../../../utils/index';
  import { fileDownload } from '../../../../../../../utils/file';
  import { CONFIG_FILE_TYPE } from '../../../../../../../constants/config';
  import ConfigContentEditor from '../../components/config-content-editor.vue';

  interface IUploadFile {
    file: any;
    status: string;
    isExist: boolean;
    errorMessage?: string;
  }

  const { t } = useI18n();

  const PRIVILEGE_GROUPS = [t('属主（own）'), t('属组（group）'), t('其他人（other）')];
  const PRIVILEGE_VALUE_MAP = {
    0: [],
    1: [1],
    2: [2],
    3: [1, 2],
    4: [4],
    5: [1, 4],
    6: [2, 4],
    7: [1, 2, 4],
  };

  const props = withDefaults(
    defineProps<{
      config: IConfigEditParams;
      isEdit: boolean;
      content?: string | IFileConfigContentSummary;
      variables?: IVariableEditParams[];
      bkBizId: string;
      id: number; // 服务ID或者模板空间ID
      fileUploading?: boolean;
      fileSizeLimit?: number;
      isTpl?: boolean; // 是否未模板配置文件，非模板配置文件和模板配置文件的上传、下载接口参数有差异
    }>(),
    {
      isEdit: false,
      fileSizeLimit: 100,
    },
  );

  const emits = defineEmits(['change', 'update:fileUploading']);
  const localVal = ref({ ...props.config, fileAP: '' });
  const privilegeInputVal = ref('');
  const showPrivilegeErrorTips = ref(false);
  const stringContent = ref('');
  const fileContent = ref<IFileConfigContentSummary | File>();
  const uploadFileSignature = ref(''); // 新上传文件的sha256
  const isFileChanged = ref(false); // 标识文件是否被修改，编辑配置文件时若文件未修改，不重新上传文件
  const formRef = ref();
  const uploadProgress = ref({
    percent: 0,
    status: '',
  });
  const fileDownloading = ref(false);
  const uploadFile = ref<IUploadFile>();
  const rules = {
    // 配置文件名校验规则，path+filename
    fileAP: [
      {
        validator: (val: string) => {
          // 必须为绝对路径, 且不能以/结尾
          if (!val.startsWith('/') || val.endsWith('/')) {
            return false;
          }

          const parts = val.split('/').slice(1);
          const fileName = parts.pop() as string;

          // 文件名称校验
          // 文件名和路径不能全由.组成
          if (!/^((?!\.{1,}$)[\u4e00-\u9fa5A-Za-z0-9.\-_#%,:?!@$^+=\\[\]{}]+)$/.test(fileName)) {
            return false;
          }

          let isValid = true;
          // 文件路径校验
          parts.some((part) => {
            if (!/^((?!\.{1,}$)[\u4e00-\u9fa5A-Za-z0-9.\-_#%,:?!@$^+=\\[\]{}]+)$/.test(part)) {
              isValid = false;
              return true;
            }
            return false;
          });

          return isValid;
        },
        message: t('无效的路径,路径不符合Unix文件路径格式规范'),
        trigger: 'change',
      },
    ],
    privilege: [
      {
        required: true,
        validator: () => {
          const type = typeof privilegeInputVal.value;
          return type === 'number' || (type === 'string' && privilegeInputVal.value.length > 0);
        },
        message: t('文件权限 不能为空'),
        trigger: 'change',
      },
      {
        validator: () => {
          const privilege = parseInt(privilegeInputVal.value[0], 10);
          return privilege >= 4;
        },
        message: t('文件own必须有读取权限'),
        trigger: 'blur',
      },
    ],
    memo: [
      {
        validator: (value: string) => value.length <= 200,
        message: t('最大长度200个字符'),
      },
    ],
    revision_name: [
      {
        validator: (value: string) => value.length <= 128,
        message: t('最大长度128个字符'),
      },
      {
        validator: (value: string) => {
          if (value.length > 0) {
            return /^[\u4e00-\u9fa5a-zA-Z0-9][\u4e00-\u9fa5a-zA-Z0-9_-]*[\u4e00-\u9fa5a-zA-Z0-9]?$/.test(value);
          }
          return true;
        },
        message: t('仅允许使用中文、英文、数字、下划线、中划线，且必须以中文、英文、数字开头和结尾'),
      },
    ],
  };

  // 传入到bk-upload组件的文件对象
  // const fileList = computed(() => (fileContent.value ? [transFileToObject(fileContent.value as File)] : []));

  // 将权限数字拆分成三个分组配置
  const privilegeGroupsValue = computed(() => {
    const data: { [index: string]: number[] } = { 0: [], 1: [], 2: [] };
    if (typeof localVal.value.privilege === 'string' && localVal.value.privilege.length > 0) {
      const valArr = localVal.value.privilege.split('').map((i) => parseInt(i, 10));
      valArr.forEach((item, index) => {
        data[index as keyof typeof data] = PRIVILEGE_VALUE_MAP[item as keyof typeof PRIVILEGE_VALUE_MAP];
      });
    }
    return data;
  });

  watch(
    () => props.config.privilege,
    (val) => {
      privilegeInputVal.value = val as string;
    },
    { immediate: true },
  );

  watch(
    () => props.config,
    () => {
      const { path, name } = props.config;
      if (!path) return;
      localVal.value.fileAP = path.endsWith('/') ? `${path}${name}` : `${path}/${name}`;
    },
    { immediate: true, deep: true },
  );

  onMounted(() => {
    if (props.config.file_type === 'binary') {
      fileContent.value = cloneDeep(props.content as IFileConfigContentSummary);
      if (props.isEdit) {
        if (fileContent.value.signature) {
          uploadFile.value = {
            file: { ...fileContent.value },
            status: 'success',
            isExist: false,
          };
          uploadFileSignature.value = fileContent.value.signature;
        }
      }
    } else {
      stringContent.value = props.content as string;
    }
  });

  // 权限输入框失焦后，校验输入是否合法，如不合法回退到上次输入
  const handlePrivilegeInputBlur = () => {
    const val = String(privilegeInputVal.value);
    if (/^[0-7]{3}$/.test(val)) {
      localVal.value.privilege = val;
      showPrivilegeErrorTips.value = false;
      change();
    } else {
      privilegeInputVal.value = String(localVal.value.privilege);
      showPrivilegeErrorTips.value = true;
    }
  };

  // 选择文件权限
  const handleSelectPrivilege = (index: number, val: number[]) => {
    const groupsValue = { ...privilegeGroupsValue.value };
    groupsValue[index] = val;
    const digits = [];
    for (let i = 0; i < 3; i++) {
      let sum = 0;
      if (groupsValue[i].length > 0) {
        sum = groupsValue[i].reduce((acc, crt) => acc + crt, 0);
      }
      digits.push(sum);
    }
    const newVal = digits.join('');
    privilegeInputVal.value = newVal;
    localVal.value.privilege = newVal;
    showPrivilegeErrorTips.value = false;
    change();
  };

  const handleStringContentChange = (val: string) => {
    stringContent.value = val;
    change();
  };

  // 选择文件后上传
  const handleFileUpload = async (option: { file: File }) => {
    emits('update:fileUploading', true);
    fileContent.value = option.file;
    uploadFile.value = {
      file: option.file,
      status: 'checking',
      isExist: false,
    };
    const fileSize = option.file.size / 1024 / 1024;
    if (fileSize > props.fileSizeLimit) {
      uploadFile.value!.status = 'fail';
      uploadFile.value.errorMessage = t('请确保文件大小不超过 {n} MB', { n: props.fileSizeLimit });
      return;
    }
    isFileChanged.value = true;
    if (localVal.value.fileAP === '') {
      localVal.value.fileAP = `/${option.file.name}`;
    }
    // 文件存在 无需重复上传
    const res = await checkFileExist();
    if (res.exists) {
      uploadFile.value.status = 'success';
      uploadFile.value.isExist = true;
      fileContent.value = {
        name: option.file.name,
        signature: res.metadata.sha256,
        size: res.metadata.byte_size,
      };
      change();
      emits('update:fileUploading', false);
      return Promise.resolve();
    }
    uploadFile.value.status = 'uploading';
    uploadFile.value.isExist = false;
    return new Promise((resolve, reject) => {
      uploadContent()
        .then((res) => {
          uploadFile.value!.status = 'success';
          fileContent.value = {
            name: option.file.name,
            signature: res.sha256,
            size: res.byte_size,
          };
          change();
          resolve(res);
        })
        .catch((err) => {
          console.error(err);
          uploadFile.value!.status = 'fail';
          uploadFile.value!.errorMessage = '';
          reject(err);
        })
        .finally(() => {
          emits('update:fileUploading', false);
          uploadProgress.value.status = 'success';
          uploadProgress.value.percent = 0;
        });
    });
  };

  // 上传配置内容
  const uploadContent = async () => {
    uploadProgress.value.status = 'uploading';
    if (props.isTpl) {
      return updateTemplateContent(
        props.bkBizId,
        props.id,
        fileContent.value as File,
        uploadFileSignature.value,
        (progress: number) => {
          uploadProgress.value.percent = progress;
        },
      );
    }
    return updateConfigContent(
      props.bkBizId,
      props.id,
      fileContent.value as File,
      uploadFileSignature.value,
      (progress: number) => {
        uploadProgress.value.percent = progress;
      },
    );
  };

  // 判断上传的文件是否存在
  const checkFileExist = async () => {
    const signature = await getSignature();
    uploadFileSignature.value = signature;
    if (props.isTpl) {
      return getTemplateUploadFileIsExist(props.bkBizId, props.id, signature);
    }
    return getConfigUploadFileIsExist(props.bkBizId, props.id, signature);
  };

  // 生成文件或文本的sha256
  const getSignature = async () => {
    if (localVal.value.file_type === 'binary') {
      const CHUNK_SIZE = 1024 * 1024; // 1MB
      // 初始化第一个切片的处理
      let start = 0;
      let end = Math.min(CHUNK_SIZE, fileContent.value!.size as number);
      if (isFileChanged.value) {
        return new Promise((resolve) => {
          const reader = new FileReader();
          const hash = CryptoJS.algo.SHA256.create();
          const processChunk = () => {
            // @ts-ignore
            const slice = fileContent.value.slice(start, end);
            reader.readAsArrayBuffer(slice);
          };
          reader.onload = function () {
            const wordArray = WordArray.create(reader.result);
            hash.update(wordArray);
            if (end < (fileContent.value!.size as number)) {
              start += CHUNK_SIZE;
              end = Math.min(start + CHUNK_SIZE, fileContent.value!.size as number);
              processChunk();
            } else {
              const sha256Hash = hash.finalize();
              resolve(sha256Hash.toString());
            }
          };
          // 开始处理第一个切片
          processChunk();
        });
      }
      return (fileContent.value as IFileConfigContentSummary).signature;
    }
    if (!stringContent.value.endsWith('\n')) stringContent.value += '\n';
    return SHA256(stringContent.value).toString();
  };

  // 下载已上传文件
  const handleDownloadFile = async () => {
    if (uploadProgress.value.status === 'uploading') return;
    try {
      fileDownloading.value = true;
      const { signature, name } = fileContent.value as IFileConfigContentSummary;
      const fileSignature = signature || uploadFileSignature.value;
      const getContent = props.isTpl ? downloadTemplateContent : downloadConfigContent;
      const res = await getContent(props.bkBizId, props.id, fileSignature, true);
      fileDownload(res, name);
    } catch (error) {
      console.error(error);
    } finally {
      fileDownloading.value = false;
    }
  };

  const validate = async () => {
    await formRef.value.validate();
    if (localVal.value.file_type === 'binary') {
      if (!uploadFile.value) {
        BkMessage({ theme: 'error', message: t('请上传文件') });
        return false;
      }
      if (uploadFile.value.status === 'fail') {
        BkMessage({ theme: 'error', message: t('文件上传失败，请重新上传文件') });
        return false;
      }
    } else if (localVal.value.file_type === 'text') {
      if (stringLengthInBytes(stringContent.value) > 1024 * 1024 * props.fileSizeLimit) {
        BkMessage({ theme: 'error', message: t('配置内容不能超过{size}M', { size: props.fileSizeLimit }) });
        return false;
      }
    }
    return true;
  };

  const change = () => {
    const content = localVal.value.file_type === 'binary' ? fileContent.value : stringContent.value;
    const { fileAP } = localVal.value;
    const lastSlashIndex = fileAP.lastIndexOf('/');
    localVal.value.name = fileAP.slice(lastSlashIndex + 1);
    localVal.value.path = fileAP.slice(0, lastSlashIndex + 1);
    emits('change', localVal.value, content);
  };

  const handleFileAPInput = () => {
    // 用户输入文件名 补全路径
    if (localVal.value.fileAP && !localVal.value.fileAP.startsWith('/')) {
      localVal.value.fileAP = `/${localVal.value.fileAP}`;
    }
    change();
  };

  defineExpose({
    getSignature: () => {
      if (localVal.value.file_type === 'binary') {
        return uploadFileSignature.value;
      }
      return getSignature();
    },
    validate,
  });
</script>
<style lang="scss" scoped>
  .user-settings {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
  }
  .perm-input {
    display: flex;
    align-items: center;
    width: 172px;
    :deep(.bk-input) {
      width: 140px;
      border-right: none;
      border-top-right-radius: 0;
      border-bottom-right-radius: 0;
      .bk-input--number-control {
        display: none;
      }
    }
    .perm-panel-trigger {
      width: 32px;
      height: 32px;
      text-align: center;
      background: #fafcfe;
      color: #3a84ff;
      border: 1px solid #3a84ff;
      cursor: pointer;
      &.disabled {
        color: #dcdee5;
        border-color: #dcdee5;
        cursor: not-allowed;
      }
    }
  }

  .privilege-tips-btn-area {
    margin-top: 8px;
    text-align: right;
  }
  .privilege-select-panel {
    display: flex;
    align-items: top;
    border: 1px solid #dcdee5;
    .group-item {
      .header {
        padding: 0 16px;
        height: 42px;
        line-height: 42px;
        color: #313238;
        font-size: 12px;
        background: #fafbfd;
        border-bottom: 1px solid #dcdee5;
      }
      &:not(:last-of-type) {
        .header,
        .checkbox-area {
          border-right: 1px solid #dcdee5;
        }
      }
    }
    .checkbox-area {
      padding: 10px 16px 12px;
      background: #ffffff;
      &:not(:last-child) {
        border-right: 1px solid #dcdee5;
      }
    }
    .group-checkboxs {
      font-size: 12px;
      .bk-checkbox ~ .bk-checkbox {
        margin-left: 16px;
      }
      :deep(.bk-checkbox-label) {
        font-size: 12px;
      }
    }
  }
  :deep(.config-uploader) {
    .bk-upload-list {
      display: none;
    }
  }
  .file-wrapper {
    margin: 8px 0;
    position: relative;
    display: flex;
    align-items: center;
    color: #979ba5;
    font-size: 12px;
    line-height: 12px;
    width: 100%;
    border: 1px solid #c4c6cc;
    padding: 10px;
    cursor: pointer;
    &.upload-fail {
      border-color: #ff5656;
      background: #fedddc66;
    }
    &:hover .name {
      color: #3a84ff;
    }
    .file-content {
      width: 400px;
      line-height: 20px;
      .spinner-icon {
        font-size: 14px;
        color: #3a84ff;
      }
    }
    .status-icon-area {
      display: flex;
      align-items: center;
      &.success-text {
        color: #2dcb56;
      }
      &.error-text {
        color: #ea3636;
      }
      .success-icon {
        font-size: 20px;
      }
      .error-icon {
        font-size: 14px;
      }
    }
    .file-icon {
      margin: 0 6px 0 0;
      font-size: 32px;
    }
    .name {
      max-width: 360px;
      white-space: nowrap;
      text-overflow: ellipsis;
      overflow: hidden;
      line-height: normal;
    }
    .size {
      position: absolute;
      right: 10px;
      top: 50%;
      transform: translateY(-50%);
    }
  }
  .config-content-label {
    display: flex;
    align-items: center;
    span {
      margin-right: 5px;
    }
  }
  .file-down-loading {
    width: 100%;
    :deep(.bk-loading-indicator) {
      align-items: center;
      flex-direction: row;
      .bk-loading-title {
        margin-top: 0px;
        margin-left: 8px;
        color: #979ba5;
        font-size: 12px;
      }
    }
  }
</style>
<style lang="scss">
  .privilege-select-popover.bk-popover {
    padding: 0;
    .bk-pop2-arrow {
      border-left: 1px solid #dcdee5;
      border-top: 1px solid #dcdee5;
    }
  }
  .privilege-tips-wrap {
    border: 1px solid #dcdee5;
    .bk-pop2-arrow {
      border-right: 1px solid #dcdee5;
      border-bottom: 1px solid #dcdee5;
    }
  }
</style>
