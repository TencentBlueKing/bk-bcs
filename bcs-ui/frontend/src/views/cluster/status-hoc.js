/**
 * Tencent is pleased to support the open source community by making 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community Edition) available.
 * Copyright (C) 2017-2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

export default function statusHoc (WrappedComponent) {
    return {
        data () {
            return {
            }
        },
        watch: {
            // loading (v) {
            //     if (!v) {
            //         // this.curCluster = Object.assign({}, this.overviewList[this.curClusterIndex])
            //     }
            // }
        },
        created () {
            // if (this.overviewList[this.curClusterIndex]) {
            //     this.curCluster = Object.assign({}, this.overviewList[this.curClusterIndex])
            // }
        },
        props: WrappedComponent.props,
        methods: {
            /**
             * 转换百分比
             *
             * @param {number} used 使用量
             * @param {number} total 总量
             *
             * @return {number} 百分比数字
             */
            conversionPercent (used, total) {
                if (!total || parseFloat(total) === 0) {
                    return 0
                }
                let ret = parseFloat(used) / parseFloat(total) * 100
                if (ret !== 0 && ret !== 100) {
                    ret = ret.toFixed(2)
                }
                return ret
            }
        },
        render (h) {
            const slots = Object.keys(this.$slots)
                .reduce((arr, key) => arr.concat(this.$slots[key]), [])
                .map(vnode => {
                    vnode.context = this._self
                    return vnode
                })

            return h(WrappedComponent, {
                on: this.$listeners,
                props: this.$props,
                scopedSlots: this.$scopedSlots,
                attrs: this.$attrs
            }, slots)
        }
    }
}
