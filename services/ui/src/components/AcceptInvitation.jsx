import { Alert, Button, Card, Form, Input, Spin, Typography } from "antd";
import { useEffect, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { fetchCsrf, registerInvitation, validateInvitation } from "../api";

const ROLE_LABELS = {
  admin: "Administrator",
  partner: "Partner",
};

function roleHeading(roleName) {
  const label = ROLE_LABELS[roleName] || roleName;
  return label
    ? `You've been invited as a Animal Pride Partner ${label}`
    : "You've been invited to join Animal Pride Partners";
}

function roleSubtext(roleName) {
  if (roleName === "admin") {
    return "As an Administrator you will have full access to manage content, users, and settings for the Animal Pride partners platform.";
  }
  if (roleName === "partner") {
    return "Your partner account will give you access to the Animal Pride partner portal and resources.";
  }
  return "Complete your registration below to get started.";
}

export function AcceptInvitation() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const token = searchParams.get("token") || "";

  const [validating, setValidating] = useState(true);
  const [validateError, setValidateError] = useState("");
  const [invitation, setInvitation] = useState(null); // { email, role_name, expires_at }

  const [submitting, setSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState("");

  useEffect(() => {
    if (!token) {
      setValidateError(
        "No invitation token provided. Please use the link from your invitation email.",
      );
      setValidating(false);
      return;
    }

    let cancelled = false;

    async function load() {
      try {
        await fetchCsrf();
        const data = await validateInvitation(token);
        if (!cancelled) {
          setInvitation(data);
        }
      } catch (err) {
        if (!cancelled) {
          setValidateError(
            err.message || "This invitation is no longer valid.",
          );
        }
      } finally {
        if (!cancelled) setValidating(false);
      }
    }

    load();

    return () => {
      cancelled = true;
    };
  }, [token]);

  async function onFinish(values) {
    setSubmitting(true);
    setSubmitError("");
    try {
      await registerInvitation({
        token,
        first_name: values.first_name.trim(),
        last_name: values.last_name.trim(),
        password: values.password,
        password_confirm: values.password_confirm,
      });
      navigate(invitation?.role_name === "admin" ? "/admin" : "/", { replace: true });
    } catch (err) {
      setSubmitError(err.message || "Registration failed. Please try again.");
    } finally {
      setSubmitting(false);
    }
  }

  if (validating) {
    return (
      <div className="accept-invitation-page">
        <div className="accept-invitation-center">
          <Spin size="large" />
        </div>
      </div>
    );
  }

  if (validateError) {
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
            message="Invitation Not Valid"
            description={validateError}
            style={{ marginBottom: 0 }}
          />
        </Card>
      </div>
    );
  }

  const roleName = invitation?.role_name || "";

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
          {roleHeading(roleName)}
        </Typography.Title>
        <Typography.Paragraph type="secondary" style={{ marginBottom: 24 }}>
          {roleSubtext(roleName)}
        </Typography.Paragraph>

        <Typography.Text
          type="secondary"
          style={{ display: "block", marginBottom: 20 }}
        >
          Registering for <strong>{invitation?.email}</strong>
        </Typography.Text>

        <Form layout="vertical" onFinish={onFinish} disabled={submitting}>
          <div style={{ display: "flex", gap: 12 }}>
            <Form.Item
              label="First Name"
              name="first_name"
              style={{ flex: 1 }}
              rules={[{ required: true, message: "First name is required" }]}
            >
              <Input autoComplete="given-name" />
            </Form.Item>
            <Form.Item
              label="Last Name"
              name="last_name"
              style={{ flex: 1 }}
              rules={[{ required: true, message: "Last name is required" }]}
            >
              <Input autoComplete="family-name" />
            </Form.Item>
          </div>

          <Form.Item
            label="Password"
            name="password"
            rules={[
              { required: true, message: "Password is required" },
              { min: 8, message: "Password must be at least 8 characters" },
            ]}
          >
            <Input.Password autoComplete="new-password" />
          </Form.Item>

          <Form.Item
            label="Confirm Password"
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

          {submitError ? (
            <Alert
              type="error"
              message={submitError}
              showIcon
              style={{ marginBottom: 16 }}
            />
          ) : null}

          <Button type="primary" htmlType="submit" loading={submitting} block>
            Complete Registration
          </Button>
        </Form>
      </Card>
    </div>
  );
}
