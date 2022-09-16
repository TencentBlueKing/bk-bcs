<template src="./tmpl-instance.html"></template>

<script>
    import mixinBaseInstance from '../mixins/mixin-base-instance'

    export default {
        mixins: [mixinBaseInstance],
        data () {
            return {
                CATEGORY: 'job'
            }
        },
        methods: {
            /**
             * 跳转到容器详情
             *
             * @param {Object} taskgroup 当前容器所属的 taskgroup 对象
             * @param {Object} container 当前容器对象
             */
            goContainerDetail (taskgroup, container) {
                this.$router.push({
                    name: 'jobContainerDetail2',
                    params: {
                        projectId: this.projectId,
                        projectCode: this.projectCode,
                        instanceId: this.instanceId,
                        namespaceId: this.namespaceId,
                        taskgroupName: taskgroup.name,
                        containerId: container.container_id,
                        searchParamsList: this.searchParamsList
                    },
                    query: {
                        cluster_id: this.clusterId
                    }
                })
            },

            /**
             * 返回应用列表
             * 有 namespaceId 时，会到应用列表页，对应的 namespace 会展开
             * 有 tplsetId 时，会到应用列表页，对应的模版集会展开
             */
            goAppList () {
                const viewMode = localStorage.getItem('appViewMode')
                this.$router.push({
                    name: 'job',
                    params: {
                        projectId: this.projectId,
                        projectCode: this.projectCode,
                        instanceId: this.instanceId,
                        templateId: this.templateId,
                        namespaceId: viewMode === 'namespace' ? this.namespaceId : '',
                        tplsetId: viewMode === 'template' ? this.instanceInfo.template_id : '',
                        searchParamsList: this.searchParamsList
                    }
                })
            }
        }
    }
</script>

<style scoped>
    @import '../instance.css';
</style>
