<template>
    <div class="content-header">
        <div class="title">
            <i class="bcs-icon bcs-icon-arrows-left back" @click="goBack"></i>
            {{ title }}
        </div>
        <bk-guide></bk-guide>
    </div>
</template>
<script>
    import { defineComponent } from '@vue/composition-api'

    export default defineComponent({
        name: 'ContentHeader',
        props: {
            title: {
                type: String,
                default: ''
            }
        },
        setup (props, ctx) {
            const { $router } = ctx.root
            const goBack = () => {
                $router.back()
            }

            return {
                goBack
            }
        }
    })
</script>
<style lang="postcss" scoped>
.content-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    height: 60px;
    border-bottom: 1px solid #dde4eb;
    background: #fff;
    box-shadow: 4px 0 3px rgb(0 0 0 / 10%);
    padding-left: 20px;
    .title {
        font-size: 16px;
        i {
            cursor: pointer;
            font-weight: 700;
            color: #3a84ff;
        }
    }
}
</style>
