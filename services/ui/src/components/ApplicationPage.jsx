import { Alert, Button, Card, Col, Form, Input, Row, Space, Typography } from 'antd'
import { useEffect, useMemo, useState } from 'react'
import { getPage, submitApplication } from '../api'
import { CMSPageView } from './CMSPageView'

const REQUIRED_CORE_FIELDS = [
  { name: 'organization_name', label: 'Organization Name', type: 'text', required: true, placeholder: '' },
  { name: 'contact_name', label: 'Contact Name', type: 'text', required: true, placeholder: '' },
  { name: 'email', label: 'Email', type: 'email', required: true, placeholder: '' },
]

function normalizeField(field) {
  return {
    name: String(field?.name || '').trim(),
    label: String(field?.label || '').trim(),
    type: ['text', 'email', 'tel', 'url', 'textarea'].includes(field?.type) ? field.type : 'text',
    required: Boolean(field?.required),
    placeholder: String(field?.placeholder || '').trim(),
  }
}

function ensureCoreFields(fields) {
  const byName = new Map((fields || []).map((field) => [field.name, field]))
  REQUIRED_CORE_FIELDS.forEach((field) => {
    if (!byName.has(field.name)) {
      byName.set(field.name, field)
    }
  })

  return Array.from(byName.values()).sort((a, b) => {
    const ai = REQUIRED_CORE_FIELDS.findIndex((field) => field.name === a.name)
    const bi = REQUIRED_CORE_FIELDS.findIndex((field) => field.name === b.name)
    if (ai === -1 && bi === -1) return 0
    if (ai === -1) return 1
    if (bi === -1) return -1
    return ai - bi
  })
}

function buildInitialForm(fields) {
  const initial = {}
  fields.forEach((field) => {
    initial[field.name] = ''
  })
  return initial
}

function renderFieldInput(field, value, onChange) {
  if (field.type === 'textarea') {
    return (
      <Input.TextArea
        rows={4}
        required={field.required}
        value={value}
        placeholder={field.placeholder}
        onChange={(event) => onChange(event.target.value)}
      />
    )
  }

  return (
    <Input
      type={field.type}
      required={field.required}
      value={value}
      placeholder={field.placeholder}
      onChange={(event) => onChange(event.target.value)}
    />
  )
}

export function ApplicationPage({ canEdit, refreshToken = 0 }) {
  const emptyConfig = useMemo(
    () => ({
      intro: '',
      submitLabel: 'Submit Application',
      fields: [],
      configured: false,
    }),
    [],
  )

  const [formConfig, setFormConfig] = useState(emptyConfig)
  const [formValues, setFormValues] = useState({})
  const [status, setStatus] = useState('')
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    getPage('application-contact')
      .then((page) => {
        try {
          const parsed = JSON.parse(page.content_json || '{}')
          const sections = Array.isArray(parsed.sections) ? parsed.sections : []
          const textSection = sections.find((section) => section.type === 'text')
          const formSection = sections.find((section) => section.type === 'application_form')

          if (!formSection) {
            setFormConfig({
              intro: textSection?.body || '',
              submitLabel: 'Submit Application',
              fields: [],
              configured: false,
            })
            setFormValues({})
            return
          }

          const nextFields = ensureCoreFields(
            Array.isArray(formSection.fields)
              ? formSection.fields.map(normalizeField).filter((field) => field.name)
              : [],
          )

          setFormConfig({
            intro: textSection?.body || '',
            submitLabel: formSection.submit_label || 'Submit Application',
            fields: nextFields,
            configured: true,
          })
          setFormValues(buildInitialForm(nextFields))
        } catch {
          setFormConfig(emptyConfig)
          setFormValues({})
        }
      })
      .catch(() => {
        setFormConfig(emptyConfig)
        setFormValues({})
      })
  }, [emptyConfig, refreshToken])

  async function submit(event) {
    event.preventDefault()
    if (!formConfig.configured) {
      setStatus('Application form is not configured yet.')
      return
    }

    setLoading(true)
    setStatus('')

    try {
      const payload = {
        organization_name: formValues.organization_name || '',
        contact_name: formValues.contact_name || '',
        email: formValues.email || '',
        phone: formValues.phone || '',
        website: formValues.website || '',
        monthly_traffic: formValues.monthly_traffic || '',
        current_store: formValues.current_store || '',
        goals: formValues.goals || '',
        notes: formValues.notes || '',
      }

      const extraEntries = Object.entries(formValues).filter(
        ([key, value]) =>
          ![
            'organization_name',
            'contact_name',
            'email',
            'phone',
            'website',
            'monthly_traffic',
            'current_store',
            'goals',
            'notes',
          ].includes(key) &&
          String(value || '').trim(),
      )

      if (extraEntries.length) {
        const extraText = extraEntries.map(([key, value]) => `${key}: ${value}`).join('\n')
        payload.notes = payload.notes ? `${payload.notes}\n\nAdditional Fields:\n${extraText}` : `Additional Fields:\n${extraText}`
      }

      await submitApplication(payload)
      setStatus('Application submitted successfully.')
      setFormValues(buildInitialForm(formConfig.fields))
    } catch (err) {
      setStatus(err.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="application-layout">
      {canEdit ? <CMSPageView slug="application-contact" canEdit={canEdit} /> : null}
      <Card className="page-card">
        <Space direction="vertical" size={16} style={{ width: '100%' }}>
          <div>
            <Typography.Title level={3}>Partner Application Form</Typography.Title>
            {formConfig.intro ? <Typography.Paragraph type="secondary">{formConfig.intro}</Typography.Paragraph> : null}
          </div>

          {!formConfig.configured ? (
            <Alert
              type="info"
              showIcon
              message="Application form is not configured yet."
              description="An editor can add an Application Form Fields section on this page to publish the form."
            />
          ) : (
            <Form layout="vertical" onSubmitCapture={submit} className="lead-form">
              <Row gutter={[16, 0]}>
                {formConfig.fields.map((field) => (
                  <Col key={field.name} xs={24} md={field.type === 'textarea' ? 24 : 12}>
                    <Form.Item label={field.label || field.name} required={field.required}>
                      {renderFieldInput(field, formValues[field.name] || '', (nextValue) =>
                        setFormValues((prev) => ({
                          ...prev,
                          [field.name]: nextValue,
                        }))
                      )}
                    </Form.Item>
                  </Col>
                ))}
              </Row>

              <Button type="primary" htmlType="submit" loading={loading}>
                {formConfig.submitLabel || 'Submit Application'}
              </Button>
              {status ? (
                <Alert
                  style={{ marginTop: 12 }}
                  type={status.toLowerCase().includes('success') ? 'success' : 'error'}
                  message={status}
                  showIcon
                />
              ) : null}
            </Form>
          )}
        </Space>
      </Card>
    </div>
  )
}
