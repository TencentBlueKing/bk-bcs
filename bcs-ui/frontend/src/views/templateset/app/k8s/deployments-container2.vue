<template src="./tmpl-container.html"></template>

<script>
import mixinBaseContainer from '../mixins/mixin-base-container';

export default {
  mixins: [mixinBaseContainer],
  data() {
    return {
      CATEGORY: 'deployment',
    };
  },
  methods: {
    /**
             * 返回实例详情
             */
    goInstanceDetail() {
      const params = {
        projectId: this.projectId,
        projectCode: this.projectCode,
        instanceId: this.instanceId,
        instanceName: this.instanceName,
        instanceNamespace: this.instanceNamespace,
        instanceCategory: this.instanceCategory,
        searchParamsList: this.searchParamsList,
      };
      if (this.namespaceId) {
        params.namespaceId = this.namespaceId;
      }
      this.$router.push({
        name: 'deploymentsInstanceDetail2',
        params,
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
