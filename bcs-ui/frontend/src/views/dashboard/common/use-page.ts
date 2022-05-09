import { reactive, computed, Ref, ComputedRef } from '@vue/composition-api'

export interface IPageConf {
    current: number;
    limit: number;
}

export interface IPagination extends IPageConf {
    count: number;
}

export interface IOptions extends IPageConf {
    onPageChange?: (page: number) => any;
    onPageSizeChange?: (pageSize: number) => any;
}

export interface IPageConfResult {
    curPageData: ComputedRef<any[]>;
    pageChange: (page: number) => void;
    pageSizeChange: (size: number) => void;
    pageConf: IPageConf;
    pagination: ComputedRef<IPagination>;
    handleResetPage: Function;
}

/**
 * 前端分页
 * @param data 全量数据
 * @param options 配置数据
 */
export default function usePageConf (data: Ref<any[]>, options: IOptions = {
    current: 1,
    limit: 10
}): IPageConfResult {
    const pageConf = reactive<IPageConf>({
        current: options.current,
        limit: options.limit
    })

    const curPageData = computed(() => {
        const { limit, current } = pageConf
        return data.value.slice(limit * (current - 1), limit * current)
    })

    const pageChange = (page = 1) => {
        pageConf.current = page
        const { onPageChange = null } = options
        onPageChange && typeof onPageChange === 'function' && onPageChange(page)
    }

    const pageSizeChange = (pageSize = 10) => {
        pageConf.limit = pageSize
        pageConf.current = 1
        const { onPageSizeChange = null } = options
        onPageSizeChange && typeof onPageSizeChange === 'function' && onPageSizeChange(pageSize)
    }

    const pagination = computed<IPagination>(() => {
        return {
            ...pageConf,
            count: data.value.length
        }
    })

    const handleResetPage = () => {
        pageConf.current = 1
    }

    return {
        pageConf,
        pagination,
        curPageData,
        pageChange,
        pageSizeChange,
        handleResetPage
    }
}
