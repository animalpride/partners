import { WarningOutlined } from '@ant-design/icons'
import {
  Alert, Button, Card, Collapse, Divider, Dropdown, Form, Grid, Input, Layout, Menu,
  Modal, Select, Slider, Space, Spin, Switch, Tag, Tooltip, Typography,
} from 'antd'
import { useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { getPage, updatePage } from '../api'
import {
  CMS_PAGES, CONTAINER_SIZE_OPTIONS, CONTENT_WITH_IMAGE_SUPPORTED_TYPES, FIELD_TYPE_OPTIONS, IMAGE_ASPECT_RATIO_OPTIONS, IMAGE_POSITION_OPTIONS,
  SECTION_ALIGNMENT_OPTIONS, SECTION_BACKGROUND_OPTIONS, SECTION_TYPES,
  TWO_COLUMN_IMAGE_SIDE_OPTIONS, TWO_COLUMN_VARIANT_OPTIONS,
  makeDefaultSection, normalizeComparisonRow, normalizeContent, normalizeField, normalizeGridItem, normalizeIconItem,
  normalizeMatrixColumn, normalizeMatrixRow,
  normalizeSections, toContentJSON,
} from '../cmsTypes'
import { CMSContentRenderer } from './CMSPageView'
import { UsersRoles } from './UsersRoles'

// ── Section mutation helpers ───────────────────────────────────────────────────
function mutateSectionAt(prev, index, fn) {
  return prev.map((s, i) => (i === index ? fn(s) : s))
}

const CONTAINER_STEPS = CONTAINER_SIZE_OPTIONS.map((o) => o.value)
const CONTAINER_STEP_INDEX = Object.fromEntries(CONTAINER_STEPS.map((v, i) => [v, i]))
const CONTAINER_MARKS = Object.fromEntries(CONTAINER_SIZE_OPTIONS.map((o, i) => [i, o.label]))

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
        <Form.Item label="Type" style={{ marginBottom: 8 }}>
          <Select
            value={section.type}
            options={SECTION_TYPES}
            onChange={(v) => update(makeDefaultSection(v))}
          />
        </Form.Item>
        <div className="admin-form-row">
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

        <Form.Item label="Section Width" style={{ marginBottom: 24 }}>
          <Slider
            min={0}
            max={3}
            step={1}
            marks={CONTAINER_MARKS}
            value={CONTAINER_STEP_INDEX[section.container_size] ?? 0}
            onChange={(v) => update({ container_size: CONTAINER_STEPS[v] })}
            tooltip={{ formatter: (v) => CONTAINER_MARKS[v] }}
          />
        </Form.Item>

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
            <Input.TextArea autoSize={{ minRows: 3, maxRows: 12 }} value={section.body || ''} onChange={(e) => update({ body: e.target.value })} />
          </Form.Item>
        ) : null}

        {/* ── Bullets ── */}
        {section.type === 'bullets' ? (
          <Form.Item label="Bullet points" style={{ marginBottom: 8 }}>
            <BulletsEditor
              items={section.items || []}
              onChange={(newItems) => update({ items: newItems })}
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
              <Input.TextArea autoSize={{ minRows: 2, maxRows: 8 }} value={section.body || ''} onChange={(e) => update({ body: e.target.value })} />
            </Form.Item>
            <Form.Item label="Buttons" style={{ marginBottom: 0 }}>
              <ButtonsEditor section={section} sectionIndex={index} onChange={onChange} />
            </Form.Item>
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
            <Form.Item label="Image Label (placeholder text, shown when no URL)" style={{ marginBottom: 8 }}>
              <Input value={section.image_label || ''} onChange={(e) => update({ image_label: e.target.value })} placeholder="Describes the intended image" />
            </Form.Item>
            <Form.Item label="Alt Text" style={{ marginBottom: 8 }}>
              <Input value={section.image_alt || ''} onChange={(e) => update({ image_alt: e.target.value })} />
            </Form.Item>
            <Form.Item label="Caption / Body" style={{ marginBottom: 8 }}>
              <Input.TextArea autoSize={{ minRows: 2, maxRows: 6 }} value={section.body || ''} onChange={(e) => update({ body: e.target.value })} />
            </Form.Item>
            <Form.Item label="Buttons" style={{ marginBottom: 0 }}>
              <ButtonsEditor section={section} sectionIndex={index} onChange={onChange} />
            </Form.Item>
          </>
        ) : null}

        {/* ── Image Grid ── */}
        {section.type === 'image_grid' ? (
          <ImageGridEditor section={section} sectionIndex={index} onChange={onChange} />
        ) : null}

        {/* ── Two Column ── */}
        {section.type === 'two_column' ? (
          <TwoColumnEditor section={section} sectionIndex={index} onChange={onChange} />
        ) : null}

        {/* ── Comparison Table ── */}
        {section.type === 'comparison_table' ? (
          <ComparisonTableEditor section={section} sectionIndex={index} onChange={onChange} />
        ) : null}

        {/* ── Icon Grid ── */}
        {section.type === 'icon_grid' ? (
          <IconGridEditor section={section} sectionIndex={index} onChange={onChange} />
        ) : null}

        {/* ── Matrix / Pivot Table ── */}
        {section.type === 'matrix_table' ? (
          <MatrixTableEditor section={section} sectionIndex={index} onChange={onChange} />
        ) : null}

        {/* ── Content + Image (paired) ── */}
        {section.type === 'content_with_image' ? (
          <ContentWithImageEditor section={section} sectionIndex={index} onChange={onChange} />
        ) : null}
      </Form>
    </div>
  )
}

function ButtonsEditor({ section, sectionIndex, onChange }) {
  function addButton() {
    onChange(sectionIndex, { buttons: [...(section.buttons || []), { label: '', url: '', variant: 'primary', visibility: 'both' }] })
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
                value={btn.variant || 'primary'}
                options={[{ label: 'Primary', value: 'primary' }, { label: 'Ghost', value: 'default' }]}
                onChange={(v) => updateButton(bi, { variant: v })}
              />
            </Form.Item>
            <Form.Item label="Show On" style={{ flex: 1, marginBottom: 6 }}>
              <Select
                size="small"
                value={btn.visibility || 'both'}
                options={[
                  { label: 'Both', value: 'both' },
                  { label: 'Desktop Only', value: 'desktop' },
                  { label: 'Mobile Only', value: 'mobile' },
                ]}
                onChange={(v) => updateButton(bi, { visibility: v })}
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

function BulletsEditor({ items = [], onChange }) {
  function addItem() { onChange([...items, '']) }
  function updateItem(i, val) { onChange(items.map((x, j) => (j === i ? val : x))) }
  function removeItem(i) { onChange(items.filter((_, j) => j !== i)) }
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
      {items.map((item, i) => (
        <div key={i} style={{ display: 'flex', gap: 6, alignItems: 'center' }}>
          <Input
            size="small"
            value={item}
            placeholder={`Bullet ${i + 1}`}
            onChange={(e) => updateItem(i, e.target.value)}
            style={{ flex: 1, minWidth: 0 }}
          />
          <Button size="small" danger type="text" onClick={() => removeItem(i)} style={{ flexShrink: 0 }}>✕</Button>
        </div>
      ))}
      <Button size="small" style={{ alignSelf: 'flex-start', marginTop: 2 }} onClick={addItem}>+ Add Bullet</Button>
    </div>
  )
}

function TwoColumnEditor({ section, sectionIndex, onChange }) {
  function update(patch) { onChange(sectionIndex, patch) }
  return (
    <>
      <div className="admin-form-row">
        <Form.Item label="Image Side" style={{ flex: 1, marginBottom: 8 }}>
          <Select
            value={section.image_side || 'right'}
            options={TWO_COLUMN_IMAGE_SIDE_OPTIONS}
            onChange={(v) => update({ image_side: v })}
          />
        </Form.Item>
        <Form.Item label="Variant" style={{ flex: 1, marginBottom: 8 }}>
          <Select
            value={section.variant || 'default'}
            options={TWO_COLUMN_VARIANT_OPTIONS}
            onChange={(v) => update({ variant: v })}
          />
        </Form.Item>
      </div>
      <Form.Item label="Image Aspect Ratio" style={{ marginBottom: 8 }}>
        <Select
          value={section.image_aspect_ratio || '16/9'}
          options={IMAGE_ASPECT_RATIO_OPTIONS}
          onChange={(v) => update({ image_aspect_ratio: v })}
        />
      </Form.Item>
      <Form.Item label="Body" style={{ marginBottom: 8 }}>
        <Input.TextArea autoSize={{ minRows: 2, maxRows: 8 }} value={section.body || ''} onChange={(e) => update({ body: e.target.value })} />
      </Form.Item>
      <Form.Item label="Bullet points" style={{ marginBottom: 8 }}>
        <BulletsEditor
          items={section.bullets || []}
          onChange={(newBullets) => update({ bullets: newBullets })}
        />
      </Form.Item>
      <Form.Item label="Pull Quote" style={{ marginBottom: 8 }}>
        <Input value={section.pull_quote || ''} onChange={(e) => update({ pull_quote: e.target.value })} placeholder="Optional callout quote" />
      </Form.Item>
      <Form.Item label="Buttons" style={{ marginBottom: 8 }}>
        <ButtonsEditor section={section} sectionIndex={sectionIndex} onChange={onChange} />
      </Form.Item>
      <Form.Item label="Image Label (placeholder text)" style={{ marginBottom: 8 }}>
        <Input value={section.image_label || ''} onChange={(e) => update({ image_label: e.target.value })} placeholder="Describes the intended image" />
      </Form.Item>
      <div className="admin-form-row">
        <Form.Item label="Image URL (leave empty to show placeholder)" style={{ flex: 2, marginBottom: 8 }}>
          <Input value={section.image_url || ''} onChange={(e) => update({ image_url: e.target.value })} />
        </Form.Item>
        <Form.Item label="Image Alt" style={{ flex: 1, marginBottom: 8 }}>
          <Input value={section.image_alt || ''} onChange={(e) => update({ image_alt: e.target.value })} />
        </Form.Item>
      </div>
    </>
  )
}

function ComparisonTableEditor({ section, sectionIndex, onChange }) {
  function update(patch) { onChange(sectionIndex, patch) }
  function addRow() {
    onChange(sectionIndex, { rows: [...(section.rows || []), { left: '', right: '' }] })
  }
  function updateRow(ri, patch) {
    onChange(sectionIndex, {
      rows: (section.rows || []).map((r, i) => (i === ri ? normalizeComparisonRow({ ...r, ...patch }) : r)),
    })
  }
  function removeRow(ri) {
    onChange(sectionIndex, { rows: (section.rows || []).filter((_, i) => i !== ri) })
  }
  return (
    <>
      <div className="admin-form-row">
        <Form.Item label="Left Column Label" style={{ flex: 1, marginBottom: 8 }}>
          <Input value={section.left_label || ''} onChange={(e) => update({ left_label: e.target.value })} />
        </Form.Item>
        <Form.Item label="Right Column Label" style={{ flex: 1, marginBottom: 8 }}>
          <Input value={section.right_label || ''} onChange={(e) => update({ right_label: e.target.value })} />
        </Form.Item>
      </div>
      {(section.rows || []).map((row, ri) => (
        <Card
          key={ri}
          size="small"
          className="admin-sub-card"
          title={`Row ${ri + 1}`}
          extra={<Button size="small" danger type="text" onClick={() => removeRow(ri)}>Remove</Button>}
        >
          <div className="admin-form-row">
            <Form.Item label="Left" style={{ flex: 1, marginBottom: 0 }}>
              <Input size="small" value={row.left || ''} onChange={(e) => updateRow(ri, { left: e.target.value })} />
            </Form.Item>
            <Form.Item label="Right" style={{ flex: 1, marginBottom: 0 }}>
              <Input size="small" value={row.right || ''} onChange={(e) => updateRow(ri, { right: e.target.value })} />
            </Form.Item>
          </div>
        </Card>
      ))}
      <Button size="small" style={{ marginTop: 4, marginBottom: 8 }} onClick={addRow}>+ Add Row</Button>
      <Form.Item label="Note (shown below table)" style={{ marginBottom: 8 }}>
        <Input.TextArea autoSize={{ minRows: 2, maxRows: 6 }} value={section.note || ''} onChange={(e) => update({ note: e.target.value })} />
      </Form.Item>
    </>
  )
}

function IconGridEditor({ section, sectionIndex, onChange }) {
  function update(patch) { onChange(sectionIndex, patch) }
  function addItem() {
    onChange(sectionIndex, { items: [...(section.items || []), { icon: '', text: '' }] })
  }
  function updateItem(ii, patch) {
    onChange(sectionIndex, {
      items: (section.items || []).map((item, i) => (i === ii ? normalizeIconItem({ ...item, ...patch }) : item)),
    })
  }
  function removeItem(ii) {
    onChange(sectionIndex, { items: (section.items || []).filter((_, i) => i !== ii) })
  }
  return (
    <>
      <Form.Item label="Body" style={{ marginBottom: 8 }}>
        <Input.TextArea autoSize={{ minRows: 2, maxRows: 6 }} value={section.body || ''} onChange={(e) => update({ body: e.target.value })} />
      </Form.Item>
      {(section.items || []).map((item, ii) => (
        <Card
          key={ii}
          size="small"
          className="admin-sub-card"
          title={`Item ${ii + 1}`}
          extra={<Button size="small" danger type="text" onClick={() => removeItem(ii)}>Remove</Button>}
        >
          <div className="admin-form-row">
            <Form.Item label="Icon / Emoji" style={{ flex: 1, marginBottom: 0 }}>
              <Input size="small" value={item.icon || ''} onChange={(e) => updateItem(ii, { icon: e.target.value })} />
            </Form.Item>
            <Form.Item label="Text" style={{ flex: 3, marginBottom: 0 }}>
              <Input size="small" value={item.text || ''} onChange={(e) => updateItem(ii, { text: e.target.value })} />
            </Form.Item>
          </div>
        </Card>
      ))}
      <Button size="small" style={{ marginTop: 4, marginBottom: 8 }} onClick={addItem}>+ Add Item</Button>
      <Form.Item label="Image Label (photo collage placeholder)" style={{ marginBottom: 8 }}>
        <Input value={section.image_label || ''} onChange={(e) => update({ image_label: e.target.value })} placeholder="Describes the intended photo collage" />
      </Form.Item>
    </>
  )
}

function MatrixTableEditor({ section, sectionIndex, onChange }) {
  function update(patch) { onChange(sectionIndex, patch) }

  const colCount = (section.columns || []).length

  function addColumn() {
    const newCols = [...(section.columns || []), normalizeMatrixColumn({})]
    const newRows = (section.rows || []).map((r) => ({ ...r, cells: [...(r.cells || []), ''] }))
    onChange(sectionIndex, { columns: newCols, rows: newRows })
  }
  function updateColumn(ci, patch) {
    onChange(sectionIndex, {
      columns: (section.columns || []).map((c, i) => (i === ci ? normalizeMatrixColumn({ ...c, ...patch }) : c)),
    })
  }
  function removeColumn(ci) {
    const newCols = (section.columns || []).filter((_, i) => i !== ci)
    const newRows = (section.rows || []).map((r) => ({
      ...r,
      cells: (r.cells || []).filter((_, i) => i !== ci),
    }))
    onChange(sectionIndex, { columns: newCols, rows: newRows })
  }
  function updateColumnFeatures(ci, newFeatures) {
    onChange(sectionIndex, {
      columns: (section.columns || []).map((c, i) =>
        i === ci ? { ...c, features: newFeatures } : c
      ),
    })
  }

  function addRow() {
    const newRow = normalizeMatrixRow({ label: '', subtext: '', cells: [] }, colCount)
    onChange(sectionIndex, { rows: [...(section.rows || []), newRow] })
  }
  function updateRow(ri, patch) {
    onChange(sectionIndex, {
      rows: (section.rows || []).map((r, i) =>
        i === ri ? normalizeMatrixRow({ ...r, ...patch }, colCount) : r
      ),
    })
  }
  function updateCell(ri, ci, value) {
    onChange(sectionIndex, {
      rows: (section.rows || []).map((r, i) => {
        if (i !== ri) return r
        const cells = [...(r.cells || [])]
        cells[ci] = value
        return { ...r, cells }
      }),
    })
  }
  function removeRow(ri) {
    onChange(sectionIndex, { rows: (section.rows || []).filter((_, i) => i !== ri) })
  }

  return (
    <>
      <div className="admin-form-row">
        <Form.Item label="Subheading" style={{ flex: 1, marginBottom: 8 }}>
          <Input value={section.subheading || ''} onChange={(e) => update({ subheading: e.target.value })} />
        </Form.Item>
        <Form.Item label="Cell Label (e.g. Yearly Revenue)" style={{ flex: 1, marginBottom: 8 }}>
          <Input value={section.cell_label || ''} onChange={(e) => update({ cell_label: e.target.value })} />
        </Form.Item>
      </div>

      <Typography.Text strong style={{ display: 'block', marginBottom: 6 }}>Columns</Typography.Text>
      {(section.columns || []).map((col, ci) => (
        <Card
          key={ci}
          size="small"
          className="admin-sub-card"
          title={`Column ${ci + 1}${col.label ? ` — ${col.label}` : ''}`}
          extra={<Button size="small" danger type="text" onClick={() => removeColumn(ci)}>Remove</Button>}
        >
          <Form.Item label="Label" style={{ marginBottom: 6 }}>
            <Input size="small" value={col.label || ''} onChange={(e) => updateColumn(ci, { label: e.target.value })} />
          </Form.Item>
          <Form.Item label="Subtext" style={{ marginBottom: 6 }}>
            <Input size="small" value={col.subtext || ''} onChange={(e) => updateColumn(ci, { subtext: e.target.value })} />
          </Form.Item>
          <Form.Item label="Features" style={{ marginBottom: 0 }}>
            <BulletsEditor
              items={col.features || []}
              onChange={(newFeatures) => updateColumnFeatures(ci, newFeatures)}
            />
          </Form.Item>
        </Card>
      ))}
      <Button size="small" style={{ marginBottom: 12 }} onClick={addColumn}>+ Add Column</Button>

      <Typography.Text strong style={{ display: 'block', marginBottom: 6 }}>Rows</Typography.Text>
      {(section.rows || []).map((row, ri) => (
        <Card
          key={ri}
          size="small"
          className="admin-sub-card"
          title={`Row ${ri + 1}${row.label ? ` — ${row.label}` : ''}`}
          extra={<Button size="small" danger type="text" onClick={() => removeRow(ri)}>Remove</Button>}
        >
          <div className="admin-form-row">
            <Form.Item label="Label" style={{ flex: 1, marginBottom: 6 }}>
              <Input size="small" value={row.label || ''} onChange={(e) => updateRow(ri, { label: e.target.value })} />
            </Form.Item>
            <Form.Item label="Subtext" style={{ flex: 1, marginBottom: 6 }}>
              <Input size="small" value={row.subtext || ''} onChange={(e) => updateRow(ri, { subtext: e.target.value })} />
            </Form.Item>
          </div>
          {(section.columns || []).map((col, ci) => (
            <Form.Item key={ci} label={`Cell: ${col.label || `Column ${ci + 1}`}`} style={{ marginBottom: 6 }}>
              <Input
                size="small"
                value={(row.cells || [])[ci] || ''}
                onChange={(e) => updateCell(ri, ci, e.target.value)}
                placeholder="e.g. $1,000 - $5,000"
              />
            </Form.Item>
          ))}
        </Card>
      ))}
      <Button size="small" style={{ marginBottom: 8 }} onClick={addRow}>+ Add Row</Button>

      <Form.Item label="Note (shown below table)" style={{ marginBottom: 8 }}>
        <Input.TextArea autoSize={{ minRows: 2, maxRows: 6 }} value={section.note || ''} onChange={(e) => update({ note: e.target.value })} />
      </Form.Item>

      <Divider orientation="left" orientationMargin={0} style={{ marginBottom: 8, fontSize: '0.82rem' }}>Annotation (optional callout shown top-right of table)</Divider>
      <Form.Item label="Annotation Heading" style={{ marginBottom: 8 }}>
        <Input value={section.annotation_heading || ''} onChange={(e) => update({ annotation_heading: e.target.value })} placeholder="e.g. Key Assumptions" />
      </Form.Item>
      <Form.Item label="Annotation Items" style={{ marginBottom: 8 }}>
        <BulletsEditor
          items={section.annotation_items || []}
          onChange={(newItems) => update({ annotation_items: newItems })}
        />
      </Form.Item>
    </>
  )
}

function ContentWithImageEditor({ section, sectionIndex, onChange }) {
  function update(patch) { onChange(sectionIndex, patch) }

  const inner = section.content_section || makeDefaultSection('text')

  function updateInner(patch) {
    update({ content_section: { ...inner, ...patch } })
  }
  function changeInnerType(newType) {
    update({ content_section: makeDefaultSection(newType) })
  }
  // Adapter so sub-editors can call onChange(_, patch) and updateInner(patch) is triggered
  function innerOnChange(_, patch) { updateInner(patch) }

  return (
    <>
      <div className="admin-form-row">
        <Form.Item label="Image Side" style={{ flex: 1, marginBottom: 8 }}>
          <Select
            value={section.image_side || 'right'}
            options={TWO_COLUMN_IMAGE_SIDE_OPTIONS}
            onChange={(v) => update({ image_side: v })}
          />
        </Form.Item>
        <Form.Item label="Aspect Ratio" style={{ flex: 1, marginBottom: 8 }}>
          <Select
            value={section.image_aspect_ratio || '4/3'}
            options={IMAGE_ASPECT_RATIO_OPTIONS}
            onChange={(v) => update({ image_aspect_ratio: v })}
          />
        </Form.Item>
      </div>
      <Form.Item label="Image URL" style={{ marginBottom: 8 }}>
        <Input value={section.image_url || ''} onChange={(e) => update({ image_url: e.target.value })} />
      </Form.Item>
      <Form.Item label="Image Label (placeholder text, shown when no URL)" style={{ marginBottom: 8 }}>
        <Input value={section.image_label || ''} onChange={(e) => update({ image_label: e.target.value })} placeholder="Describes the intended image" />
      </Form.Item>
      <Form.Item label="Image Alt Text" style={{ marginBottom: 8 }}>
        <Input value={section.image_alt || ''} onChange={(e) => update({ image_alt: e.target.value })} />
      </Form.Item>

      <Typography.Text strong style={{ display: 'block', margin: '12px 0 6px' }}>Content Column</Typography.Text>
      <Card size="small" className="admin-sub-card">
        <Form.Item label="Content Type" style={{ marginBottom: 8 }}>
          <Select
            value={inner.type || 'text'}
            options={CONTENT_WITH_IMAGE_SUPPORTED_TYPES}
            onChange={changeInnerType}
          />
        </Form.Item>
        <Form.Item label="Heading" style={{ marginBottom: 8 }}>
          <Input value={inner.heading || ''} onChange={(e) => updateInner({ heading: e.target.value })} />
        </Form.Item>
        {inner.type === 'text' ? (
          <Form.Item label="Body" style={{ marginBottom: 0 }}>
            <Input.TextArea autoSize={{ minRows: 3, maxRows: 8 }} value={inner.body || ''} onChange={(e) => updateInner({ body: e.target.value })} />
          </Form.Item>
        ) : null}
        {inner.type === 'bullets' ? (
          <Form.Item label="Items" style={{ marginBottom: 0 }}>
            <BulletsEditor items={inner.items || []} onChange={(items) => updateInner({ items })} />
          </Form.Item>
        ) : null}
        {inner.type === 'comparison_table' ? (
          <ComparisonTableEditor section={inner} sectionIndex={0} onChange={innerOnChange} />
        ) : null}
        {inner.type === 'icon_grid' ? (
          <IconGridEditor section={inner} sectionIndex={0} onChange={innerOnChange} />
        ) : null}
        {inner.type === 'matrix_table' ? (
          <MatrixTableEditor section={inner} sectionIndex={0} onChange={innerOnChange} />
        ) : null}
      </Card>
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
  const [containerSize, setContainerSize] = useState('standard')
  const [showHeader, setShowHeader] = useState(false)
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
        try {
          const parsed = JSON.parse(d.content_json)
          const { sections, container_size, show_header } = normalizeContent(parsed)
          setSectionsDraft(sections)
          setContainerSize(container_size)
          setShowHeader(show_header)
        } catch { setSectionsDraft([]) }
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
    try {
      const parsed = JSON.parse(d.content_json)
      const { sections, container_size, show_header } = normalizeContent(parsed)
      setSectionsDraft(sections)
      setContainerSize(container_size)
      setShowHeader(show_header)
    } catch { setSectionsDraft([]) }
    setIsDirty(false)
    setSaveError('')
    setSaveSuccess(false)
  }

  async function saveChanges() {
    if (!page) return
    setIsSaving(true)
    setSaveError('')
    try {
      const payload = { ...draft, content_json: toContentJSON(sectionsDraft, containerSize, showHeader) }
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
  const collapseItems = sectionsDraft.map((section, index) => {
    const typeLabel = SECTION_TYPES.find((t) => t.value === section.type)?.label || section.type
    const heading = section.heading || `Section ${index + 1}`
    return {
      key: String(index),
      label: (
        <span className="admin-section-label">
          <span className="admin-section-label__meta">
            <span className="admin-section-label__index">#{index + 1}</span>
            <Tag className="admin-section-label__type">{typeLabel}</Tag>
          </span>
          <span className="admin-section-label__heading">{heading}</span>
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
    }
  })

  const selectedPageLabel = CMS_PAGES.find((p) => p.slug === selectedSlug)?.label || selectedSlug

  const screens = Grid.useBreakpoint()
  const [mobileWarnSeen, setMobileWarnSeen] = useState(false)
  const isMobile = screens.md === false

  return (
    <Layout className="admin-layout">
      <Modal
        open={isMobile && !mobileWarnSeen}
        onOk={() => setMobileWarnSeen(true)}
        onCancel={() => setMobileWarnSeen(true)}
        cancelButtonProps={{ style: { display: 'none' } }}
        okText="Got it"
        title={
          <Space>
            <WarningOutlined style={{ color: '#faad14' }} />
            Desktop recommended
          </Space>
        }
      >
        The admin panel is designed for desktop use. Some features may be difficult or impossible to use on a small screen.
      </Modal>

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
          <div style={{ width: 160, display: 'flex', justifyContent: 'flex-end', alignItems: 'center' }}>
            {isMobile ? (
              <Tooltip title="Admin panel works best on desktop">
                <WarningOutlined
                  style={{ color: '#faad14', fontSize: 18, cursor: 'pointer' }}
                  onClick={() => setMobileWarnSeen(false)}
                />
              </Tooltip>
            ) : null}
          </div>
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
                      <Form.Item label="Show page header" style={{ marginBottom: 0, marginTop: 12 }}>
                        <Switch
                          checked={showHeader}
                          onChange={(v) => { setShowHeader(v); markDirty() }}
                          checkedChildren="Visible"
                          unCheckedChildren="Hidden"
                        />
                      </Form.Item>
                      <Form.Item label="Container Width" style={{ marginBottom: 0, marginTop: 12, paddingBottom: 8 }}>
                        <Slider
                          min={0}
                          max={3}
                          step={1}
                          marks={CONTAINER_MARKS}
                          value={CONTAINER_STEP_INDEX[containerSize] ?? 0}
                          onChange={(v) => { setContainerSize(CONTAINER_STEPS[v]); markDirty() }}
                          tooltip={{ formatter: (v) => CONTAINER_MARKS[v] }}
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
                <div className={`admin-preview-inner admin-preview-page--${containerSize}`}>
                  <div className="cms-page-header" style={{ paddingLeft: 0, paddingRight: 0 }}>
                    <Typography.Title level={2} className="cms-page-title">{draft.title}</Typography.Title>
                    {draft.description ? (
                      <Typography.Paragraph className="cms-page-desc">{draft.description}</Typography.Paragraph>
                    ) : null}
                  </div>
                  <CMSContentRenderer sections={previewSections} containerSize={containerSize} />
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
          <UsersRoles />
        ) : null}
      </Layout.Content>
    </Layout>
  )
}
