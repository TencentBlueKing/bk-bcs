export default {
    name: 'bk-table',
    props: {
        pagination: Object,
        pageParams: Object
    },
    render (h) {
        const slots = Object.keys(this.$slots)
            .reduce((arr, key) => arr.concat(this.$slots[key]), [])
            .map(vnode => {
                vnode.context = this._self
                return vnode
            })

        // 兼容旧分页参数
        if (!this.$props.pagination && this.$props.pageParams) {
            this.$props.pagination = Object.assign({}, this.$props.pageParams, {
                count: this.$props.pageParams.allCount || this.$props.pageParams.total,
                current: this.$props.pageParams.curPage,
                limit: this.$props.pageParams.pageSize
            })
        }
        return h('bcs-table', {
            on: this.$listeners, // 透传事件
            props: Object.assign({}, this.$props, this.$attrs), // 透传props
            scopedSlots: this.$scopedSlots, // 透传scopedSlots
            attrs: this.$attrs // 透传属性，非props
        }, slots)
    }
}
