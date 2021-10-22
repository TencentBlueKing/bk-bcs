<template>
    <component v-bind:is="currentView" v-if="curProject.project_id" ref="loadbalance"></component>
</template>

<script>
    import k8sLoadBalance from './loadbalance/k8s/index'
    import globalMixin from '@open/mixins/global'

    export default {
        components: {
            k8sLoadBalance
        },
        mixins: [globalMixin],
        data () {
            return {
                curProject: {},
                currentView: k8sLoadBalance
            }
        },
        computed: {
            onlineProjectList () {
                return this.$store.state.sideMenu.onlineProjectList
            }
        },
        mounted () {
            this.curProject = this.initCurProject()
            this.$store.commit('network/updateLoadBalanceList', [])
        },
        beforeDestroy () {
            this.$refs.loadbalance.leaveCallback()
        },
        beforeRouteLeave (to, from, next) {
            this.$refs.loadbalance.leaveCallback()
            next(true)
        }
    }
</script>
