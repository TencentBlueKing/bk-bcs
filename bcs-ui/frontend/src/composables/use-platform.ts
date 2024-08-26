
import { getPlatformConfig, setDocumentTitle, setShortcutIcon } from '@blueking/platform-config';

import $i18n from '@/i18n/i18n-setup';
import { config } from '@/store/modules/platform-config';
export default function usePlatform() {
  const getPlatformInfo = async () => {
    const version = localStorage.getItem('__bcs_latest_version__');
    const defaultFooterInfoHTML = `<a href="https://wpa1.qq.com/KziXGWJs?_type=wpa&qidian=true" target="_blank">{{ $t('blueking.support') }}</a> |
    <a href="https://bk.tencent.com/s-mart/community/" target="_blank">{{ $t('blueking.community') }}</a> |
    <a href="https://bk.tencent.com/index/" target="_blank">{{ $t('blueking.website') }}</a>`;
    const defaults = {
      name: $i18n.t('bcs.TKEx.title'),
      nameEn: 'BCS',
      // eslint-disable-next-line @typescript-eslint/no-require-imports
      appLogo: require('@/images/bcs.svg'),
      brandName: '蓝鲸智云',
      brandNameEn: 'Tencent BlueKing',
      productName: '容器管理平台',
      productNameEn: 'BK Container Service',
      favicon: '/static/images/favicon.icon',
      helperLink: 'wxwork://message?uin=8444252571319680',
      footerCopyright: `Copyright © 2012 Tencent BlueKing. All Rights Reserved. ${version}`,
      helperText: $i18n.t('blueking.onCall'),
      footerInfoHTML: defaultFooterInfoHTML,
      version,
      i18n: {
        footerInfoHTML: defaultFooterInfoHTML,
      },
    };
    let data = {};
    if (window.BK_SHARED_RES_BASE_JS_URL) {
      data = await getPlatformConfig(window.BK_SHARED_RES_BASE_JS_URL, defaults);
    } else {
      data = await getPlatformConfig(defaults);
    }
    Object.keys(config).forEach((key) => {
      config[key] = data[key];
    });
    return data;
  };
  return {
    config,
    getPlatformInfo,
    setDocumentTitle,
    setShortcutIcon,
  };
}
