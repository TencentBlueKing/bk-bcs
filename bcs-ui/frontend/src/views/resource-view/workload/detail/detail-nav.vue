<template>
  <div class="header-router">
    <span class="icon-wrapper" @click="handleBack">
      <i class="bcs-icon bcs-icon-arrows-left icon-back"></i>
    </span>
    <span class="title-wrapper" v-for="(item, index) in curNavList" :key="index">
      <span class="title-item" @click="routeHop(item, index)">{{item.name}}</span>
      <span class="separator" v-if="index < (curNavList.length - 1)">/</span>
    </span>
  </div>
</template>
<script lang="ts">
import { computed, defineComponent, PropType, toRefs } from 'vue';

export default defineComponent({
  name: 'DetailTopNav',
  props: {
    list: {
      type: Array as PropType<any[]>,
      default: () => [],
    },
    clusterId: {
      type: String,
      default: '',
    },
    active: {
      type: String,
      default: '',
    },
  },
  setup(props, ctx) {
    const { list } = toRefs(props);
    const curNavList = computed(() => {
      const index = list.value.findIndex(item => item.id === props.active);
      return index > -1 ? list.value.slice(0, index + 1) : list.value;
    });

    const handleBack = () => {
      const index = curNavList.value.length - 2;
      if (!curNavList.value[index]) return;
      routeHop(curNavList.value[index], index);
    };

    const routeHop = (item, index) => {
      if (index === (curNavList.value.length - 1)) return;

      ctx.emit('change', item);
    };

    return {
      handleBack,
      routeHop,
      curNavList,
    };
  },
});
</script>
<style lang="postcss" scoped>
.header-router .title-wrapper {
    &:last-of-type {
        color: #979BA5;
    }
    &:last-of-type .title-item {
        cursor: default;
    }
}
.header-router {
    display: flex;
    height: 52px;
    align-items: center;
    font-size: 16px;
    padding-left: 12px;
    border-bottom: 1px solid #F3F3F3;
    background: #fff;
    .icon-wrapper {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 32px;
        height: 32px;
        cursor: pointer;
        .icon-back {
            font-size: 16px;
            font-weight: bold;
            color: #3A84FF;
        }
    }
    .title-wrapper {
        display: flex;
        color: #3A84FF;
        .title-item {
            cursor: pointer;
        }
        .separator {
            padding: 0 6px;
            color: #979ba5;
        }
    }
}
</style>
