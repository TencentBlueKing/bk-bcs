<template>
    <div class="bk-file-upload">
        <div class="file-list" v-if="uploadQueue.length">
            <div class="file-item" v-for="(file, index) of uploadQueue" :key="index">
                <div :class="['file-item-wrapper', { 'error': file.status === 'error', 'success': file.status === 'success' }]">
                    <div class="file-icon">
                        <img src="@/images/placeholder.svg" alt="">
                    </div>
                    <div class="file-info">
                        <div class="file-metedata">
                            <p class="file-name">{{file.name}}</p>
                            <span class="file-status" v-if="file.status === 'uploading'">{{$t('上传中')}}...</span>
                            <span class="file-status success" v-if="file.status === 'success'">{{$t('上传成功')}}</span>
                            <span class="file-status error" v-if="file.status === 'error'">{{$t('上传失败')}}</span>
                        </div>
                        <div :class="['file-progress']">
                            <div :class="['file-progress-bar', { 'running': file.status === 'uploading' }]" style=""></div>
                        </div>
                    </div>
                    <i class="bcs-icon bcs-icon-close" @click="removeFile(index)" v-if="file.status !== 'success'"></i>
                </div>
                <p class="tip" v-if="file.statusText && file.status === 'error'">{{file.statusText}}</p>
            </div>
        </div>
        <div class="file-input" v-else>
            <bk-button class="trigger-btn">
                <img src="@/images/upload.svg" alt="" class="upload-icon">
                点击上传
            </bk-button>
            <p class="tip" v-if="tip">{{tip}}</p>
            <input type="file" @change="selectFile" :multiple="multiple" :accept="accept">
        </div>
    </div>
</template>
<script>
    import cookie from 'cookie'

    const CSRFToken = cookie.parse(document.cookie).bcs_csrftoken

    export default {
        props: {
            // 必选参数，上传的地址
            postUrl: {
                type: String,
                default: ''
            },
            name: {
                type: String,
                default: 'upload-file'
            },
            // 设置上传的请求头部
            headers: {
                type: Object,
                default () {
                    return {}
                }
            },
            tip: {
                type: String,
                default: ''
            },
            multiple: {
                type: Boolean,
                default: false
            },
            accept: {
                type: String
            },
            maxSize: {
                type: Number
            },
            dragable: {
                type: Boolean,
                default: false
            },
            disabled: {
                type: Boolean,
                default: false
            }
        },
        data () {
            return {
                uploadQueue: [],
                isUploadLoading: false
            }
        },
        methods: {
            selectFile (event) {
                const target = event.target
                const files = target.files

                if (!files.length) {
                    return
                }

                for (const file of files) {
                    const fileObj = {
                        name: file.name,
                        size: file.size / 1000 / 1000,
                        type: file.type,
                        origin: file,
                        isUploadLoading: false,
                        status: '',
                        statusText: ''
                    }
                    this.uploadQueue.push(fileObj)
                    if (this.maxSize && (fileObj.size > this.maxSize)) {
                        fileObj.status = 'error'
                        fileObj.statusText = `${this.$t('文件不能超过')}${this.maxSize}M`
                    } else {
                        this.uploadFile(fileObj)
                    }
                }
            },
            uploadFile (fileObj) {
                this.isUploadLoading = true

                const formData = new FormData()
                formData.append('image', fileObj.origin)
                fileObj.status = 'uploading'
                fileObj.statusText = this.$t('上传中')

                const xhr = new XMLHttpRequest()
                fileObj.xhr = xhr // 保存，用于中断请求

                xhr.withCredentials = true
                xhr.open('POST', this.postUrl, true)
                xhr.onreadystatechange = () => {
                    if (xhr.readyState === 4) {
                        if (xhr.status === 200) {
                            const response = JSON.parse(xhr.responseText)

                            if (response.code === 0) {
                                this.isUploadLoading = false
                                fileObj.status = 'success'
                                fileObj.statusText = this.$t('上传成功')
                                setTimeout(() => {
                                    this.$emit('uploadSuccess', response.data)
                                    this.clearUploadQueue()
                                }, 1000)
                            } else {
                                this.isUploadLoading = false
                                fileObj.status = 'error'
                                fileObj.statusText = response.message
                                this.$emit('uploadFail', response)
                            }
                        }
                    }
                }
                xhr.setRequestHeader('X-CSRFToken', CSRFToken)
                xhr.send(formData)
            },
            removeFile (index) {
                if (this.uploadQueue[index].xhr) {
                    this.uploadQueue[index].xhr.abort()
                }

                this.uploadQueue.splice(index, 1)
            },
            clearUploadQueue () {
                this.uploadQueue.forEach(item => {
                    item.xhr.abort()
                })
                this.uploadQueue = []
            }
        }
    }
</script>
<style>
    @import './index.css';
</style>
