<template>
  <bcs-exception :type="type" scene="part">
    <div class="text-[14px]">{{ typeMap[type] }}</div>
    <template v-if="type === 'search-empty'">
      <i18n
        tag="div"
        path="可以尝试 调整关键词 或{action}"
        class="mt-[8px] text-[12px] text-[#979BA5]">
        <button place="action" class="bk-text-button" @click="handleClear">{{$t('清空筛选条件')}}</button>
      </i18n>
    </template>
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
  },
  setup(props, ctx) {
    const typeMap = ref({
      empty: $i18n.t('暂无数据'),
      'search-empty': $i18n.t('搜索结果为空'),
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
