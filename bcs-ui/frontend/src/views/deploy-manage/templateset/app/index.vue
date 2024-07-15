<template>
  <keep-alive>
    <component
      :is="componentName"
      :cur-project="curProject"
      :category="category"
      :namespace="namespace"
      :name="name"
      :kind="kind"
      :crd="crd"
      :hidden-operate="true"
      :cluster-id="clusterId"
    ></component>
  </keep-alive>
  <!-- <router-view v-else :key="$route.path"></router-view> -->
</template>

<script>
const deployments = () => import(
  /* webpackChunkName: 'app-list' */'./k8s/deployments');
const deploymentsInstanceDetail = () => import(
  /* webpackChunkName: 'app-instance' */'@/views/resource-view/workload/detail/old-index.vue');
const deploymentsInstanceDetail2 = () => import(
  /* webpackChunkName: 'app-instance' */'@/views/resource-view/workload/detail/old-index.vue');
const deploymentsInstantiation = () => import(
  /* webpackChunkName: 'app-instantiation' */'./k8s/deployments-instantiation');
const daemonset = () => import(
  /* webpackChunkName: 'app-list' */'./k8s/daemonset');
const daemonsetInstanceDetail = () => import(
  /* webpackChunkName: 'app-instance' */'@/views/resource-view/workload/detail/old-index.vue');
const daemonsetInstanceDetail2 = () => import(
  /* webpackChunkName: 'app-instance' */'@/views/resource-view/workload/detail/old-index.vue');
const daemonsetInstantiation = () => import(
  /* webpackChunkName: 'app-instantiation' */'./k8s/daemonset-instantiation');

const job = () => import(
  /* webpackChunkName: 'app-list' */'./k8s/job');
const jobInstanceDetail = () => import(
  /* webpackChunkName: 'app-instance' */'@/views/resource-view/workload/detail/old-index.vue');
const jobInstanceDetail2 = () => import(
  /* webpackChunkName: 'app-instance' */'@/views/resource-view/workload/detail/old-index.vue');
const jobInstantiation = () => import(
  /* webpackChunkName: 'app-instantiation' */'./k8s/job-instantiation');

const statefulset = () => import(
  /* webpackChunkName: 'app-list' */'./k8s/statefulset');
const statefulsetInstanceDetail = () => import(
  /* webpackChunkName: 'app-instance' */'@/views/resource-view/workload/detail/old-index.vue');
const statefulsetInstanceDetail2 = () => import(
  /* webpackChunkName: 'app-instance' */'@/views/resource-view/workload/detail/old-index.vue');
const statefulsetInstantiation = () => import(
  /* webpackChunkName: 'app-instantiation' */'./k8s/statefulset-instantiation');
const gamestatefulset = () => import(
  /* webpackChunkName: 'app-list' */'./k8s/gamestatefulset');
const gamestatefulSetsInstanceDetail = () => import(
  /* webpackChunkName: 'app-instance' */'@/views/resource-view/workload/detail/old-index.vue');
const gamedeployments = () => import(
  /* webpackChunkName: 'app-list' */'./k8s/gamedeployments');
const gamedeploymentsInstanceDetail = () => import(
  /* webpackChunkName: 'app-instance' */'@/views/resource-view/workload/detail/old-index.vue');
const customobjects = () => import(
  /* webpackChunkName: 'app-list' */'./k8s/customobjects');

export default {
  name: 'DetailIndex',
  components: {

    deployments,
    deploymentsInstanceDetail,
    deploymentsInstanceDetail2,
    deploymentsInstantiation,

    daemonset,
    daemonsetInstanceDetail,
    daemonsetInstanceDetail2,
    daemonsetInstantiation,

    job,
    jobInstanceDetail,
    jobInstanceDetail2,
    jobInstantiation,

    statefulset,
    statefulsetInstanceDetail,
    statefulsetInstanceDetail2,
    statefulsetInstantiation,

    gamestatefulset,
    gamestatefulSetsInstanceDetail,
    gamedeployments,
    gamedeploymentsInstanceDetail,
    customobjects,
  },
  data() {
    return {
      k8sPathNameList: [
        'deployments',
        'deploymentsInstanceDetail',
        'deploymentsInstanceDetail2',
        'deploymentsInstantiation',
        'daemonset',
        'daemonsetInstanceDetail',
        'daemonsetInstanceDetail2',
        'daemonsetInstantiation',
        'job',
        'jobInstanceDetail',
        'jobInstanceDetail2',
        'jobInstantiation',
        'statefulset',
        'statefulsetInstanceDetail',
        'statefulsetInstanceDetail2',
        'statefulsetInstantiation',

        'gamestatefulset',
        'gamestatefulSetsInstanceDetail',
        'gamedeployments',
        'gamedeploymentsInstanceDetail',
        'customobjects',
      ],
      componentName: '',
    };
  },
  computed: {
    curProject() {
      return this.$store.state.curProject;
    },
    projectCode() {
      return this.$store.getters.curProjectCode;
    },
    // curProjectCode() {
    //   return this.$store.getters.curProjectCode;
    // },
    projectId() {
      return this.$store.getters.curProjectId;
    },
    // curProjectId() {
    //   return this.$store.getters.curProjectId;
    // },
    category() {
      const categoryMap = {
        deploymentsInstanceDetail: 'deployments',
        deploymentsInstanceDetail2: 'deployments',
        daemonsetInstanceDetail: 'daemonsets',
        daemonsetInstanceDetail2: 'daemonsets',
        statefulsetInstanceDetail: 'statefulsets',
        statefulsetInstanceDetail2: 'statefulsets',
        jobInstanceDetail: 'jobs',
        jobInstanceDetail2: 'jobs',
        gamedeploymentsInstanceDetail: 'custom_objects',
        gamestatefulSetsInstanceDetail: 'custom_objects',
      };
      return categoryMap[this.$route.name];
    },
    namespace() {
      return this.$route.query?.namespace;
    },
    name() {
      return this.$route.query?.name;
    },
    kind() {
      const kindMap = {
        deploymentsInstanceDetail: 'Deployment',
        deploymentsInstanceDetail2: 'Deployment',
        daemonsetInstanceDetail: 'DaemonSet',
        daemonsetInstanceDetail2: 'DaemonSet',
        statefulsetInstanceDetail: 'StatefulSet',
        statefulsetInstanceDetail2: 'StatefulSet',
        jobInstanceDetail: 'Job',
        jobInstanceDetail2: 'Job',
        gamedeploymentsInstanceDetail: 'GameDeployment',
        gamestatefulSetsInstanceDetail: 'GameStatefulSet',
      };
      return kindMap[this.$route.name];
    },
    crd() {
      const crdMap = {
        gamedeploymentsInstanceDetail: 'gamedeployments.tkex.tencent.com',
        gamestatefulSetsInstanceDetail: 'gamestatefulsets.tkex.tencent.com',
      };
      return crdMap[this.$route.name];
    },
    clusterId() {
      return this.$route.query?.cluster_id;
    },
  },
  mounted() {
    this.setCurProject();
  },
  methods: {
    /**
             * 设置 curProject
             */
    setCurProject() {
      if (this.curProject) {
        this.setComponent();
      }
    },

    /**
             * 设置动态组件
             */
    setComponent() {
      const routeName = this.$route.name;
      if (this.k8sPathNameList.indexOf(routeName) <= -1) {
        this.componentName = 'deployments';
      } else {
        this.componentName = routeName;
      }
    },
  },
};
</script>
