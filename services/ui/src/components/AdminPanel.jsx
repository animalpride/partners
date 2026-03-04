import {
  Alert, Button, Card, Collapse, Dropdown, Form, Input, Layout, Menu,
  Select, Space, Spin, Tag, Typography,
} from 'antd'
import { useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { getPage, updatePage } from '../api'
import {
  CMS_PAGES, FIELD_TYPE_OPTIONS, IMAGE_POSITION_OPTIONS, SECTION_ALIGNMENT_OPTIONS,
  SECTION_BACKGROUND_OPTIONS, SECTION_TYPES,
  makeDefaultSection, normalizeField, normalizeGridItem,
  normalizeSections, toContentJSON,
} from '../cmsTypes'
import { CMSContentRenderer } from './CMSPageView'

// ── Section mutation helpers ───────────────────────────────────────────────────
function mutateSectionAt(prev, index, fn) {
  return prev.map((s, i) => (i === index ? fn(s) : s))
}

// ── Per-section editor panel ───────────────────────────────────────────────────
function SectionEditor({ section, index, total, onChange, onRemove, onMoveUp, onMoveDown }) {
  function update(patch) { onChange(index, patch) }

  return (
    <div className="admin-section-editor">
      {/* Controls row */}
      <div className="admin-section-controls">
        <Space size={4}>
          <Button size="small" onClick={() => onMoveUp(index)} disabled={index === 0}>↑</Button>
          <Button size="small" onClick={() => onMoveDown(index)} disabled={index === total - 1}>↓</Button>
        </Space>
        <Button size="small" danger onClick={() => onRemove(index)}>Remove</Button>
      </div>

      <Form layout="vertical" size="small">
        <div className="admin-form-row">
          <Form.Item label="Type" style={{ flex: 1, marginBottom: 8 }}>
            <Select
              value={section.type}
              options={SECTION_TYPES}
              onChange={(v) => update(makeDefaultSection(v))}
            />
          </Form.Item>
          <Form.Item label="Background" style={{ flex: 1, marginBottom: 8 }}>
            <Select
              value={section.background || ''}
              options={SECTION_BACKGROUND_OPTIONS}
              onChange={(v) => update({ background: v })}
            />
          </Form.Item>
          <Form.Item label="Alignment" style={{ flex: 1, marginBottom: 8 }}>
            <Select
              value={section.alignment || 'left'}
              options={SECTION_ALIGNMENT_OPTIONS}
              onChange={(v) => update({ alignment: v })}
            />
          </Form.Item>
        </div>

        {section.type === 'image' ? (
          <Form.Item label="Image Focus" style={{ marginBottom: 8 }}>
            <Select
              value={section.image_position || 'center'}
              options={IMAGE_POSITION_OPTIONS}
              onChange={(v) => update({ image_position: v })}
            />
          </Form.Item>
        ) : null}

        <Form.Item label="Heading" style={{ marginBottom: 8 }}>
          <Input value={section.heading || ''} onChange={(e) => update({ heading: e.target.value })} />
        </Form.Item>

        {/* ── Text ── */}
        {section.type === 'text' ? (
          <Form.Item label="Body" style={{ marginBottom: 8 }}>
            <Input.TextArea rows={4} value={section.body || ''} onChange={(e) => update({ body: e.target.value })} />
          </Form.Item>
        ) : null}

        {/* ── Bullets ── */}
        {section.type === 'bullets' ? (
          <Form.Item label="Bullet points (one per line)" style={{ marginBottom: 8 }}>
            <Input.TextArea
              rows={5}
              value={(section.items || []).join('\n')}
              onChange={(e) =>
                update({ items: e.target.value.split('\n').map((l) => l.trim()).filter(Boolean) })
              }
            />
          </Form.Item>
        ) : null}

        {/* ── Buttons ── */}
        {section.type === 'buttons' ? (
          <ButtonsEditor section={section} sectionIndex={index} onChange={onChange} />
        ) : null}

        {/* ── Form CTA ── */}
        {section.type === 'form_cta' ? (
          <>
            <Form.Item label="Description" style={{ marginBottom: 8 }}>
              <Input.TextArea rows={3} value={section.body || ''} onChange={(e) => update({ body: e.target.value })} />
            </Form.Item>
            <div className="admin-form-row">
              <Form.Item label="Button Label" style={{ flex: 1, marginBottom: 8 }}>
                <Input value={section.button_label || ''} onChange={(e) => update({ button_label: e.target.value })} />
              </Form.Item>
              <Form.Item label="Button Link" style={{ flex: 1, marginBottom: 8 }}>
                <Input value={section.button_link || ''} onChange={(e) => update({ button_link: e.target.value })} />
              </Form.Item>
            </div>
          </>
        ) : null}

        {/* ── Application Form ── */}
        {section.type === 'application_form' ? (
          <ApplicationFormEditor section={section} sectionIndex={index} onChange={onChange} />
        ) : null}

        {/* ── Image Block ── */}
        {section.type === 'image' ? (
          <>
            <Form.Item label="Image URL" style={{ marginBottom: 8 }}>
              <Input value={section.image_url || ''} onChange={(e) => update({ image_url: e.target.value })} />
            </Form.Item>
            <Form.Item label="Alt Text" style={{ marginBottom: 8 }}>
              <Input value={section.image_alt || ''} onChange={(e) => update({ image_alt: e.target.value })} />
            </Form.Item>
            <Form.Item label="Caption / Body" style={{ marginBottom: 8 }}>
              <Input.TextArea rows={2} value={section.body || ''} onChange={(e) => update({ body: e.target.value })} />
            </Form.Item>
            <div className="admin-form-row">
              <Form.Item label="Button Label" style={{ flex: 1, marginBottom: 8 }}>
                <Input value={section.button_label || ''} onChange={(e) => update({ button_label: e.target.value })} />
              </Form.Item>
              <Form.Item label="Button Link" style={{ flex: 1, marginBottom: 8 }}>
                <Input value={section.button_link || ''} onChange={(e) => update({ button_link: e.target.value })} />
              </Form.Item>
            </div>
          </>
        ) : null}

        {/* ── Image Grid ── */}
        {section.type === 'image_grid' ? (
          <ImageGridEditor section={section} sectionIndex={index} onChange={onChange} />
        ) : null}
      </Form>
    </div>
  )
}

function ButtonsEditor({ section, sectionIndex, onChange }) {
  function addButton() {
    onChange(sectionIndex, { buttons: [...(section.buttons || []), { label: '', url: '', variant: 'default' }] })
  }
  function updateButton(bi, patch) {
    onChange(sectionIndex, {
      buttons: (section.buttons || []).map((b, i) => (i === bi ? { ...b, ...patch } : b)),
    })
  }
  function removeButton(bi) {
    onChange(sectionIndex, { buttons: (section.buttons || []).filter((_, i) => i !== bi) })
  }

  return (
    <>
      {(section.buttons || []).map((btn, bi) => (
        <Card
          key={bi}
          size="small"
          className="admin-sub-card"
          title={`Button ${bi + 1}`}
          extra={<Button size="small" danger type="text" onClick={() => removeButton(bi)}>Remove</Button>}
        >
          <div className="admin-form-row">
            <Form.Item label="Label" style={{ flex: 1, marginBottom: 6 }}>
              <Input size="small" value={btn.label || ''} onChange={(e) => updateButton(bi, { label: e.target.value })} />
            </Form.Item>
            <Form.Item label="URL" style={{ flex: 2, marginBottom: 6 }}>
              <Input size="small" value={btn.url || ''} onChange={(e) => updateButton(bi, { url: e.target.value })} />
            </Form.Item>
            <Form.Item label="Style" style={{ flex: 1, marginBottom: 6 }}>
              <Select
                size="small"
                value={btn.variant || 'default'}
                options={[{ label: 'Primary', value: 'primary' }, { label: 'Ghost', value: 'default' }]}
                onChange={(v) => updateButton(bi, { variant: v })}
              />
            </Form.Item>
          </div>
        </Card>
      ))}
      <Button size="small" style={{ marginTop: 4 }} onClick={addButton}>+ Add Button</Button>
    </>
  )
}

function ApplicationFormEditor({ section, sectionIndex, onChange }) {
  function addField() {
    onChange(sectionIndex, {
      fields: [...(section.fields || []), { name: '', label: '', type: 'text', required: false, placeholder: '' }],
    })
  }
  function updateField(fi, patch) {
    onChange(sectionIndex, {
      fields: (section.fields || []).map((f, i) => (i === fi ? normalizeField({ ...f, ...patch }) : f)),
    })
  }
  function removeField(fi) {
    onChange(sectionIndex, { fields: (section.fields || []).filter((_, i) => i !== fi) })
  }

  return (
    <>
      <Form.Item label="Submit Button Label" style={{ marginBottom: 8 }}>
        <Input
          size="small"
          value={section.submit_label || 'Submit Application'}
          onChange={(e) => onChange(sectionIndex, { submit_label: e.target.value })}
        />
      </Form.Item>
      {(section.fields || []).map((field, fi) => (
        <Card
          key={fi}
          size="small"
          className="admin-sub-card"
          title={field.label || `Field ${fi + 1}`}
          extra={<Button size="small" danger type="text" onClick={() => removeField(fi)}>Remove</Button>}
        >
          <div className="admin-form-row">
            <Form.Item label="Key" style={{ flex: 1, marginBottom: 6 }}>
              <Input size="small" value={field.name || ''} onChange={(e) => updateField(fi, { name: e.target.value })} />
            </Form.Item>
            <Form.Item label="Label" style={{ flex: 1, marginBottom: 6 }}>
              <Input size="small" value={field.label || ''} onChange={(e) => updateField(fi, { label: e.target.value })} />
            </Form.Item>
            <Form.Item label="Type" style={{ flex: 1, marginBottom: 6 }}>
              <Select size="small" value={field.type || 'text'} options={FIELD_TYPE_OPTIONS} onChange={(v) => updateField(fi, { type: v })} />
            </Form.Item>
            <Form.Item label="Required" style={{ flex: 1, marginBottom: 6 }}>
              <Select
                size="small"
                value={field.required ? 'true' : 'false'}
                options={[{ label: 'Required', value: 'true' }, { label: 'Optional', value: 'false' }]}
                onChange={(v) => updateField(fi, { required: v === 'true' })}
              />
            </Form.Item>
          </div>
          <Form.Item label="Placeholder" style={{ marginBottom: 0 }}>
            <Input size="small" value={field.placeholder || ''} onChange={(e) => updateField(fi, { placeholder: e.target.value })} />
          </Form.Item>
        </Card>
      ))}
      <Button size="small" style={{ marginTop: 4 }} onClick={addField}>+ Add Field</Button>
    </>
  )
}

function ImageGridEditor({ section, sectionIndex, onChange }) {
  function addItem() {
    onChange(sectionIndex, { items: [...(section.items || []), { title: '', image_url: '', link_url: '' }] })
  }
  function updateItem(ii, patch) {
    onChange(sectionIndex, {
      items: (section.items || []).map((item, i) => (i === ii ? normalizeGridItem({ ...item, ...patch }) : item)),
    })
  }
  function removeItem(ii) {
    onChange(sectionIndex, { items: (section.items || []).filter((_, i) => i !== ii) })
  }

  return (
    <>
      {(section.items || []).map((item, ii) => (
        <Card
          key={ii}
          size="small"
          className="admin-sub-card"
          title={item.title || `Grid Item ${ii + 1}`}
          extra={<Button size="small" danger type="text" onClick={() => removeItem(ii)}>Remove</Button>}
        >
          <div className="admin-form-row">
            <Form.Item label="Title" style={{ flex: 1, marginBottom: 6 }}>
              <Input size="small" value={item.title || ''} onChange={(e) => updateItem(ii, { title: e.target.value })} />
            </Form.Item>
            <Form.Item label="Image URL" style={{ flex: 2, marginBottom: 6 }}>
              <Input size="small" value={item.image_url || ''} onChange={(e) => updateItem(ii, { image_url: e.target.value })} />
            </Form.Item>
            <Form.Item label="Link URL" style={{ flex: 2, marginBottom: 6 }}>
              <Input size="small" value={item.link_url || ''} onChange={(e) => updateItem(ii, { link_url: e.target.value })} />
            </Form.Item>
          </div>
        </Card>
      ))}
      <Button size="small" style={{ marginTop: 4 }} onClick={addItem}>+ Add Grid Item</Button>
    </>
  )
}

// ── Main AdminPanel ────────────────────────────────────────────────────────────
export function AdminPanel() {
  const navigate = useNavigate()
  const [adminSection, setAdminSection] = useState('content')

  // Content Editor state
  const [selectedSlug, setSelectedSlug] = useState(CMS_PAGES[0].slug)
  const [page, setPage] = useState(null)
  const [loadError, setLoadError] = useState('')
  const [isSaving, setIsSaving] = useState(false)
  const [saveError, setSaveError] = useState('')
  const [saveSuccess, setSaveSuccess] = useState(false)
  const [draft, setDraft] = useState({ title: '', description: '', content_json: '{}' })
  const [sectionsDraft, setSectionsDraft] = useState([])
  const [isDirty, setIsDirty] = useState(false)

  // Load page on slug change
  useEffect(() => {
    let mounted = true
    setPage(null)
    setLoadError('')
    setSaveError('')
    setSaveSuccess(false)
    setIsDirty(false)
    getPage(selectedSlug)
      .then((data) => {
        if (!mounted) return
        setPage(data)
        const d = { title: data.title, description: data.description || '', content_json: data.content_json || '{}' }
        setDraft(d)
        try { setSectionsDraft(normalizeSections(JSON.parse(d.content_json))) }
        catch { setSectionsDraft([]) }
      })
      .catch((err) => { if (mounted) setLoadError(err.message) })
    return () => { mounted = false }
  }, [selectedSlug])

  // Preview sections live
  const previewSections = useMemo(() => sectionsDraft, [sectionsDraft])

  function markDirty() { setIsDirty(true); setSaveSuccess(false) }

  function updateSection(index, patch) {
    setSectionsDraft((prev) => mutateSectionAt(prev, index, (s) => ({ ...s, ...patch })))
    markDirty()
  }
  function removeSection(index) {
    setSectionsDraft((prev) => prev.filter((_, i) => i !== index))
    markDirty()
  }
  function moveSectionUp(index) {
    if (index <= 0) return
    setSectionsDraft((prev) => { const a = [...prev]; [a[index - 1], a[index]] = [a[index], a[index - 1]]; return a })
    markDirty()
  }
  function moveSectionDown(index) {
    setSectionsDraft((prev) => {
      if (index >= prev.length - 1) return prev
      const a = [...prev]; [a[index], a[index + 1]] = [a[index + 1], a[index]]; return a
    })
    markDirty()
  }
  function addSection(type) {
    setSectionsDraft((prev) => [...prev, makeDefaultSection(type)])
    markDirty()
  }
  function discardChanges() {
    if (!page) return
    const d = { title: page.title, description: page.description || '', content_json: page.content_json || '{}' }
    setDraft(d)
    try { setSectionsDraft(normalizeSections(JSON.parse(d.content_json))) }
    catch { setSectionsDraft([]) }
    setIsDirty(false)
    setSaveError('')
    setSaveSuccess(false)
  }

  async function saveChanges() {
    if (!page) return
    setIsSaving(true)
    setSaveError('')
    try {
      const payload = { ...draft, content_json: toContentJSON(sectionsDraft) }
      const updated = await updatePage(selectedSlug, payload)
      setPage(updated)
      setIsDirty(false)
      setSaveSuccess(true)
    } catch (err) {
      setSaveError(err.message)
    } finally {
      setIsSaving(false)
    }
  }

  // Add-section dropdown items
  const addSectionItems = {
    items: SECTION_TYPES.map((t) => ({ key: t.value, label: t.label })),
    onClick: ({ key }) => addSection(key),
  }

  // Collapse items for section list
  const collapseItems = sectionsDraft.map((section, index) => ({
    key: String(index),
    label: (
      <span className="admin-section-label">
        <Tag color="blue" style={{ marginRight: 6, fontSize: 11 }}>
          {SECTION_TYPES.find((t) => t.value === section.type)?.label || section.type}
        </Tag>
        {section.heading || `Section ${index + 1}`}
      </span>
    ),
    children: (
      <SectionEditor
        section={section}
        index={index}
        total={sectionsDraft.length}
        onChange={updateSection}
        onRemove={removeSection}
        onMoveUp={moveSectionUp}
        onMoveDown={moveSectionDown}
      />
    ),
  }))

  const selectedPageLabel = CMS_PAGES.find((p) => p.slug === selectedSlug)?.label || selectedSlug

  return (
    <Layout className="admin-layout">
      {/* ── Admin header bar ── */}
      <Layout.Header className="admin-header">
        <div className="admin-header-inner">
          <Button type="text" className="admin-back-btn" onClick={() => navigate('/')}>
            ← Back to Site
          </Button>
          <Menu
            mode="horizontal"
            selectedKeys={[adminSection]}
            className="admin-nav"
            onClick={({ key }) => setAdminSection(key)}
            items={[
              { key: 'content', label: 'Content Editor' },
              { key: 'portals', label: 'Partner Portals' },
              { key: 'users', label: 'Users & Roles' },
            ]}
          />
          <div style={{ width: 160 }} />
        </div>
      </Layout.Header>

      {/* ── Main split area ── */}
      <Layout.Content className="admin-body">
        {adminSection === 'content' ? (
          <div className="admin-split">
            {/* ── LEFT: Editor sidebar ── */}
            <aside className="admin-sidebar">
              {/* Page selector */}
              <div className="admin-sidebar-section">
                <div className="admin-sidebar-label">Page</div>
                <div className="admin-page-list">
                  {CMS_PAGES.map((p) => (
                    <button
                      key={p.slug}
                      type="button"
                      className={`admin-page-btn${selectedSlug === p.slug ? ' admin-page-btn--active' : ''}`}
                      onClick={() => {
                        if (isDirty && !window.confirm('Discard unsaved changes?')) return
                        setSelectedSlug(p.slug)
                      }}
                    >
                      {p.label}
                    </button>
                  ))}
                </div>
              </div>

              {/* Section list */}
              <div className="admin-sidebar-section admin-sidebar-section--grow">
                <div className="admin-sidebar-label" style={{ marginBottom: 8 }}>
                  Sections
                  {isDirty ? <Tag color="gold" style={{ marginLeft: 8, fontSize: 10 }}>Unsaved</Tag> : null}
                </div>

                {loadError ? <Alert type="error" message={loadError} showIcon /> : null}

                {!page && !loadError ? (
                  <div style={{ textAlign: 'center', padding: 24 }}><Spin /></div>
                ) : (
                  <>
                    {/* Page meta fields */}
                    <Form layout="vertical" size="small" style={{ marginBottom: 12 }}>
                      <Form.Item label="Page Title" style={{ marginBottom: 6 }}>
                        <Input
                          value={draft.title}
                          onChange={(e) => { setDraft((d) => ({ ...d, title: e.target.value })); markDirty() }}
                        />
                      </Form.Item>
                      <Form.Item label="Description" style={{ marginBottom: 0 }}>
                        <Input.TextArea
                          rows={2}
                          value={draft.description}
                          onChange={(e) => { setDraft((d) => ({ ...d, description: e.target.value })); markDirty() }}
                        />
                      </Form.Item>
                    </Form>

                    {collapseItems.length > 0 ? (
                      <Collapse
                        accordion
                        size="small"
                        className="admin-sections-collapse"
                        items={collapseItems}
                      />
                    ) : (
                      <div className="admin-empty-sections">No sections yet. Add one below.</div>
                    )}

                    <Dropdown menu={addSectionItems} trigger={['click']}>
                      <Button block style={{ marginTop: 10 }}>+ Add Section ▾</Button>
                    </Dropdown>
                  </>
                )}
              </div>

              {/* Save / Discard */}
              <div className="admin-sidebar-footer">
                {saveError ? <Alert type="error" message={saveError} showIcon style={{ marginBottom: 8 }} /> : null}
                {saveSuccess ? <Alert type="success" message="Saved successfully!" showIcon style={{ marginBottom: 8 }} /> : null}
                <Space style={{ width: '100%' }}>
                  <Button
                    type="primary"
                    block
                    loading={isSaving}
                    disabled={!isDirty}
                    onClick={saveChanges}
                    style={{ flex: 1 }}
                  >
                    Save Changes
                  </Button>
                  <Button block disabled={!isDirty || isSaving} onClick={discardChanges} style={{ flex: 1 }}>
                    Discard
                  </Button>
                </Space>
              </div>
            </aside>

            {/* ── RIGHT: Live preview ── */}
            <div className="admin-preview">
              <div className="admin-preview-bar">
                <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                  Live preview — {selectedPageLabel}
                  {isDirty ? <Tag color="gold" style={{ marginLeft: 8, fontSize: 10 }}>Unsaved changes</Tag> : null}
                </Typography.Text>
              </div>
              <div className="admin-preview-content">
                <div className="admin-preview-inner app-shell">
                  <div className="cms-page-header" style={{ paddingLeft: 0, paddingRight: 0 }}>
                    <Typography.Title level={2} className="cms-page-title">{draft.title}</Typography.Title>
                    {draft.description ? (
                      <Typography.Paragraph className="cms-page-desc">{draft.description}</Typography.Paragraph>
                    ) : null}
                  </div>
                  <CMSContentRenderer sections={previewSections} />
                </div>
              </div>
            </div>
          </div>
        ) : null}

        {adminSection === 'portals' ? (
          <div className="admin-stub">
            <Typography.Title level={3}>Partner Portals</Typography.Title>
            <Typography.Paragraph type="secondary">
              Manage partner portal configurations, custom branding, and access settings.
              This section is under construction.
            </Typography.Paragraph>
          </div>
        ) : null}

        {adminSection === 'users' ? (
          <div className="admin-stub">
            <Typography.Title level={3}>Users & Roles</Typography.Title>
            <Typography.Paragraph type="secondary">
              Manage user accounts, assign roles, and configure permissions.
              This section is under construction.
            </Typography.Paragraph>
          </div>
        ) : null}
      </Layout.Content>
    </Layout>
  )
}
