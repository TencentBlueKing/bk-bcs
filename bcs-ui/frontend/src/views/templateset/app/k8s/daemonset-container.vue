<template src="./tmpl-container.html"></template>

<script>
import mixinBaseContainer from '../mixins/mixin-base-container';

export default {
  mixins: [mixinBaseContainer],
  data() {
    return {
      CATEGORY: 'daemonset',
    };
  },
  methods: {
    /**
             * 返回实例详情
             */
    goInstanceDetail() {
      this.$router.push({
        name: 'daemonsetInstanceDetail',
        params: {
          projectId: this.projectId,
          projectCode: this.projectCode,
          instanceId: this.instanceId,
          searchParamsList: this.searchParamsList,
        },
        query: {
          cluster_id: this.clusterId,
        },
      });
    },
  },
};
</script>

<style scoped>
    @import '../container.css';
</style>
