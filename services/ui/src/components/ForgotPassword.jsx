import { Alert, Button, Card, Form, Input, Typography } from 'antd'
import { useState } from 'react'
import { fetchCsrf, requestPasswordReset } from '../api'

export function ForgotPassword() {
  const [loading, setLoading] = useState(false)
  const [submitted, setSubmitted] = useState(false)
  const [error, setError] = useState('')

  async function onFinish({ email }) {
    setLoading(true)
    setError('')
    try {
      await fetchCsrf()
      await requestPasswordReset(email.trim())
      setSubmitted(true)
    } catch (err) {
      setError(err.message || 'Something went wrong. Please try again.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="accept-invitation-page">
      <Card className="accept-invitation-card">
        <div className="accept-invitation-logo">
          <img src="/AnimalPridePartnerLogotrans.png" alt="Animal Pride Partners" style={{ height: 48 }} />
        </div>

        {submitted ? (
          <>
            <Typography.Title level={3} style={{ marginTop: 16, marginBottom: 8 }}>
              Check your email
            </Typography.Title>
            <Typography.Paragraph type="secondary">
              If that address is registered and active, a password reset link has been sent. The link will expire in 15
              minutes.
            </Typography.Paragraph>
            <div style={{ textAlign: 'center', marginTop: 16 }}>
              <a href="/">Back to site</a>
            </div>
          </>
        ) : (
          <>
            <Typography.Title level={3} style={{ marginTop: 16, marginBottom: 4 }}>
              Forgot your password?
            </Typography.Title>
            <Typography.Paragraph type="secondary" style={{ marginBottom: 24 }}>
              Enter your email address and we&apos;ll send you a link to reset your password.
            </Typography.Paragraph>

            {error ? <Alert type="error" message={error} showIcon style={{ marginBottom: 16 }} /> : null}

            <Form layout="vertical" onFinish={onFinish}>
              <Form.Item
                label="Email"
                name="email"
                rules={[
                  { required: true, message: 'Email is required' },
                  { type: 'email', message: 'Must be a valid email address' },
                ]}
              >
                <Input type="email" autoComplete="email" autoFocus />
              </Form.Item>
              <Button type="primary" htmlType="submit" loading={loading} block>
                Send reset link
              </Button>
            </Form>

            <div style={{ textAlign: 'center', marginTop: 16 }}>
              <a href="/">Back to site</a>
            </div>
          </>
        )}
      </Card>
    </div>
  )
}
