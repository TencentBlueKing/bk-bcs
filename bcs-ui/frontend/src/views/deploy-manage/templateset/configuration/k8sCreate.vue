<template>
  <router-view :key="time" v-if="isK8sCreate"></router-view>
</template>
<script>
export default {
  data() {
    return {
      time: Date.now(),
      isProjectChange: false,
      isK8sCreate: true,
    };
  },
  computed: {
    curProject() {
      const project = this.$store.state.curProject;
      return project;
    },
    deployments() {
      return this.$store.state.k8sTemplate.deployments;
    },
    services() {
      return this.$store.state.k8sTemplate.services;
    },
    configmaps() {
      return this.$store.state.k8sTemplate.configmaps;
    },
    secrets() {
      return this.$store.state.k8sTemplate.secrets;
    },
    daemonsets() {
      return this.$store.state.k8sTemplate.daemonsets;
    },
    jobs() {
      return this.$store.state.k8sTemplate.jobs;
    },
    statefulsets() {
      return this.$store.state.k8sTemplate.statefulsets;
    },
    ingresss() {
      return this.$store.state.k8sTemplate.ingresss;
    },
    projectId() {
      return this.$route.params.projectId;
    },
  },
  mounted() {
    const createRoutes = [
      'k8sTemplatesetDeployment',
      'k8sTemplatesetService',
      'k8sTemplatesetConfigmap',
      'k8sTemplatesetSecret',
      'k8sTemplatesetDaemonset',
      'k8sTemplatesetJob',
      'k8sTemplatesetStatefulset',
      'k8sTemplatesetIngress',
      'k8sTemplatesetHPA',
    ];
    const routeName = this.$route.name;
    if (createRoutes.join(',').indexOf(routeName) < 0) {
      this.clearData();
    }
    this.getExistConfigmap();
  },
  beforeRouteLeave(to, from, next) {
    this.clearData();
    next(true);
  },
  methods: {
    /**
             * 获取已经存在的configmap
             */
    async getExistConfigmap() {
      try {
        const res = await this.$store.dispatch('k8sTemplate/getExistConfigmap', { projectId: this.projectId });
        this.$store.commit('k8sTemplate/updateExistConfigmap', res.data);
      } catch (e) {
        this.$store.commit('k8sTemplate/updateExistConfigmap', []);
      }
    },

    /**
             * 清空模板集数据
             */
    clearData() {
      this.$store.commit('k8sTemplate/clearCurTemplateData');
    },

    /**
             * 重新刷新
             */
    reloadTemplateset() {
      // this.time = Date.now()
      this.isK8sCreate = false;
      this.$nextTick(() => {
        this.isK8sCreate = true;
      });
    },
  },
};
</script>
<style>
    .biz-template-tip {
        font-size: 12px;
        margin-bottom: 10px;
        color: #979BA5;
    }
    .biz-tip {
        font-size: 12px;
    }
    .biz-configuration-create-box {
        width: 100%;
    }
</style>
