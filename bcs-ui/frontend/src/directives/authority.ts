/* eslint-disable @typescript-eslint/camelcase */
/* eslint-disable camelcase */
/* eslint-disable no-unused-expressions */
import { DirectiveBinding } from 'vue/types/options'
import { VueConstructor, VNode } from 'vue'
import { bus } from '@/common/bus'
import bkTooltips from 'bk-magic-vue/lib/directives/tooltips'
import { userPerms, userPermsByAction } from '@/api/base'
interface IElement extends HTMLElement {
    [prop: string]: any;
}
interface IOptions {
    clickable: boolean;
    offset: number[];
    cls: string;
    autoUpdatePerms?: boolean; // 是否在指令更新的时候重新发送权限请求（disablePerms: false时生效）
    disablePerms?: boolean; // 是否禁用权限请求（完全交个外部控制clickable的值决定状态）
    resourceName?: string;
    actionId?: string | string[];
    permCtx?: {
        project_id: string; // 项目权限 如果实例无关，可不传
        cluster_id: string; // 集群权限 如果实例无关，可不传cluster_id
        name: string; // 命名空间相关权限 如果实例无关，可不传name
        template_id: string; // 模板集相关权限  果实例无关，可不传template_id
        resource_type?: string; // 资源类型
    };
}

const DEFAULT_OPTIONS: IOptions = {
    clickable: false,
    offset: [12, 0],
    cls: 'bcs-cursor-element',
    disablePerms: false
}

function init (el: IElement, binding: DirectiveBinding) {
    const parent = el.parentNode
    const options: IOptions = Object.assign({}, DEFAULT_OPTIONS, binding.value)
    if (options.clickable || el.dataset.clickable || !parent) return

    if (!el.cloneEl) {
        el.cloneEl = el.cloneNode(true)
    }
    // 替换当前节点（为了移除节点的所有事件）
    parent?.replaceChild(el.cloneEl, el)

    const cloneEl = el.cloneEl
    cloneEl.style.filter = 'grayscale(100%)'
    bkTooltips.update(cloneEl, binding)
    cloneEl.mouseEnterHandler = function () {
        const element = document.createElement('div')
        element.id = 'directive-ele'
        element.style.position = 'absolute'
        element.style.zIndex = '9999'
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
    cloneEl.clickHandler = async (e: Event) => {
        e.stopPropagation()
        const { actionId, permCtx, resourceName } = options
        if (!actionId || actionId.length === 0) return

        delete permCtx?.resource_type
        const data = await userPermsByAction({
            $actionId: Array.isArray(actionId) ? actionId[0] : actionId,
            perm_ctx: permCtx
        }).catch(() => ({}))
        bus.$emit('show-apply-perm-modal', {
            perms: {
                apply_url: data?.perms?.apply_url,
                action_list: [
                    {
                        action_id: actionId,
                        resource_name: resourceName
                    }
                ]
            }
        })
    }

    cloneEl.addEventListener('mouseenter', cloneEl.mouseEnterHandler)
    cloneEl.addEventListener('mouseleave', cloneEl.mouseLeaveHandler)
    cloneEl.addEventListener('click', cloneEl.clickHandler)
}

function destroy (el: IElement) {
    const cloneEl = el.cloneEl
    if (!cloneEl) return

    // 还原原始节点
    const parent = cloneEl.parentNode
    parent?.replaceChild(el, el.cloneEl)

    bkTooltips.unbind(cloneEl)
    cloneEl.removeEventListener('mouseenter', cloneEl.mouseEnterHandler)
    cloneEl.removeEventListener('mousemove', cloneEl.mouseMoveHandler)
    cloneEl.removeEventListener('mouseleave', cloneEl.mouseLeaveHandler)
    cloneEl.removeEventListener('click', cloneEl.clickHandler)
    cloneEl.element?.remove()
    cloneEl.element = null
    delete el.cloneEl
}

async function updatePerms (el: IElement, binding: DirectiveBinding) {
    const { actionId = '', permCtx } = binding.value as IOptions
    const { resource_type, cluster_id, project_id } = permCtx || {}
    // 校验数据完整性
    if (!actionId
        || !resource_type
        || (resource_type === 'cluster' && !cluster_id)
        || (resource_type === 'project' && !project_id)) return

    const actionIds = Array.isArray(actionId) ? actionId : [actionId]
    const data = await userPerms({
        action_ids: actionIds,
        perm_ctx: permCtx
    }).catch(() => ({}))
    const clickable = actionIds.every(actionId => data?.perms?.[actionId])
    el.dataset.clickable = clickable ? 'true' : ''

    const cloneBinding = JSON.parse(JSON.stringify(binding))
    cloneBinding.value.clickable = clickable
    destroy(el)
    init(el, cloneBinding)
}

export default class AuthorityDirective {
    public static install (Vue: VueConstructor) {
        Vue.directive('authority', {
            bind (el: IElement, binding: DirectiveBinding, vNode: VNode) {
                el.cloneEl = el.cloneNode(true) as IElement
                const { actionId, disablePerms, clickable } = binding.value as IOptions
                if (actionId && !disablePerms && !clickable) {
                    updatePerms(el, binding)
                }
            },
            inserted (el: IElement, binding: DirectiveBinding) {
                init(el, binding)
            },
            update (el: IElement, binding: DirectiveBinding) {
                setTimeout(() => {
                    const { autoUpdatePerms, disablePerms } = binding.value as IOptions
                    if (autoUpdatePerms && !disablePerms) {
                        updatePerms(el, binding)
                    } else {
                        destroy(el)
                        init(el, binding)
                    }
                }, 0)
            },
            unbind (el: IElement) {
                destroy(el)
            }
        })
    }
}
