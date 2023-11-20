<template>
  <bk-sideslider
    width="640"
    :is-show="props.show"
    :title="t('新建服务')"
    :before-close="handleBeforeClose"
    @closed="close">
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
</template>
<script setup lang="ts">
import { ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { InfoBox } from 'bkui-vue/lib';
import { storeToRefs } from 'pinia';
import useGlobalStore from '../../../../../store/global';
import { createApp } from '../../../../../api';
import { IServiceEditForm } from '../../../../../../types/service';
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
  config_type: 'file',
  kv_type: '',
  reload_type: 'file',
  reload_file_path: '/bscp_test', // @todo 待确认
  mode: 'normal',
  deploy_type: 'common', // @todo 待和后台确认是否需要
  memo: '', // @todo 包含换行符后接口会报错
});
const formCompRef = ref();
const pending = ref(false);
const isFormChange = ref(false);

watch(
  () => props.show,
  (val) => {
    if (val) {
      isFormChange.value = false;
      serviceData.value = {
        name: '',
        config_type: 'file',
        kv_type: '',
        reload_type: 'file',
        reload_file_path: '/bscp_test',
        mode: 'normal',
        deploy_type: 'common',
        memo: '',
      };
    }
  },
);

const handleChange = (val: IServiceEditForm) => {
  isFormChange.value = true
  serviceData.value = val
};

const handleCreateConfirm = async () => {
  await formCompRef.value.validate();
  debugger
  pending.value = false;
  try {
    const resp = await createApp(spaceId.value, serviceData.value);
    InfoBox({
      type: 'success',
      title: '服务新建成功',
      subTitle: '接下来你可以在服务下新增并使用配置文件',
      headerAlign: 'center',
      footerAlign: 'center',
      confirmText: '新增配置文件',
      cancelText: '稍后再说',
      onConfirm() {
        router.push({
          name: 'service-config',
          params: {
            spaceId: spaceId.value,
            appId: resp.id,
          },
        });
      },
      onClosed() {
        emits('reload');
      },
    } as any);
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
</style>
