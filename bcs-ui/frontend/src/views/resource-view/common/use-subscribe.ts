/* eslint-disable no-unused-expressions */
/* eslint-disable camelcase */
import { computed, onBeforeUnmount, reactive, Ref, toRef } from 'vue';

import { crPrefix } from '@/api/base';
import BCSWebSocket from '@/components/bcs-log/common/websocket';
import $router from '@/router';
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
  total: number;
  webAnnotations?: any;
}

export interface IUseSubscribeResult {
  handleSubscribe: (params: Record<string, any>) => Promise<void>;
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
export default function useSubscribe(data: Ref<ISubscribeData>): IUseSubscribeResult {
  const $route = computed(() => toRef(reactive($router), 'currentRoute').value);
  let bcsWebSocket: BCSWebSocket | null = null;

  // 添加事件
  const handleAddSubscribe = (event: IEvent) => {
    const { manifest, uid, manifestExt } = event;
    data.value.manifest.items?.unshift(manifest);
    data.value.manifestExt[uid] = manifestExt;
  };
  // 删除事件
  const handleDeleteSubscribe = (event: IEvent) => {
    const { uid } = event;
    const index = data.value.manifest.items?.findIndex(item => item.metadata.uid === uid);

    if (index !== undefined && index !== -1) {
      data.value.manifest.items?.splice(index, 1);
    }

    if (data.value.manifestExt[uid]) {
      delete data.value.manifestExt[uid];
    }
  };
  // 修改事件
  const handleModifySubscribe = (event: IEvent) => {
    const { manifest, uid, manifestExt } = event;
    const index = data.value.manifest.items?.findIndex(item => item.metadata.uid === uid);

    if (index !== undefined && index !== -1) {
      data.value.manifest.items?.splice(index, 1, manifest);
      data.value.manifestExt[uid] = manifestExt;
    }
  };

  const projectId = computed(() => $route.value.params.projectId);
  const clusterId = computed(() => $route.value.params.clusterId);
  const subscribeURL = computed(() => {
    const host = window.BCS_API_HOST.replace(/(^\w+:|^)\/\//, '');
    const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws';
    return `${protocol}://${host}${crPrefix}/projects/${projectId.value}/clusters/${clusterId.value}/subscribe`;
  });

  // 订阅事件处理
  const handleSubscribe = async (params: any) => {
    const url = `${subscribeURL.value}?${new URLSearchParams(params as Record<string, any>)}`;
    bcsWebSocket?.ws?.close();
    bcsWebSocket = null;
    bcsWebSocket = new BCSWebSocket(url);

    bcsWebSocket.ws.onmessage = (event: MessageEvent) => {
      try {
        const { result } = JSON.parse(event.data);

        if (!result) return;
        console.log(result);
        // 处理具体订阅事件
        switch (result.type) {
          case 'DELETED':
            handleDeleteSubscribe(result);
            break;
          case 'ADDED':
            handleAddSubscribe(result);
            break;
          case 'MODIFIED':
            handleModifySubscribe(result);
            break;
        }
      } catch (err) {
        console.log(err);
      }
    };
  };

  onBeforeUnmount(() => {
    bcsWebSocket?.ws?.close();
    bcsWebSocket = null;
  });

  return {
    handleSubscribe,
    handleAddSubscribe,
    handleDeleteSubscribe,
    handleModifySubscribe,
  };
}
