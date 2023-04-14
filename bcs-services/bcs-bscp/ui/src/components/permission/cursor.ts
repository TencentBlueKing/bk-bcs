// @ts-nocheck
import lockIcon from '../../assets/'

const init = (el, binding) => {
  el.mouseEnterHandler = function () {
      const element = document.createElement('div')
      element.id = 'directive-ele'
      element.style.position = 'absolute'
      element.style.zIndex = '999999'
      el.element = element
      document.body.appendChild(element)

      element.classList.add(binding.value.cls || DEFAULT_OPTIONS.cls)
      el.addEventListener('mousemove', el.mouseMoveHandler)
  }
  el.mouseMoveHandler = function (event: Event) {
      const { pageX, pageY } = event
      const elLeft = pageX + DEFAULT_OPTIONS.offset[0]
      const elTop = pageY + DEFAULT_OPTIONS.offset[1]
      el.element.style.left = elLeft + 'px'
      el.element.style.top = elTop + 'px'
  }
  el.mouseLeaveHandler = function (event) {
      el.element && el.element.remove()
      el.element = null
      el.removeEventListener('mousemove', el.mouseMoveHandler)
  }
  if (binding.value.active) {
      el.addEventListener('mouseenter', el.mouseEnterHandler)
      el.addEventListener('mouseleave', el.mouseLeaveHandler)
  }
}

const DEFAULT_OPTIONS = {
  active: true,
  offset: [12, 0],
  cls: 'cursor-element'
}

const destroy = (el) => {
  el.element && el.element.remove()
  el.element = null
  el.removeEventListener('mouseenter', el.mouseEnterHandler)
  el.removeEventListener('mousemove', el.mouseMoveHandler)
  el.removeEventListener('mouseleave', el.mouseLeaveHandler)
}

const cursor = {
  mounted(el, binding) {
    binding.value = Object.assign({}, DEFAULT_OPTIONS, binding.value)
    init(el, binding)
  },
  updated(el, binding) {
    binding.value = Object.assign({}, DEFAULT_OPTIONS, binding.value)
    destroy(el)
    init(el, binding)
  },
  // 绑定元素的父组件卸载前调用
  beforeUnmount(el) {
    destroy(el)
  }
}

export default cursor
