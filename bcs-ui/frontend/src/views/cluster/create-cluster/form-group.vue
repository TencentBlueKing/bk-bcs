<template>
    <section class="form-group">
        <div class="form-group-title" @click="toggleActive">
            <span class="left">
                <span class="icon" :style="!active ? 'transform: rotate(-90deg);' : 'transform: rotate(0deg);'">
                    <i class="bcs-icon bcs-icon-down-shape"></i>
                </span>
                <span class="label">{{ title }}</span>
                <span class="desc">{{ desc }}</span>
            </span>
            <slot name="title"></slot>
        </div>
        <div class="form-group-content" v-show="active">
            <slot></slot>
        </div>
    </section>
</template>
<script lang="ts">
    import { defineComponent, ref } from '@vue/composition-api'

    export default defineComponent({
        name: 'FormGroup',
        props: {
            title: {
                type: String,
                default: ''
            },
            desc: {
                type: String,
                default: ''
            },
            defaultActive: {
                type: Boolean,
                default: true
            }
        },
        setup (props, ctx) {
            const { emit } = ctx
            const active = ref(props.defaultActive)
            const toggleActive = () => {
                active.value = !active.value
                emit('toggle', active.value)
            }
            return {
                active,
                toggleActive
            }
        }
    })
</script>
<style lang="postcss" scoped>
.form-group {
    background: #fff;
    border-radius: 2px;
    font-size: 12px;
    padding: 20px 16px;
    &-title {
        display: flex;
        align-items: center;
        justify-content: space-between;
        cursor: pointer;
        .left {
            display: flex;
            align-items: center;
        }
        .icon {
            display: flex;
            align-items: center;
            justify-content: center;
            width: 24px;
            height: 24px;
            transition: all 0.2s ease;
        }
        .label {
            font-size: 14px;
            font-weight: 700;
            line-height: 22px;
        }
        .desc {
            margin-left: 16px;

        }
    }
    &-content {
        padding-top: 32px;
    }
}
</style>
