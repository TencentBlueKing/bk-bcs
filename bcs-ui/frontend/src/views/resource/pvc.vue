<template>
    <div class="biz-content">
        <div class="biz-top-bar">
            <div class="biz-topbar-title">
                {{$t('存储')}}
            </div>
        </div>
        <div class="biz-content-wrapper p0">
            <app-exception :type="'building'"></app-exception>
        </div>
    </div>
</template>

<script>
    export default {
        computed: {
            curProject () {
                return this.$store.state.curProject
            },
            projectId () {
                return this.$route.params.projectId
            }
        },
        watch: {
            '$route' () {
                this.checkProjectKind()
            }
        },
        mounted () {
            this.checkProjectKind()
        },
        methods: {
            checkProjectKind () {
                if (this.curProject && this.curProject.kind !== 1) {
                    this.$router.push({
                        name: 'clusterMain',
                        params: {
                            projectId: this.projectId
                        }
                    })
                }
            }
        }
    }
</script>
