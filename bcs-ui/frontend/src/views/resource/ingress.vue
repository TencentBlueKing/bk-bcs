<template>
    <component v-bind:is="currentView" v-if="curProject.project_id" ref="ingress"></component>
</template>

<script>
    import k8sIngress from './ingress/k8s/index'
    import globalMixin from '@open/mixins/global'

    export default {
        components: {
            k8sIngress
        },
        mixins: [globalMixin],
        data () {
            return {
                curProject: {},
                currentView: k8sIngress
            }
        },
        computed: {
            onlineProjectList () {
                return this.$store.state.sideMenu.onlineProjectList
            }
        },
        mounted () {
            this.curProject = this.initCurProject()
            this.$store.commit('network/updateIngressList', [])
        }
    }
</script>
