export default {
    name: 'bk-paging',
    props: {
        curPage: {
            type: Number,
            default: 1
        },
        totalPage: {
            type: Number,
            default: 1
        },
        count: {
            type: Number,
            default: 1
        },
        showLimit: {
            type: Boolean,
            default: false
        },
        limit: {
            type: Boolean,
            default: 10
        }
    },
    render (h) {
        const slots = Object.keys(this.$slots)
            .reduce((arr, key) => arr.concat(this.$slots[key]), [])
            .map(vnode => {
                vnode.context = this._self
                return vnode
            })
        if (this.$props.curPage) {
            this.$props.current = this.$props.curPage
        }

        if (this.$listeners && this.$listeners['page-change']) {
            this.$listeners['change'] = this.$listeners['page-change']
        }

        return h('bcs-pagination', {
            on: this.$listeners, // 透传事件
            props: Object.assign({}, this.$props, this.$attrs), // 透传props
            scopedSlots: this.$scopedSlots, // 透传scopedSlots
            attrs: this.$attrs // 透传属性，非props
        }, slots)
    }
}
