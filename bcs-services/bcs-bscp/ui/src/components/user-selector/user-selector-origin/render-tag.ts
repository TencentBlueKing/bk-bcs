// import * as Vue from 'vue';
// export default {
//   name: 'render-tag',
//   props: ['username', 'user', 'index'],
//   render() {
//     return this.$parent.renderTag(Vue.h, {
//       username: this.username,
//       index: this.index,
//       user: this.user,
//     });
//   },
// };

import { h, inject } from 'vue';

export default {
  name: 'render-tag',
  props: ['username', 'user', 'index'],
  // props: {
  //   selector: {
  //     type: Object,
  //   },
  //   user: {
  //     type: Object,
  //   },
  //   keyword: {
  //     type: String,
  //   },
  //   index: {
  //     type: Number,
  //   },
  //   disabled: {
  //     type: Boolean,
  //   },
  // },
  setup(props: any) {
    const parentSelector = inject('parentSelector');
    return () => (parentSelector as any).renderTag(h, {
      username: props.username,
      index: props.index,
      user: props.user,
    });
  },
};
