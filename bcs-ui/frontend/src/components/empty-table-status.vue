<template>
  <bcs-exception :type="type" scene="part">
    <div class="text-[14px]">{{ typeMap[type] }}</div>
    <template v-if="type === 'search-empty'">
      <i18n
        tag="div"
        path="generic.msg.empty.searchEmpty.subTitle"
        class="mt-[8px] text-[12px] text-[#979BA5]">
        <button place="action" class="bk-text-button" @click="handleClear">
          {{ buttonText || $t('generic.button.clearSearch')}}
        </button>
      </i18n>
    </template>
    <slot></slot>
  </bcs-exception>
</template>
<script lang="ts">
import { defineComponent, ref } from 'vue';

import $i18n from '@/i18n/i18n-setup';

export default defineComponent({
  name: 'EmptyTableStatus',
  props: {
    type: {
      type: String,
      default: 'empty',
    },
    buttonText: {
      type: String,
      default: '',
    },
  },
  setup(props, ctx) {
    const typeMap = ref({
      empty: $i18n.t('generic.msg.empty.noData'),
      'search-empty': $i18n.t('generic.msg.empty.searchEmpty.text'),
    });
    const handleClear = () => {
      ctx.emit('clear');
    };

    return {
      typeMap,
      handleClear,
    };
  },
});
</script>
