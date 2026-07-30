import {
  Alert,
  Button,
  Card,
  Descriptions,
  Form,
  Input,
  Popconfirm,
  Progress,
  Space,
  Spin,
  Steps,
  Tag,
  Typography,
  message,
} from 'antd'
import { useCallback, useEffect, useMemo, useState } from 'react'
import {
  domainApi,
  type DomainBindState,
  type DomainBindStatus,
} from '../api/domain'
import PageHeader from '../components/PageHeader'

const statusLabels: Record<DomainBindState, string> = {
  none: '未绑定',
  pending_verification: '正在验证 DNS',
  verified: 'DNS 验证通过',
  creating_site: '正在创建站点',
  configuring: '正在配置域名',
  ready: '已就绪',
  verification_failed: 'DNS 验证失败',
  setup_failed: '绑定失败',
}

const statusColors: Record<DomainBindState, string> = {
  none: 'default',
  pending_verification: 'processing',
  verified: 'success',
  creating_site: 'processing',
  configuring: 'processing',
  ready: 'success',
  verification_failed: 'error',
  setup_failed: 'error',
}

const flowSteps = [
  { title: '验证 DNS', description: '确认 CNAME 或 A 记录已指向平台服务器' },
  { title: '创建站点', description: '由系统自动在宝塔中创建租户站点' },
  { title: '配置访问', description: '添加反向代理并禁止访问 /admin' },
  { title: '申请 HTTPS', description: '自动申请并等待 Let’s Encrypt 证书生效' },
  { title: '完成绑定', description: '保存域名并开始提供租户学习站点' },
]

function isWorking(state?: DomainBindState) {
  return state === 'pending_verification'
    || state === 'creating_site'
    || state === 'configuring'
}

function isFailed(state?: DomainBindState) {
  return state === 'verification_failed' || state === 'setup_failed'
}

export default function DomainSettings() {
  const [form] = Form.useForm<{ domain: string }>()
  const [status, setStatus] = useState<DomainBindStatus>()
  const [loading, setLoading] = useState(false)
  const [statusError, setStatusError] = useState(false)
  const domainValue = Form.useWatch('domain', form)

  const refreshStatus = useCallback(async () => {
    try {
      const next = await domainApi.status()
      setStatus(next)
      setStatusError(false)
      if (next.domain) {
        form.setFieldValue('domain', next.domain)
      }
      return next
    } catch (error) {
      setStatusError(true)
      throw error
    }
  }, [form])

  useEffect(() => {
    void refreshStatus().catch(() => undefined)
  }, [refreshStatus])

  useEffect(() => {
    if (!isWorking(status?.state)) return
    const timer = window.setInterval(() => {
      void refreshStatus().catch(() => undefined)
    }, 1500)
    return () => window.clearInterval(timer)
  }, [refreshStatus, status?.state])

  const currentStep = useMemo(
    () => Math.max(0, Math.min(flowSteps.length - 1, (status?.current_step || 1) - 1)),
    [status?.current_step],
  )

  const validateDomain = async () => {
    const { domain } = await form.validateFields()
    const normalized = domain.trim().toLowerCase()
    setStatus({
      state: 'pending_verification',
      domain: normalized,
      message: '正在查询 CNAME 和 A 记录',
      current_step: 1,
      total_steps: 5,
      cname_target: status?.cname_target || 'play.imai.work',
    })
    setLoading(true)
    try {
      const next = await domainApi.verify(normalized)
      setStatus(next)
      message.success('DNS 验证通过，可以开始自动绑定')
    } catch {
      try {
        await refreshStatus()
      } catch {
        setStatus({
          state: 'verification_failed',
          domain: normalized,
          message: '验证请求失败，请检查网络后重试',
          current_step: 1,
          total_steps: 5,
          cname_target: status?.cname_target || 'play.imai.work',
        })
      }
    } finally {
      setLoading(false)
    }
  }

  const bindDomain = async () => {
    const { domain } = await form.validateFields()
    setLoading(true)
    try {
      const next = await domainApi.bind(domain.trim())
      setStatus(next)
      message.success('已开始自动配置，请稍候')
    } catch {
      await refreshStatus().catch(() => undefined)
    } finally {
      setLoading(false)
    }
  }

  const unbindDomain = async () => {
    setLoading(true)
    try {
      const next = await domainApi.unbind()
      setStatus(next)
      form.resetFields()
      message.success('域名已解绑')
    } catch {
      await refreshStatus().catch(() => undefined)
    } finally {
      setLoading(false)
    }
  }

  const bindEnabled =
    status?.state === 'verified'
    && status.domain === form.getFieldValue('domain')?.trim().toLowerCase()

  return (
    <>
      <PageHeader
        title="域名设置"
        description="只需配置一次 CNAME，系统会自动完成站点、反向代理和 HTTPS 证书配置。"
      />
      <Card style={{ maxWidth: 860 }}>
        <Space direction="vertical" size={24} style={{ width: '100%' }}>
          <Descriptions column={1} size="small">
            <Descriptions.Item label="当前状态">
              {status ? (
                <Tag color={statusColors[status.state]}>
                  {statusLabels[status.state]}
                </Tag>
              ) : statusError ? <Tag color="error">状态加载失败</Tag> : <Spin size="small" />}
            </Descriptions.Item>
            {status?.domain && (
              <Descriptions.Item label="当前域名">
                <Typography.Text copyable>{status.domain}</Typography.Text>
              </Descriptions.Item>
            )}
            <Descriptions.Item label="CNAME 记录值">
              <Typography.Text copyable>
                {status?.cname_target || 'play.imai.work'}
              </Typography.Text>
            </Descriptions.Item>
            {status?.state === 'ready' && (
              <Descriptions.Item label="HTTPS 证书">
                <Tag color="success">已启用，宝塔自动续期</Tag>
              </Descriptions.Item>
            )}
          </Descriptions>

          <Alert
            type={statusError || isFailed(status?.state) ? 'error' : status?.state === 'ready' ? 'success' : 'info'}
            showIcon
            message={statusError ? '域名状态加载失败，请刷新页面重试' : status?.message || '请先在域名服务商配置 CNAME，然后回来验证'}
            description={
              !statusError && status?.state === 'none'
                ? `将你的子域名 CNAME 到 ${status.cname_target || 'play.imai.work'}，DNS 生效后点击“验证域名”。`
                : undefined
            }
          />

          {status?.state !== 'ready' && (
            <Card size="small" title="请在域名服务商添加一条解析记录">
              <Descriptions column={1} size="small">
                <Descriptions.Item label="记录类型">CNAME</Descriptions.Item>
                <Descriptions.Item label="完整域名">
                  {domainValue?.trim() || '例如 academy.example.com'}
                </Descriptions.Item>
                <Descriptions.Item label="记录值">
                  <Typography.Text copyable>
                    {status?.cname_target || 'play.imai.work'}
                  </Typography.Text>
                </Descriptions.Item>
                <Descriptions.Item label="TTL">600 秒（或使用默认值）</Descriptions.Item>
              </Descriptions>
              <Typography.Text type="secondary">
                DNS 控制台中的“主机记录”请按所选根域名填写相对前缀；不确定时以域名服务商提示为准。
                DNS 通常几分钟生效，部分服务商最长可能需要 24–48 小时。无需在宝塔手动建站或上传证书。
              </Typography.Text>
            </Card>
          )}

          <Form form={form} layout="vertical">
            <Form.Item
              name="domain"
              label="自定义域名"
              rules={[
                { required: true, message: '请输入自定义域名' },
                {
                  pattern: /^(?=.{1,253}$)(?!-)(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$/i,
                  message: '请输入合法域名，例如 academy.example.com',
                },
              ]}
            >
              <Input
                placeholder="academy.example.com"
                disabled={isWorking(status?.state) || status?.state === 'ready'}
                onChange={() => {
                  if (status?.state === 'verified') {
                    setStatus({ ...status, state: 'none', message: '域名已修改，请重新验证' })
                  }
                }}
              />
            </Form.Item>
            <Space wrap>
              <Button
                type="primary"
                loading={status?.state === 'pending_verification'}
                disabled={isWorking(status?.state) || status?.state === 'ready'}
                onClick={() => void validateDomain()}
              >
                验证域名
              </Button>
              <Button
                type="primary"
                loading={loading || status?.state === 'creating_site' || status?.state === 'configuring'}
                disabled={!bindEnabled || isWorking(status?.state)}
                onClick={() => void bindDomain()}
              >
                自动绑定
              </Button>
              {(status?.state === 'ready' || status?.state === 'setup_failed') && (
                <Popconfirm
                  title={status.state === 'ready' ? '确认解绑该域名？' : '确认清理失败的绑定？'}
                  description="系统将删除可能存在的宝塔站点并清除租户域名。"
                  okText="确认解绑"
                  cancelText="取消"
                  onConfirm={() => void unbindDomain()}
                >
                  <Button danger loading={loading}>解绑</Button>
                </Popconfirm>
              )}
            </Space>
          </Form>

          <div>
            <Typography.Title level={5}>自动配置进度</Typography.Title>
            {isWorking(status?.state) && (
              <Progress
                percent={Math.round(((status?.current_step || 1) / (status?.total_steps || 5)) * 100)}
                status="active"
                showInfo={false}
                style={{ marginBottom: 16 }}
              />
            )}
            <Steps
              direction="vertical"
              size="small"
              current={currentStep}
              status={isFailed(status?.state) ? 'error' : status?.state === 'ready' ? 'finish' : 'process'}
              items={flowSteps}
            />
          </div>
        </Space>
      </Card>
    </>
  )
}
