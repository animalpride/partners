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

export function makeDefaultSection(type = 'text') {
  if (type === 'bullets') {
    return { type, background: '', alignment: 'left', heading: 'Key Benefits', items: ['First point'] }
  }
  if (type === 'buttons') {
    return { type, background: '', alignment: 'center', heading: 'Next Steps', buttons: [{ label: 'Learn More', url: '/', variant: 'primary' }] }
  }
  if (type === 'form_cta') {
    return { type, background: '', alignment: 'center', heading: 'Ready to Partner?', body: 'Tell us about your organization and goals.', button_label: 'Start Application', button_link: '/apply' }
  }
  if (type === 'application_form') {
    return {
      type, background: '#f6f7fe', alignment: 'left', heading: 'Application Fields', submit_label: 'Submit Application',
      fields: [
        { name: 'organization_name', label: 'Organization Name', type: 'text', required: true, placeholder: '' },
        { name: 'contact_name', label: 'Contact Name', type: 'text', required: true, placeholder: '' },
        { name: 'email', label: 'Email', type: 'email', required: true, placeholder: '' },
      ],
    }
  }
  if (type === 'image') {
    return { type, background: '', alignment: 'left', heading: 'Image Section', body: '', image_url: '', image_label: '', image_alt: '', image_position: 'center', button_label: '', button_link: '' }
  }
  if (type === 'image_grid') {
    return { type, background: '', alignment: 'left', heading: 'Image Grid', items: [{ title: '', image_url: '', link_url: '' }] }
  }
  if (type === 'two_column') {
    return {
      type, variant: 'default', image_side: 'right', background: '', alignment: 'left',
      heading: 'Section Heading', body: '', bullets: [], pull_quote: '',
      button_label: '', button_link: '',
      image_url: '', image_label: '', image_alt: '', image_aspect_ratio: '16/9',
    }
  }
  if (type === 'comparison_table') {
    return {
      type, background: '', alignment: 'center',
      heading: 'Why We Are Different', left_label: 'Traditional', right_label: 'Our Approach',
      rows: [{ left: '', right: '' }], note: '',
    }
  }
  if (type === 'icon_grid') {
    return {
      type, background: '', alignment: 'center',
      heading: 'Impact', body: '', items: [{ icon: '✓', text: '' }], image_label: '',
    }
  }
  return { type: 'text', background: '', alignment: 'left', heading: 'Overview', body: '' }
}

export function normalizeSections(content) {
  if (Array.isArray(content?.sections)) {
    return content.sections.map((section) => {
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
      return normalized
    })
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

export function toContentJSON(sections) {
  return JSON.stringify({ sections }, null, 2)
}
