import { Alert, Spin, Typography } from 'antd'
import { useEffect, useMemo, useState } from 'react'
import { Helmet } from 'react-helmet-async'
import { getPage } from '../api'
import { getContrastTextColor, normalizeContent } from '../cmsTypes'

const SITE_URL = 'https://partners.animalpride.com'
const DEFAULT_OG_IMAGE = `${SITE_URL}/Logo-Wordmark-Dog-PartnersPlatform2.png`

const SLUG_TO_PATH = {
  'partnership-overview': '/',
  'how-it-works': '/how-it-works',
  'case-studies': '/case-studies',
  'pricing-revenue-share': '/pricing',
  'partner-faq': '/faq',
  'application-contact': '/apply',
}




const OVERLAY_JUSTIFY = { left: 'flex-start', center: 'center', right: 'flex-end' }
const BTN_ROW_JUSTIFY  = { left: 'flex-start', center: 'center', right: 'flex-end' }
const BULLET_ALIGN     = { left: 'flex-start', center: 'center', right: 'flex-end' }
const SECTION_INNER_CLS = {
  standard: 'cms-section-inner',
  wide:     'cms-section-inner cms-section-inner--wide',
  wider:    'cms-section-inner cms-section-inner--wider',
  full:     'cms-section-inner cms-section-inner--full',
}
const ASPECT_CLASS_MAP = {
  '16/9': 'cms-img-placeholder--16-9',
  '4/3':  'cms-img-placeholder--4-3',
  '3/2':  'cms-img-placeholder--3-2',
  '1/1':  'cms-img-placeholder--1-1',
}

// Parse inline markdown (**bold**, *italic*, _italic_) within a single line of text.
function parseInline(text) {
  const re = /\*\*(.*?)\*\*|\*(.*?)\*|_(.*?)_|[^*_]+|[*_]/g
  const nodes = []
  let m
  while ((m = re.exec(text)) !== null) {
    if (m[1] !== undefined) nodes.push(<strong key={nodes.length}>{m[1]}</strong>)
    else if (m[2] !== undefined) nodes.push(<em key={nodes.length}>{m[2]}</em>)
    else if (m[3] !== undefined) nodes.push(<em key={nodes.length}>{m[3]}</em>)
    else nodes.push(m[0])
  }
  return nodes
}

// Render user-typed text with \n line breaks and inline markdown.
function renderMarkdown(text) {
  if (!text) return null
  const lines = String(text).split('\\n')
  return lines.flatMap((line, li) => {
    const nodes = parseInline(line)
    return li === 0 ? nodes : [<br key={`br-${li}`} />, ...nodes]
  })
}

// Render a vertical stack of CMS buttons from a buttons[] array.
// alignment: 'left' | 'center' | 'right' — controls horizontal alignment of the stack.
// sizeClass: optional extra class added to each button (e.g. 'cms-cta-btn--lg').
function renderButtons(buttons, alignment, sizeClass) {
  const validButtons = (buttons || []).filter((btn) => btn?.label && btn?.url)
  if (!validButtons.length) return null
  const alignItems = alignment === 'center' ? 'center' : alignment === 'right' ? 'flex-end' : 'flex-start'
  return (
    <div className="cms-btn-row" style={{ alignItems }}>
      <div className="cms-btn-stack">
        {validButtons.map((button, i) => {
          const isExternal = /^https?:\/\//i.test(button.url)
          const variantCls = button.variant === 'primary' ? 'cms-cta-btn--primary' : 'cms-cta-btn--ghost'
          const visibilityCls =
            button.visibility === 'desktop' ? ' cms-btn--desktop-only' :
            button.visibility === 'mobile' ? ' cms-btn--mobile-only' : ''
          return (
            <a
              key={i}
              href={button.url}
              target={isExternal ? '_blank' : undefined}
              rel={isExternal ? 'noreferrer' : undefined}
              className={`cms-cta-btn ${variantCls}${sizeClass ? ` ${sizeClass}` : ''}${visibilityCls}`}
            >
              {button.label}
            </a>
          )
        })}
      </div>
    </div>
  )
}

// ── SectionContent ────────────────────────────────────────────────────────────
// Renders the inner content of any "simple" section type.
// embedded=false (default): wraps content in cms-section-inner (max-width, centering).
// embedded=true: renders content directly — used inside a content_with_image column.
function SectionContent({ section, textColor, hasColor, embedded = false }) {
  const alignment = section.alignment || 'left'
  const innerStyle = { textAlign: alignment }
  const bulletAccent = hasColor && textColor === '#ffffff' ? '#ffffff' : '#00698f'

  function W(children) {
    if (embedded) return <div style={innerStyle}>{children}</div>
    const cls = SECTION_INNER_CLS[section.container_size] || SECTION_INNER_CLS.standard
    return (
      <div className={cls} style={innerStyle}>
        {children}
      </div>
    )
  }

  // ── Bullets ──────────────────────────────────────────────────────────────
  if (section.type === 'bullets') {
    const bulletCls = embedded ? undefined : (SECTION_INNER_CLS[section.container_size] || SECTION_INNER_CLS.standard)
    return (
      <div
        className={bulletCls}
        style={embedded ? {} : { ...innerStyle, ...(section.heading ? {} : { marginLeft: 0 }) }}
      >
        {section.heading ? (
          <h2 className="cms-section-heading" style={{ color: textColor }}>{section.heading}</h2>
        ) : null}
        <ul className="cms-bullets" style={{ alignItems: BULLET_ALIGN[alignment] || 'flex-start' }}>
          {(section.items || []).filter(Boolean).map((item, i) => (
            <li key={i} className="cms-bullet-item" style={{ color: textColor }}>
              <span className="cms-bullet-icon" style={{ color: bulletAccent }}>✓</span>
              {renderMarkdown(item)}
            </li>
          ))}
        </ul>
      </div>
    )
  }

  // ── Buttons / Links ───────────────────────────────────────────────────────
  if (section.type === 'buttons') {
    return W(
      <>
        {section.heading ? (
          <h2 className="cms-section-heading" style={{ color: textColor }}>{section.heading}</h2>
        ) : null}
        {renderButtons(section.buttons, alignment)}
      </>
    )
  }

  // ── Form CTA ──────────────────────────────────────────────────────────────
  if (section.type === 'form_cta') {
    return W(
      <>
        {section.heading ? (
          <h2 className="cms-section-heading" style={{ color: textColor }}>{section.heading}</h2>
        ) : null}
        {section.body ? <p className="cms-section-body" style={{ color: textColor }}>{renderMarkdown(section.body)}</p> : null}
        {(section.buttons && section.buttons.length > 0)
          ? renderButtons(section.buttons, alignment, 'cms-cta-btn--lg')
          : (section.button_label && section.button_link
            ? <div style={{ display: 'flex', justifyContent: BTN_ROW_JUSTIFY[alignment] || 'center' }}>
                <a href={section.button_link} className="cms-cta-btn cms-cta-btn--primary cms-cta-btn--lg">
                  {section.button_label}
                </a>
              </div>
            : null)
        }
      </>
    )
  }

  // ── Application Form (preview only) ──────────────────────────────────────
  if (section.type === 'application_form') {
    return W(
      <>
        {section.heading ? (
          <h2 className="cms-section-heading" style={{ color: textColor }}>{section.heading}</h2>
        ) : null}
        <ul className="cms-bullets">
          {(section.fields || []).map((field, i) => (
            <li key={i} className="cms-bullet-item" style={{ color: textColor }}>
              <span className="cms-bullet-icon" style={{ color: bulletAccent }}>✓</span>
              {field.label || field.name}{field.required ? ' *' : ''}
            </li>
          ))}
        </ul>
      </>
    )
  }

  // ── Comparison Table ──────────────────────────────────────────────────────
  if (section.type === 'comparison_table') {
    return W(
      <>
        {section.heading ? (
          <h2 className="cms-section-heading" style={{ color: textColor }}>{section.heading}</h2>
        ) : null}
        <div className="cms-ct-wrap">
          <div className="cms-ct-header">
            <div className="cms-ct-header-cell">{section.left_label}</div>
            <div className="cms-ct-header-cell">{section.right_label}</div>
          </div>
          {(section.rows || []).map((row, ri) => (
            <div key={ri} className="cms-ct-row">
              <div className="cms-ct-cell cms-ct-cell--left">{row.left}</div>
              <div className="cms-ct-cell cms-ct-cell--right">{row.right}</div>
            </div>
          ))}
        </div>
        {section.note ? <p className="cms-ct-note">{section.note}</p> : null}
      </>
    )
  }

  // ── Icon Grid ─────────────────────────────────────────────────────────────
  if (section.type === 'icon_grid') {
    return W(
      <>
        {section.heading ? (
          <h2 className="cms-section-heading" style={{ color: textColor }}>{section.heading}</h2>
        ) : null}
        {section.body ? <p className="cms-section-body" style={{ color: textColor }}>{renderMarkdown(section.body)}</p> : null}
        <div className="cms-icon-grid">
          {(section.items || []).map((item, i) => (
            <div key={i} className="cms-icon-item">
              {item.icon ? <span className="cms-icon-item__icon">{item.icon}</span> : null}
              <span className="cms-icon-item__text">{item.text}</span>
            </div>
          ))}
        </div>
        {section.image_label ? (
          <div className="cms-img-placeholder cms-img-placeholder--16-9" style={{ marginTop: 32 }}>
            <span className="cms-img-placeholder__label">{section.image_label}</span>
          </div>
        ) : null}
      </>
    )
  }

  // ── Matrix / Pivot Table ──────────────────────────────────────────────────
  if (section.type === 'matrix_table') {
    const columns = section.columns || []
    const rows = section.rows || []
    const annotationItems = section.annotation_items || []
    const hasAnnotation = section.annotation_heading || annotationItems.length > 0
    return W(
      <>
        <div className="cms-matrix-top-area">
          <div className="cms-matrix-meta">
            {section.heading ? (
              <h2 className="cms-section-heading" style={{ color: textColor }}>{section.heading}</h2>
            ) : null}
            {section.subheading ? (
              <p className="cms-section-body" style={{ color: textColor }}>{renderMarkdown(section.subheading)}</p>
            ) : null}
          </div>
          {hasAnnotation ? (
            <div className="cms-matrix-annotation">
              {section.annotation_heading ? (
                <p className="cms-matrix-annotation__heading">{section.annotation_heading}</p>
              ) : null}
              {annotationItems.length > 0 ? (
                <ul className="cms-matrix-annotation__list">
                  {annotationItems.map((item, i) => <li key={i}>{renderMarkdown(item)}</li>)}
                </ul>
              ) : null}
            </div>
          ) : null}
        </div>
        <div className="cms-matrix-scroll">
          <div className="cms-matrix-grid" style={{ '--matrix-cols': columns.length }}>
            <div className="cms-matrix-header-row">
              <div className="cms-matrix-corner" />
              {columns.map((col, ci) => (
                <div key={ci} className="cms-matrix-col-header">
                  <p className="cms-matrix-col-header__label">{col.label}</p>
                  {col.subtext ? <p className="cms-matrix-col-header__subtext">{col.subtext}</p> : null}
                  {(col.features || []).length > 0 ? (
                    <ul className="cms-matrix-features">
                      {col.features.map((f, fi) => (
                        <li key={fi} className="cms-matrix-features__item">
                          <span className="cms-bullet-icon cms-bullet-icon--light">✓</span>{f}
                        </li>
                      ))}
                    </ul>
                  ) : null}
                </div>
              ))}
            </div>
            {rows.map((row, ri) => (
              <div key={ri} className="cms-matrix-data-row">
                <div className="cms-matrix-row-header">
                  <p className="cms-matrix-row-header__label">{row.label}</p>
                  {row.subtext ? <p className="cms-matrix-row-header__subtext">{row.subtext}</p> : null}
                </div>
                {columns.map((col, ci) => (
                  <div key={ci} className="cms-matrix-cell" data-label={col.label}>
                    <p className="cms-matrix-cell__value">{(row.cells || [])[ci] || ''}</p>
                    {section.cell_label ? <p className="cms-matrix-cell__label">{section.cell_label}</p> : null}
                  </div>
                ))}
              </div>
            ))}
          </div>
        </div>
        {section.note ? (
          <p className="cms-ct-note" style={{ textAlign: 'center', marginTop: 24 }}>{section.note}</p>
        ) : null}
      </>
    )
  }

  // ── Default: Text ─────────────────────────────────────────────────────────
  return W(
    <>
      {section.heading ? (
        <h2 className="cms-section-heading" style={{ color: textColor }}>{section.heading}</h2>
      ) : null}
      {section.body ? (
        <p className="cms-section-body" style={{ color: textColor }}>{renderMarkdown(section.body)}</p>
      ) : null}
    </>
  )
}

export function CMSContentRenderer({ sections, containerSize = 'standard' }) {
  if (!sections.length) {
    return (
      <div className="cms-empty">
        <Typography.Paragraph type="secondary">No content yet. Use "Edit this page" to add sections.</Typography.Paragraph>
      </div>
    )
  }

  return (
    <div className={`cms-page${containerSize !== 'standard' ? ` cms-page--${containerSize}` : ''}`}>
      {sections.map((section, index) => {
        const bg = section.background || ''
        const textColor = getContrastTextColor(bg)
        const hasColor = Boolean(bg)
        const alignment = section.alignment || 'left'
        const innerStyle = { textAlign: alignment }

        // ── Image / Hero ──────────────────────────────────────────────────
        if (section.type === 'image') {
          const hasButtons = (section.buttons && section.buttons.some((b) => b?.label && b?.url))
          const hasOverlay = section.heading || section.body || (section.button_label && section.button_link) || hasButtons
          return (
            <section key={index} className="cms-section cms-section--hero">
              {section.image_url ? (
                <div className="cms-hero-wrap">
                  <img
                    src={section.image_url}
                    alt={section.image_alt || section.heading || 'Section image'}
                    className="cms-hero-img"
                    style={{ objectPosition: section.image_position || 'center' }}
                  />
                  {hasOverlay ? (
                    <div className="cms-hero-overlay" style={{ justifyContent: OVERLAY_JUSTIFY[alignment] || 'flex-start' }}>
                      <div className="cms-hero-content" style={{ textAlign: alignment }}>
                        {section.heading ? <h1 className="cms-hero-heading">{section.heading}</h1> : null}
                        {section.body ? <p className="cms-hero-body">{section.body}</p> : null}
                        {hasButtons
                          ? renderButtons(section.buttons, alignment, 'cms-cta-btn--lg')
                          : (section.button_label && section.button_link
                            ? <a href={section.button_link} className="cms-hero-cta">{section.button_label}</a>
                            : null)
                        }
                      </div>
                    </div>
                  ) : null}
                </div>
              ) : (
                <div className="cms-section-inner" style={{ ...innerStyle, backgroundColor: bg || undefined, color: textColor }}>
                  {section.heading ? <h2 className="cms-section-heading" style={{ color: textColor }}>{section.heading}</h2> : null}
                  {section.body ? <p className="cms-section-body" style={{ color: textColor }}>{renderMarkdown(section.body)}</p> : null}
                  <div className="cms-img-placeholder cms-img-placeholder--16-9" style={{ marginTop: 16 }}>
                    <span className="cms-img-placeholder__label">{section.image_label || 'Image placeholder'}</span>
                  </div>
                  {hasButtons
                    ? <div style={{ marginTop: 16 }}>{renderButtons(section.buttons, alignment)}</div>
                    : (section.button_label && section.button_link
                      ? <div style={{ display: 'flex', justifyContent: BTN_ROW_JUSTIFY[alignment] || 'flex-start', marginTop: 16 }}>
                          <a href={section.button_link} className="cms-cta-btn cms-cta-btn--primary">{section.button_label}</a>
                        </div>
                      : null)
                  }
                </div>
              )}
            </section>
          )
        }

        // ── Image Grid ────────────────────────────────────────────────────
        if (section.type === 'image_grid') {
          return (
            <section
              key={index}
              className={`cms-section cms-section--grid${hasColor ? ' cms-section--colored' : ''}`}
              style={{ backgroundColor: bg || undefined, color: textColor }}
            >
              <div className={SECTION_INNER_CLS[section.container_size] || SECTION_INNER_CLS.standard} style={innerStyle}>
                {section.heading ? (
                  <h2 className="cms-section-heading" style={{ color: textColor }}>{section.heading}</h2>
                ) : null}
                <div className="cms-grid">
                  {(section.items || []).map((item, itemIndex) => {
                    const isExternal = /^https?:\/\//i.test(item.link_url)
                    const Tag = item.link_url ? 'a' : 'div'
                    const linkProps = item.link_url
                      ? { href: item.link_url, target: isExternal ? '_blank' : undefined, rel: isExternal ? 'noreferrer' : undefined }
                      : {}
                    return (
                      <Tag key={itemIndex} className="cms-grid-card" {...linkProps}>
                        <div className="cms-grid-img-wrap">
                          {item.image_url ? (
                            <img src={item.image_url} alt={item.title || `Partner ${itemIndex + 1}`} className="cms-grid-img" />
                          ) : (
                            <div className="cms-grid-img-placeholder" />
                          )}
                        </div>
                        {item.title ? <span className="cms-grid-label">{item.title}</span> : null}
                      </Tag>
                    )
                  })}
                </div>
              </div>
            </section>
          )
        }

        // ── Two Column ────────────────────────────────────────────────────
        if (section.type === 'two_column') {
          const isHero = section.variant === 'hero'
          const imgLeft = section.image_side === 'left'
          const aspectClass = ASPECT_CLASS_MAP[section.image_aspect_ratio] || 'cms-img-placeholder--16-9'

          const textCol = (
            <div className="cms-two-col__text">
              {section.heading ? (
                <h2 className="cms-section-heading" style={{ color: textColor }}>{section.heading}</h2>
              ) : null}
              {section.body ? (
                <p className="cms-section-body" style={{ color: textColor }}>{renderMarkdown(section.body)}</p>
              ) : null}
              {(section.bullets || []).length > 0 ? (
                <ul className="cms-bullets">
                  {section.bullets.filter(Boolean).map((item, bi) => (
                    <li key={bi} className="cms-bullet-item" style={{ color: textColor }}>
                      <span className="cms-bullet-icon" style={{ color: hasColor && textColor === '#ffffff' ? '#ffffff' : '#00698f' }}>✓</span>
                      {renderMarkdown(item)}
                    </li>
                  ))}
                </ul>
              ) : null}
              {section.pull_quote ? (
                <blockquote className="cms-pull-quote">{section.pull_quote}</blockquote>
              ) : null}
              {(section.buttons && section.buttons.length > 0)
                ? <div style={{ marginTop: 20 }}>{renderButtons(section.buttons, section.alignment || 'left')}</div>
                : (section.button_label && section.button_link
                  ? <div className="cms-btn-row" style={{ alignItems: 'flex-start', marginTop: 20 }}>
                      <a href={section.button_link} className="cms-cta-btn cms-cta-btn--primary">
                        {section.button_label}
                      </a>
                    </div>
                  : null)
              }
            </div>
          )

          const imgCol = (
            <div className="cms-two-col__img">
              {section.image_url ? (
                <img
                  src={section.image_url}
                  alt={section.image_alt || section.heading || 'Section image'}
                  className="cms-two-col__img-el"
                />
              ) : (
                <div className={`cms-img-placeholder ${aspectClass}`}>
                  <span className="cms-img-placeholder__label">{section.image_label || 'Image placeholder'}</span>
                </div>
              )}
            </div>
          )

          return (
            <section
              key={index}
              className={`cms-section${hasColor ? ' cms-section--colored' : ''}${isHero ? ' cms-section--twocol-hero' : ''}`}
              style={{ backgroundColor: bg || undefined, color: textColor }}
            >
              <div className={SECTION_INNER_CLS[section.container_size] || SECTION_INNER_CLS.standard}>
                <div className={`cms-two-col${isHero ? ' cms-two-col--hero' : ''}${imgLeft ? ' cms-two-col--img-left' : ''}`}>
                  {textCol}
                  {imgCol}
                </div>
              </div>
            </section>
          )
        }

        // ── Content + Image (paired) ──────────────────────────────────────
        if (section.type === 'content_with_image') {
          const inner = section.content_section || { type: 'text' }
          const imgLeft = section.image_side === 'left'
          const aspectClass = ASPECT_CLASS_MAP[section.image_aspect_ratio] || 'cms-img-placeholder--4-3'
          return (
            <section
              key={index}
              className={`cms-section${hasColor ? ' cms-section--colored' : ''}`}
              style={{ backgroundColor: bg || undefined, color: textColor }}
            >
              <div className={SECTION_INNER_CLS[section.container_size] || SECTION_INNER_CLS.standard}>
                <div className={`cms-two-col${imgLeft ? ' cms-two-col--img-left' : ''}`}>
                  <div className="cms-two-col__text">
                    <SectionContent section={inner} textColor={textColor} hasColor={hasColor} embedded />
                  </div>
                  <div className="cms-two-col__img">
                    {section.image_url ? (
                      <img
                        src={section.image_url}
                        alt={section.image_alt || inner.heading || 'Section image'}
                        className="cms-two-col__img-el"
                      />
                    ) : (
                      <div className={`cms-img-placeholder ${aspectClass}`}>
                        <span className="cms-img-placeholder__label">{section.image_label || 'Image placeholder'}</span>
                      </div>
                    )}
                  </div>
                </div>
              </div>
            </section>
          )
        }

        // ── All other types via SectionContent ────────────────────────────
        const typeExtraClass =
          section.type === 'bullets' ? ' cms-section--bullets' :
          section.type === 'buttons' ? ' cms-section--buttons' :
          section.type === 'form_cta' ? ' cms-section--cta' : ''
        return (
          <section
            key={index}
            className={`cms-section${typeExtraClass}${hasColor ? ' cms-section--colored' : ''}`}
            style={{ backgroundColor: bg || undefined, color: textColor }}
          >
            <SectionContent section={section} textColor={textColor} hasColor={hasColor} embedded={false} />
          </section>
        )
      })}
    </div>
  )
}

export function CMSPageView({ slug }) {
  const [page, setPage] = useState(null)
  const [error, setError] = useState('')

  useEffect(() => {
    let mounted = true
    setError('')
    setPage(null)
    getPage(slug)
      .then((data) => { if (mounted) setPage(data) })
      .catch((err) => { if (mounted) setError(err.message) })
    return () => { mounted = false }
  }, [slug])

  const { sections, container_size: containerSize, show_header: showHeader } = useMemo(() => {
    try { return normalizeContent(JSON.parse(page?.content_json || '{}')) }
    catch { return { sections: [], container_size: 'standard', show_header: false } }
  }, [page?.content_json])

  const canonicalPath = SLUG_TO_PATH[slug] ?? '/'
  const canonicalUrl = `${SITE_URL}${canonicalPath}`
  const pageTitle = page ? `${page.title} | Animal Pride Partners` : 'Animal Pride Partners'
  const pageDescription = page?.description || ''

  // Build JSON-LD based on page slug
  const jsonLd = useMemo(() => {
    if (!page) return null

    const base = {
      '@context': 'https://schema.org',
      '@type': 'WebPage',
      name: page.title,
      description: page.description || '',
      url: canonicalUrl,
      publisher: {
        '@type': 'Organization',
        name: 'Animal Pride',
        url: 'https://animalpride.com',
      },
    }

    if (slug === 'partner-faq') {
      // Extract FAQ items from bullets sections for FAQPage schema
      const faqSections = sections.filter((s) => s.type === 'bullets' && s.heading && (s.items || []).length > 0)
      if (faqSections.length > 0) {
        return {
          '@context': 'https://schema.org',
          '@type': 'FAQPage',
          mainEntity: faqSections.flatMap((s) =>
            (s.items || []).filter(Boolean).map((item) => ({
              '@type': 'Question',
              name: s.heading,
              acceptedAnswer: { '@type': 'Answer', text: item },
            }))
          ),
        }
      }
    }

    if (slug === 'partnership-overview') {
      return {
        '@context': 'https://schema.org',
        '@type': 'Organization',
        name: 'Animal Pride Partners',
        url: SITE_URL,
        description: page.description || '',
        logo: `${SITE_URL}/animalpride-fav.png`,
        sameAs: ['https://animalpride.com'],
      }
    }

    return base
  }, [page, slug, sections, canonicalUrl])

  if (error) return <Alert type="error" message={error} showIcon style={{ margin: '24px 0' }} />
  if (!page) return <div style={{ padding: '60px 0', textAlign: 'center' }}><Spin size="large" /></div>

  return (
    <>
      <Helmet>
        <title>{pageTitle}</title>
        {pageDescription ? <meta name="description" content={pageDescription} /> : null}
        <link rel="canonical" href={canonicalUrl} />
        <meta property="og:title" content={pageTitle} />
        {pageDescription ? <meta property="og:description" content={pageDescription} /> : null}
        <meta property="og:url" content={canonicalUrl} />
        <meta property="og:type" content="website" />
        <meta property="og:image" content={DEFAULT_OG_IMAGE} />
        <meta name="twitter:card" content="summary_large_image" />
        <meta name="twitter:title" content={pageTitle} />
        {pageDescription ? <meta name="twitter:description" content={pageDescription} /> : null}
        <meta name="twitter:image" content={DEFAULT_OG_IMAGE} />
        {jsonLd ? (
          <script type="application/ld+json">{JSON.stringify(jsonLd)}</script>
        ) : null}
      </Helmet>
      <div className="cms-page-root">
        {showHeader ? (
          <div className="cms-page-header">
            <Typography.Title level={2} className="cms-page-title">{page.title}</Typography.Title>
            {page.description ? <Typography.Paragraph className="cms-page-desc">{page.description}</Typography.Paragraph> : null}
          </div>
        ) : null}
        <CMSContentRenderer sections={sections} containerSize={containerSize} />
      </div>
    </>
  )
}
