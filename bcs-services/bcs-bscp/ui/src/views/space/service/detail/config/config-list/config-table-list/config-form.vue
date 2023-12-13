<template>
  <bk-form ref="formRef" form-type="vertical" :model="localVal" :rules="rules">
    <bk-form-item label="配置文件名称" property="name" :required="true">
      <bk-input
        v-model="localVal.name"
        placeholder="请输入1~64个字符，只允许英文、数字、下划线、中划线或点"
        :disabled="!editable"
        @change="change"
      />
    </bk-form-item>
    <bk-form-item label="配置文件路径" property="path" :required="true">
      <template #label>
        <span
          v-bk-tooltips="{
            content:
              '客户端拉取配置文件后存放路径为：临时目录/业务ID/服务名称/files/配置文件路径，除了配置文件路径其它参数都在客户端sidecar中配置',
            placement: 'top',
          }"
          >配置文件路径</span>
      </template>
      <bk-input v-model="localVal.path" placeholder="请输入绝对路径" :disabled="!editable" @change="change" />
    </bk-form-item>
    <bk-form-item label="配置文件描述" property="memo">
      <bk-input
        v-model="localVal.memo"
        type="textarea"
        :maxlength="200"
        :disabled="!editable"
        @change="change"
        :resize="true"
      />
    </bk-form-item>
    <bk-form-item label="配置文件格式">
      <bk-radio-group v-model="localVal.file_type" :required="true" @change="change">
        <bk-radio v-for="typeItem in CONFIG_FILE_TYPE" :key="typeItem.id" :label="typeItem.id" :disabled="!editable">{{
          typeItem.name
        }}</bk-radio>
      </bk-radio-group>
    </bk-form-item>
    <div class="user-settings">
      <bk-form-item label="文件权限" property="privilege" required>
        <div class="perm-input">
          <bk-popover
            ext-cls="privilege-tips-wrap"
            theme="light"
            trigger="manual"
            placement="top"
            :is-show="showPrivilegeErrorTips"
          >
            <bk-input
              v-model="privilegeInputVal"
              type="number"
              placeholder="请输入三位权限数字"
              :disabled="!editable"
              @blur="handlePrivilegeInputBlur"
            />
            <template #content>
              <div>只能输入三位 0~7 数字</div>
              <div class="privilege-tips-btn-area">
                <bk-button text theme="primary" @click="showPrivilegeErrorTips = false">我知道了</bk-button>
              </div>
            </template>
          </bk-popover>
          <bk-popover
            ext-cls="privilege-select-popover"
            theme="light"
            trigger="click"
            placement="bottom"
            :disabled="!editable"
          >
            <div :class="['perm-panel-trigger', { disabled: !editable }]">
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
                      @change="handleSelectPrivilege(index, $event)"
                    >
                      <bk-checkbox size="small" :label="4" :disabled="privilegeGroupsValue[0]">读</bk-checkbox>
                      <bk-checkbox size="small" :label="2">写</bk-checkbox>
                      <bk-checkbox size="small" :label="1">执行</bk-checkbox>
                    </bk-checkbox-group>
                  </div>
                </div>
              </div>
            </template>
          </bk-popover>
        </div>
      </bk-form-item>
      <bk-form-item label="用户" property="user" :required="true">
        <bk-input v-model="localVal.user" :disabled="!editable" @change="change"></bk-input>
      </bk-form-item>
      <bk-form-item label="用户组" property="user_group" :required="true">
        <bk-input v-model="localVal.user_group" :disabled="!editable" @change="change"></bk-input>
      </bk-form-item>
    </div>
    <bk-form-item v-if="localVal.file_type === 'binary'" label="配置内容" :required="true">
      <bk-upload
        class="config-uploader"
        url=""
        theme="button"
        tip="文件大小100M以内"
        :size="100"
        :disabled="!editable"
        :multiple="false"
        :files="fileList"
        :custom-request="handleFileUpload"
      >
        <template #file="{ file }">
          <div>
            <div class="file-wrapper">
              <div class="status-icon-area">
                <Done v-if="file.status === 'success'" class="success-icon" />
                <Error v-if="file.status === 'fail'" class="error-icon" />
              </div>
              <TextFill class="file-icon" />
              <div class="name" :title="file.name" @click="handleDownloadFile">{{ file.name }}</div>
              ({{ file.status === 'fail' ? byteUnitConverse(file.size) : file.size }})
            </div>
            <div v-if="file.status === 'fail'" class="error-msg">{{ file.statusText }}</div>
          </div>
        </template>
      </bk-upload>
    </bk-form-item>
    <bk-form-item v-else>
      <template #label
        ><div class="config-content-label">
          <span>配置内容</span
          ><info v-bk-tooltips="{ content: configContentTip, placement: 'top' }" fill="#3a84ff" /></div
      ></template>
      <ConfigContentEditor
        :content="stringContent"
        :editable="editable"
        :variables="props.variables"
        @change="handleStringContentChange"
      />
    </bk-form-item>
  </bk-form>
</template>
<script setup lang="ts">
import { ref, computed, watch, getCurrentInstance } from 'vue';
import SHA256 from 'crypto-js/sha256';
import WordArray from 'crypto-js/lib-typedarrays';
import { TextFill, Done, Info, Error } from 'bkui-vue/lib/icon';
import BkMessage from 'bkui-vue/lib/message';
import { IConfigEditParams, IFileConfigContentSummary } from '../../../../../../../../types/config';
import { IVariableEditParams } from '../../../../../../../../types/variable';
import { updateConfigContent, downloadConfigContent } from '../../../../../../../api/config';
import { downloadTemplateContent, updateTemplateContent } from '../../../../../../../api/template';
import { stringLengthInBytes, byteUnitConverse } from '../../../../../../../utils/index';
import { transFileToObject, fileDownload } from '../../../../../../../utils/file';
import { CONFIG_FILE_TYPE } from '../../../../../../../constants/config';
import ConfigContentEditor from '../../components/config-content-editor.vue';

const PRIVILEGE_GROUPS = ['属主（own）', '属组（group）', '其他人（other）'];
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
    editable: boolean;
    content: string | IFileConfigContentSummary;
    variables?: IVariableEditParams[];
    bkBizId: string;
    id: number; // 服务ID或者模板空间ID
    fileUploading?: boolean;
    isTpl?: boolean; // 是否未模板配置文件，非模板配置文件和模板配置文件的上传、下载接口参数有差异
  }>(),
  {
    editable: true,
  },
);

const emits = defineEmits(['change', 'update:fileUploading']);
const configContentTip = `配置文件内支持引用全局变量与定义新的BSCP变量，变量规则如下
                          1.需是要go template语法， 例如 {{ .bk_bscp_appid }}
                          2.变量名需以 “bk_bscp_” 或 “BK_BSCP_” 开头`;
const localVal = ref({ ...props.config });
const privilegeInputVal = ref('');
const showPrivilegeErrorTips = ref(false);
const stringContent = ref('');
const fileContent = ref<IFileConfigContentSummary | File>();
const isFileChanged = ref(false); // 标识文件是否被修改，编辑配置文件时若文件未修改，不重新上传文件
const uploadPending = ref(false);
const formRef = ref();
const rules = {
  name: [
    {
      validator: (value: string) => value.length <= 64,
      message: '最大长度64个字符',
    },
    {
      validator: (value: string) => /^[\u4e00-\u9fa5A-Za-z0-9_\-#%,@^+=[\]{}]+[\u4e00-\u9fa5A-Za-z0-9_\-#%,.@^+=[\]{}]*$/.test(value),
      message: '请使用中文、英文、数字、下划线、中划线或点',
    },
  ],
  privilege: [
    {
      required: true,
      validator: () => {
        const type = typeof privilegeInputVal.value;
        return type === 'number' || (type === 'string' && privilegeInputVal.value.length > 0);
      },
      message: '文件权限 不能为空',
      trigger: 'change',
    },
    {
      validator: () => {
        const privilege = parseInt(privilegeInputVal.value[0], 10);
        return privilege >= 4;
      },
      message: '文件own必须有读取权限',
      trigger: 'blur',
    },
  ],
  path: [
    {
      validator: (value: string) => value.length <= 1024,
      message: '最大长度1024个字符',
      trigger: 'change',
    },
    {
      validator: (value: string) => /^\/([\u4e00-\u9fa5A-Za-z0-9_\-#%,@^+=[\]{}]+[\u4e00-\u9fa5A-Za-z0-9_\-#%,.@^+=[\]{}]*\/?)*$/.test(value),
      message: '无效的路径,路径不符合Unix文件路径格式规范',
      trigger: 'change',
    },
  ],
  memo: [
    {
      validator: (value: string) => value.length <= 200,
      message: '最大长度200个字符',
    },
  ],
};

// 传入到bk-upload组件的文件对象
const fileList = computed(() => (fileContent.value ? [transFileToObject(fileContent.value as File)] : []));

// 将权限数字拆分成三个分组配置
const privilegeGroupsValue = computed(() => {
  const data: { [index: string]: number[] } = { 0: [], 1: [], 2: [] };
  if (typeof localVal.value.privilege === 'string' && localVal.value.privilege.length > 0) {
    const valArr = localVal.value.privilege.split('').map(i => parseInt(i, 10));
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
  () => props.content,
  () => {
    if (props.config.file_type === 'binary') {
      fileContent.value = props.content as IFileConfigContentSummary;
    } else {
      stringContent.value = props.content as string;
    }
  },
  { immediate: true },
);

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
const handleFileUpload = (option: { file: File }) => {
  isFileChanged.value = true;
  return new Promise((resolve) => {
    uploadPending.value = true;
    emits('update:fileUploading', true);
    fileContent.value = option.file;
    uploadContent().then((res) => {
      uploadPending.value = false;
      emits('update:fileUploading', false);
      change();
      resolve(res);
    });
  });
};

// 上传配置内容
const uploadContent = async () => {
  const signature = await getSignature();
  if (props.isTpl) {
    return updateTemplateContent(props.bkBizId, props.id, fileContent.value as File, signature as string);
  }
  return updateConfigContent(props.bkBizId, props.id, fileContent.value as File, signature as string);
};

// 生成文件或文本的sha256
const getSignature = async () => {
  if (localVal.value.file_type === 'binary') {
    if (isFileChanged.value) {
      return new Promise((resolve) => {
        const reader = new FileReader();
        // @ts-ignore
        reader.readAsArrayBuffer(fileContent.value);
        reader.onload = () => {
          const wordArray = WordArray.create(reader.result);
          resolve(SHA256(wordArray).toString());
        };
      });
    }
    return (fileContent.value as IFileConfigContentSummary).signature;
  }
  return SHA256(stringContent.value).toString();
};

// 下载已上传文件
const handleDownloadFile = async () => {
  const { signature, name } = fileContent.value as IFileConfigContentSummary;
  const getContent = props.isTpl ? downloadTemplateContent : downloadConfigContent;
  const res = await getContent(props.bkBizId, props.id, signature);
  fileDownload(res, `${name}.bin`);
};

const validate = async () => {
  await formRef.value.validate();
  if (localVal.value.file_type === 'binary') {
    if (fileList.value.length === 0) {
      BkMessage({ theme: 'error', message: '请上传文件' });
      return false;
    }
  } else if (localVal.value.file_type === 'text') {
    if (stringLengthInBytes(stringContent.value) > 1024 * 1024 * 50) {
      BkMessage({ theme: 'error', message: '配置内容不能超过50M' });
      return false;
    }
  }
  return true;
};
const instance = getCurrentInstance();

const change = () => {
  console.log('change', instance?.uid);
  const content = localVal.value.file_type === 'binary' ? fileContent.value : stringContent.value;
  emits('change', localVal.value, content);
};

defineExpose({
  getSignature,
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
.config-uploader {
  :deep(.bk-upload-list__item) {
    padding: 0;
    border: none;
  }
  :deep(.bk-upload-list--disabled .bk-upload-list__item) {
    pointer-events: inherit;
  }
  .file-wrapper {
    display: flex;
    align-items: center;
    color: #979ba5;
    font-size: 12px;
    .status-icon-area {
      display: flex;
      width: 20px;
      height: 100%;
      align-items: center;
      justify-content: center;
      .success-icon {
        font-size: 20px;
        color: #2dcb56;
      }
      .error-icon {
        font-size: 14px;
        color: #ea3636;
      }
    }
    .file-icon {
      margin: 0 6px 0 0;
    }
    .name {
      max-width: 360px;
      margin-right: 4px;
      color: #63656e;
      white-space: nowrap;
      text-overflow: ellipsis;
      overflow: hidden;
      cursor: pointer;
      &:hover {
        color: #3a84ff;
        text-decoration: underline;
      }
    }
  }
  .error-msg {
    padding: 0 0 10px 38px;
    line-height: 1;
    font-size: 12px;
    color: #ff5656;
  }
}
.config-content-label {
  display: flex;
  align-items: center;
  span {
    margin-right: 5px;
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
