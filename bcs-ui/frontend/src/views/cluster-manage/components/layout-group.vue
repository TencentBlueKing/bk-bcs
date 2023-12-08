<template>
  <div :class="['layout-group', size, { collapsible }]">
    <div class="title" @click="handleToggleActive">
      <span
        class="icon"
        :style="!active ? 'transform: rotate(-90deg);' : 'transform: rotate(0deg);'"
        v-if="collapsible">
        <i class="bcs-icon bcs-icon-down-shape"></i>
      </span>
      <span class="name">
        <slot name="title">{{title}}</slot>
      </span>
    </div>
    <div class="content" v-show="active">
      <slot></slot>
    </div>
  </div>
</template>
<script lang="ts">
import { defineComponent, ref, toRefs, watch } from 'vue';

export default defineComponent({
  name: 'LayoutGroup',
  props: {
    size: {
      type: String,
      default: 'default',
    },
    collapsible: {
      type: Boolean,
      default: false,
    },
    title: {
      type: String,
      default: '',
    },
    expanded: {
      type: Boolean,
      default: true,
    },
  },
  setup(props) {
    const { expanded, collapsible } = toRefs(props);
    const active = ref(expanded.value);

    watch(expanded, () => {
      active.value = expanded.value;
    });
    function handleToggleActive() {
      if (!collapsible.value) return;
      active.value = !active.value;
    }
    return {
      active,
      handleToggleActive,
    };
  },
});
</script>
<style lang="postcss" scoped>
.layout-group {
  .title {
    position: relative;
    display: flex;
    align-items: center;
    background: #F5F7FA;
    color: #313238;
    .icon {
      color: #DCDEE5;
      display: flex;
      align-items: center;
      justify-content: center;
      position: absolute;
      transition: all linear .2s;
    }
  }
  .content {
    padding: 10px 24px;
  }
  &.collapsible .title {
    cursor: pointer;
  }
  &.default {
    .title {
      padding: 0 24px;
      height: 32px;
      font-size: 12px;
      .icon {
        font-size: 14px;
        width: 28px;
        height: 28px;
        left: 0;
      }
    }
  }
}
</style>
