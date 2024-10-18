// @ts-nocheck
/* eslint-disable */
import type { DirectiveBinding } from 'vue';

import Tippy, { ReferenceElement } from 'tippy.js';
import 'tippy.js/dist/tippy.css';
import 'tippy.js/themes/light.css';
export default {
  mounted(el: ReferenceElement, binding: DirectiveBinding) {
    const props = typeof binding.value === 'object'
      ? binding.value
      : { disabled: false, content: binding.value };
    const instance = Tippy(
      el,
      Object.assign({ appendTo: document.body }, props),
    );
    if (props.disabled) {
      instance.disable();
    }
  },
  unmounted(el: ReferenceElement) {
    el._tippy?.destroy();
  },
  updated(el: ReferenceElement, binding: DirectiveBinding) {
    const props = typeof binding.value === 'object'
      ? binding.value
      : { disabled: false, content: binding.value };
    props.disabled ? el._tippy.disable() : el._tippy.enable();
    el._tippy.setContent(props.content);
  },
};
