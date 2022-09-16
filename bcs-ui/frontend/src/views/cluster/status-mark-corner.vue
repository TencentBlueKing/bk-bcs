<template>
    <div class="bk-mark-corner bk-warning" v-if="
        $parent.conversionPercent(curCluster.cpu_usage.used, curCluster.cpu_usage.total) >= 80
            || $parent.conversionPercent(curCluster.mem_usage.used_bytes, curCluster.mem_usage.total_bytes) >= 80
            || $parent.conversionPercent(curCluster.disk_usage.used_bytes, curCluster.disk_usage.total_bytes) >= 80
    ">
        <p>!</p>
    </div>
</template>

<script>
    export default {
        props: {
            curCluster: {
                type: Object
            }
        },
        computed: {
        },
        beforeDestroy () {
        },
        methods: {
        }
    }
</script>

<style lang="postcss">
    @import './status-mark-corner.css';
</style>
