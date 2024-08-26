<template>
  <div
    :class="[
      'flex items-center mb-[8px]',
      {
        'px-[12px] bg-[#F5F7FA] h-[32px] rounded-sm hover:bg-[#F0F1F5]': !edit
      }
    ]">
    <!-- 编辑态 -->
    <template v-if="edit">
      <Validate
        class="flex-1 z-10"
        :value="label.key"
        :rules="[{
          message: $i18n.t('generic.validate.labelKey'),
          validator: '^[A-Za-z0-9._/-]+$',
        }]"
        required
        ref="labelValidateRef">
        <bcs-input class="flex-1" ref="labelKeyRef" v-model="label.key"></bcs-input>
      </Validate>
      <span class="text-[#FF9C01] px-[8px]">{{ label.operator || '=' }}</span>
      <Validate
        class="flex-1 hover:z-20 ml-[-1px]"
        :value="label.value"
        :rules="[{
          message: $i18n.t('generic.validate.labelKey'),
          validator: '^[A-Za-z0-9._/-]+$',
        }]"
        required
        ref="valueValidateRef">
        <bcs-input class="flex-1" v-model="label.value"></bcs-input>
      </Validate>
      <bcs-icon
        type="check-1"
        class="text-[#2DCB56] ml-[12px] !text-[24px] cursor-pointer"
        @click="handleAddLabel" />
      <bcs-icon
        type="close"
        class="text-[#C4C6CC] ml-[12px] !text-[24px] cursor-pointer"
        @click="handleCancelAddLabel" />
    </template>
    <!-- 查看态 -->
    <template v-else>
      <span class="flex-1" v-bk-overflow-tips>{{ label.key }}</span>
      <span
        :class="[
          'flex items-center justify-center rounded-sm px-[8px]',
          'h-[24px] text-[#FF9C01] mx-[12px] bg-[#fff]'
        ]">
        {{ label.operator || '=' }}
      </span>
      <span class="flex-1 bcs-ellipsis" v-bk-overflow-tips>{{ label.value }}</span>
      <!-- <i
        class="bk-icon icon-edit-line text-[16px] text-[#979BA5] cursor-pointer"
        @click="handleEditLabel">
      </i> -->
      <i
        class="bcs-icon bcs-icon-close-5 text-[14px] text-[#979BA5] cursor-pointer ml-[20px]"
        v-if="deleteable"
        @click="handleDeleteLabel">
      </i>
    </template>
  </div>
</template>
<script setup lang="ts">
import { onMounted, PropType, ref, watch } from 'vue';

import Validate from '@/components/validate.vue';

const props = defineProps({
  value: {
    type: Object as PropType<{key: string, value: string, operator?: string}>,
    default: () => ({}),
  },
  editable: {
    type: Boolean,
    default: false,
  },
  deleteable: {
    type: Boolean,
    default: true,
  },
});

const emits = defineEmits(['input', 'confirm', 'cancel', 'delete']);

const edit = ref(props.editable);
const label = ref(props.value);

watch(() => props.value, () => {
  label.value = props.value;
});

const handleChange = () => {
  emits('input', label.value);
};

const labelValidateRef = ref();
const valueValidateRef = ref();
const handleAddLabel = async () => {
  const labelValidateResult = await labelValidateRef.value?.validate('blur');
  const valueValidateResult = await valueValidateRef.value?.validate('blur');
  if (!labelValidateResult || !valueValidateResult) return;

  emits('confirm', label.value);
  handleChange();
};
const handleCancelAddLabel = () => {
  emits('cancel');
};
// const handleEditLabel = () => {
//   edit.value = true;
// };
const handleDeleteLabel = () => {
  emits('delete');
};

const labelKeyRef = ref();
onMounted(() => {
  labelKeyRef.value?.focus();
});
</script>
