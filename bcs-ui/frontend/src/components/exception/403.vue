<template>
    <bk-exception :type="type">
        <span>{{$t('无权限')}}</span>
        <a class="bk-text-button text-wrap" @click="handleGotoIAM">{{$t('去申请')}}</a>
    </bk-exception>
</template>
<script>
    import { defineComponent, onMounted, ref } from '@vue/composition-api'
    import { userPermsByAction } from '@/api/base'

    export default defineComponent({
        props: {
            type: {
                type: String,
                default: '403'
            },
            actionId: {
                type: String,
                default: ''
            },
            permCtx: {
                type: Object,
                default: () => ({})
            }
        },
        setup (props) {
            const href = ref('')
            const handleGotoIAM = () => {
                window.open(href.value)
            }
            onMounted(async () => {
                if (!props.actionId) return
                const data = await userPermsByAction({
                    $actionId: [props.actionId],
                    perm_ctx: props.permCtx
                }).catch(() => ({}))
                // eslint-disable-next-line camelcase
                href.value = data?.perms?.apply_url
            })
            return {
                handleGotoIAM
            }
        }
    })
</script>
