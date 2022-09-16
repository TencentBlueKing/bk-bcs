export default {
    name: 'bk-dialog',
    props: {
        isShow: {
            type: Boolean,
            default: false
        },
        value: {
            type: Boolean,
            default: false
        },
        hasFooter: {
            type: Boolean,
            default: true
        },
        confirm: {
            type: String,
            default: window.i18n.t('确定')
        },
        cancel: {
            type: String,
            default: window.i18n.t('取消')
        },
        maskClose: {
            type: Boolean,
            default: true
        },
        quickClose: {
            type: Boolean,
            default: true
        },
        headerPosition: {
            type: String,
            default: 'left'
        },
        closeIcon: {
            type: Boolean,
            default: true
        },
        title: {
            type: String
        }
    },
    
    render (h) {
        const slots = Object.keys(this.$slots)
            .reduce((arr, key) => arr.concat(this.$slots[key]), [])
            .map(vnode => {
                vnode.context = this._self
                return vnode
            })
        this.$props.value = this.$props.isShow
        this.$props.showFooter = this.$props.hasFooter
        if (this.$props.confirm !== this.$t('确定')) {
            this.$props.okText = this.$props.confirm
        }
        if (this.$props.cancel !== this.$t('取消')) {
            this.$props.cancelText = this.$props.cancel
        }
        this.$props.maskClose = this.$props.quickClose

        return h('bcs-dialog', {
            on: this.$listeners, // 透传事件
            props: Object.assign({}, this.$props, this.$attrs), // 透传props
            scopedSlots: this.$scopedSlots, // 透传scopedSlots
            attrs: this.$attrs // 透传属性，非props
        }, slots)
    }
}
