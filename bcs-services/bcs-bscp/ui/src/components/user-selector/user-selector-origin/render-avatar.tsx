// export default {
//   name: 'render-avatar',
//   props: ['user', 'urlMethod'],
//   render() {
//     return <span class="user-selector-avatar"></span>;
//   },
//   async created() {
//     try {
//       let avatar = null;
//       if (typeof this.user === 'string') {
//         avatar = await this.urlMethod(this.user);
//       } else if (typeof this.user === 'object') {
//         avatar = this.user.avatar
//           || this.user.logo
//           || (await this.urlMethod(this.user.username));
//       }
//       if (avatar) {
//         this.$el.style.backgroundImage = `url(${avatar})`;
//       }
//     } catch (e) {}
//   },
// };

import { onMounted, ref, watch } from 'vue';

export default {
  name: 'render-avatar',
  props: ['user', 'urlMethod'],
  setup(props: any) {
    const avatar = ref('');
    onMounted(async () => {
      try {
        if (typeof props.user === 'string') {
          avatar.value = await props.urlMethod(props.user);
        } else if (typeof props.user === 'object') {
          avatar.value = props.user.avatar || props.user.logo || (await props.urlMethod(props.user.username));
        }
      } catch (e) {}
    });

    const userSelectorAvatarRef = ref(null);

    watch(avatar, (v: string) => {
      if (v) {
        userSelectorAvatarRef.value.style.backgroundImage = `url(${avatar.value})`;
      }
    });

    return () => <span ref={userSelectorAvatarRef} class="user-selector-avatar"></span>;
  },
};
