<template src="./tmpl-instantiation.html"></template>

<script>
import mixinBaseInstantiation from '../mixins/mixin-base-instantiation';

import BcsContent from '@/components/layout/Content.vue';
import Row from '@/components/layout/Row.vue';

export default {
  components: { BcsContent, Row },
  mixins: [mixinBaseInstantiation],
  data() {
    return {
      CATEGORY: 'statefulset',
    };
  },
  methods: {
    /**
             * 返回模板集列表
             *
             * @param {boolean} needConfirm 是否需要 confirm 提示
             */
    goTemplateset(needConfirm) {
      const params = {
        projectId: this.projectId,
        projectCode: this.projectCode,
        tplsetId: this.templateId,
        searchParamsList: this.searchParamsList,
      };
      if (needConfirm) {
        // eslint-disable-next-line @typescript-eslint/no-this-alias
        const me = this;
        const h = me.$createElement;
        me.$bkInfo({
          title: '',
          content: h('p', this.$t('deploy.templateset.confirmTemplate', { tmplAppName: me.tmplAppName })),
          confirmFn() {
            me.$router.push({
              name: 'statefulset',
              params,
            });
          },
        });
      } else {
        this.$router.push({
          name: 'statefulset',
          params,
        });
      }
    },
  },
};
</script>

<style scoped>
    @import '../instantiation.css';
</style>
