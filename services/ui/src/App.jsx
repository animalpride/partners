import { MenuOutlined } from '@ant-design/icons'
import { Alert, Button, Drawer, Form, Grid, Input, Layout, Menu, Modal, Space, Typography } from 'antd'
import { useEffect, useState } from 'react'
import { Navigate, Route, Routes, useLocation, useNavigate } from 'react-router-dom'
import { getAuthState, login, logout } from './api'
import { AdminPanel } from './components/AdminPanel'
import { ApplicationPage } from './components/ApplicationPage'
import { CMSPageView } from './components/CMSPageView'

const NAV_ITEMS = [
  { key: '/', label: 'Partnership Overview' },
  { key: '/how-it-works', label: 'How it Works' },
  { key: '/case-studies', label: 'Case Studies' },
  { key: '/pricing', label: 'Pricing' },
  { key: '/faq', label: 'Partner FAQ' },
  { key: '/apply', label: 'Application' },
]

function Nav() {
  const navigate = useNavigate()
  const location = useLocation()
  const screens = Grid.useBreakpoint()
  const [mobileOpen, setMobileOpen] = useState(false)

  const selectedKey = NAV_ITEMS.some((item) => item.key === location.pathname) ? location.pathname : '/'

  const onNavigate = ({ key }) => {
    navigate(key)
    setMobileOpen(false)
  }

  if (!screens.md) {
    return (
      <>
        <Button icon={<MenuOutlined />} className="nav-toggle" onClick={() => setMobileOpen(true)}>
          Sections
        </Button>
        <Drawer title="Partner Sections" open={mobileOpen} onClose={() => setMobileOpen(false)} placement="left">
          <Menu mode="inline" selectedKeys={[selectedKey]} items={NAV_ITEMS} onClick={onNavigate} />
        </Drawer>
      </>
    )
  }

  return (
    <Menu mode="horizontal" selectedKeys={[selectedKey]} items={NAV_ITEMS} onClick={onNavigate} className="top-nav" />
  )
}

function PublicLayout({ canEdit, isAuthenticated, onLogin, onLogout, logoutLoading, loginOpen, setLoginOpen, loginLoading, loginError, setLoginError }) {
  const navigate = useNavigate()

  return (
    <Layout className="app-layout">
      <Layout.Header className="app-header">
        <div className="app-shell app-header-inner">
          <div className="site-brand">
            <div className="site-brand-logo">AP</div>
            <div>
              <Typography.Title level={3} className="site-title">
                Animal Pride Partners
              </Typography.Title>
              <span className="site-subtitle">Mission-aligned commerce for partner organizations</span>
            </div>
          </div>
          <Space>
            {canEdit ? (
              <Button onClick={() => navigate('/admin')}>Admin Panel</Button>
            ) : null}
            {!isAuthenticated ? (
              <Button
                type="primary"
                onClick={() => {
                  setLoginError('')
                  setLoginOpen(true)
                }}
              >
                Login
              </Button>
            ) : (
              <Button onClick={onLogout} loading={logoutLoading}>
                Logout
              </Button>
            )}
          </Space>
        </div>
      </Layout.Header>

      <Layout.Content className="app-shell app-content">
        <Nav />
        <Routes>
          <Route path="/" element={<CMSPageView slug="partnership-overview" />} />
          <Route path="/how-it-works" element={<CMSPageView slug="how-it-works" />} />
          <Route path="/case-studies" element={<CMSPageView slug="case-studies" />} />
          <Route path="/pricing" element={<CMSPageView slug="pricing-revenue-share" />} />
          <Route path="/faq" element={<CMSPageView slug="partner-faq" />} />
          <Route path="/apply" element={<ApplicationPage />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </Layout.Content>

      <Modal title="Login" open={loginOpen} onCancel={() => setLoginOpen(false)} footer={null} destroyOnClose>
        <Form layout="vertical" onFinish={onLogin}>
          <Form.Item label="Email" name="email" rules={[{ required: true, message: 'Email is required' }]}>
            <Input type="email" autoComplete="email" />
          </Form.Item>
          <Form.Item label="Password" name="password" rules={[{ required: true, message: 'Password is required' }]}>
            <Input.Password autoComplete="current-password" />
          </Form.Item>
          {loginError ? <Alert type="error" message={loginError} showIcon style={{ marginBottom: 12 }} /> : null}
          <Button type="primary" htmlType="submit" loading={loginLoading} block>
            Sign in
          </Button>
        </Form>
      </Modal>
    </Layout>
  )
}

export default function App() {
  const [canEdit, setCanEdit] = useState(false)
  const [isAuthenticated, setIsAuthenticated] = useState(false)
  const [loginOpen, setLoginOpen] = useState(false)
  const [loginLoading, setLoginLoading] = useState(false)
  const [loginError, setLoginError] = useState('')
  const [logoutLoading, setLogoutLoading] = useState(false)

  async function refreshAuth() {
    try {
      const authState = await getAuthState()
      setIsAuthenticated(authState.authenticated)
      setCanEdit(authState.canEditCms)
    } catch {
      setIsAuthenticated(false)
      setCanEdit(false)
    }
  }

  useEffect(() => {
    refreshAuth()
  }, [])

  async function onLogin(values) {
    setLoginLoading(true)
    setLoginError('')
    try {
      await login(values)
      await refreshAuth()
      setLoginOpen(false)
    } catch (err) {
      setLoginError(err.message)
    } finally {
      setLoginLoading(false)
    }
  }

  async function onLogout() {
    setLogoutLoading(true)
    try {
      await logout()
    } catch {
      // best-effort
    } finally {
      await refreshAuth()
      setLogoutLoading(false)
    }
  }

  const publicProps = { canEdit, isAuthenticated, onLogin, onLogout, logoutLoading, loginOpen, setLoginOpen, loginLoading, loginError, setLoginError }

  return (
    <Routes>
      <Route path="/admin/*" element={<AdminPanel />} />
      <Route path="/*" element={<PublicLayout {...publicProps} />} />
    </Routes>
  )
}
