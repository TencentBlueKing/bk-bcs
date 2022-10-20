<template>
  <div class="login-modal" v-if="visible">
    <iframe
      :src="src"
      frameborder="0"
      border="0"
      :width="width"
      :height="height"
      scrolling="none"
      allowtransparency="yes"
      :referrerpolicy="referrerpolicy">
    </iframe>
  </div>
</template>

<script>
export default {
  name: 'BkMagicLogin',
  props: {
    loginUrl: {
      type: String,
      default: '',
    },
    successUrl: {
      type: String,
      default: '',
    },
    width: {
      type: [Number, String],
      default: 400,
    },
    height: {
      type: [Number, String],
      default: 400,
    },
    referrerpolicy: {
      type: String,
      default: 'strict-origin-when-cross-origin',
    },
  },
  data() {
    return {
      visible: false,
    };
  },
  computed: {
    src() {
      if (this.successUrl) {
        return `${this.loginUrl}?c_url=${this.successUrl}`;
      }
      return this.loginUrl;
    },
  },
  methods: {
    show() {
      this.visible = true;
    },
    hide() {
      this.visible = false;
    },
  },
};
</script>

<style lang="postcss" scoped>
  .login-modal {
      position: fixed;
      top: 0;
      right: 0;
      bottom: 0;
      left: 0;
      background-color: rgba(0, 0, 0, .6);
      z-index: 99999;
      font-size: 0;
      iframe {
          display: block;
          margin: calc((100vh - 470px) / 2) auto;
          background-color: #fff;
          border-radius: 2px;
      }
  }
</style>
