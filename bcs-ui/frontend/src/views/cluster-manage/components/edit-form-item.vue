<template>
  <div class="flex items-center">
    <template v-if="!isEdit">
      <span class="break-all">{{ value || '--' }}</span>
      <span
        class="hover:text-[#3a84ff] cursor-pointer ml-[8px]"
        v-if="!disableEdit"
        @click="handleEdit">
        <i class="bk-icon icon-edit-line"></i>
      </span>
    </template>
    <template v-else>
      <Validate :rules="rules" :value="innerValue" class="flex-1 max-w-[300px]" ref="validateRef">
        <bcs-input
          :maxlength="maxlength"
          :placeholder="placeholder"
          v-model="innerValue"
          ref="inputRef"
          :type="type"
          clearable
          v-clickoutside="handleSave"
          @enter="handleEnter"
          @change="handleChange">
        </bcs-input>
      </Validate>
    </template>
  </div>
</template>
<script lang="ts">
import { defineComponent, PropType, ref, toRefs, watch } from 'vue';

import Validate, { IValidate } from '@/components/validate.vue';
import clickoutside from '@/directives/clickoutside';

export default defineComponent({
  name: 'EditFormItem',
  directives: {
    clickoutside,
  },
  components: { Validate },
  props: {
    value: {
      type: String,
      default: '',
    },
    type: {
      type: String,
      default: 'text',
    },
    placeholder: {
      type: String,
      default: '',
    },
    maxlength: Number,
    editable: {
      type: Boolean,
      default: false,
    },
    rules: {
      type: Array as PropType<IValidate[]>,
      default: () => [],
    },
    disableEdit: {
      type: Boolean,
      default: false,
    },
  },
  setup(props, ctx) {
    const { editable, value, type } = toRefs(props);
    watch(editable, () => {
      isEdit.value = editable.value;
    });
    watch(value, () => {
      innerValue.value = value.value;
    });
    const validateRef = ref<any>(null);
    const isEdit = ref(editable.value);
    const inputRef = ref<any>(null);
    const handleEdit = () => {
      isEdit.value = true;
      innerValue.value = value.value;
      setTimeout(() => {
        inputRef.value?.focus();
      });
    };
    const innerValue = ref(value.value);
    const handleChange = (value) => {
      innerValue.value = value;
      ctx.emit('input', value);
    };
    const handleSave = async () => {
      const result = await validateRef.value?.validate();
      if (!result) return;

      isEdit.value = false;
      // 值未变更不做保存
      if (innerValue.value === value.value) return;

      ctx.emit('save', innerValue.value);
    };
    const handleCancel = () => {
      isEdit.value = false;
      innerValue.value = value.value;
      ctx.emit('cancel');
    };
    const handleEnter = () => {
      if (type.value === 'textarea') return;

      handleSave();
    };

    return {
      isEdit,
      innerValue,
      validateRef,
      inputRef,
      handleEdit,
      handleChange,
      handleSave,
      handleCancel,
      handleEnter,
    };
  },
});
</script>
<style lang="postcss" scoped>
>>> textarea::-webkit-scrollbar {
  width: 4px;
}

>>> textarea::-webkit-scrollbar-thumb {
  background: #ddd;
  border-radius: 20px;
}
</style>
