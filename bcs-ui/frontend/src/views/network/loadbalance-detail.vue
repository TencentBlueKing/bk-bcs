<template>
    <component v-bind:is="currentView"></component>
</template>

<script>
    import k8sLoadBalance from './loadbalance/k8s/detail'
    import globalMixin from '@open/mixins/global'

    export default {
        components: {
            k8sLoadBalance
        },
        mixins: [globalMixin],
        data () {
            return {
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
        }
    }
</script>
