// Shared CMS constants and helpers used by both the public renderer and admin editor.

export const CMS_PAGES = [
  { slug: 'partnership-overview', label: 'Partnership Overview' },
  { slug: 'how-it-works', label: 'How it Works' },
  { slug: 'case-studies', label: 'Case Studies' },
  { slug: 'pricing-revenue-share', label: 'Pricing & Revenue Share' },
  { slug: 'partner-faq', label: 'Partner FAQ' },
  { slug: 'application-contact', label: 'Application / Contact' },
]

export const SECTION_TYPES = [
  { label: 'Text Block', value: 'text' },
  { label: 'Bullet List', value: 'bullets' },
  { label: 'Buttons / Links', value: 'buttons' },
  { label: 'Form Call-To-Action', value: 'form_cta' },
  { label: 'Application Form Fields', value: 'application_form' },
  { label: 'Image Block', value: 'image' },
  { label: 'Image Grid', value: 'image_grid' },
  { label: 'Two-Column (Text + Image)', value: 'two_column' },
  { label: 'Comparison Table', value: 'comparison_table' },
  { label: 'Icon Grid', value: 'icon_grid' },
  { label: 'Matrix / Pivot Table', value: 'matrix_table' },
  { label: 'Content + Image (paired)', value: 'content_with_image' },
]

// Section types allowed as the content column inside content_with_image.
// Excludes layout types (two_column, image, image_grid, content_with_image).
export const CONTENT_WITH_IMAGE_SUPPORTED_TYPES = [
  { label: 'Text Block', value: 'text' },
  { label: 'Bullet List', value: 'bullets' },
  { label: 'Comparison Table', value: 'comparison_table' },
  { label: 'Icon Grid', value: 'icon_grid' },
  { label: 'Matrix / Pivot Table', value: 'matrix_table' },
]

export const SECTION_BACKGROUND_OPTIONS = [
  { label: 'Default', value: '' },
  { label: 'White', value: '#ffffff' },
  { label: 'Soft Blue-Gray', value: '#f6f7fe' },
  { label: 'Brand Teal', value: '#00698f' },
]

export const SECTION_ALIGNMENT_OPTIONS = [
  { label: 'Left', value: 'left' },
  { label: 'Center', value: 'center' },
  { label: 'Right', value: 'right' },
]

export const CONTAINER_SIZE_OPTIONS = [
  { label: 'Standard', value: 'standard' },
  { label: 'Wide',     value: 'wide'     },
  { label: 'Wider',    value: 'wider'    },
  { label: 'Full',     value: 'full'     },
]

export const IMAGE_POSITION_OPTIONS = [
  { label: 'Center (default)', value: 'center' },
  { label: 'Top', value: 'top' },
  { label: 'Bottom', value: 'bottom' },
]

export const TWO_COLUMN_IMAGE_SIDE_OPTIONS = [
  { label: 'Right (default)', value: 'right' },
  { label: 'Left', value: 'left' },
]

export const TWO_COLUMN_VARIANT_OPTIONS = [
  { label: 'Default', value: 'default' },
  { label: 'Hero (full height)', value: 'hero' },
]

export const IMAGE_ASPECT_RATIO_OPTIONS = [
  { label: '16 : 9  (wide landscape)', value: '16/9' },
  { label: '4 : 3', value: '4/3' },
  { label: '3 : 2', value: '3/2' },
  { label: '1 : 1  (square)', value: '1/1' },
]

export const FIELD_TYPE_OPTIONS = [
  { label: 'Text', value: 'text' },
  { label: 'Email', value: 'email' },
  { label: 'Phone', value: 'tel' },
  { label: 'URL', value: 'url' },
  { label: 'Text Area', value: 'textarea' },
]

export const ALLOWED_FIELD_TYPES = new Set(FIELD_TYPE_OPTIONS.map((o) => o.value))

export function getContrastTextColor(background) {
  if (!background || !background.startsWith('#') || (background.length !== 7 && background.length !== 4)) {
    return '#121212'
  }
  let r, g, b
  if (background.length === 7) {
    r = parseInt(background.slice(1, 3), 16)
    g = parseInt(background.slice(3, 5), 16)
    b = parseInt(background.slice(5, 7), 16)
  } else {
    r = parseInt(background[1] + background[1], 16)
    g = parseInt(background[2] + background[2], 16)
    b = parseInt(background[3] + background[3], 16)
  }
  return (0.299 * r + 0.587 * g + 0.114 * b) / 255 < 0.5 ? '#ffffff' : '#121212'
}

export function normalizeField(field) {
  return {
    name: String(field?.name || '').trim(),
    label: String(field?.label || '').trim(),
    type: ALLOWED_FIELD_TYPES.has(field?.type) ? field.type : 'text',
    required: Boolean(field?.required),
    placeholder: String(field?.placeholder || '').trim(),
  }
}

export function normalizeGridItem(item) {
  return {
    title: String(item?.title || '').trim(),
    image_url: String(item?.image_url || '').trim(),
    link_url: String(item?.link_url || '').trim(),
  }
}

export function normalizeIconItem(item) {
  return {
    icon: String(item?.icon || '').trim(),
    text: String(item?.text || '').trim(),
  }
}

export function normalizeComparisonRow(row) {
  return {
    left: String(row?.left || '').trim(),
    right: String(row?.right || '').trim(),
  }
}

export function normalizeMatrixColumn(col) {
  return {
    label: String(col?.label || '').trim(),
    subtext: String(col?.subtext || '').trim(),
    features: Array.isArray(col?.features) ? col.features.map((f) => String(f).trim()).filter(Boolean) : [],
  }
}

export function normalizeMatrixRow(row, colCount) {
  const cells = Array.isArray(row?.cells) ? row.cells.map((c) => String(c).trim()) : []
  while (cells.length < colCount) cells.push('')
  return {
    label: String(row?.label || '').trim(),
    subtext: String(row?.subtext || '').trim(),
    cells: cells.slice(0, colCount),
  }
}

export function makeDefaultSection(type = 'text') {
  if (type === 'bullets') {
    return { type, container_size: 'standard', background: '', alignment: 'left', heading: 'Key Benefits', items: ['First point'] }
  }
  if (type === 'buttons') {
    return { type, container_size: 'standard', background: '', alignment: 'center', heading: 'Next Steps', buttons: [{ label: 'Learn More', url: '/', variant: 'primary' }] }
  }
  if (type === 'form_cta') {
    return { type, container_size: 'standard', background: '', alignment: 'center', heading: 'Ready to Partner?', body: 'Tell us about your organization and goals.', button_label: 'Start Application', button_link: '/apply' }
  }
  if (type === 'application_form') {
    return {
      type, container_size: 'standard', background: '#f6f7fe', alignment: 'left', heading: 'Application Fields', submit_label: 'Submit Application',
      fields: [
        { name: 'organization_name', label: 'Organization Name', type: 'text', required: true, placeholder: '' },
        { name: 'contact_name', label: 'Contact Name', type: 'text', required: true, placeholder: '' },
        { name: 'email', label: 'Email', type: 'email', required: true, placeholder: '' },
      ],
    }
  }
  if (type === 'image') {
    return { type, container_size: 'full', background: '', alignment: 'left', heading: 'Image Section', body: '', image_url: '', image_label: '', image_alt: '', image_position: 'center', button_label: '', button_link: '' }
  }
  if (type === 'image_grid') {
    return { type, container_size: 'wide', background: '', alignment: 'left', heading: 'Image Grid', items: [{ title: '', image_url: '', link_url: '' }] }
  }
  if (type === 'two_column') {
    return {
      type, container_size: 'standard', variant: 'default', image_side: 'right', background: '', alignment: 'left',
      heading: 'Section Heading', body: '', bullets: [], pull_quote: '',
      button_label: '', button_link: '',
      image_url: '', image_label: '', image_alt: '', image_aspect_ratio: '16/9',
    }
  }
  if (type === 'comparison_table') {
    return {
      type, container_size: 'wide', background: '', alignment: 'center',
      heading: 'Why We Are Different', left_label: 'Traditional', right_label: 'Our Approach',
      rows: [{ left: '', right: '' }], note: '',
    }
  }
  if (type === 'icon_grid') {
    return {
      type, container_size: 'wide', background: '', alignment: 'center',
      heading: 'Impact', body: '', items: [{ icon: '✓', text: '' }], image_label: '',
    }
  }
  if (type === 'matrix_table') {
    return {
      type, container_size: 'wider', background: '', alignment: 'center',
      heading: 'Matrix Heading', subheading: '',
      annotation_heading: '', annotation_items: [],
      columns: [
        { label: 'Column A', subtext: '', features: [] },
        { label: 'Column B', subtext: '', features: [] },
      ],
      rows: [
        { label: 'Row 1', subtext: '', cells: ['', ''] },
        { label: 'Row 2', subtext: '', cells: ['', ''] },
      ],
      cell_label: '',
      note: '',
    }
  }
  if (type === 'content_with_image') {
    return {
      type, container_size: 'standard', background: '', image_side: 'right',
      image_url: '', image_alt: '', image_label: '', image_aspect_ratio: '4/3',
      content_section: makeDefaultSection('text'),
    }
  }
  return { type: 'text', container_size: 'standard', background: '', alignment: 'left', heading: 'Overview', body: '' }
}

export function normalizeSection(section) {
  const normalized = { ...makeDefaultSection(section?.type), ...section }
  if (normalized.type === 'bullets') normalized.items = Array.isArray(normalized.items) ? normalized.items : []
  if (normalized.type === 'buttons') normalized.buttons = Array.isArray(normalized.buttons) ? normalized.buttons : []
  if (normalized.type === 'application_form') {
    normalized.submit_label = String(normalized.submit_label || 'Submit Application')
    normalized.fields = Array.isArray(normalized.fields) ? normalized.fields.map(normalizeField).filter((f) => f.name) : []
  }
  if (normalized.type === 'image_grid') {
    normalized.items = Array.isArray(normalized.items) ? normalized.items.map(normalizeGridItem) : []
  }
  if (normalized.type === 'two_column') {
    normalized.bullets = Array.isArray(normalized.bullets) ? normalized.bullets : []
  }
  if (normalized.type === 'comparison_table') {
    normalized.rows = Array.isArray(normalized.rows) ? normalized.rows.map(normalizeComparisonRow) : []
  }
  if (normalized.type === 'icon_grid') {
    normalized.items = Array.isArray(normalized.items) ? normalized.items.map(normalizeIconItem) : []
  }
  if (normalized.type === 'matrix_table') {
    normalized.columns = Array.isArray(normalized.columns) ? normalized.columns.map(normalizeMatrixColumn) : []
    const colCount = normalized.columns.length
    normalized.rows = Array.isArray(normalized.rows) ? normalized.rows.map((r) => normalizeMatrixRow(r, colCount)) : []
    normalized.annotation_items = Array.isArray(normalized.annotation_items) ? normalized.annotation_items : []
  }
  if (normalized.type === 'content_with_image') {
    const inner = normalizeSection(normalized.content_section || {})
    normalized.content_section = inner.type === 'content_with_image' ? makeDefaultSection('text') : inner
  }
  return normalized
}

export function normalizeSections(content) {
  if (Array.isArray(content?.sections)) {
    return content.sections.map(normalizeSection)
  }

  const entries = Object.entries(content || {})
  if (!entries.length) return []

  return entries.map(([key, value]) => {
    if (Array.isArray(value)) {
      return { type: 'bullets', heading: key.replaceAll('_', ' '), items: value.map(String), background: '', alignment: 'left' }
    }
    return { type: 'text', heading: key.replaceAll('_', ' '), body: typeof value === 'string' ? value : JSON.stringify(value, null, 2), background: '', alignment: 'left' }
  })
}

export function toContentJSON(sections, containerSize = 'standard', showHeader = false) {
  return JSON.stringify({ container_size: containerSize, show_header: showHeader, sections }, null, 2)
}

export function normalizeContent(raw) {
  return {
    sections: normalizeSections(raw),
    container_size: raw?.container_size || 'standard',
    show_header: Boolean(raw?.show_header ?? false),
  }
}
