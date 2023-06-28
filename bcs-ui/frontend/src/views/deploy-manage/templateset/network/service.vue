<!-- eslint-disable vue/multi-word-component-names -->
<template>
  <component v-bind:is="currentView" ref="service"></component>
</template>

<script>
import k8sService from './service/k8s';

export default {
  beforeRouteLeave(to, from, next) {
    this.$refs.service?.leaveCallback();
    next(true);
  },
  components: {
    k8sService,
  },
  data() {
    return {
      currentView: 'k8sService',
    };
  },
  beforeDestroy() {
    this.$refs.service.leaveCallback();
  },
};
</script>
