import { Button, Card, Form, Input, Radio, Space, Typography, message } from 'antd'
import { useEffect, useState } from 'react'
import { storageApi, type StorageConfig } from '../api/storage'
import PageHeader from '../components/PageHeader'

const empty: StorageConfig = { driver: 'local', local: { root: '', url: '' }, s3: { endpoint: '', bucket: '', access_key: '', region: 'us-east-1', prefix: '' } }
export default function StorageSettings() {
  const [form] = Form.useForm<StorageConfig>(); const [config, setConfig] = useState<StorageConfig>(empty); const [loading, setLoading] = useState(false)
  useEffect(() => { void storageApi.get().then((value) => { setConfig({ ...empty, ...value, s3: { ...empty.s3, ...value.s3 } }); form.setFieldsValue({ ...empty, ...value, s3: { ...empty.s3, ...value.s3 } }) }) }, [form])
  const save = async () => { setLoading(true); try { const value = await storageApi.save(form.getFieldsValue(true)); setConfig(value); form.setFieldsValue(value); message.success('存储配置已保存') } finally { setLoading(false) } }
  const test = async () => { await storageApi.test(form.getFieldsValue(true)); message.success('连接测试成功') }
  return <><PageHeader title="存储配置" description="切换本地存储或 S3 兼容对象存储，密钥只写入不回显。" /><Card><Form form={form} layout="vertical" initialValues={config}><Form.Item name="driver" label="存储后端"><Radio.Group options={[{ label: '本地存储', value: 'local' }, { label: 'S3 / OSS / MinIO', value: 's3' }]} /></Form.Item><Typography.Title level={5}>S3 兼容配置</Typography.Title><Form.Item name={['s3','endpoint']} label="Endpoint"><Input placeholder="https://s3.example.com" /></Form.Item><Form.Item name={['s3','bucket']} label="Bucket"><Input /></Form.Item><Form.Item name={['s3','access_key']} label="Access Key"><Input /></Form.Item><Form.Item name={['s3','secret_key']} label="Secret"><Input.Password placeholder={config.s3.secret_key === '' ? '已配置则留空不修改' : ''} /></Form.Item><Form.Item name={['s3','region']} label="Region"><Input /></Form.Item><Form.Item name={['s3','prefix']} label="路径前缀"><Input /></Form.Item><Space><Button type="primary" loading={loading} onClick={() => void save()}>保存配置</Button><Button onClick={() => void test()}>测试连接</Button></Space></Form></Card></>
}
