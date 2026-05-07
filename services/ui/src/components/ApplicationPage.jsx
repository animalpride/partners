import { Alert, AutoComplete, Button, Card, Col, Form, Input, Row, Select, Space, Typography } from 'antd'
import { useEffect, useMemo, useRef, useState } from 'react'
import { getLocationCityStates, getLocationCountries, getPage, submitApplication } from '../api'
import { CMSPageView } from './CMSPageView'

const PHONE_REGEX = /^\+?[\d\s\-(). ]{7,20}$/

const REQUIRED_CORE_FIELDS = [
  { name: 'organization_name', label: 'Organization Name', type: 'text', required: true, placeholder: '' },
  { name: 'contact_name', label: 'Contact Name', type: 'text', required: true, placeholder: '' },
  { name: 'email', label: 'Email', type: 'email', required: true, placeholder: '' },
  { name: 'phone', label: 'Phone Number', type: 'tel', required: true, placeholder: '' },
  { name: 'address_line1', label: 'Address Line 1', type: 'text', required: true, placeholder: '', col: 24 },
  { name: 'address_line2', label: 'Address Line 2', type: 'text', required: false, placeholder: '', col: 24 },
  { name: 'country', label: 'Country', type: 'text', required: true, placeholder: '' },
  { name: 'city_state', label: 'City / State', type: 'text', required: true, placeholder: 'Start typing a city...', col: 24 },
  { name: 'postal_code', label: 'Postal Code', type: 'text', required: true, placeholder: '' },
]

const KNOWN_FIELD_NAMES = new Set([
  'organization_name', 'contact_name', 'email', 'phone',
  'address_line1', 'address_line2', 'city_state', 'city', 'state', 'state_code', 'postal_code',
  'country', 'country_code', 'city_lookup_id',
  'website', 'monthly_traffic', 'current_store', 'goals', 'notes',
])

const HIDDEN_LOCATION_FIELDS = new Set(['city', 'state', 'state_code', 'country_code', 'city_lookup_id'])

function normalizeField(field) {
  return {
    name: String(field?.name || '').trim(),
    label: String(field?.label || '').trim(),
    type: ['text', 'email', 'tel', 'url', 'textarea'].includes(field?.type) ? field.type : 'text',
    required: Boolean(field?.required),
    placeholder: String(field?.placeholder || '').trim(),
    col: field?.col ? Number(field.col) : undefined,
  }
}

function ensureCoreFields(fields) {
  const byName = new Map((fields || []).filter((field) => !HIDDEN_LOCATION_FIELDS.has(field.name)).map((field) => [field.name, field]))
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

function buildInitialForm(fields, defaults = {}) {
  const initial = {}
  fields.forEach((field) => {
    initial[field.name] = ''
  })
  return {
    ...initial,
    ...defaults,
  }
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
  const [errors, setErrors] = useState({})
  const [status, setStatus] = useState('')
  const [loading, setLoading] = useState(false)
  const [countries, setCountries] = useState([])
  const [cityOptions, setCityOptions] = useState([])
  const [cityLoading, setCityLoading] = useState(false)
  const citySearchTimerRef = useRef(null)
  const citySearchNonceRef = useRef(0)

  useEffect(() => {
    return () => {
      if (citySearchTimerRef.current) {
        clearTimeout(citySearchTimerRef.current)
      }
    }
  }, [])

  useEffect(() => {
    getLocationCountries()
      .then((rows) => {
        setCountries(Array.isArray(rows) ? rows : [])
        const us = (Array.isArray(rows) ? rows : []).find((row) => row.code === 'US')
        const fallback = us || (Array.isArray(rows) ? rows[0] : null)
        if (!fallback) return
        setFormValues((prev) => ({
          ...prev,
          country: prev.country || fallback.name,
          country_code: prev.country_code || fallback.code,
        }))
      })
      .catch(() => {
        setCountries([])
      })
  }, [])

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
          setFormValues(buildInitialForm(nextFields, { country_code: 'US' }))
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

    // Client-side validation
    const newErrors = {}
    formConfig.fields.forEach((field) => {
      if (field.required && !String(formValues[field.name] || '').trim()) {
        newErrors[field.name] = `${field.label || field.name} is required`
      }
    })
    const phoneVal = String(formValues.phone || '').trim()
    if (phoneVal && !PHONE_REGEX.test(phoneVal)) {
      newErrors.phone = 'Please enter a valid phone number (e.g. +1 555-123-4567)'
    }
    if (!String(formValues.city_lookup_id || '').trim()) {
      newErrors.city_state = 'Please choose a city from the suggestions'
    }
    if (Object.keys(newErrors).length > 0) {
      setErrors(newErrors)
      return
    }
    setErrors({})

    setLoading(true)
    setStatus('')

    try {
      const payload = {
        organization_name: formValues.organization_name || '',
        contact_name: formValues.contact_name || '',
        email: formValues.email || '',
        phone: formValues.phone || '',
        address_line1: formValues.address_line1 || '',
        address_line2: formValues.address_line2 || '',
        city: formValues.city || '',
        city_state: formValues.city_state || '',
        city_lookup_id: Number(formValues.city_lookup_id || 0),
        state: formValues.state || '',
        state_code: formValues.state_code || '',
        postal_code: formValues.postal_code || '',
        country: formValues.country || '',
        country_code: formValues.country_code || '',
        website: formValues.website || '',
        monthly_traffic: formValues.monthly_traffic || '',
        current_store: formValues.current_store || '',
        goals: formValues.goals || '',
        notes: formValues.notes || '',
      }

      const extraEntries = Object.entries(formValues).filter(
        ([key, value]) =>
          !KNOWN_FIELD_NAMES.has(key) &&
          String(value || '').trim(),
      )

      if (extraEntries.length) {
        const extraText = extraEntries.map(([key, value]) => `${key}: ${value}`).join('\n')
        payload.notes = payload.notes ? `${payload.notes}\n\nAdditional Fields:\n${extraText}` : `Additional Fields:\n${extraText}`
      }

      await submitApplication(payload)
      setStatus('Application submitted successfully.')
      setFormValues(buildInitialForm(formConfig.fields, {
        country: formValues.country,
        country_code: formValues.country_code,
      }))
      setCityOptions([])
    } catch (err) {
      setStatus(err.message)
    } finally {
      setLoading(false)
    }
  }

  function onCountryChange(nextCountryCode) {
    const nextCountry = countries.find((item) => item.code === nextCountryCode)
    setFormValues((prev) => ({
      ...prev,
      country_code: nextCountryCode,
      country: nextCountry?.name || prev.country,
      city_state: '',
      city_lookup_id: '',
      city: '',
      state: '',
      state_code: '',
    }))
    setCityOptions([])
  }

  function onCitySearch(query) {
    const trimmed = String(query || '').trim()
    const countryCode = String(formValues.country_code || '').trim()

    if (!countryCode || trimmed.length < 2) {
      setCityOptions([])
      return
    }

    if (citySearchTimerRef.current) {
      clearTimeout(citySearchTimerRef.current)
    }

    const nonce = citySearchNonceRef.current + 1
    citySearchNonceRef.current = nonce

    citySearchTimerRef.current = setTimeout(() => {
      setCityLoading(true)
      getLocationCityStates(countryCode, trimmed, 25)
        .then((rows) => {
          if (nonce !== citySearchNonceRef.current) return
          const options = (Array.isArray(rows) ? rows : []).map((row) => ({
            value: row.label,
            city: row.city,
            state: row.state,
            state_code: row.state_code,
            country: row.country,
            country_code: row.country_code,
            city_lookup_id: row.city_lookup_id,
          }))
          setCityOptions(options)
        })
        .catch(() => {
          if (nonce !== citySearchNonceRef.current) return
          setCityOptions([])
        })
        .finally(() => {
          if (nonce === citySearchNonceRef.current) {
            setCityLoading(false)
          }
        })
    }, 250)
  }

  function onCityInputChange(query) {
    setFormValues((prev) => ({
      ...prev,
      city_state: query,
      city_lookup_id: '',
      city: '',
      state: '',
      state_code: '',
    }))
  }

  function onCitySelect(_value, option) {
    setFormValues((prev) => ({
      ...prev,
      city_state: option.value,
      city_lookup_id: option.city_lookup_id,
      city: option.city,
      state: option.state,
      state_code: option.state_code,
      country: option.country,
      country_code: option.country_code,
    }))
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
                  <Col key={field.name} xs={24} md={field.col || (field.type === 'textarea' ? 24 : 12)}>
                    <Form.Item
                      label={field.label || field.name}
                      required={field.required}
                      validateStatus={errors[field.name] ? 'error' : ''}
                      help={errors[field.name] || undefined}
                    >
                      {field.name === 'country' ? (
                        <Select
                          showSearch
                          placeholder="Select a country"
                          value={formValues.country_code || undefined}
                          options={countries.map((item) => ({
                            value: item.code,
                            label: `${item.name} (${item.code})`,
                          }))}
                          optionFilterProp="label"
                          onChange={onCountryChange}
                        />
                      ) : field.name === 'city_state' ? (
                        <AutoComplete
                          value={formValues.city_state || ''}
                          options={cityOptions}
                          placeholder={field.placeholder || 'Start typing a city'}
                          onSearch={onCitySearch}
                          onSelect={onCitySelect}
                          onChange={onCityInputChange}
                        />
                      ) : (
                        renderFieldInput(field, formValues[field.name] || '', (nextValue) =>
                          setFormValues((prev) => ({
                            ...prev,
                            [field.name]: nextValue,
                          }))
                        )
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
