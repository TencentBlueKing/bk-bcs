<template>
  <!-- 添加子网 -->
  <bk-dialog
    :is-show="modelValue"
    :title="$t('tke.title.addSubnets')"
    :width="width || 480"
    @cancel="cancel">
    <bk-form
      form-type="vertical"
      :ref="(v) => formRef = v"
      :rules="subnetsRules">
      <bk-form-item
        label="Pod IP"
        property="podIP"
        error-display-type="normal"
        required>
        <VpcCni
          :subnets="newSubnets"
          :region="clusterData.region"
          :cloud-account-i-d="clusterData.cloudAccountID"
          :cloud-i-d="clusterData.provider"
          @change="handleSetSubnets" />
      </bk-form-item>
    </bk-form>
    <template #footer>
      <div>
        <bk-button
          :disabled="!isSubnetsValidate"
          :loading="pending"
          theme="primary"
          @click="handleAddSubnets">{{ $t('generic.button.confirm') }}</bk-button>
        <bk-button @click="cancel">{{ $t('generic.button.cancel') }}</bk-button>
      </div>
    </template>
  </bk-dialog>
</template>
<script setup lang="ts">
import { computed, ref, watch } from 'vue';

import { addSubnets } from '@/api/modules/cluster-manager';
import { ICluster } from '@/composables/use-app';
import $i18n from '@/i18n/i18n-setup';
import VpcCni from '@/views/cluster-manage/components/vpc-cni.vue';

interface Props {
  modelValue: boolean
  clusterData: ICluster
  width?: number
  validateMsg?: string
  confirmFn?: Function
  validateFn?: Function
}

const props = defineProps<Props>();
const emits = defineEmits(['cancel', 'confirm']);
const formRef = ref();

const tableKey = ref('');
const subnetsRules = computed(() => ({
  podIP: [
    {
      trigger: 'change',
      validator: () => {
        if (props?.validateFn) {
          return props?.validateFn?.(newSubnets.value);
        }
        return true;
      },
      message: props.validateMsg || $i18n.t('generic.validate.required'),
    },
  ],
}));

// 添加子网
const newSubnets = ref([{
  ipCnt: '',
  zone: '',
}]);
const isSubnetsValidate = computed(() => newSubnets.value.every(item => item.ipCnt && item.zone));
const handleSetSubnets = (data) => {
  newSubnets.value = data;
};
const pending = ref(false);
const handleAddSubnets = async () => {
  const valid = await formRef.value?.validate()
    .then(() => true)
    .catch(() => false);
  if (!valid) {
    return;
  }
  pending.value = true;
  if (props.confirmFn) {
    await props.confirmFn(newSubnets.value.filter(item => item.ipCnt && item.zone));
  } else {
    const result = await addSubnets({
      $clusterId: props.clusterData?.clusterID,
      subnet: {
        new: newSubnets.value,
      },
    }).then(() => true)
      .catch(() => false);
    if (result) {
      emits('confirm');
    }
  }
  pending.value = false;
};

// 取消
function cancel() {
  emits('cancel');
}

watch(() => props.modelValue, () => {
  tableKey.value = new Date().getTime()
    .toString();
});
</script>
