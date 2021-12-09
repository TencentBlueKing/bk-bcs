<template>
    <bcs-dialog
        class="sharedClusterDialog"
        v-model="isShow"
        width="720"
        :mask-close="false"
        :close-icon="false"
        title="用户须知">
        <li>共享集群功能当前上线 <b>beta</b> 版本，仅供用户进行功能体验，应 <b>避免部署正式服务</b>，非必要应 <b>避免添加敏感配置</b>（如云 API 密钥等），仍需添加请加密后使用</li>
        <li>平台将对所有用户操作进行审计。请勿尝试提权及访问无权限资源，请勿在集群中运行挖矿程序、病毒木马等恶意或者不确定的程序，在未备案情况下，私自利用安全测试等手段对集群发起探测攻击属于违反安全高压线！</li>
        <li>共享集群体验中 <b>发现问题</b> 或 <b>有优化建议</b>，欢迎随时联系我们进行反馈</li>
        <div slot="footer">
            <bcs-button theme="primary" @click="handleConfirm">{{ $t('确定') }}</bcs-button>
        </div>
    </bcs-dialog>
</template>

<script>
    export default {
        data () {
            return {
                isShow: false
            }
        },
        methods: {
            show () {
                this.isShow = true
            },
            handleConfirm () {
                this.isShow = false
            }
        }
    }
</script>

<style>
    .sharedClusterDialog {
        z-index: 9999;
        .bk-dialog-body {
            li {
                list-style: disc;
                text-align: left;
                margin-bottom: 25px;
            }
        }
    }
</style>
