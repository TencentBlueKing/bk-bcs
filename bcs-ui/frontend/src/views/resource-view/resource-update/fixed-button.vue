<template>
  <div
    class="fixed-button"
    ref="fixedButtonRef"
    :style="{ position: position, cursor: drag ? 'move' : 'pointer' }"
    @click="handleClick">
    <slot>
      <i class="bcs-icon bcs-icon-qiehuan mr5"></i>
      {{title}}
    </slot>
  </div>
</template>
<script lang="ts">
import { Property } from 'csstype';
import { PropType, defineComponent, onMounted, ref } from 'vue';

export default defineComponent({
  name: 'FixedButton',
  props: {
    title: {
      type: String,
      default: '',
    },
    drag: {
      type: Boolean,
      default: false,
    },
    position: {
      type: String as PropType<Property.Position>,
      default: 'absolute',
    },
  },
  setup(props, ctx) {
    const fixedButtonRef = ref<any>(null);

    const dragElement = (elmnt) => {
      let pos1 = 0; let pos2 = 0; let pos3 = 0; let pos4 = 0;
      elmnt.onmousedown = dragMouseDown;

      function dragMouseDown(e) {
        e.preventDefault();
        pos3 = e.clientX;
        pos4 = e.clientY;
        document.onmouseup = closeDragElement;
        document.onmousemove = elementDrag;
      }

      function elementDrag(e) {
        e.preventDefault();
        pos1 = pos3 - e.clientX;
        pos2 = pos4 - e.clientY;
        pos3 = e.clientX;
        pos4 = e.clientY;
        elmnt.style.top = `${elmnt.offsetTop - pos2}px`;
        elmnt.style.left = `${elmnt.offsetLeft - pos1}px`;
      }

      function closeDragElement() {
        document.onmouseup = null;
        document.onmousemove = null;
      }
    };

    const handleClick = () => {
      ctx.emit('click');
    };

    onMounted(() => {
      props.drag && dragElement(fixedButtonRef.value);
    });

    return {
      fixedButtonRef,
      handleClick,
    };
  },
});
</script>
<style lang="postcss" scoped>
.fixed-button {
    display: flex;
    align-items: center;
    min-width: 148px;
    padding: 0 16px;
    height: 40px;
    background: #fff;
    box-shadow: 0 2px 8px 0 rgba(0,0,0,0.16);
    border-radius: 20px;
    z-index: 200;
    font-size: 12px;
    color: #63656E;
    cursor: pointer;

    &:hover {
        box-shadow: 0 2px 12px 0 rgba(0,0,0,0.20);
        color: #3A84FF;
    }
}
</style>
