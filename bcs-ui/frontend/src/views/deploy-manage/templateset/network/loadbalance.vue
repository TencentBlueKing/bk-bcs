<!-- eslint-disable vue/multi-word-component-names -->
<template>
  <component v-bind:is="currentView" ref="loadbalance"></component>
</template>

<script>
import k8sLoadBalance from './loadbalance/k8s/index';

export default {
  components: {
    k8sLoadBalance,
  },
  data() {
    return {
      currentView: k8sLoadBalance,
    };
  },
  mounted() {
    this.$store.commit('network/updateLoadBalanceList', []);
  },
  beforeDestroy() {
    this.$refs.loadbalance.leaveCallback();
  },
  beforeRouteLeave(to, from, next) {
    this.$refs.loadbalance.leaveCallback();
    next(true);
  },
};
</script>
