import client from './client'

export interface StorageConfig { driver: 'local' | 's3'; local: { root: string; url: string }; s3: { endpoint: string; bucket: string; access_key: string; secret_key?: string; region: string; prefix: string } }
export const storageApi = {
  get: () => client.get<StorageConfig>('/backend/v1/storage-config').then((response) => response.data),
  save: (data: StorageConfig) => client.put<StorageConfig>('/backend/v1/storage-config', data).then((response) => response.data),
  test: (data: StorageConfig) => client.post('/backend/v1/storage-config/test', data).then((response) => response.data),
}
