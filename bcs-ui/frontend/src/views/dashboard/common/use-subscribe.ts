/* eslint-disable no-unused-expressions */
/* eslint-disable @typescript-eslint/camelcase */
/* eslint-disable camelcase */
import { SetupContext, Ref, onBeforeUnmount } from '@vue/composition-api'
import BCSWebSocket from '@/components/bcs-log/common/websocket'
export interface ISubscribeParams {
    kind: string;
    resourceVersion: string;
    apiVersion?: string;
    namespace?: string;
    CRDName?: string;
}

export interface IManifestExt {
    [uid: string]: any;
}

export interface IMetaData {
    resourceVersion: string;
}

export interface IManifest {
    metadata?: IMetaData;
    items?: any[];
}

export interface IEvent {
    type: 'ADDED' | 'DELETED' | 'MODIFIED';
    uid: string;
    manifestExt: IManifestExt;
    manifest: any[]; // 订阅事件的 manifest 跟 ISubscribeData 的 manifest 不一样
}

export interface ISubscribeData {
    manifest: IManifest;
    manifestExt: IManifestExt;
    webAnnotations?: any;
}

export interface IUseSubscribeResult {
    handleSubscribe: (url: string) => Promise<void>;
    handleAddSubscribe: (event: IEvent) => void;
    handleDeleteSubscribe: (event: IEvent) => void;
    handleModifySubscribe: (event: IEvent) => void;
}

/**
 * 订阅事件处理
 * @param params
 * @param ctx
 * @returns
 */
export default function useSubscribe (data: Ref<ISubscribeData>, ctx: SetupContext): IUseSubscribeResult {
    let bcsWebSocket: BCSWebSocket | null = null

    // 添加事件
    const handleAddSubscribe = (event: IEvent) => {
        const { manifest, uid, manifestExt } = event
        data.value.manifest.items?.unshift(manifest)
        data.value.manifestExt[uid] = manifestExt
    }
    // 删除事件
    const handleDeleteSubscribe = (event: IEvent) => {
        const { uid } = event
        const index = data.value.manifest.items?.findIndex(item => item.metadata.uid === uid)

        if (index !== undefined && index !== -1) {
            data.value.manifest.items?.splice(index, 1)
        }

        if (data.value.manifestExt[uid]) {
            delete data.value.manifestExt[uid]
        }
    }
    // 修改事件
    const handleModifySubscribe = (event: IEvent) => {
        const { manifest, uid, manifestExt } = event
        const index = data.value.manifest.items?.findIndex(item => item.metadata.uid === uid)

        if (index !== undefined && index !== -1) {
            data.value.manifest.items?.splice(index, 1, manifest)
            data.value.manifestExt[uid] = manifestExt
        }
    }

    // 订阅事件处理
    const handleSubscribe = async (url) => {
        bcsWebSocket?.ws?.close()
        bcsWebSocket = new BCSWebSocket(url)

        bcsWebSocket.ws.onmessage = (event: MessageEvent) => {
            try {
                const data = JSON.parse(event.data)
        
                if (!data || !data.length) {
                    return
                }
                // 处理具体订阅事件
                data.forEach((event: IEvent) => {
                    switch (event.type) {
                        case 'DELETED':
                            handleDeleteSubscribe(event)
                            break
                        case 'ADDED':
                            handleAddSubscribe(event)
                            break
                        case 'MODIFIED':
                            handleModifySubscribe(event)
                            break
                    }
                })
            } catch (err) {
                console.log(err)
            }
        }
    }

    onBeforeUnmount(() => {
        bcsWebSocket?.ws?.close()
    })

    return {
        handleSubscribe,
        handleAddSubscribe,
        handleDeleteSubscribe,
        handleModifySubscribe
    }
}
