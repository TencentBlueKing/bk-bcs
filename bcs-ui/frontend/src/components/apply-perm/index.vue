<template>
    <bk-dialog
        :ext-cls="'apply-perm-dialog'"
        :is-show.sync="dialogConf.isShow"
        :width="dialogConf.width"
        :quick-close="false"
        @cancel="hide"
        :title="dialogConf.title">
        <template slot="content">
            <table class="bk-table has-table-hover biz-table biz-apply-perm-table">
                <thead>
                    <tr>
                        <th style="width: 260px; padding-left: 20px;">{{$t('资源')}}</th>
                        <th style="width: 180px;">{{$t('需要的权限')}}</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="(perm, index) in permList" :key="index">
                        <td style="padding-left: 20px;">
                            <span v-if="perm.policy_code !== 'create'">{{perm.resource_type_name}}：</span>{{perm.resource_name || perm.resource_type_name}}
                        </td>
                        <td>{{perm.policy_name}}</td>
                    </tr>
                </tbody>
            </table>
        </template>
        <div slot="footer">
            <div class="bk-dialog-outer">
                <bk-button type="primary" @click="goApplyUrl">{{$t('去申请')}}</bk-button>
                <bk-button type="button" @click="hide">{{$t('取消')}}</bk-button>
            </div>
        </div>
    </bk-dialog>
</template>

<script>
    export default {
        name: 'apply-perm',
        data () {
            return {
                dialogConf: {
                    isShow: false,
                    width: 640,
                    title: this.$t('无权限操作'),
                    closeIcon: false
                },
                applyUrl: '',
                permList: []
            }
        },
        destroyed () {
            this.applyUrl = ''
        },
        methods: {
            hide () {
                this.dialogConf.isShow = false
            },
            show (projectCode, data) {
                this.applyUrl = `${data.apply_url}&project_code=${projectCode}`

                this.permList.splice(0, this.permList.length, ...(data.perms || []))
                this.$nextTick(() => {
                    this.dialogConf.isShow = true
                })
            },
            goApplyUrl () {
                this.hide()
                setTimeout(() => {
                    window.open(this.applyUrl)
                }, 300)
            }
        }
    }
</script>

<style>
    @import './index.css';
</style>
