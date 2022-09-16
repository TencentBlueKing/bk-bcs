<template>
    <div class="bk-form biz-configuration-form">
        <div class="bk-form-item">
            <div class="bk-form-item">
                <div class="bk-form-content" style="margin-left: 0;">
                    <div class="bk-form-item is-required">
                        <label class="bk-label" style="width: 130px;">{{$t('名称')}}：</label>
                        <div class="bk-form-content" style="margin-left: 130px;">
                            <input type="text" :class="['bk-form-input',{ 'is-danger': errors.has('applicationName') }]" :placeholder="$t('请输入64个字符以内')" style="width: 310px;" v-model="curIngress.config.metadata.name" maxlength="64" name="applicationName" v-validate="{ required: true, regex: /^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$/ }">
                        </div>
                        <div class="bk-form-tip is-danger" style="margin-left: 130px;" v-if="errors.has('applicationName')">
                            <p class="bk-tip-text">{{$t('名称必填，以小写字母或数字开头和结尾，只能包含：小写字母、数字、连字符(-)、点(.)')}}</p>
                        </div>
                    </div>
                </div>
            </div>

            <div class="bk-form-item">
                <div class="bk-form-content" style="margin-left: 130px;">
                    <bk-button :class="['bk-text-button f12 mb10 pl0', { 'rotate': isTlsPanelShow }]" @click.stop.prevent="toggleTlsPanel">
                        {{$t('TLS设置')}}<i class="bcs-icon bcs-icon-angle-double-down ml5"></i>
                    </bk-button>
                    <bk-button :class="['bk-text-button f12 mb10 pl0', { 'rotate': isPanelShow }]" @click.stop.prevent="togglePanel">
                        {{$t('更多设置')}}<i class="bcs-icon bcs-icon-angle-double-down ml5"></i>
                    </bk-button>
                </div>
            </div>
            <div class="bk-form-item mt0" v-show="isTlsPanelShow">
                <div class="bk-form-content" style="margin-left: 130px;">
                    <bk-tab :type="'fill'" :active-name="'tls'" :size="'small'">
                        <bk-tab-panel name="tls" title="TLS">
                            <div class="p20">
                                <table class="biz-simple-table">
                                    <tbody>
                                        <tr v-for="(computer, index) in curIngress.config.spec.tls" :key="index">
                                            <td>
                                                <bkbcs-input
                                                    type="text"
                                                    :placeholder="$t('主机名，多个用英文逗号分隔')"
                                                    style="width: 310px;"
                                                    :value.sync="computer.hosts"
                                                    :list="varList"
                                                >
                                                </bkbcs-input>
                                            </td>
                                            <td>
                                                <bkbcs-input
                                                    type="text"
                                                    :placeholder="$t('请输入证书')"
                                                    style="width: 350px;"
                                                    :value.sync="computer.secretName"
                                                >
                                                </bkbcs-input>
                                            </td>
                                            <td>
                                                <bk-button class="action-btn ml5" @click.stop.prevent="addTls">
                                                    <i class="bcs-icon bcs-icon-plus"></i>
                                                </bk-button>
                                                <bk-button class="action-btn" v-if="curIngress.config.spec.tls.length > 1" @click.stop.prevent="removeTls(index, computer)">
                                                    <i class="bcs-icon bcs-icon-minus"></i>
                                                </bk-button>
                                            </td>
                                        </tr>
                                    </tbody>
                                </table>
                            </div>
                        </bk-tab-panel>
                    </bk-tab>
                </div>
            </div>

            <div class="bk-form-item mt0" v-show="isPanelShow">
                <div class="bk-form-content" style="margin-left: 130px;">
                    <bk-tab :type="'fill'" :active-name="'remark'" :size="'small'">
                        <bk-tab-panel name="remark" :title="$t('注解')">
                            <div class="biz-tab-wrapper m20">
                                <bk-keyer :key-list.sync="curRemarkList" :var-list="varList" ref="remarkKeyer"></bk-keyer>
                            </div>
                        </bk-tab-panel>
                        <bk-tab-panel name="label" :title="$t('标签')">
                            <div class="biz-tab-wrapper m20">
                                <bk-keyer :key-list.sync="curLabelList" :var-list="varList" ref="labelKeyer"></bk-keyer>
                            </div>
                        </bk-tab-panel>
                    </bk-tab>
                </div>
            </div>

            <!-- part2 start -->
            <div class="biz-part-header">
                <div class="bk-button-group">
                    <div class="item" v-for="(rule, index) in curIngress.config.spec.rules" :key="index">
                        <bk-button :class="['bk-button bk-default is-outline', { 'is-selected': curRuleIndex === index }]" @click.stop="setCurRule(rule, index)">
                            {{rule.host || $t('未命名')}}
                        </bk-button>
                        <span class="bcs-icon bcs-icon-close-circle" @click.stop="removeRule(index)" v-if="curIngress.config.spec.rules.length > 1"></span>
                    </div>
                    <bcs-popover ref="containerTooltip" :content="$t('添加Rule')" placement="top">
                        <bk-button type="button" class="bk-button bk-default is-outline is-icon" @click.stop.prevent="addLocalRule">
                            <i class="bcs-icon bcs-icon-plus"></i>
                        </bk-button>
                    </bcs-popover>
                </div>
            </div>

            <div class="bk-form biz-configuration-form pb15">
                <div class="biz-span">
                    <span class="title">{{$t('基础信息')}}</span>
                </div>
                <div class="bk-form-item is-required">
                    <label class="bk-label" style="width: 130px;">{{$t('虚拟主机名')}}：</label>
                    <div class="bk-form-content" style="margin-left: 130px;">
                        <bk-input :placeholder="$t('请输入')" style="width: 310px;" v-model="curRule.host" name="ruleName" />
                    </div>
                </div>
                <div class="bk-form-item">
                    <label class="bk-label" style="width: 130px;">{{$t('路径组')}}：</label>
                    <div class="bk-form-content" style="margin-left: 130px;">
                        <table class="biz-simple-table">
                            <tbody>
                                <tr v-for="(pathRule, index) of curRule.http.paths" :key="index">
                                    <td>
                                        <bkbcs-input
                                            type="text"
                                            :placeholder="$t('路径')"
                                            style="width: 310px;"
                                            :value.sync="pathRule.path"
                                            :list="varList"
                                        >
                                        </bkbcs-input>
                                    </td>
                                    <td style="text-align: center;">
                                        <i class="bcs-icon bcs-icon-arrows-right"></i>
                                    </td>
                                    <td>
                                        <bk-selector
                                            style="width: 180px;"
                                            :placeholder="$t('Service名称')"
                                            :disabled="isLoadBalanceEdited"
                                            :setting-key="'_name'"
                                            :display-key="'_name'"
                                            :selected.sync="pathRule.backend.serviceName"
                                            :list="linkServices || []"
                                            @item-selected="handlerSelectService(pathRule)">
                                        </bk-selector>
                                    </td>
                                    <td>
                                        <bk-selector
                                            style="width: 180px;"
                                            :placeholder="$t('端口')"
                                            :disabled="isLoadBalanceEdited"
                                            :setting-key="'_id'"
                                            :display-key="'_name'"
                                            :selected.sync="pathRule.backend.servicePort"
                                            :list="linkServices[pathRule.backend.serviceName] || []">
                                        </bk-selector>
                                    </td>
                                    <td>
                                        <bk-button class="action-btn ml5" @click.stop.prevent="addRulePath">
                                            <i class="bcs-icon bcs-icon-plus"></i>
                                        </bk-button>
                                        <bk-button class="action-btn" v-if="curRule.http.paths.length > 1" @click.stop.prevent="removeRulePath(pathRule, index)">
                                            <i class="bcs-icon bcs-icon-minus"></i>
                                        </bk-button>
                                    </td>
                                </tr>
                            </tbody>
                        </table>
                        <p class="biz-tip">{{$t('提示：同一个虚拟主机名可以有多个路径')}}</p>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>
<script>
    import bkKeyer from '@/components/keyer'
    import ingressParams from '@/json/k8s-ingress.json'
    import ruleParams from '@/json/k8s-ingress-rule.json'

    export default {
        components: {
            bkKeyer
        },
        props: {
            ingressData: {
                type: Object,
                default () {
                    return JSON.parse(JSON.stringify(ingressParams))
                }
            }
        },
        data () {
            return {
                curRuleIndex: 0,
                isPanelShow: false,
                isTlsPanelShow: false,
                curIngress: this.ingressData,
                curRule: this.ingressData.config.spec.rules[0],
                computerList: [{
                    name: '',
                    cert: ''
                }]
            }
        },
        computed: {
            projectId () {
                return this.$route.params.projectId
            },
            linkServices () {
                const list = this.$store.state.k8sTemplate.linkServices.map(item => {
                    item._id = item.service_tag
                    item._name = item.service_name
                    return item
                })
                list.forEach(item => {
                    list[item.service_name] = []
                    item.service_ports.forEach(port => {
                        list[item.service_name].push({
                            _id: port,
                            _name: port
                        })
                    })
                })
                return list
            },
            serviceNames () {
                return this.$store.state.k8sTemplate.linkServices.map(item => {
                    return item.service_name
                })
            },
            varList () {
                const list = this.$store.state.variable.varList.map(item => {
                    item._id = item.key
                    item._name = item.key
                    return item
                })
                return list
            },
            curLabelList () {
                const list = []
                // 如果有缓存直接使用
                if (this.curIngress.config.webCache && this.curIngress.config.webCache.labelListCache) {
                    return this.curIngress.config.webCache.labelListCache
                }
                const labels = this.curIngress.config.metadata.labels
                for (const [key, value] of Object.entries(labels)) {
                    list.push({
                        key: key,
                        value: value
                    })
                }
                if (!list.length) {
                    list.push({
                        key: '',
                        value: ''
                    })
                }
                return list
            },
            curRemarkList () {
                const list = []
                // 如果有缓存直接使用
                if (this.curIngress.config.webCache && this.curIngress.config.webCache.remarkListCache) {
                    return this.curIngress.config.webCache.remarkListCache
                }
                const annotations = this.curIngress.config.metadata.annotations
                for (const [key, value] of Object.entries(annotations)) {
                    list.push({
                        key: key,
                        value: value
                    })
                }
                if (!list.length) {
                    list.push({
                        key: '',
                        value: ''
                    })
                }
                return list
            }
        },
        mounted () {
            this.$nextTick(() => {
                this.checkService()
            })
        },
        methods: {
            handlerSelectCert (computer, index, data) {
                computer.certType = data.certType
            },
            addTls () {
                this.curIngress.config.spec.tls.push({
                    hosts: '',
                    secretName: ''
                })
            },
            removeTls (index, curTls) {
                this.curIngress.config.spec.tls.splice(index, 1)
            },
            togglePanel () {
                this.isTlsPanelShow = false
                this.isPanelShow = !this.isPanelShow
            },
            toggleTlsPanel () {
                this.isPanelShow = false
                this.isTlsPanelShow = !this.isTlsPanelShow
            },
            setCurRule (rule, index) {
                this.curRule = rule
                this.curRuleIndex = index
            },
            removeRule (index) {
                const rules = this.curIngress.config.spec.rules
                rules.splice(index, 1)
                if (this.curRuleIndex === index) {
                    this.curRuleIndex = 0
                } else {
                    this.curRuleIndex = this.curRuleIndex - 1
                }

                this.curRule = rules[this.curRuleIndex]
            },
            addLocalRule () {
                const rule = JSON.parse(JSON.stringify(ruleParams))
                const rules = this.curIngress.config.spec.rules
                const index = rules.length
                rule.host = 'rule-' + (index + 1)
                rules.push(rule)
                this.setCurRule(rule, index)
                this.$refs.containerTooltip.visible = false
            },
            addRulePath () {
                const params = {
                    backend: {
                        serviceName: '',
                        servicePort: ''
                    },
                    path: ''
                }

                this.curRule.http.paths.push(params)
            },
            removeRulePath (pathRule, index) {
                this.curRule.http.paths.splice(index, 1)
            },
            handlerSelectService (pathRule) {
                pathRule.backend.servicePort = ''
            },
            checkService (pathRule) {
                const rules = this.curIngress.config.spec.rules
                for (const rule of rules) {
                    const paths = rule.http.paths
                    for (const path of paths) {
                        if (path.backend.serviceName && !this.serviceNames.includes(path.backend.serviceName)) {
                            this.$bkMessage({
                                theme: 'error',
                                message: this.$t('{name}中路径组：关联的Service【{serviceName}】不存在，请重新绑定', {
                                    name: this.curIngress.config.metadata.name,
                                    serviceName: path.backend.serviceName
                                }),
                                delay: 5000
                            })
                            return false
                        }
                    }
                }
                return true
            }
        }
    }
</script>

<style scoped>
    @import '@/css/variable.css';
    @import '@/css/mixins/clearfix.css';

    .biz-simple-table {
        width: auto;
    }
    .action-btn {
        height: 36px;
        text-align: center;
        display: inline-block;
        border: none;
        background: transparent;
        outline: none;
        float: left;
        .bcs-icon {
            width: 24px;
            height: 24px;
            line-height: 24px;
            border-radius: 50%;
            vertical-align: middle;
            border: 1px solid #dde4eb;
            color: #737987;
            font-size: 14px;
            display: inline-block;
            &.icon-minus {
                font-size: 15px;
            }
        }
    }
    .biz-part-header {
        margin-top: 40px;
        text-align: center;
    }
    .bk-text-button {
        .bcs-icon {
            transition: all ease 0.3s;
        }
        &.rotate {
            .bcs-icon {
                transform: rotate(180deg);
            }
        }
    }

    .bk-form .bk-form-content .bk-form-tip {
        overflow: hidden;
        padding: 0;
        margin: 10px 0 0 0;
        position: relative;
        height: auto;
        line-height: 1;
        left: 0;
    }
</style>
