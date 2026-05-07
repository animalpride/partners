import {
  Alert, Badge, Button, Card, Form, Input, Modal, Popconfirm, Select, Space,
  Spin, Table, Tag, Typography,
} from 'antd'
import { useEffect, useMemo, useState } from 'react'
import {
  createOAuthClient,
  getOAuthClients,
  rotateOAuthClientSecret,
  setOAuthClientStatus,
} from '../api'

const SCOPE_OPTIONS = [
  { label: 'Partner Applications: Read', value: 'partners_applications:read' },
  { label: 'Partner Applications: Write', value: 'partners_applications:write' },
]

function formatDate(iso) {
  if (!iso) return '—'
  return new Date(iso).toLocaleDateString('en-US', {
    month: 'short', day: 'numeric', year: 'numeric',
  })
}

export function OAuthClients() {
  const [clients, setClients] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [actionLoading, setActionLoading] = useState({})

  const [createOpen, setCreateOpen] = useState(false)
  const [creating, setCreating] = useState(false)
  const [createError, setCreateError] = useState('')
  const [createForm] = Form.useForm()

  const [secretOpen, setSecretOpen] = useState(false)
  const [secretValue, setSecretValue] = useState('')
  const [secretContext, setSecretContext] = useState('')

  async function loadClients() {
    setLoading(true)
    setError('')
    try {
      const data = await getOAuthClients()
      setClients(data || [])
    } catch (err) {
      setError(err.message || 'Failed to fetch OAuth clients')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadClients()
  }, [])

  async function onCreate(values) {
    setCreating(true)
    setCreateError('')
    try {
      const payload = {
        client_id: values.client_id.trim(),
        name: values.name.trim(),
        description: (values.description || '').trim(),
        scopes: values.scopes || [],
      }
      const result = await createOAuthClient(payload)
      setCreateOpen(false)
      createForm.resetFields()
      await loadClients()

      if (result?.client_secret) {
        setSecretContext(`Created client ${result.client?.client_id || payload.client_id}`)
        setSecretValue(result.client_secret)
        setSecretOpen(true)
      }
    } catch (err) {
      setCreateError(err.message || 'Failed to create OAuth client')
    } finally {
      setCreating(false)
    }
  }

  async function onToggleStatus(client) {
    const nextActive = client.active === 1 ? 0 : 1
    setActionLoading((prev) => ({ ...prev, [client.id]: 'status' }))
    setError('')
    try {
      await setOAuthClientStatus(client.id, nextActive)
      await loadClients()
    } catch (err) {
      setError(err.message || 'Failed to update OAuth client status')
    } finally {
      setActionLoading((prev) => ({ ...prev, [client.id]: undefined }))
    }
  }

  async function onRotateSecret(client) {
    setActionLoading((prev) => ({ ...prev, [client.id]: 'rotate' }))
    setError('')
    try {
      const result = await rotateOAuthClientSecret(client.id)
      if (result?.client_secret) {
        setSecretContext(`Rotated secret for ${client.client_id}`)
        setSecretValue(result.client_secret)
        setSecretOpen(true)
      }
    } catch (err) {
      setError(err.message || 'Failed to rotate OAuth client secret')
    } finally {
      setActionLoading((prev) => ({ ...prev, [client.id]: undefined }))
    }
  }

  const columns = useMemo(() => [
    {
      title: 'Client ID',
      dataIndex: 'client_id',
      key: 'client_id',
      width: 220,
      render: (value) => <Typography.Text code>{value}</Typography.Text>,
    },
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
      width: 220,
      render: (value) => value || <Typography.Text type="secondary">—</Typography.Text>,
    },
    {
      title: 'Description',
      dataIndex: 'description',
      key: 'description',
      render: (value) => value || <Typography.Text type="secondary">—</Typography.Text>,
    },
    {
      title: 'Status',
      dataIndex: 'active',
      key: 'active',
      width: 120,
      render: (value) => (
        <Badge color={value === 1 ? 'green' : 'red'} text={value === 1 ? 'Active' : 'Inactive'} />
      ),
    },
    {
      title: 'Created',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 120,
      render: formatDate,
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 230,
      render: (_, record) => {
        const loadingState = actionLoading[record.id]
        const isActive = record.active === 1
        return (
          <Space size="small">
            <Popconfirm
              title={isActive ? 'Deactivate this machine client?' : 'Activate this machine client?'}
              description={
                isActive
                  ? 'New tokens will no longer be issued for this client.'
                  : 'The client can request access tokens again.'
              }
              onConfirm={() => onToggleStatus(record)}
              okText={isActive ? 'Deactivate' : 'Activate'}
              okButtonProps={{ danger: isActive }}
            >
              <Button
                size="small"
                danger={isActive}
                loading={loadingState === 'status'}
                disabled={!!loadingState}
              >
                {isActive ? 'Deactivate' : 'Activate'}
              </Button>
            </Popconfirm>
            <Popconfirm
              title="Rotate client secret?"
              description="The old secret will stop working immediately."
              onConfirm={() => onRotateSecret(record)}
              okText="Rotate"
              okButtonProps={{ danger: true }}
            >
              <Button
                size="small"
                loading={loadingState === 'rotate'}
                disabled={!!loadingState}
              >
                Rotate Secret
              </Button>
            </Popconfirm>
          </Space>
        )
      },
    },
  ], [actionLoading])

  return (
    <div className="users-roles-panel">
      <Typography.Title level={3} style={{ marginBottom: 4 }}>Machine Clients</Typography.Title>
      <Typography.Paragraph type="secondary" style={{ marginBottom: 24 }}>
        Create and manage OAuth machine clients used for server-to-server integrations.
      </Typography.Paragraph>

      <Card
        title="OAuth Clients"
        extra={
          <Button type="primary" onClick={() => { setCreateError(''); setCreateOpen(true) }}>
            Create Client
          </Button>
        }
      >
        <Space direction="vertical" size={12} style={{ width: '100%' }}>
          <Alert
            type="warning"
            showIcon
            message="Client secrets are shown only at create and rotate time. Store them securely."
          />

          {error ? <Alert type="error" message={error} showIcon /> : null}

          {loading ? (
            <div style={{ textAlign: 'center', padding: 32 }}><Spin /></div>
          ) : (
            <Table
              dataSource={clients}
              columns={columns}
              rowKey="id"
              size="small"
              pagination={{ pageSize: 20, hideOnSinglePage: true }}
              locale={{ emptyText: 'No OAuth clients found' }}
            />
          )}
        </Space>
      </Card>

      <Modal
        title="Create OAuth Machine Client"
        open={createOpen}
        onCancel={() => {
          if (!creating) {
            setCreateOpen(false)
            setCreateError('')
          }
        }}
        onOk={() => createForm.submit()}
        okText="Create"
        confirmLoading={creating}
        destroyOnClose
      >
        {createError ? <Alert type="error" message={createError} showIcon style={{ marginBottom: 12 }} /> : null}

        <Form
          form={createForm}
          layout="vertical"
          onFinish={onCreate}
          initialValues={{ scopes: ['partners_applications:read'] }}
        >
          <Form.Item
            name="client_id"
            label="Client ID"
            rules={[
              { required: true, message: 'Client ID is required' },
              { pattern: /^[a-z0-9_-]{3,120}$/, message: 'Use lowercase letters, numbers, underscore, or hyphen' },
            ]}
          >
            <Input autoComplete="off" placeholder="partner-sync-prod" />
          </Form.Item>

          <Form.Item
            name="name"
            label="Display Name"
            rules={[{ required: true, message: 'Name is required' }]}
          >
            <Input autoComplete="off" placeholder="Partner Sync Production" />
          </Form.Item>

          <Form.Item name="description" label="Description">
            <Input.TextArea rows={3} placeholder="Machine client for partner application sync" />
          </Form.Item>

          <Form.Item
            name="scopes"
            label="Scopes"
            rules={[{ required: true, message: 'Select at least one scope' }]}
          >
            <Select
              mode="multiple"
              options={SCOPE_OPTIONS}
              placeholder="Select scopes"
              optionFilterProp="label"
            />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title="Client Secret"
        open={secretOpen}
        onCancel={() => setSecretOpen(false)}
        footer={<Button onClick={() => setSecretOpen(false)}>Done</Button>}
        destroyOnClose
      >
        <Alert
          type="warning"
          showIcon
          message="This secret is shown once. Copy it now and store it in a secure secret manager."
          style={{ marginBottom: 12 }}
        />
        {secretContext ? <Typography.Paragraph>{secretContext}</Typography.Paragraph> : null}
        <Typography.Text code copyable>{secretValue}</Typography.Text>
        <div style={{ marginTop: 12 }}>
          <Tag color="gold">Do not commit this secret to source control.</Tag>
        </div>
      </Modal>
    </div>
  )
}
