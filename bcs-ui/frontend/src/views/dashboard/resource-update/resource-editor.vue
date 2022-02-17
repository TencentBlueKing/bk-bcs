<!-- 资源视图yaml代码编辑器组件 -->
<template>
    <div class="resource-editor" :style="style" ref="editorRef"></div>
</template>
<script lang="ts">
    /* eslint-disable no-unused-expressions */
    import { computed, defineComponent, ref, toRefs, watch, onMounted, onBeforeMount } from '@vue/composition-api'
    import * as monaco from 'monaco-editor/esm/vs/editor/editor.main'
    import yamljs from 'js-yaml'
    import BcsEditorTheme from './theme.json'

    self.MonacoEnvironment = {
        getWorkerUrl (moduleId, label) {
            if (label === 'json') {
                return `${window.DEVOPS_BCS_HOST}${window.STATIC_URL}${window.VERSION_STATIC_URL}/json.worker.js`
            }
            if (label === 'css') {
                return `${window.DEVOPS_BCS_HOST}${window.STATIC_URL}${window.VERSION_STATIC_URL}/css.worker.js`
            }
            if (label === 'html') {
                return `${window.DEVOPS_BCS_HOST}${window.STATIC_URL}${window.VERSION_STATIC_URL}/html.worker.js`
            }
            if (label === 'typescript' || label === 'javascript') {
                return `${window.DEVOPS_BCS_HOST}${window.STATIC_URL}${window.VERSION_STATIC_URL}/ts.worker.js`
            }
            return `${window.DEVOPS_BCS_HOST}${window.STATIC_URL}${window.VERSION_STATIC_URL}/editor.worker.js`
        }
    }

    export default defineComponent({
        name: 'ResourceEditor',
        props: {
            value: { type: [String, Object], default: () => ({}) },
            diffEditor: { type: Boolean, default: false }, // 是否使用diff模式
            width: { type: [String, Number], default: '100%' },
            height: { type: [String, Number], default: '100%' },
            original: { type: [String, Object], default: () => ({}) }, // 只有在diff模式下有效
            language: { type: String, default: 'yaml' },
            theme: { type: String, default: 'bcs-theme' },
            readonly: { type: Boolean, default: false },
            options: { type: Object, default: () => ({}) },
            ignoreKeys: { type: [Array, String], default: () => '' }
        },
        setup (props, ctx) {
            const { value, diffEditor, width, height, original, language, theme, options, readonly, ignoreKeys } = toRefs(props)
            const editorRef = ref<any>(null)
            const editorErr = ref<any>('')
            // diff统计
            const diffStat = ref({
                insert: 0,
                delete: 0
            })
            let editor: any | null = null

            const handleIgnoreKeys = (data) => {
                const cloneData = JSON.parse(JSON.stringify(data))
                const keys: string[] = typeof ignoreKeys.value === 'string' ? [ignoreKeys.value] : ignoreKeys.value as string[]
                keys.forEach(key => {
                    const props = key.split('.')
                    props.reduce((pre, prop, index) => {
                        if (index === (props.length - 1) && pre) {
                            delete pre[prop]
                        }

                        if (pre && (prop in pre)) {
                            return pre[prop]
                        }

                        return ''
                    }, cloneData)
                })
                return cloneData
            }
            // 原始数据统一转换为字符串
            const yaml = computed<string>(() => {
                if (typeof value.value === 'object') {
                    return Object.keys(value.value).length ? yamljs.dump(handleIgnoreKeys(value.value)) : ''
                }
                return value.value
            })
            const diffYaml = computed<string>(() => {
                if (typeof original.value === 'object') {
                    return Object.keys(original.value).length ? yamljs.dump(handleIgnoreKeys(original.value)) : ''
                }
                return value.value
            })

            const style = computed(() => {
                return {
                    width: typeof width.value === 'number' ? `${width.value}px` : width.value,
                    height: typeof height.value === 'number' ? `${height.value}px` : height.value
                }
            })

            watch(options, (opt) => {
                editor?.updateOptions(opt)
            }, { deep: true })

            // 非只读模式时，建议手动调用setValue方法，watch在双向绑定时会让编辑器抖动
            watch(value, () => {
                if (readonly.value && yaml.value !== getValue()) {
                    setValue(yaml.value)
                }
            })

            watch(language, () => {
                if (!editor) return
                if (diffEditor.value) {
                    // diff模式下更新language
                    const { original, modified } = editor.getModel()
                    monaco.editor.setModelLanguage(original, language.value)
                    monaco.editor.setModelLanguage(modified, language.value)
                } else {
                    monaco.editor.setModelLanguage(editor.getModel(), language.value)
                }
            })

            watch(theme, () => {
                editor && monaco.editor.setTheme(theme.value)
            })

            watch([width, height], () => {
                editor?.layout()
            })

            watch(editorErr, (err) => {
                ctx.emit('error', err)
            })

            onMounted(() => {
                initMonaco()
            })

            onBeforeMount(() => {
                editor?.dispose()
                editor = null
            })

            const initMonaco = () => {
                if (!editorRef.value) return

                monaco.editor.defineTheme('bcs-theme', BcsEditorTheme)
                const opt = {
                    value: yaml.value,
                    language: language.value,
                    theme: theme.value,
                    minimap: {
                        enabled: false
                    },
                    readOnly: readonly.value,
                    automaticLayout: true,
                    ...options.value
                }
                if (diffEditor.value) {
                    editor = monaco.editor.createDiffEditor(editorRef.value, {
                        ...opt,
                        readOnly: true
                    })
                    setModel(yaml.value, diffYaml?.value)
                } else {
                    editor = monaco.editor.create(editorRef.value, opt)
                }
                editorMounted() // 编辑器初始化后
            }

            const editorMounted = () => {
                if (diffEditor.value) {
                    editor.onDidUpdateDiff((event) => {
                        const value = getValue()
                        handleSetDiffStat()
                        emitChange(value, event)
                    })
                } else {
                    editor.onDidChangeModelContent(event => {
                        const yamlValue = getValue()
                        try {
                            let emitValue = {}
                            const tmpObj = yamljs.load(yamlValue)
                            emitValue = typeof tmpObj === 'object' ? tmpObj : {}
                            editorErr.value = ''

                            emitChange(emitValue, event)
                        } catch (err: any) {
                            editorErr.value = err?.message || String(err)
                        }
                    })
                }
            }

            const handleSetDiffStat = () => {
                diffStat.value = {
                    insert: 0,
                    delete: 0
                }
                const changes = editor.getLineChanges() || []
                changes.forEach(item => {
                    if ((item.originalEndLineNumber >= item.originalStartLineNumber) && item.originalEndLineNumber > 0) {
                        diffStat.value.delete += item.originalEndLineNumber - item.originalStartLineNumber + 1
                    }

                    if ((item.modifiedEndLineNumber >= item.modifiedStartLineNumber) && item.modifiedEndLineNumber > 0) {
                        diffStat.value.insert += item.modifiedEndLineNumber - item.modifiedStartLineNumber + 1
                    }
                })
                ctx.emit('diff-stat', diffStat.value)
            }

            const emitChange = (emitValue, event) => {
                ctx.emit('change', emitValue, event)
                ctx.emit('input', emitValue, event)
            }

            const getEditor = () => {
                if (!editor) return null
                return diffEditor.value ? editor.modifiedEditor : editor
            }

            const getValue = (): string => {
                const editor = getEditor()
                if (!editor) return ''
                return editor.getValue()
            }

            const setValue = (value) => {
                try {
                    if (typeof value !== 'string') {
                        value = yamljs.dump(value)
                    }
                    const editor = getEditor()
                    if (editor) return editor.setValue(value)
                } catch (err: any) {
                    editorErr.value = err?.message || String(err)
                }
            }

            const setModel = (value, original) => {
                // diff模式下设置model
                const originalModel = monaco.editor.createModel(original, language.value)
                const modifiedModel = monaco.editor.createModel(value, language.value)
                editor.setModel({
                    original: originalModel,
                    modified: modifiedModel
                })
            }

            return {
                diffStat,
                editorRef,
                editorErr,
                yaml,
                diffYaml,
                style,
                getEditor,
                getValue,
                setValue,
                handleSetDiffStat
            }
        }
    })
</script>
<style scoped>
.resource-editor {
    width: 100%;
    height: 100%;
    border-radius: 2px;
}
</style>
