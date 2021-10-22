<template>
    <component v-bind:is="currentView" ref="service"></component>
</template>

<script>
    import k8sService from './service/k8s'
    import globalMixin from '@open/mixins/global'

    export default {
        beforeRouteLeave (to, from, next) {
            this.$refs.service && this.$refs.service.leaveCallback()
            next(true)
        },
        components: {
            k8sService
        },
        mixins: [globalMixin],
        data () {
            return {
                currentView: 'k8sService'
            }
        },
        computed: {
            onlineProjectList () {
                return this.$store.state.sideMenu.onlineProjectList
            }
        },
        mounted () {
            this.curProject = this.initCurProject()
        },
        beforeDestroy () {
            this.$refs.service.leaveCallback()
        }
    }
</script>
