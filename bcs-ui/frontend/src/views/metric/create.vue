<template>
    <bk-sideslider
        :is-show.sync="isVisible"
        :title="createMetricConf.title"
        :width="createMetricConf.width"
        :quick-close="false"
        class="biz-metric-manage-create-sideslider"
        @hidden="hideCreateMetric">
        <div slot="content">
            <div class="wrapper" style="position: relative;">
                <form class="bk-form bk-form-vertical create-form">
                    <div class="bk-form-item">
                        <label class="bk-label label">{{$t('名称')}}：<span class="red">*</span></label>
                        <div class="bk-form-content">
                            <bk-input style="width: 282px;" v-model="createParams.name" :placeholder="$t('请输入')" maxlength="253" />
                        </div>
                    </div>
                    <div class="bk-form-item">
                        <label class="bk-label label">
                            {{$t('选择Service')}}：<span class="red">*</span>
                            <span class="tip">{{$t('选择Service以获取Label')}}</span>
                        </label>
                        <div class="bk-form-content flex-item">
                            <div class="left">
                                <div class="bk-form-content">
                                    <bkbcs-input style="cursor: not-allowed;" type="text" disabled :value.sync="clusterName"></bkbcs-input>
                                </div>
                            </div>
                            <div class="right">
                                <div class="bk-form-content">
                                    <bk-selector
                                        searchable
                                        :selected.sync="createParams.displayName"
                                        :list="serviceList"
                                        :search-key="'displayName'"
                                        :setting-key="'displayName'"
                                        :display-key="'displayName'"
                                        @item-selected="changeSelectedService">
                                    </bk-selector>
                                </div>
                            </div>
                        </div>
                    </div>
                    <div class="bk-form-item" v-if="keyValueData">
                        <label class="bk-label label">
                            {{$t('选择关联Label')}}：<span class="red">*</span>
                        </label>
                        <div class="bk-form-content" v-if="Object.keys(keyValueData).length">
                            <div class="bk-keyer http-header">
                                <div class="biz-keys-list mb10">
                                    <div class="biz-key-item" v-for="(key, keyIndex) in Object.keys(keyValueData)" :key="keyIndex">
                                        <bk-input style="width: 235px;" :disabled="true" :value="key" />
                                        <span class="operator">=</span>
                                        <bk-input style="width: 235px;" :disabled="true" :value="keyValueData[key]" />
                                        <bk-checkbox class="ml10" :value="!!createParams.selector[key]" @change="valueChange(arguments[0], key)"></bk-checkbox>
                                    </div>
                                </div>
                            </div>
                        </div>
                        <div class="bk-form-content" v-else>
                            <i18n path="当前Service没有设置Labels，{action}" style="color: #737987;font-size: 14px;" tag="div">
                                <a place="action" class="bk-text-button metric-query" @click="goService" href="javascript:void(0)">{{$t('请先添加')}}</a>
                            </i18n>
                        </div>
                    </div>
                    <div class="bk-form-item" v-if="portList.length">
                        <label class="bk-label label">{{$t('选择PortName')}}：<span class="red">*</span></label>
                        <div class="bk-form-content">
                            <bk-selector
                                searchable
                                :selected.sync="createParams.port"
                                :list="portList"
                                :setting-key="'name'"
                                :display-key="'name'"
                                @item-selected="changeSelectedPort">
                            </bk-selector>
                        </div>
                    </div>
                    <div class="bk-form-item">
                        <label class="bk-label label">
                            {{$t('Metric路径')}}：<span class="red">*</span>
                        </label>
                        <div class="bk-form-content">
                            <bk-input v-model="createParams.path" :placeholder="$t('请输入')" />
                        </div>
                    </div>
                    <div class="bk-form-item">
                        <label class="bk-label label">
                            {{$t('Metric参数')}}：
                        </label>
                        <div class="bk-form-content">
                            <metric-keyer :key-list="metricParamsList"></metric-keyer>
                        </div>
                    </div>
                    <div class="bk-form-item flex-item">
                        <div class="left">
                            <label class="bk-label label">{{$t('采集周期（秒）')}}：<span class="red">*</span></label>
                            <div class="bk-form-content">
                                <bk-number-input
                                    class="text-input-half"
                                    :value.sync="createParams.interval"
                                    :min="0"
                                    :max="999999999"
                                    :debounce-timer="0"
                                    :placeholder="$t('请输入')">
                                </bk-number-input>
                            </div>
                        </div>
                        <div class="right">
                            <label class="bk-label label">{{$t('允许最大Sample数')}}：<span class="red">*</span></label>
                            <div class="bk-form-content">
                                <bk-number-input
                                    class="text-input-half"
                                    :value.sync="createParams.sample_limit"
                                    :min="0"
                                    :debounce-timer="0"
                                    :placeholder="$t('请输入')">
                                </bk-number-input>
                            </div>
                        </div>
                    </div>
                    <div class="action-inner">
                        <bk-button type="primary" :loading="isLoading" @click="confirmCreateMetric">
                            {{$t('提交')}}
                        </bk-button>
                        <bk-button type="button" :disabled="isLoading" @click="hideCreateMetric">
                            {{$t('取消')}}
                        </bk-button>
                    </div>
                </form>
            </div>
        </div>
    </bk-sideslider>
</template>

<script>
    import MetricKeyer from './keyer'

    export default {
        name: 'CreateMetric',
        components: {
            MetricKeyer
        },
        props: {
            isShow: {
                type: Boolean,
                default: false
            },
            clusterId: {
                type: String
            },
            clusterName: {
                type: String
            }
        },
        data () {
            return {
                bkMessageInstance: null,
                createMetricConf: {
                    isShow: false,
                    title: this.$t('新建Metric'),
                    timer: null,
                    width: 644,
                    loading: false
                },
                // 创建的参数
                createParams: {
                    name: '',
                    cluster_id: '',
                    namespace: '',
                    service_name: '',
                    displayName: '',
                    path: '/metrics',
                    selector: {},
                    interval: 30,
                    port: '',
                    sample_limit: 10000
                },

                isVisible: false,
                isLoading: false,
                keyValueData: null,
                portList: [],
                serviceList: [],
                metricParamsList: [{ key: '', value: '' }]
            }
        },
        computed: {
            projectId () {
                return this.$route.params.projectId
            },
            projectCode () {
                return this.$route.params.projectCode
            },
            isEn () {
                return this.$store.state.isEn
            }
        },
        watch: {
            isShow: {
                async handler (newVal, oldVal) {
                    this.isVisible = newVal
                    if (this.isVisible) {
                        this.isLoading = true
                        await this.fetchService()
                    }
                },
                immediate: true
            }
        },
        mounted () {
        },
        destroyed () {
            this.bkMessageInstance && this.bkMessageInstance.close()
        },
        methods: {
            /**
             * 获取 service 数据
             */
            async fetchService () {
                try {
                    const res = await this.$store.dispatch('metric/listServices', {
                        projectId: this.projectId,
                        clusterId: this.clusterId
                    })
                    const list = res.data || []
                    list.forEach(item => {
                        item.displayName = `${item.namespace}/${item.resourceName}`
                    })
                    this.serviceList.splice(0, this.serviceList.length, ...list)
                } catch (e) {
                    console.error(e)
                } finally {
                    this.isLoading = false
                }
            },

            /**
             * 切换 service 下拉框改变事件
             *
             * @param {string} displayName 选择的 service 的 displayName
             */
            changeSelectedService (displayName) {
                this.createParams.service_name = displayName.split('/')[1]
                this.createParams.displayName = displayName
                const curSelectedService = this.serviceList.find(
                    item => item.displayName === this.createParams.displayName
                )
                if (curSelectedService) {
                    this.keyValueData = Object.assign({}, curSelectedService.data.metadata.labels)
                    this.portList.splice(0, this.portList.length, ...(curSelectedService.data.spec.ports || []))
                    this.createParams.namespace = curSelectedService.namespace
                }
            },

            /**
             * 切换 port 下拉框改变事件
             *
             * @param {string} port 选择的 port name
             */
            changeSelectedPort (port) {
                this.createParams.port = port
            },

            /**
             * 选择关联 Label 多选框勾选
             *
             * @param {Object} e 事件对象
             * @param {string} key 勾选的 key
             */
            valueChange (checked, key) {
                if (checked) {
                    this.createParams.selector[key] = this.keyValueData[key]
                } else {
                    delete this.createParams.selector[key]
                }
            },

            /**
             * 重置添加 metric 的参数
             */
            resetParams () {
                this.keyValueData = null
                this.portList.splice(0, this.portList.length, ...[])
                this.createParams = Object.assign({}, {
                    name: '',
                    cluster_id: '',
                    namespace: '',
                    service_name: '',
                    displayName: '',
                    path: '/metrics',
                    selector: {},
                    interval: 30,
                    port: '',
                    sample_limit: 10000
                })
            },

            /**
             * 隐藏创建 metric sideslider
             */
            hideCreateMetric () {
                this.createMetricConf.isShow = false

                this.resetParams()
                this.$emit('hide-create-metric', false)
            },

            /**
             * 创建 Metric 确定按钮
             */
            async confirmCreateMetric () {
                const name = this.createParams.name.trim()
                const path = this.createParams.path.trim()
                const port = this.createParams.port
                const serviceName = this.createParams.service_name
                const selector = this.createParams.selector
                const interval = this.createParams.interval
                const sampleLimit = this.createParams.sample_limit

                if (!name) {
                    this.$bkMessage({ theme: 'error', message: this.$t('请输入名称') })
                    return
                }

                if (name.length < 3) {
                    this.$bkMessage({ theme: 'error', message: this.$t('名称不得小于三个字符') })
                    return
                }

                if (!serviceName) {
                    this.$bkMessage({ theme: 'error', message: this.$t('请选择Service') })
                    return
                }

                if (!Object.keys(selector).length) {
                    this.$bkMessage({ theme: 'error', message: this.$t('至少要关联一个Label') })
                    return
                }

                if (!port) {
                    this.$bkMessage({ theme: 'error', message: this.$t('请选择PortName') })
                    return
                }

                if (!path) {
                    this.$bkMessage({ theme: 'error', message: this.$t('请输入Metric路径') })
                    return
                }

                if (path.length < 2) {
                    this.$bkMessage({ theme: 'error', message: this.$t('Metric路径不得小于两个字符') })
                    return
                }

                if (interval === null || interval === undefined || interval === '') {
                    this.$bkMessage({ theme: 'error', message: this.$t('请输入采集周期') })
                    return
                }

                if (sampleLimit === '' || sampleLimit === null || sampleLimit === undefined) {
                    this.$bkMessage({ theme: 'error', message: this.$t('请输入允许最大Sample数') })
                    return
                }

                this.createParams.cluster_id = this.clusterId
                this.createParams.projectId = this.projectId

                const metricParamsList = []
                const len = this.metricParamsList.length
                for (let i = 0; i < len; i++) {
                    const metricParams = this.metricParamsList[i]
                    const key = metricParams.key.trim()
                    const value = metricParams.value.trim()
                    if (!key && value) {
                        this.bkMessageInstance && this.bkMessageInstance.close()
                        this.bkMessageInstance = this.$bkMessage({
                            theme: 'error',
                            message: this.$t('请填写参数名称')
                        })
                        return
                    } else if (key && !value) {
                        this.bkMessageInstance && this.bkMessageInstance.close()
                        this.bkMessageInstance = this.$bkMessage({
                            theme: 'error',
                            message: this.$t('请填写参数值')
                        })
                        return
                    }
                    if (key && value) {
                        metricParamsList.push({
                            key,
                            value
                        })
                    }
                }

                if (metricParamsList.length) {
                    this.createParams.params = {}
                    metricParamsList.forEach(m => {
                        this.createParams.params[m.key] = [m.value]
                    })
                }

                try {
                    this.isLoading = true
                    await this.$store.dispatch('metric/createServiceMonitor', Object.assign({}, this.createParams))
                    this.$emit('create-success')
                } catch (e) {
                    console.error(e)
                } finally {
                    this.isLoading = false
                }
            },

            /**
             * 跳转到 service 页面
             */
            goService () {
                this.$router.push({
                    name: 'service'
                })
            }
        }
    }
</script>

<style>
    @import './create.css';
</style>
