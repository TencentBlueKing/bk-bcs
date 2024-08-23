import { h } from 'vue';

// export default {
//   name: 'render-list',
//   props: ['selector', 'user', 'index', 'keyword', 'disabled'],
//   render() {
//     return this.selector.renderList(h, {
//       user: this.user,
//       index: this.index,
//       keyword: this.keyword,
//       disabled: this.disabled,
//     });
//   },
// };

export default {
  name: 'render-list',
  props: {
    selector: {
      type: Object,
    },
    user: {
      type: Object,
    },
    keyword: {
      type: String,
    },
    index: {
      type: Number,
    },
    disabled: {
      type: Boolean,
    },
  },
  setup(props: any): any {
    return () => props.selector.renderList(h, {
      user: props.user,
      index: props.index,
      keyword: props.keyword,
      disabled: props.disabled,
    });
  },
};
