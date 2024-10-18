// import * as Vue from 'vue';
// export default {
//   name: 'render-alternate',
//   data() {
//     return { selector: null };
//   },
//   render() {
//     return this.selector.defaultAlternate(Vue.h);
//   },
// };

import { h } from 'vue';

export default {
  name: 'render-alternate',
  props: ['selector'],
  setup(props: any): any {
    return () => props.selector.defaultAlternate(h);
  },
};
