import { MenuOutlined } from "@ant-design/icons";
import {
  Alert,
  Button,
  Drawer,
  Form,
  Grid,
  Input,
  Layout,
  Menu,
  Modal,
  Space,
  Typography,
} from "antd";
import { useEffect, useState } from "react";
import {
  Navigate,
  Route,
  Routes,
  useLocation,
  useNavigate,
} from "react-router-dom";
import {
  getAuthState,
  getComingSoonState,
  login,
  logout,
  unlockComingSoonPreview,
} from "./api";
import { AcceptInvitation } from "./components/AcceptInvitation";
import { AdminPanel } from "./components/AdminPanel";
import { ForgotPassword } from "./components/ForgotPassword";
import { ResetPassword } from "./components/ResetPassword";
import { ApplicationPage } from "./components/ApplicationPage";
import { CMSPageView } from "./components/CMSPageView";
import { ComingSoonPage } from "./components/ComingSoonPage";

const PREVIEW_PREFIX = "/preview/";

const NAV_ITEMS = [
  { key: "/", label: "Partnership Overview" },
  { key: "/how-it-works", label: "How it Works" },
  { key: "/case-studies", label: "Case Studies" },
  { key: "/pricing", label: "Revenue Share" },
  // { key: "/faq", label: "Partner FAQ" },
  { key: "/apply", label: "Application" },
];

function Nav({ mobileOpen, setMobileOpen, mobileExtra }) {
  const navigate = useNavigate();
  const location = useLocation();
  const screens = Grid.useBreakpoint();

  const selectedKey = NAV_ITEMS.some((item) => item.key === location.pathname)
    ? location.pathname
    : "/";

  const onNavigate = ({ key }) => {
    navigate(key);
    setMobileOpen(false);
  };

  if (!screens.md) {
    return (
      <Drawer
        title="Partner Sections"
        open={mobileOpen}
        onClose={() => setMobileOpen(false)}
        placement="left"
      >
        <Menu
          mode="inline"
          selectedKeys={[selectedKey]}
          items={NAV_ITEMS}
          onClick={onNavigate}
        />
        {mobileExtra ? (
          <div style={{ marginTop: 16 }}>{mobileExtra}</div>
        ) : null}
      </Drawer>
    );
  }

  return (
    <Menu
      mode="horizontal"
      selectedKeys={[selectedKey]}
      items={NAV_ITEMS}
      onClick={onNavigate}
      className="top-nav"
    />
  );
}

function PublicLayout({
  canEdit,
  isAuthenticated,
  onLogin,
  onLogout,
  logoutLoading,
  loginOpen,
  setLoginOpen,
  loginLoading,
  loginError,
  setLoginError,
}) {
  const navigate = useNavigate();
  const screens = Grid.useBreakpoint();
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  const mobileExtra = (
    <Space direction="vertical" style={{ width: "100%" }}>
      {canEdit ? (
        <Button
          block
          onClick={() => {
            navigate("/admin");
            setMobileMenuOpen(false);
          }}
        >
          Admin Panel
        </Button>
      ) : null}
      {!isAuthenticated ? (
        <Button
          type="primary"
          block
          onClick={() => {
            setLoginError("");
            setLoginOpen(true);
            setMobileMenuOpen(false);
          }}
        >
          Login
        </Button>
      ) : (
        <Button block onClick={onLogout} loading={logoutLoading}>
          Logout
        </Button>
      )}
    </Space>
  );

  return (
    <Layout className="app-layout">
      <Layout.Header className="app-header">
        <div className="app-shell app-header-inner">
          <div className="site-brand">
            <img
              src="/Logo-Wordmark-Dog-PartnersPlatform2.png"
              alt="Animal Pride Partners Platform"
              className="site-brand-logo"
            />

          </div>
          {screens.md ? (
            <Space>
              {canEdit ? (
                <Button onClick={() => navigate("/admin")}>Admin Panel</Button>
              ) : null}
              {!isAuthenticated ? (
                <Button
                  type="primary"
                  onClick={() => {
                    setLoginError("");
                    setLoginOpen(true);
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
          ) : (
            <Button
              icon={<MenuOutlined />}
              type="text"
              onClick={() => setMobileMenuOpen(true)}
            />
          )}
        </div>
      </Layout.Header>

      <Layout.Content className="app-shell app-content">
        <Nav
          mobileOpen={mobileMenuOpen}
          setMobileOpen={setMobileMenuOpen}
          mobileExtra={mobileExtra}
        />
        <Routes>
          <Route
            path="/"
            element={<CMSPageView slug="partnership-overview" />}
          />
          <Route
            path="/how-it-works"
            element={<CMSPageView slug="how-it-works" />}
          />
          <Route
            path="/case-studies"
            element={<CMSPageView slug="case-studies" />}
          />
          <Route
            path="/pricing"
            element={<CMSPageView slug="pricing-revenue-share" />}
          />
          <Route path="/faq" element={<CMSPageView slug="partner-faq" />} />
          <Route path="/apply" element={<ApplicationPage />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </Layout.Content>

      <Modal
        title="Login"
        open={loginOpen}
        onCancel={() => setLoginOpen(false)}
        footer={null}
        destroyOnClose
      >
        <Form layout="vertical" onFinish={onLogin}>
          <Form.Item
            label="Email"
            name="email"
            rules={[{ required: true, message: "Email is required" }]}
          >
            <Input type="email" autoComplete="email" />
          </Form.Item>
          <Form.Item
            label="Password"
            name="password"
            rules={[{ required: true, message: "Password is required" }]}
          >
            <Input.Password autoComplete="current-password" />
          </Form.Item>
          {loginError ? (
            <Alert
              type="error"
              message={loginError}
              showIcon
              style={{ marginBottom: 12 }}
            />
          ) : null}
          <Button type="primary" htmlType="submit" loading={loginLoading} block>
            Sign in
          </Button>
          <div style={{ textAlign: "center", marginTop: 12 }}>
            <Typography.Link
              onClick={() => {
                setLoginOpen(false);
                navigate("/forgot-password");
              }}
            >
              Forgot password?
            </Typography.Link>
          </div>
        </Form>
      </Modal>
    </Layout>
  );
}

export default function App() {
  const location = useLocation();
  const navigate = useNavigate();
  const [canEdit, setCanEdit] = useState(false);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [loginOpen, setLoginOpen] = useState(false);
  const [loginLoading, setLoginLoading] = useState(false);
  const [loginError, setLoginError] = useState("");
  const [logoutLoading, setLogoutLoading] = useState(false);
  const [bootLoading, setBootLoading] = useState(true);
  const [comingSoonState, setComingSoonState] = useState({
    enabled: false,
    preview_unlocked: false,
    message: "",
  });

  async function refreshAuth() {
    try {
      const authState = await getAuthState();
      setIsAuthenticated(authState.authenticated);
      setCanEdit(authState.canEditCms);
    } catch {
      setIsAuthenticated(false);
      setCanEdit(false);
    }
  }

  async function refreshSiteState() {
    const state = await getComingSoonState();
    setComingSoonState(state);
  }

  function getPreviewToken(pathname) {
    if (!pathname.startsWith(PREVIEW_PREFIX)) {
      return "";
    }

    return pathname.slice(PREVIEW_PREFIX.length).trim();
  }

  useEffect(() => {
    let cancelled = false;

    async function bootstrap() {
      await Promise.all([refreshAuth(), refreshSiteState()]);

      if (!cancelled) {
        setBootLoading(false);
      }
    }

    bootstrap();

    return () => {
      cancelled = true;
    };
  }, []);

  useEffect(() => {
    const previewToken = getPreviewToken(location.pathname);
    if (!previewToken) {
      return;
    }

    let cancelled = false;

    async function unlockPreview() {
      try {
        await unlockComingSoonPreview(previewToken);
      } catch {
        // best-effort: keep gate closed for invalid tokens
      }

      await Promise.all([refreshAuth(), refreshSiteState()]);

      if (!cancelled) {
        setBootLoading(false);
        navigate("/", { replace: true });
      }
    }

    unlockPreview();

    return () => {
      cancelled = true;
    };
  }, [location.pathname, navigate]);

  useEffect(() => {
    if (bootLoading) {
      return;
    }

    if (!comingSoonState.enabled || comingSoonState.preview_unlocked) {
      return;
    }

    if (
      location.pathname !== "/" &&
      !location.pathname.startsWith(PREVIEW_PREFIX)
    ) {
      navigate("/", { replace: true });
    }
  }, [
    bootLoading,
    comingSoonState.enabled,
    comingSoonState.preview_unlocked,
    location.pathname,
    navigate,
  ]);

  async function onLogin(values) {
    setLoginLoading(true);
    setLoginError("");
    try {
      await login(values);
      await refreshAuth();
      setLoginOpen(false);
    } catch (err) {
      setLoginError(err.message);
    } finally {
      setLoginLoading(false);
    }
  }

  async function onLogout() {
    setLogoutLoading(true);
    try {
      await logout();
    } catch {
      // best-effort
    } finally {
      await refreshAuth();
      setLogoutLoading(false);
    }
  }

  const publicProps = {
    canEdit,
    isAuthenticated,
    onLogin,
    onLogout,
    logoutLoading,
    loginOpen,
    setLoginOpen,
    loginLoading,
    loginError,
    setLoginError,
  };

  if (bootLoading) {
    return null;
  }

  if (comingSoonState.enabled && !comingSoonState.preview_unlocked) {
    return <ComingSoonPage message={comingSoonState.message} />;
  }

  return (
    <Routes>
      <Route path="/admin/*" element={<AdminPanel />} />
      <Route path="/accept-invitation" element={<AcceptInvitation />} />
      <Route path="/forgot-password" element={<ForgotPassword />} />
      <Route path="/reset-password" element={<ResetPassword />} />
      <Route path="/*" element={<PublicLayout {...publicProps} />} />
    </Routes>
  );
}
