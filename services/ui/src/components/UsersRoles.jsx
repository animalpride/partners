import {
  Alert, Badge, Button, Card, Form, Input, Modal, Popconfirm, Select, Space,
  Spin, Table, Tag, Typography,
} from 'antd'
import { useEffect, useState } from 'react'
import {
  assignRole, createInvitation, getInvitations, getRoles, getUsers,
  removeRole, resendInvitation, revokeInvitation, setUserActive,
} from '../api'

const INV_STATUS_COLORS = {
  pending: 'blue',
  expired: 'orange',
  revoked: 'red',
  accepted: 'green',
}

function capitalize(str) {
  if (!str) return '—'
  return str.charAt(0).toUpperCase() + str.slice(1)
}

function formatDate(iso) {
  if (!iso) return '—'
  return new Date(iso).toLocaleDateString('en-US', {
    month: 'short', day: 'numeric', year: 'numeric',
  })
}

function ExpiresCell({ expiresAt }) {
  const hoursLeft = Math.round((new Date(expiresAt).getTime() - Date.now()) / 1000 / 3600)
  if (hoursLeft <= 0) return <span style={{ color: '#d97706' }}>Expired</span>
  if (hoursLeft < 6) return <span style={{ color: '#d97706' }}>{hoursLeft}h left</span>
  return <span>{formatDate(expiresAt)}</span>
}

// ── Role Management Modal ─────────────────────────────────────────────────────
function RolesModal({ user, allRoles, open, onClose, onChanged }) {
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState('')
  const [addRoleId, setAddRoleId] = useState(null)

  const currentRoleIds = new Set((user?.roles || []).map((r) => r.id))
  const assignableRoles = allRoles.filter((r) => r.active && !currentRoleIds.has(r.id))

  async function handleRemove(roleId) {
    setBusy(true)
    setError('')
    try {
      await removeRole(user.id, roleId)
      await onChanged()
    } catch (err) {
      setError(err.message || 'Failed to remove role')
    } finally {
      setBusy(false)
    }
  }

  async function handleAdd() {
    if (!addRoleId) return
    setBusy(true)
    setError('')
    try {
      await assignRole(user.id, addRoleId)
      setAddRoleId(null)
      await onChanged()
    } catch (err) {
      setError(err.message || 'Failed to assign role')
    } finally {
      setBusy(false)
    }
  }

  return (
    <Modal
      title={`Manage Roles — ${user?.first_name || ''} ${user?.last_name || ''} (${user?.email || ''})`}
      open={open}
      onCancel={onClose}
      footer={<Button onClick={onClose}>Done</Button>}
      destroyOnClose
    >
      {error ? <Alert type="error" message={error} showIcon style={{ marginBottom: 12 }} /> : null}

      <Typography.Text type="secondary" style={{ display: 'block', marginBottom: 8 }}>Current roles</Typography.Text>
      <div style={{ marginBottom: 20, minHeight: 32 }}>
        {(user?.roles || []).length === 0 ? (
          <Typography.Text type="secondary">No roles assigned</Typography.Text>
        ) : (
          (user?.roles || []).map((role) => (
            <Tag
              key={role.id}
              closable
              onClose={(e) => { e.preventDefault(); handleRemove(role.id) }}
              style={{ marginBottom: 4 }}
            >
              {capitalize(role.name)}
            </Tag>
          ))
        )}
      </div>

      <Typography.Text type="secondary" style={{ display: 'block', marginBottom: 8 }}>Add a role</Typography.Text>
      <Space>
        <Select
          placeholder="Select role"
          style={{ width: 180 }}
          value={addRoleId}
          onChange={setAddRoleId}
          options={assignableRoles.map((r) => ({ value: r.id, label: capitalize(r.name) }))}
          disabled={busy || assignableRoles.length === 0}
        />
        <Button
          type="primary"
          onClick={handleAdd}
          loading={busy}
          disabled={!addRoleId}
        >
          Assign
        </Button>
      </Space>
      {assignableRoles.length === 0 && (user?.roles || []).length > 0 ? (
        <Typography.Text type="secondary" style={{ display: 'block', marginTop: 8, fontSize: 12 }}>
          All available roles are already assigned.
        </Typography.Text>
      ) : null}
    </Modal>
  )
}

// ── Main component ────────────────────────────────────────────────────────────
export function UsersRoles() {
  const [roles, setRoles] = useState([])
  const [rolesLoading, setRolesLoading] = useState(true)
  const [rolesError, setRolesError] = useState('')

  const [users, setUsers] = useState([])
  const [usersLoading, setUsersLoading] = useState(true)
  const [usersError, setUsersError] = useState('')

  const [invitations, setInvitations] = useState([])
  const [invLoading, setInvLoading] = useState(true)
  const [invError, setInvError] = useState('')

  const [inviteForm] = Form.useForm()
  const [inviting, setInviting] = useState(false)
  const [inviteError, setInviteError] = useState('')
  const [inviteSuccess, setInviteSuccess] = useState('')

  const [invActionLoading, setInvActionLoading] = useState({}) // keyed by email
  const [userActionLoading, setUserActionLoading] = useState({}) // keyed by user id

  const [rolesModalUser, setRolesModalUser] = useState(null)

  async function loadRoles() {
    setRolesLoading(true)
    setRolesError('')
    try {
      const data = await getRoles()
      setRoles(data || [])
    } catch (err) {
      setRolesError(err.message)
    } finally {
      setRolesLoading(false)
    }
  }

  async function loadUsers() {
    setUsersLoading(true)
    setUsersError('')
    try {
      const data = await getUsers()
      setUsers(data || [])
    } catch (err) {
      setUsersError(err.message)
    } finally {
      setUsersLoading(false)
    }
  }

  async function loadInvitations() {
    setInvLoading(true)
    setInvError('')
    try {
      const data = await getInvitations()
      setInvitations(data || [])
    } catch (err) {
      setInvError(err.message)
    } finally {
      setInvLoading(false)
    }
  }

  useEffect(() => {
    loadRoles()
    loadUsers()
    loadInvitations()
  }, [])

  // Pre-select admin role once roles load.
  useEffect(() => {
    if (roles.length > 0 && !inviteForm.getFieldValue('role_id')) {
      const adminRole = roles.find((r) => r.name === 'admin')
      if (adminRole) inviteForm.setFieldValue('role_id', adminRole.id)
    }
  }, [roles, inviteForm])

  async function onInvite(values) {
    setInviting(true)
    setInviteError('')
    setInviteSuccess('')
    try {
      await createInvitation(values.email.trim(), values.role_id)
      setInviteSuccess(`Invitation sent to ${values.email.trim()}`)
      inviteForm.resetFields(['email'])
      await loadInvitations()
    } catch (err) {
      setInviteError(err.message || 'Failed to send invitation')
    } finally {
      setInviting(false)
    }
  }

  async function onResend(email) {
    setInvActionLoading((prev) => ({ ...prev, [email]: 'resend' }))
    try {
      await resendInvitation(email)
      await loadInvitations()
    } catch (err) {
      setInvError(err.message || 'Failed to resend invitation')
    } finally {
      setInvActionLoading((prev) => ({ ...prev, [email]: undefined }))
    }
  }

  async function onRevoke(email) {
    setInvActionLoading((prev) => ({ ...prev, [email]: 'revoke' }))
    try {
      await revokeInvitation(email)
      await loadInvitations()
    } catch (err) {
      setInvError(err.message || 'Failed to revoke invitation')
    } finally {
      setInvActionLoading((prev) => ({ ...prev, [email]: undefined }))
    }
  }

  async function onToggleActive(user) {
    const newActive = user.active === 1 ? 0 : 1
    setUserActionLoading((prev) => ({ ...prev, [user.id]: 'active' }))
    try {
      await setUserActive(user.id, newActive)
      await loadUsers()
    } catch (err) {
      setUsersError(err.message || 'Failed to update user status')
    } finally {
      setUserActionLoading((prev) => ({ ...prev, [user.id]: undefined }))
    }
  }

  // Called by the modal after a role change so the users list reflects it.
  async function onRoleChanged() {
    const data = await getUsers()
    setUsers(data || [])
    // Sync the modal user object to the updated data.
    if (rolesModalUser) {
      const updated = (data || []).find((u) => u.id === rolesModalUser.id)
      if (updated) setRolesModalUser(updated)
    }
  }

  const roleOptions = roles
    .filter((r) => r.active)
    .map((r) => ({ value: r.id, label: capitalize(r.name) }))

  // ── Users table columns ───────────────────────────────────────────────────
  const userColumns = [
    {
      title: 'Name',
      key: 'name',
      render: (_, u) => {
        const name = [u.first_name, u.last_name].filter(Boolean).join(' ')
        return name || <Typography.Text type="secondary">—</Typography.Text>
      },
    },
    {
      title: 'Email',
      dataIndex: 'email',
      key: 'email',
      ellipsis: true,
    },
    {
      title: 'Roles',
      key: 'roles',
      render: (_, u) =>
        (u.roles || []).length === 0
          ? <Typography.Text type="secondary">—</Typography.Text>
          : (u.roles || []).map((r) => <Tag key={r.id}>{capitalize(r.name)}</Tag>),
    },
    {
      title: 'Status',
      key: 'active',
      width: 100,
      render: (_, u) => (
        <Badge
          color={u.active === 1 ? 'green' : 'red'}
          text={u.active === 1 ? 'Active' : 'Inactive'}
        />
      ),
    },
    {
      title: 'Joined',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 130,
      render: formatDate,
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 190,
      render: (_, u) => {
        const loading = userActionLoading[u.id]
        const isActive = u.active === 1
        return (
          <Space size="small">
            <Button
              size="small"
              onClick={() => setRolesModalUser(u)}
            >
              Roles
            </Button>
            <Popconfirm
              title={isActive ? 'Deactivate this user?' : 'Activate this user?'}
              description={
                isActive
                  ? 'Their active session will be ended immediately.'
                  : 'The user will be able to log in again.'
              }
              onConfirm={() => onToggleActive(u)}
              okText={isActive ? 'Deactivate' : 'Activate'}
              okButtonProps={{ danger: isActive }}
            >
              <Button
                size="small"
                danger={isActive}
                loading={loading === 'active'}
                disabled={!!loading}
              >
                {isActive ? 'Deactivate' : 'Activate'}
              </Button>
            </Popconfirm>
          </Space>
        )
      },
    },
  ]

  // ── Invitations table columns ─────────────────────────────────────────────
  const invColumns = [
    {
      title: 'Email',
      dataIndex: 'email',
      key: 'email',
      ellipsis: true,
    },
    {
      title: 'Role',
      dataIndex: 'role_name',
      key: 'role_name',
      width: 120,
      render: (name) => <Tag>{capitalize(name)}</Tag>,
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      width: 110,
      render: (status) => (
        <Badge
          color={INV_STATUS_COLORS[status] || 'default'}
          text={capitalize(status)}
        />
      ),
    },
    {
      title: 'Sent',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 130,
      render: formatDate,
    },
    {
      title: 'Expires',
      dataIndex: 'expires_at',
      key: 'expires_at',
      width: 130,
      render: (val, record) =>
        record.status === 'pending' ? <ExpiresCell expiresAt={val} /> : '—',
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 160,
      render: (_, record) => {
        if (record.status !== 'pending') return null
        const loading = invActionLoading[record.email]
        return (
          <Space size="small">
            <Button
              size="small"
              loading={loading === 'resend'}
              disabled={!!loading}
              onClick={() => onResend(record.email)}
            >
              Resend
            </Button>
            <Popconfirm
              title="Revoke this invitation?"
              description="The recipient will no longer be able to use this link."
              onConfirm={() => onRevoke(record.email)}
              okText="Revoke"
              okButtonProps={{ danger: true }}
            >
              <Button
                size="small"
                danger
                loading={loading === 'revoke'}
                disabled={!!loading}
              >
                Revoke
              </Button>
            </Popconfirm>
          </Space>
        )
      },
    },
  ]

  return (
    <div className="users-roles-panel">
      <Typography.Title level={3} style={{ marginBottom: 4 }}>Users &amp; Roles</Typography.Title>
      <Typography.Paragraph type="secondary" style={{ marginBottom: 24 }}>
        Manage user accounts and roles. Invite new users by email — they receive a time-limited link to complete registration.
      </Typography.Paragraph>

      {/* ── Invite form ── */}
      <Card title="Send Invitation" className="users-roles-invite-card">
        {rolesError ? (
          <Alert type="error" message={rolesError} showIcon style={{ marginBottom: 16 }} />
        ) : null}
        <Form
          form={inviteForm}
          layout="inline"
          onFinish={onInvite}
          style={{ flexWrap: 'wrap', gap: 8 }}
        >
          <Form.Item
            name="email"
            rules={[{ required: true, type: 'email', message: 'Valid email required' }]}
            style={{ flex: '1 1 220px', minWidth: 220 }}
          >
            <Input placeholder="Email address" autoComplete="off" />
          </Form.Item>
          <Form.Item
            name="role_id"
            rules={[{ required: true, message: 'Select a role' }]}
            style={{ width: 150 }}
          >
            <Select placeholder="Role" options={roleOptions} loading={rolesLoading} />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={inviting}>
              Send Invitation
            </Button>
          </Form.Item>
        </Form>
        {inviteError ? <Alert type="error" message={inviteError} showIcon style={{ marginTop: 12 }} /> : null}
        {inviteSuccess ? <Alert type="success" message={inviteSuccess} showIcon style={{ marginTop: 12 }} /> : null}
      </Card>

      {/* ── Users table ── */}
      <Typography.Title level={5} style={{ marginBottom: 12 }}>Users</Typography.Title>

      {usersError ? (
        <Alert type="error" message={usersError} showIcon style={{ marginBottom: 12 }} />
      ) : null}

      {usersLoading ? (
        <div style={{ textAlign: 'center', padding: 32 }}><Spin /></div>
      ) : (
        <Table
          dataSource={users}
          columns={userColumns}
          rowKey="id"
          size="small"
          pagination={{ pageSize: 20, hideOnSinglePage: true }}
          locale={{ emptyText: 'No users found' }}
          style={{ marginBottom: 32 }}
        />
      )}

      {/* ── Pending invitations table ── */}
      <Typography.Title level={5} style={{ marginBottom: 12 }}>Pending Invitations</Typography.Title>

      {invError ? (
        <Alert type="error" message={invError} showIcon style={{ marginBottom: 12 }} />
      ) : null}

      {invLoading ? (
        <div style={{ textAlign: 'center', padding: 32 }}><Spin /></div>
      ) : (
        <Table
          dataSource={invitations}
          columns={invColumns}
          rowKey="id"
          size="small"
          pagination={{ pageSize: 20, hideOnSinglePage: true }}
          locale={{ emptyText: 'No pending invitations' }}
        />
      )}

      {/* ── Role management modal ── */}
      <RolesModal
        user={rolesModalUser}
        allRoles={roles}
        open={!!rolesModalUser}
        onClose={() => setRolesModalUser(null)}
        onChanged={onRoleChanged}
      />
    </div>
  )
}


