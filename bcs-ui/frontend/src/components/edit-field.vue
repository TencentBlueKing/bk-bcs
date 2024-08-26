<template>
  <section class="flex items-center">
    <template v-if="!isEdit && !editMode">
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
        @input="handleInput" />
    </slot>
  </section>
</template>
<script lang="ts" setup>
import { ref } from 'vue';

interface Props {
  readonly?: Boolean
  value?: String
  editMode?: Boolean
}

type Emits = (e: 'input', v: string) => void;

withDefaults(
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
</script>
