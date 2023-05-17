<template>
  <div :class="['editor-status bcs-ellipsis', theme]">
    <i :class="['bcs-icon', `bcs-icon-${icon}`]"></i>
    <span class="message">{{ message }}</span>
  </div>
</template>
<script>
import { computed, defineComponent, toRefs } from 'vue';

export default defineComponent({
  name: 'EditorStatus',
  props: {
    theme: {
      type: String,
      default: 'error',
    },
    message: {
      type: String,
      default: '',
    },
  },
  setup(props) {
    const { theme } = toRefs(props);
    const icon = computed(() => {
      const iconMap = {
        error: 'close-circle-shape',
        default: 'info-circle-shape',
      };
      return iconMap[theme.value] || 'info-circle-shape';
    });

    return {
      icon,
    };
  },
});
</script>
<style lang="postcss" scoped>
.editor-status {
    height: 100%;
    border-left: 4px solid;
    padding: 8px 16px;
    background: #212121;
    border-radius: 0px 0px 2px 2px;
    display: flex;
    align-items: flex-start;
    position: absolute;
    bottom: 0;
    left: 0;
    width: 100%;
    i {
        font-size: 12px;
        margin-top: 6px;
    }
    &.error {
        border-left-color: #b34747;
        i {
            color: #b34747;
        }
    }
    .message {
        margin-left: 8px;
        color: #dcdee5;
        line-height: 20px;
        font-size: 12px;
    }
}
</style>
