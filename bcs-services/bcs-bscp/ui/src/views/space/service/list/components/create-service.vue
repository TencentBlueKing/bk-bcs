<template>
  <bk-sideslider
    width="640"
    :is-show="props.show"
    :title="t('新建服务')"
    :before-close="handleBeforeClose"
    @closed="close"
  >
    <div class="create-app-form">
      <SearviceForm ref="formCompRef" :form-data="serviceData" @change="handleChange" />
    </div>
    <div class="create-app-footer">
      <bk-button theme="primary" :loading="pending" @click="handleCreateConfirm">
        {{ t('提交') }}
      </bk-button>
      <bk-button @click="close">{{ t('取消') }}</bk-button>
    </div>
  </bk-sideslider>
  <bk-dialog ext-cls="confirm-dialog" :is-show="isShowConfirmDialog" :show-mask="true" :quick-close="false">
    <div class="title-icon"><Done fill="#42C06A" /></div>
    <div class="title-info">服务新建成功</div>
    <div class="content-info">
      {{ serviceData.config_type === 'file' ? '接下来你可以在服务下新增配置文件' : '接下来你可以在服务下新增配置项' }}
    </div>
    <div class="footer-btn">
      <bk-button theme="primary" @click="handleGoCreateConfig" style="margin-right: 8px">
        {{ serviceData.config_type === 'file' ? '新增配置文件' : '新增配置项' }}
      </bk-button>
      <bk-button @click="isShowConfirmDialog = false">稍后再说</bk-button>
    </div>
  </bk-dialog>
</template>
<script setup lang="ts">
import { ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { storeToRefs } from 'pinia';
import useGlobalStore from '../../../../../store/global';
import { createApp } from '../../../../../api';
import { IServiceEditForm } from '../../../../../../types/service';
import { Done } from 'bkui-vue/lib/icon';
import useModalCloseConfirmation from '../../../../../utils/hooks/use-modal-close-confirmation';
import SearviceForm from './service-form.vue';

const router = useRouter();
const { t } = useI18n();

const props = defineProps<{
  show: boolean;
}>();
const emits = defineEmits(['update:show', 'reload']);

const { spaceId } = storeToRefs(useGlobalStore());

const serviceData = ref<IServiceEditForm>({
  name: '',
  alias: '',
  config_type: 'file',
  data_type: 'any',
  reload_type: 'file',
  reload_file_path: '/data/reload.json',
  mode: 'normal',
  memo: '', // @todo 包含换行符后接口会报错
});
const formCompRef = ref();
const pending = ref(false);
const isFormChange = ref(false);
const isShowConfirmDialog = ref(false);
const appId = ref();

watch(
  () => props.show,
  (val) => {
    if (val) {
      isFormChange.value = false;
      serviceData.value = {
        name: '',
        alias: '',
        config_type: 'file',
        data_type: 'any',
        reload_type: 'file',
        reload_file_path: '/data/reload.json',
        mode: 'normal',
        memo: '',
      };
    }
  },
);

const handleChange = (val: IServiceEditForm) => {
  isFormChange.value = true;
  serviceData.value = val;
};

const handleCreateConfirm = async () => {
  await formCompRef.value.validate();
  pending.value = false;
  try {
    let resp: { id: number };
    if (serviceData.value.config_type === 'file') {
      resp = await createApp(spaceId.value, serviceData.value);
    } else {
      resp = await createApp(spaceId.value, { ...serviceData.value, reload_type: '', reload_file_path: '' });
    }
    appId.value = resp.id;
    emits('reload');
    isShowConfirmDialog.value = true;
    close();
  } catch (e) {
    console.error(e);
  } finally {
    pending.value = false;
  }
};

const handleBeforeClose = async () => {
  if (isFormChange.value) {
    const result = await useModalCloseConfirmation();
    return result;
  }
  return true;
};

const handleGoCreateConfig = () => {
  router.push({
    name: 'service-config',
    params: {
      spaceId: spaceId.value,
      appId: appId.value,
    },
  });
};

const close = () => {
  emits('update:show', false);
};
</script>
<style lang="scss" scoped>
.create-app-form {
  padding: 20px 24px;
  height: calc(100vh - 101px);
}
.create-app-footer {
  padding: 8px 24px;
  height: 48px;
  width: 100%;
  background: #fafbfd;
  border-top: 1px solid #dcdee5;
  box-shadow: none;
  button {
    margin-right: 8px;
    min-width: 88px;
  }
}

.confirm-dialog {
  :deep(.bk-modal-body) {
    width: 400px;
    padding: 0;
    .bk-modal-header {
      display: none;
    }
    .bk-modal-footer {
      display: none;
    }
    .bk-modal-content {
      display: flex;
      flex-direction: column;
      align-items: center;
      .title-icon {
        margin: 27px 0 19px;
        width: 42px;
        height: 42px;
        border-radius: 50%;
        font-size: 42px;
        line-height: 42px;
        background-color: #e5f6e8;
      }
      .title-info {
        height: 32px;
        font-size: 20px;
        color: #313238;
        text-align: center;
        line-height: 32px;
      }
      .content-info {
        margin-top: 8px;
        height: 22px;
        font-size: 14px;
        color: #63656e;
        line-height: 22px;
      }
      .footer-btn {
        margin: 24px 0;
      }
    }
  }
}
</style>
