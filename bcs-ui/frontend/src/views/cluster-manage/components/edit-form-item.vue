<template>
  <div class="flex items-center">
    <template v-if="!isEdit">
      <span>{{ value || '--' }}</span>
      <span
        class="hover:text-[#3a84ff] cursor-pointer ml-[8px]"
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
          @change="handleChange">
        </bcs-input>
      </Validate>
      <bcs-button text class="text-[12px] ml-[10px]" @click="handleSave">{{ $t('保存') }}</bcs-button>
      <bcs-button text class="text-[12px] ml-[10px]" @click="handleCancel">{{ $t('取消') }}</bcs-button>
    </template>
  </div>
</template>
<script lang="ts">
import { defineComponent, ref, toRefs, watch } from 'vue';
import Validate from '@/components/validate.vue';

export default defineComponent({
  name: 'EditFormItem',
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
      type: Array,
      default: () => [],
    },
  },
  setup(props, ctx) {
    const { editable, value } = toRefs(props);
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
      ctx.emit('save', innerValue.value);
    };
    const handleCancel = () => {
      isEdit.value = false;
      innerValue.value = value.value;
      ctx.emit('cancel');
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
    };
  },
});
</script>
