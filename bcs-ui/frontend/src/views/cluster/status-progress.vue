<template>
    <div class="biz-cluster-content">
        <template>
            <div class="biz-progress-box">
                <div class="progress-header">
                    <span class="title">{{$t('CPU使用率')}}</span>
                    <span class="percent" v-if="!loading">
                        {{$parent.conversionPercent(curCluster.cpu_usage.used, curCluster.cpu_usage.total)}}%
                    </span>
                </div>
                <div class="progress" :class="loading ? 'loading' : ''">
                    <!-- <div class="progress loading"> -->
                    <div class="progress-bar primary" v-if="loading" style="width: 0%"></div>
                    <div class="progress-bar primary" v-else
                        :style="{ width: `${$parent.conversionPercent(curCluster.cpu_usage.used, curCluster.cpu_usage.total)}%` }"></div>
                </div>
            </div>
            <div class="biz-progress-box">
                <div class="progress-header">
                    <span class="title">{{$t('内存使用率')}}</span>
                    <span class="percent" v-if="!loading">
                        {{$parent.conversionPercent(curCluster.mem_usage.used_bytes, curCluster.mem_usage.total_bytes)}}%
                    </span>
                </div>
                <div class="progress" :class="loading ? 'loading' : ''">
                    <!-- <div class="progress loading"> -->
                    <div class="progress-bar success" v-if="loading" style="width: 0%"></div>
                    <div class="progress-bar success" v-else :style="{ width: `${$parent.conversionPercent(curCluster.mem_usage.used_bytes, curCluster.mem_usage.total_bytes)}%` }"></div>
                </div>
            </div>
            <div class="biz-progress-box">
                <div class="progress-header">
                    <span class="title">{{$t('磁盘使用率')}}</span>
                    <span class="percent" v-if="!loading">
                        {{$parent.conversionPercent(curCluster.disk_usage.used_bytes, curCluster.disk_usage.total_bytes)}}%
                    </span>
                </div>
                <div class="progress" :class="loading ? 'loading' : ''">
                    <!-- <div class="progress loading"> -->
                    <div class="progress-bar warning" v-if="loading" style="width: 0%"></div>
                    <div class="progress-bar warning" v-else :style="{ width: `${$parent.conversionPercent(curCluster.disk_usage.used_bytes, curCluster.disk_usage.total_bytes)}%` }"></div>
                </div>
            </div>
        </template>
    </div>
</template>

<script>
    export default {
        props: {
            curProject: {
                type: Object
            },
            curCluster: {
                type: Object
            },
            // curClusterIndex: {
            //     type: Number
            // },
            // overviewList: {
            //     type: Array
            // },
            loading: {
                type: Boolean
            }
        },
        computed: {
        },
        // watch: {
        //     loading (v) {
        //         console.error(v)
        //     }
        // },
        created () {
        },
        beforeDestroy () {
        },
        methods: {
        }
    }
</script>

<style lang="postcss">
    @import './status-progress.css';
</style>
