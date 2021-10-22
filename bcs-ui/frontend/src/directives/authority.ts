/* eslint-disable no-unused-expressions */
import { DirectiveBinding } from 'vue/types/options'
import { VueConstructor } from 'vue'
// import { bus } from '@/common/bus'
import bkTooltips from 'bk-magic-vue/lib/directives/tooltips'

interface IElement extends HTMLElement {
    [prop: string]: any;
}
interface IOptions {
    clickable: boolean;
    offset: number[];
    cls: string;
    data?: any;
}

const DEFAULT_OPTIONS: IOptions = {
    clickable: true,
    offset: [12, 0],
    cls: 'bcs-cursor-element'
}

function init (el: IElement, binding: DirectiveBinding) {
    const parent = el.parentNode
    const options: IOptions = Object.assign({}, DEFAULT_OPTIONS, binding.value)
    if (options.clickable || !parent) return

    // 替换当前节点（为了移除节点的所有事件）
    parent?.replaceChild(el.cloneEl, el)

    const cloneEl = el.cloneEl
    cloneEl.style.filter = 'grayscale(100%)'
    bkTooltips.update(cloneEl, binding)
    cloneEl.mouseEnterHandler = function () {
        const element = document.createElement('div')
        element.id = 'directive-ele'
        element.style.position = 'absolute'
        element.style.zIndex = '2501'
        cloneEl.element = element
        document.body.appendChild(element)

        element.classList.add(options.cls || DEFAULT_OPTIONS.cls)
        cloneEl.addEventListener('mousemove', cloneEl.mouseMoveHandler)
    }
    cloneEl.mouseMoveHandler = function (event: MouseEvent) {
        const { pageX, pageY } = event
        const elLeft = pageX + DEFAULT_OPTIONS.offset[0]
        const elTop = pageY + DEFAULT_OPTIONS.offset[1]
        cloneEl.element.style.left = `${elLeft}px`
        cloneEl.element.style.top = `${elTop}px`
    }
    cloneEl.mouseLeaveHandler = function () {
        cloneEl.element && cloneEl.element.remove()
        cloneEl.element = null
        cloneEl.removeEventListener('mousemove', cloneEl.mouseMoveHandler)
    }
    // cloneEl.clickHandler = function (e) {
    //     bus.$emit('show-apply-perm-modal', options.data || {})
    //     e.stopImmediatePropagation()
    // }

    cloneEl.addEventListener('mouseenter', cloneEl.mouseEnterHandler)
    cloneEl.addEventListener('mouseleave', cloneEl.mouseLeaveHandler)
    // cloneEl.addEventListener('click', cloneEl.clickHandler)
}

function destroy (el: IElement) {
    const cloneEl = el.cloneEl
    // 还原原始节点
    const parent = cloneEl.parentNode
    parent?.replaceChild(el, el.cloneEl)

    bkTooltips.unbind(cloneEl)
    cloneEl.removeEventListener('mouseenter', cloneEl.mouseEnterHandler)
    cloneEl.removeEventListener('mousemove', cloneEl.mouseMoveHandler)
    cloneEl.removeEventListener('mouseleave', cloneEl.mouseLeaveHandler)
    cloneEl.element?.remove()
    cloneEl.element = null
    // cloneEl.removeEventListener('click', cloneEl.clickHandler)
}

export default class AuthorityDirective {
    public static install (Vue: VueConstructor) {
        Vue.directive('authority', {
            bind (el: IElement, binding: DirectiveBinding) {
                el.cloneEl = el.cloneNode(true) as IElement
            },
            inserted (el: IElement, binding: DirectiveBinding) {
                init(el, binding)
            },
            update (el: IElement, binding: DirectiveBinding) {
                destroy(el)
                init(el, binding)
            },
            unbind (el: IElement) {
                destroy(el)
            }
        })
    }
}
