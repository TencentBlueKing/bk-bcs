interface Window {
  bus: any
  mainComponent: any
  BCS_API_HOST: string
  DEVOPS_BCS_API_URL: string
  i18n: {
    t: (word: string) => string
  }
  BCS_CONFIG: Record<string, string>
  REGION: string
  PAAS_HOST: string
  BK_IAM_HOST: string
  DEVOPS_HOST: string
  LOGIN_FULL: string
  BKMONITOR_HOST: string
  RUN_ENV: string
  PREFERRED_DOMAINS: string
  $loginModal: any
  _project_code_: string
  _project_id_: string
  readonly BK_STATIC_URL: string
}

declare const BK_BCS_WELCOME: string;

declare const BK_BCS_VERSION: string;

declare const SITE_URL: string;

declare const DEVOPS_BCS_API_URL: string;
