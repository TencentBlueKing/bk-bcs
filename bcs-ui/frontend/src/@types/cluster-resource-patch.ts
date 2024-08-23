export interface ITemplateSpaceData {
  'description': string
  'id': string
  'name': string
  'projectCode': string
}

export interface ISchemaData {
  layout: Array<Array<any>>
  rules: Record<string, any>
  schema: any
}

export interface IListTemplateMetadataItem {
  'createAt': number
  'creator': string
  'description': string
  'id': string
  'name': string
  'projectCode': string
  'resourceType': string
  'tags': string[]
  'templateSpace': string
  templateSpaceID: string
  'updateAt': number
  'updator': string
  'version': string
  versionID: string
  'versionMode': number
  draftContent: string
  draftVersion: string
  isDraft: boolean
  draftEditFormat: 'form' | 'yaml'
}

export interface ITemplateVersionItem {
  'content': string
  'createAt': number
  'creator': string
  'description': string
  'id': string
  'projectCode': string
  'templateName': string
  'templateSpace': string
  'version': string
  'draft': boolean
  'editFormat': 'form' | 'yaml'
}

export interface IVarItem {
  key: string
  value: string
  readonly: boolean
}

export interface IPreviewItem {
  content: string
  kind: string
  name: string
}
