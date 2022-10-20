<template>
  <component v-bind:is="currentView"></component>
</template>

<script>
import k8sLoadBalance from './loadbalance/k8s/detail';

export default {
  components: {
    k8sLoadBalance,
  },
  data() {
    return {
      currentView: k8sLoadBalance,
    };
  },
  computed: {
    onlineProjectList() {
      return this.$store.state.sideMenu.onlineProjectList;
    },
  },
  mounted() {
    this.$store.commit('network/updateLoadBalanceList', []);
  },
};
</script>
