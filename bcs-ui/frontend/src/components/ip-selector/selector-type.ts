import { TranslateResult } from 'vue-i18n'
import Vue, { VNode } from 'vue'

export type IpType = 'TOPO' | 'INSTANCE' | 'SERVICE_TEMPLATE' | 'SET_TEMPLATE';

// tab数据源
export interface IPanel {
    name: string; // tab唯一标识（要和components目录的文件名保持一致，作为后面动态组件的name）
    label: string | TranslateResult; // tab默认显示文本
    hidden?: boolean; // tab是否显示
    disabled?: boolean; // tab是否禁用
    keepAlive?: boolean; // 是否缓存
    tips?: string | TranslateResult;
    component?: Vue; // 组件对象（默认根据name从layout里面获取）
    type?: IpType; // 当前tab对于的类型（目前对于后端来说只有两种，只是前端选择方式不一样）
}

export interface IEventsMap {
    [key: string]: Function;
}

export interface IMenu {
    id: string | number;
    label: string | TranslateResult;
    readonly?: boolean;
    disabled?: boolean;
    hidden?: boolean;
}

export interface INodeData extends IMenu {
    data: any; // 原始数据
    children?: INodeData[];
}

export interface IPreviewData {
    id: IpType;
    name: string | TranslateResult;
    data: any[];
    dataNameKey?: string;
}

export interface ISearchData extends INodeData {
    path?: string;
}

export type IPerateFunc = (item: IPreviewData) => IMenu[];

export interface ILayoutComponents {
    [key: string]: Vue;
}

export interface ISearchDataOption {
    idKey: string;
    nameKey: string;
    pathKey: string;
}

export interface IPreviewDataOption {
    nameKey?: string;
}

export interface ITreeNode {
    id: string | number;
    name: string;
    level: string | number;
    children: ITreeNode[];
    data?: any;
    parent?: ITreeNode;
}

export interface ITableConfig {
    prop: string;
    label: string | TranslateResult;
    render?: (row: any, column: any, $index: number) => VNode;
    hidden?: boolean;
    minWidth?: number;
}

export interface IAgentStatusData {
    count?: number;
    status: string;
    display: string | TranslateResult;
    errorCount?: number;
}
// 0 未选 1 半选 2 全选
export type CheckValue = 0 | 1 | 2;

export type CheckType = 'current' | 'all';

export interface IPagination {
    limit: number;
    count: number;
    current: number;
    small?: boolean; // 小型分页器
    showTotalCount?: boolean;
    showLimit?: boolean;
    align?: 'left' | 'center' | 'right';
    limitList?: number[];
}

export interface ITemplateDataOptions {
    idKey?: string;
    labelKey?: string;
    childrenKey?: string;
}

/**
 *layout组件搜索函数的类型
 * @param params 接口参数
 */
// eslint-disable-next-line max-len
export type SearchDataFuncType = (params: any, type?: string) => Promise<{ total: number; data: any[] }>;

export interface IipListParams {
    current: number;
    limit: number;
    tableKeyword: string;
}

export interface ITableCheckData {
    excludeData?: any[];
    selections: any[];
    checkType?: CheckType;
    checkValue?: CheckValue;
}

export type IActive = 'inner' | 'outer' | 'other';

export interface IClassifyTab {
    active: string;
    list: { id: IActive; name: TranslateResult }[];
}
