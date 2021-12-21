<template>
    <bk-dialog
        :class="'biz-warning-tip'"
        :is-show.sync="tipDialogConf.isShow"
        :has-header="false"
        :has-footer="false"
        :quick-close="false"
        :close-icon="showClose"
        :width="tipDialogConf.width"
        :title="tipDialogConf.title"
        @cancel="cancel">
        <template slot="content">
            <div :class="`dialog-wrapper ${type}`">
                <div class="dialog-header">
                    <div class="logo">
                        <i :class="icon"></i>
                    </div>
                    <h2 class="dialog-title">{{titleRender}}</h2>
                    <div v-if="tips" class="dialog-tip">
                        {{ tips }}
                    </div>
                    <!-- <span v-if="showClose" class="close-btn" title="关闭" @click="cancel">╳</span> -->
                </div>
                <div class="dialog-content">
                    <strong>{{subTitleRender}}</strong>
                    <ul class="update-list">
                        <li v-for="(item, index) of noticeList" :key="index">
                            <bk-checkbox name="check" v-model="item.isChecked" :disabled="isConfirming" @change="changeCheck(item)">{{item.text}}</bk-checkbox>
                        </li>
                    </ul>
                    <template v-if="canConfirm">
                        <div class="dialog-action">
                            <bk-button :disabled="isConfirming" :loading="confirmLoading" class="bk-button bk-primary bk-button-large dialog-btn" @click="confirm">{{confirmBtnTextRender}}</bk-button>
                            <bk-button :disabled="isConfirming" class="bk-button bk-default bk-button-large dialog-btn" @click="cancel">{{cancelBtnTextRender}}</bk-button>
                        </div>

                    </template>
                    <template v-else>
                        <div class="dialog-action">
                            <bk-popover :content="$t('请确认以上内容，才可操作')" placement="top">
                                <a href="javascript:void(0)" class="bk-button bk-primary bk-button-large dialog-btn disabled">{{confirmBtnTextRender}}</a>
                            </bk-popover>
                            <a href="javascript:void(0)" class="bk-button bk-default bk-button-large dialog-btn" @click="cancel" style="margin-left: 10px;">{{cancelBtnTextRender}}</a>
                        </div>
                    </template>
                </div>
            </div>
        </template>
    </bk-dialog>
</template>

<script>
    export default {
        props: {
            type: {
                type: String,
                default: 'default'
            },
            icon: {
                type: String,
                default: 'bk-icon icon-bk'
            },
            title: {
                type: String,
                default: ''
            },
            subTitle: {
                type: String,
                default: ''
            },
            checkList: {
                type: Array,
                default () {
                    return []
                }
            },
            confirmBtnText: {
                type: String,
                default: ''
            },
            confirmingBtnText: {
                type: String,
                default: ''
            },
            cancelBtnText: {
                type: String,
                default: ''
            },
            cancelingBtnText: {
                type: String,
                default: ''
            },
            confirmCallback: {
                type: Function
            },
            cancelCallback: {
                type: Function
            },
            showClose: {
                type: Boolean,
                default: true
            },
            isConfirming: {
                type: Boolean,
                default: false
            },
            width: {
                type: Number
            },
            confirmLoading: {
                type: Boolean,
                default: false
            },
            tips: {
                type: String,
                default: ''
            }
        },
        data () {
            return {
                tipDialogConf: {
                    isShow: false,
                    width: this.width || 700
                },
                noticeList: [],
                titleRender: '',
                subTitleRender: '',
                confirmBtnTextRender: '',
                confirmingBtnTextRender: '',
                cancelBtnTextRender: '',
                cancelingBtnTextRender: ''
            }
        },
        computed: {
            canConfirm () {
                for (const item of this.noticeList) {
                    if (!item.isChecked) {
                        return false
                    }
                }
                return true
            }
        },
        created () {
            this.checkList.forEach(item => {
                item.isNeedCheck = item.isChecked
            })
            this.noticeList = JSON.parse(JSON.stringify(this.checkList))
            this.titleRender = this.title || this.$t('提示')
            this.subTitleRender = this.subTitle || this.$t('提示')
            this.confirmBtnTextRender = this.confirmBtnText || this.$t('确定')
            this.confirmingBtnTextRender = this.confirmingBtnText || this.$t('执行中...')
            this.cancelBtnTextRender = this.cancelBtnText || this.$t('取消')
            this.cancelingBtnTextRender = this.cancelingBtnText || this.$t('取消中')
        },
        methods: {
            show () {
                // this.checkList.forEach(item => {
                //     item.isNeedCheck = false
                //     item.isChecked = false
                // })
                this.noticeList = JSON.parse(JSON.stringify(this.checkList))
                this.tipDialogConf.isShow = true
            },
            hide () {
                this.tipDialogConf.isShow = false
            },
            changeCheck (item) {
                item.isNeedCheck = !item.isChecked
                this.noticeList = JSON.parse(JSON.stringify(this.noticeList))
            },
            async confirm () {
                const needCheck = []
                for (const item of this.noticeList) {
                    if (!item.isChecked) {
                        needCheck.push(item)
                        item.isNeedCheck = true
                    } else {
                        item.isNeedCheck = false
                    }
                }
                if (needCheck.length) {
                    this.noticeList = JSON.parse(JSON.stringify(this.noticeList))
                    return false
                }

                if (this.confirmCallback && typeof this.confirmCallback === 'function') {
                    await this.confirmCallback()
                }
                this.hide()
            },
            cancel () {
                this.cancelCallback && this.cancelCallback()
                this.hide()
            }
        }
    }
</script>
<style scoped lang="postcss">
    @import './index.css';
    /deep/ .bk-form-checkbox {
        display: flex;
        .bk-checkbox-text {
            flex: 1;
        }
    }
</style>
