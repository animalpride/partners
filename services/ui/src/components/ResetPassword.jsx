import { Alert, Button, Card, Form, Input, Spin, Typography } from "antd";
import { useEffect, useState } from "react";
import { useSearchParams } from "react-router-dom";
import {
  completePasswordReset,
  fetchCsrf,
  validatePasswordResetToken,
} from "../api";

export function ResetPassword() {
  const [searchParams] = useSearchParams();
  const token = searchParams.get("token") || "";

  const [tokenState, setTokenState] = useState("loading"); // 'loading' | 'valid' | 'invalid'
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState(false);

  useEffect(() => {
    if (!token) {
      setTokenState("invalid");
      return;
    }

    let cancelled = false;

    async function validate() {
      try {
        await fetchCsrf();
        await validatePasswordResetToken(token);
        if (!cancelled) setTokenState("valid");
      } catch {
        if (!cancelled) setTokenState("invalid");
      }
    }

    validate();

    return () => {
      cancelled = true;
    };
  }, [token]);

  async function onFinish({ password, password_confirm }) {
    setLoading(true);
    setError("");
    try {
      await completePasswordReset(token, password, password_confirm);
      setSuccess(true);
    } catch (err) {
      setError(err.message || "Reset failed. The link may have expired.");
    } finally {
      setLoading(false);
    }
  }

  if (tokenState === "loading") {
    return (
      <div className="accept-invitation-page">
        <div className="accept-invitation-center">
          <Spin size="large" />
        </div>
      </div>
    );
  }

  if (tokenState === "invalid") {
    return (
      <div className="accept-invitation-page">
        <Card className="accept-invitation-card">
          <div className="accept-invitation-logo">
            <img
              src="/Logo-Wordmark-Dog-PartnersPlatform2.png"
              alt="Animal Pride Partners"
              style={{ height: 48 }}
            />
          </div>
          <Alert
            type="error"
            showIcon
            message="Reset link not valid"
            description="This password reset link is invalid or has already expired. Please request a new one."
            style={{ marginBottom: 16 }}
          />
          <div style={{ textAlign: "center" }}>
            <a href="/forgot-password">Request a new link</a>
          </div>
        </Card>
      </div>
    );
  }

  if (success) {
    return (
      <div className="accept-invitation-page">
        <Card className="accept-invitation-card">
          <div className="accept-invitation-logo">
            <img
              src="/Logo-Wordmark-Dog-PartnersPlatform2.png"
              alt="Animal Pride Partners"
              style={{ height: 48 }}
            />
          </div>
          <Typography.Title
            level={3}
            style={{ marginTop: 16, marginBottom: 8 }}
          >
            Password updated
          </Typography.Title>
          <Typography.Paragraph type="secondary">
            Your password has been reset successfully. You can now sign in with
            your new password.
          </Typography.Paragraph>
          <div style={{ textAlign: "center", marginTop: 8 }}>
            <a href="/">Back to site</a>
          </div>
        </Card>
      </div>
    );
  }

  return (
    <div className="accept-invitation-page">
      <Card className="accept-invitation-card">
        <div className="accept-invitation-logo">
          <img
            src="/Logo-Wordmark-Dog-PartnersPlatform2.png"
            alt="Animal Pride Partners"
            style={{ height: 48 }}
          />
        </div>

        <Typography.Title level={3} style={{ marginTop: 16, marginBottom: 4 }}>
          Reset your password
        </Typography.Title>
        <Typography.Paragraph type="secondary" style={{ marginBottom: 24 }}>
          Choose a new password for your account.
        </Typography.Paragraph>

        {error ? (
          <Alert
            type="error"
            message={error}
            showIcon
            style={{ marginBottom: 16 }}
          />
        ) : null}

        <Form layout="vertical" onFinish={onFinish} disabled={loading}>
          <Form.Item
            label="New Password"
            name="password"
            rules={[
              { required: true, message: "Password is required" },
              { min: 8, message: "Password must be at least 8 characters" },
            ]}
          >
            <Input.Password autoComplete="new-password" autoFocus />
          </Form.Item>

          <Form.Item
            label="Confirm New Password"
            name="password_confirm"
            dependencies={["password"]}
            rules={[
              { required: true, message: "Please confirm your password" },
              ({ getFieldValue }) => ({
                validator(_, value) {
                  if (!value || getFieldValue("password") === value)
                    return Promise.resolve();
                  return Promise.reject(new Error("Passwords do not match"));
                },
              }),
            ]}
          >
            <Input.Password autoComplete="new-password" />
          </Form.Item>

          <Button type="primary" htmlType="submit" loading={loading} block>
            Reset password
          </Button>
        </Form>
      </Card>
    </div>
  );
}
