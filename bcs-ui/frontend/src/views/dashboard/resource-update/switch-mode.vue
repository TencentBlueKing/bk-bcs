<template>
    <div class="switch-button" @click="handleClick">
        <i class="bcs-icon bcs-icon-qiehuan mr5"></i>
        {{title}}
    </div>
</template>
<script>
    export default {
        props: {
            title: {
                type: String,
                default: ''
            }
        },
        methods: {
            handleClick () {
                this.$emit('click')
            }
        }
    }
</script>
<style lang="postcss" scoped>
.switch-button {
    display: flex;
    align-items: center;
    min-width: 148px;
    padding: 0 16px;
    height: 40px;
    background: #fff;
    box-shadow: 0 2px 8px 0 rgba(0,0,0,0.16);
    border-radius: 20px;
    z-index: 1;
    font-size: 12px;
    color: #63656E;
    cursor: pointer;

    &:hover {
        box-shadow: 0 2px 12px 0 rgba(0,0,0,0.20);
        color: #3A84FF;
    }
}
</style>
