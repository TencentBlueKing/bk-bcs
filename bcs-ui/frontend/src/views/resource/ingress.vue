<!-- eslint-disable vue/multi-word-component-names -->
<template>
  <component v-bind:is="currentView" ref="ingress"></component>
</template>

<script>
import k8sIngress from './ingress/k8s/index';

export default {
  components: {
    k8sIngress,
  },
  data() {
    return {
      currentView: k8sIngress,
    };
  },
  computed: {
    onlineProjectList() {
      return this.$store.state.sideMenu.onlineProjectList;
    },
  },
};
</script>
