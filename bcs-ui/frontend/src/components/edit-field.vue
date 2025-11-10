<template>
  <section class="flex items-center">
    <template v-if="!isEdit">
      <span class="break-all">{{ value || '--' }}</span>
      <span
        class="hover:text-[#3a84ff] cursor-pointer ml-[8px]"
        v-if="!readonly"
        @click="setEditStatus">
        <i class="bk-icon icon-edit-line"></i>
      </span>
    </template>
    <slot v-else>
      <bcs-input
        :value="value"
        v-bind="$attrs"
        @input="handleInput"
        @blur="handleBlur" />
    </slot>
  </section>
</template>
<script lang="ts" setup>
import { ref, watch } from 'vue';

interface Props {
  readonly?: Boolean
  value?: String
  editMode?: Boolean
}

type Emits = {
  (e: 'input'|'blur', v: string): void;
  (e: 'update:editMode', v: boolean): void;
};

const props = withDefaults(
  defineProps<Props>(),
  {
    readonly: () => false,
    editMode: () => false,
  },
);

const emit = defineEmits<Emits>();

const isEdit = ref(false);
function setEditStatus() {
  isEdit.value = true;
}

function handleInput(v: string) {
  emit('input', v);
}

function handleBlur(v: string) {
  emit('blur', v);
}

watch(() => props.editMode, () => {
  isEdit.value = !!props.editMode;
}, { immediate: true });

watch(isEdit, () => {
  emit('update:editMode', isEdit.value);
});
</script>
