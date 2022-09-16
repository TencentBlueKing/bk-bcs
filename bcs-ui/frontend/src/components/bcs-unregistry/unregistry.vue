<template>
    <!-- 容器服务未开通时引导界面 -->
    <section class="bcs-unregistry">
        <header class="header">{{ $t('开通容器服务') }}</header>
        <main class="main">
            <div class="form-item">
                <div class="form-item-label">{{ $t('业务编排类型') }}</div>
                <div class="form-item-content kind">
                    <div class="kind-panel active">
                        <div class="kind-panel-title">K8S</div>
                        <div class="kind-panel-desc mt5">{{ $t('k8s容器编排引擎') }}</div>
                    </div>
                </div>
            </div>
            <div class="form-item mt30">
                <div class="form-item-label">{{ $t('关联CMDB业务') }}</div>
                <div class="form-item-content cc-list">
                    <bk-select class="cc-selector"
                        :placeholder="$t('请选择关联业务')"
                        v-model="ccKey"
                        :disabled="!ccList.length"
                        searchable
                        @change="handleCmdbChange">
                        <bk-option v-for="item in ccList"
                            :key="item.id"
                            :id="item.id"
                            :name="item.name">
                        </bk-option>
                    </bk-select>
                </div>
                <div class="form-item-tips" v-if="!ccList.length && $INTERNAL">
                    {{ $t('请联系需要关联的CMDB业务的运维') }}
                    <bk-link theme="primary"
                        :href="PROJECT_CONFIG.doc.iam"
                        target="_blank">
                        {{ $t('申请权限') }}
                    </bk-link>
                </div>
            </div>
            <div class="form-item enable-bcs">
                <bk-button theme="primary" :disabled="!enableBtn" @click="updateProject">{{ $t('启用容器服务') }}</bk-button>
            </div>
            <div class="form-item guide" v-if="$INTERNAL">
                <div v-for="(item, index) in guideList"
                    :class="['guide-link', { mr18: index < (guideList.length - 1) }]"
                    :key="item.id">
                    <i :class="`bcs-icon bcs-icon-${item.id}`" :style="{ color: item.iconColor }"></i>
                    <span class="desc mt20">{{ item.desc }}</span>
                    <bk-link class="mt8" theme="primary" :href="item.link" target="_blank">{{ item.linkText }}</bk-link>
                </div>
            </div>
        </main>
    </section>
</template>
<script>
    import { isEmpty } from '@open/common/util'
    export default {
        name: 'bcs-unregistry',
        props: {
            ccList: {
                type: Array,
                default: () => []
            },
            defaultKind: {
                type: Number,
                default: 1
            }
        },
        data () {
            return {
                kindList: [
                    {
                        id: 1,
                        name: 'K8S',
                        desc: this.$t('k8s容器编排引擎')
                    }
                ],
                guideList: [],
                kind: this.defaultKind,
                ccKey: ''
            }
        },
        computed: {
            enableBtn () {
                return !isEmpty(this.ccKey)
            }
        },
        created () {
            this.guideList = [
                {
                    id: 'binding',
                    iconColor: '#4540DC',
                    desc: this.$t('开启容器服务时，请首先在”蓝鲸配置平台“查看业务'),
                    link: this.PROJECT_CONFIG.doc.cc,
                    linkText: this.$t('前往绑定业务')
                },
                {
                    id: 'auth',
                    iconColor: '#FFB200',
                    desc: this.$t('开启容器服务时，若没有查看业务权限，去“权限中心”申请权限'),
                    link: this.PROJECT_CONFIG.doc.iam,
                    linkText: this.$t('申请权限')
                },
                {
                    id: 'wiki',
                    iconColor: '#66EFE3',
                    desc: this.$t('如果遇到更多问题，需要了解详细信息，请前往iwiki查看'),
                    link: this.PROJECT_CONFIG.doc.quickStart,
                    linkText: this.$t('前往iwiki查看')
                }
            ]
        },
        methods: {
            handleCmdbChange (value, oldvalue) {
                this.$emit('cc-change', value, oldvalue)
            },
            updateProject () {
                this.$emit('update-project')
            }
        }
    }
</script>
<style scoped lang="postcss">
@define-mixin row-center {
    display: flex;
    justify-content: center;
}
@define-mixin column-center {
    display: flex;
    flex-direction: column;
    align-items: center;
}

.mt8 {
    margin-top: 8px;
}

.bcs-unregistry {
    overflow: hidden;
    font-size: 12px;
    .header {
        @mixin row-center;
        font-size: 20px;
        line-height: 20px;
        color: #222;
        margin-top: 86px;
        margin-bottom: 32px;
    }
    .main {
        @mixin column-center;
        .form-item {
            width: 720px;
            &-label {
                font-size: 14px;
                line-height: 14px;
                color: #000;
                margin-bottom: 8px;
            }
            &-content {
                @mixin row-center;
                .kind-panel {
                    min-width: 350px;
                    height: 64px;
                    border: 1px solid #c3cdd7;
                    border-radius: 3px;
                    background: #fff;
                    cursor: pointer;
                    padding: 12px 16px;
                    &:nth-child(1) {
                        margin-right: 24px;
                    }
                    &.active {
                        background: #E1ECFF;
                        border-color: #3A84FF;
                    }
                    &.disabled {
                        border-color: #dcdee5;
                        color: #c4c6cc;
                        cursor: not-allowed;
                    }
                    &-title {
                        font-size: 14px;
                        font-weight: 700;
                    }
                }
                .cc-selector {
                    width: 100%;
                }
                &.cc-list {
                    background: #fff;
                }
                &.kind {
                    justify-content: flex-start;
                }
            }
            &-tips {
                color: #63656e;
                margin-top: 8px;
                display: flex;
                align-items: center;
                >>> .bk-link-text {
                    font-size: 12px;
                    margin-left: 2px;
                }
            }
            &.enable-bcs {
                margin-top: 36px;
            }
            &.guide {
                display: flex;
                margin-top: 82px;
                .guide-link {
                    display: flex;
                    flex-direction: column;
                    align-items: flex-start;
                    width: 228px;
                    min-height: 160px;
                    background: #fff;
                    border-radius: 2px;
                    box-shadow: 0px 1px 1px 0px rgba(0,0,0,.09);
                    padding: 24px;
                    &.mr18 {
                        margin-right: 18px;
                    }
                    i {
                        font-size: 24px;
                    }
                }
            }
        }
    }
}
</style>
