<template>
  <PopoverSelector
    class="w-full"
    offset="0,6"
    trigger="manual"
    placement="bottom-start"
    :disabled="disableVarList"
    ref="popoverRef">
    <bcs-input v-bind="$attrs" :value="value" @input="handleInput"></bcs-input>
    <template #content>
      <ul class="max-h-[360px] overflow-auto">
        <li
          class="bcs-dropdown-item"
          v-for="item in fileStore.varList"
          :key="item.id"
          @click="handleSetVar(item.key)">
          <span>{{ `\{\{ ${item.key} \}\}` }}</span>
          <span class="text-[#979BA5]">{{ item.name }}</span>
        </li>
      </ul>
    </template>
  </PopoverSelector>
</template>
<script setup lang="ts">
import { ref } from 'vue';

import { store as fileStore } from './use-store';

import PopoverSelector from '@/components/popover-selector.vue';

interface Props {
  value?: string | number
  disableVarList?: boolean
  type?: string
}
type Emits = (e: 'input', v: string|number) => void;

const props = withDefaults(defineProps<Props>(), {
  value: () => '',
  disableVarList: () => false,
  type: () => 'text',
});
const emits = defineEmits<Emits>();

const popoverRef = ref();
function handleInput(v: string) {
  let value: string|number = v;
  if (props.type === 'number' && !isNaN(Number(v))) {
    value = Number(v);
  }
  emits('input', value);
  if (v.indexOf('{{') === 0) {
    popoverRef.value?.show();
  }
}

function handleSetVar(key: string) {
  emits('input', `{{ ${key} }}`);
  popoverRef.value?.hide();
}
</script>
<style scoped lang="postcss">
>>> .bk-tooltip-ref {
  width: 100%;
}
</style>
