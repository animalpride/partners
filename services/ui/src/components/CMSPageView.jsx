import { Alert, Spin, Typography } from 'antd'
import { useEffect, useMemo, useState } from 'react'
import { getPage } from '../api'
import { getContrastTextColor, normalizeSections } from '../cmsTypes'




const OVERLAY_JUSTIFY = { left: 'flex-start', center: 'center', right: 'flex-end' }
const BTN_ROW_JUSTIFY  = { left: 'flex-start', center: 'center', right: 'flex-end' }
const BULLET_ALIGN     = { left: 'flex-start', center: 'center', right: 'flex-end' }

export function CMSContentRenderer({ sections }) {
  if (!sections.length) {
    return (
      <div className="cms-empty">
        <Typography.Paragraph type="secondary">No content yet. Use "Edit this page" to add sections.</Typography.Paragraph>
      </div>
    )
  }

  return (
    <div className="cms-page">
      {sections.map((section, index) => {
        const bg = section.background || ''
        const textColor = getContrastTextColor(bg)
        const hasColor = Boolean(bg)
        const alignment = section.alignment || 'left'
        const innerStyle = { textAlign: alignment }

        // ── Image / Hero ────────────────────────────────────────────────────
        if (section.type === 'image') {
          const hasOverlay = section.heading || section.body || (section.button_label && section.button_link)
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
                        {section.button_label && section.button_link ? (
                          <a href={section.button_link} className="cms-hero-cta">
                            {section.button_label}
                          </a>
                        ) : null}
                      </div>
                    </div>
                  ) : null}
                </div>
              ) : (
                <div className="cms-section-inner" style={{ ...innerStyle, backgroundColor: bg || undefined, color: textColor }}>
                  {section.heading ? <h2 className="cms-section-heading" style={{ color: textColor }}>{section.heading}</h2> : null}
                  {section.body ? <p className="cms-section-body" style={{ color: textColor }}>{section.body}</p> : null}
                </div>
              )}
            </section>
          )
        }

        // ── Image Grid ───────────────────────────────────────────────────────
        if (section.type === 'image_grid') {
          return (
            <section
              key={index}
              className={`cms-section cms-section--grid${hasColor ? ' cms-section--colored' : ''}`}
              style={{ backgroundColor: bg || undefined, color: textColor }}
            >
              <div className="cms-section-inner" style={innerStyle}>
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
                      <Tag key={`grid-${itemIndex}`} className="cms-grid-card" {...linkProps}>
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

        // ── Bullets ──────────────────────────────────────────────────────────
        if (section.type === 'bullets') {
          return (
            <section
              key={index}
              className={`cms-section cms-section--bullets${hasColor ? ' cms-section--colored' : ''}`}
              style={{ backgroundColor: bg || undefined, color: textColor }}
            >
              <div className="cms-section-inner" style={{ ...innerStyle, ...(section.heading ? {} : { marginLeft: 0 }) }}>
                {section.heading ? (
                  <h2 className="cms-section-heading" style={{ color: textColor }}>{section.heading}</h2>
                ) : null}
                <ul className="cms-bullets" style={{ alignItems: BULLET_ALIGN[alignment] || 'flex-start' }}>
                  {(section.items || []).filter(Boolean).map((item, itemIndex) => (
                    <li key={`bullet-${itemIndex}`} className="cms-bullet-item" style={{ color: textColor }}>
                      <span className="cms-bullet-icon" style={{ color: hasColor && textColor === '#ffffff' ? '#ffffff' : '#00698f' }}>✓</span>
                      {item}
                    </li>
                  ))}
                </ul>
              </div>
            </section>
          )
        }

        // ── Buttons / Links ──────────────────────────────────────────────────
        if (section.type === 'buttons') {
          return (
            <section
              key={index}
              className={`cms-section cms-section--buttons${hasColor ? ' cms-section--colored' : ''}`}
              style={{ backgroundColor: bg || undefined, color: textColor }}
            >
              <div className="cms-section-inner" style={innerStyle}>
                {section.heading ? (
                  <h2 className="cms-section-heading" style={{ color: textColor }}>{section.heading}</h2>
                ) : null}
                <div className="cms-btn-row" style={{ justifyContent: BTN_ROW_JUSTIFY[alignment] || 'flex-start' }}>
                  {(section.buttons || []).filter((btn) => btn?.label && btn?.url).map((button, buttonIndex) => {
                    const isExternal = /^https?:\/\//i.test(button.url)
                    return (
                      <a
                        key={`btn-${buttonIndex}`}
                        href={button.url}
                        target={isExternal ? '_blank' : undefined}
                        rel={isExternal ? 'noreferrer' : undefined}
                        className={`cms-cta-btn${button.variant === 'primary' ? ' cms-cta-btn--primary' : ' cms-cta-btn--ghost'}`}
                      >
                        {button.label}
                      </a>
                    )
                  })}
                </div>
              </div>
            </section>
          )
        }

        // ── Form CTA ─────────────────────────────────────────────────────────
        if (section.type === 'form_cta') {
          return (
            <section
              key={index}
              className={`cms-section cms-section--cta${hasColor ? ' cms-section--colored' : ''}`}
              style={{ backgroundColor: bg || undefined, color: textColor }}
            >
              <div className="cms-section-inner" style={innerStyle}>
                {section.heading ? (
                  <h2 className="cms-section-heading" style={{ color: textColor }}>{section.heading}</h2>
                ) : null}
                {section.body ? <p className="cms-section-body" style={{ color: textColor }}>{section.body}</p> : null}
                {section.button_label && section.button_link ? (
                  <div style={{ display: 'flex', justifyContent: BTN_ROW_JUSTIFY[alignment] || 'center' }}>
                    <a href={section.button_link} className="cms-cta-btn cms-cta-btn--primary cms-cta-btn--lg">
                      {section.button_label}
                    </a>
                  </div>
                ) : null}
              </div>
            </section>
          )
        }

        // ── Application Form (preview only in renderer) ───────────────────────
        if (section.type === 'application_form') {
          return (
            <section
              key={index}
              className={`cms-section${hasColor ? ' cms-section--colored' : ''}`}
              style={{ backgroundColor: bg || undefined, color: textColor }}
            >
              <div className="cms-section-inner" style={innerStyle}>
                {section.heading ? (
                  <h2 className="cms-section-heading" style={{ color: textColor }}>{section.heading}</h2>
                ) : null}
                <ul className="cms-bullets">
                  {(section.fields || []).map((field, fieldIndex) => (
                    <li key={`field-${fieldIndex}`} className="cms-bullet-item" style={{ color: textColor }}>
                      <span className="cms-bullet-icon" style={{ color: hasColor && textColor === '#ffffff' ? '#ffffff' : '#00698f' }}>✓</span>
                      {field.label || field.name}{field.required ? ' *' : ''}
                    </li>
                  ))}
                </ul>
              </div>
            </section>
          )
        }

        // ── Default: Text ─────────────────────────────────────────────────────
        return (
          <section
            key={index}
            className={`cms-section${hasColor ? ' cms-section--colored' : ''}`}
            style={{ backgroundColor: bg || undefined, color: textColor }}
          >
            <div className="cms-section-inner" style={innerStyle}>
              {section.heading ? (
                <h2 className="cms-section-heading" style={{ color: textColor }}>{section.heading}</h2>
              ) : null}
              {section.body ? (
                <p className="cms-section-body" style={{ whiteSpace: 'pre-wrap', color: textColor }}>{section.body}</p>
              ) : null}
            </div>
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

  const sections = useMemo(() => {
    try { return normalizeSections(JSON.parse(page?.content_json || '{}')) }
    catch { return [] }
  }, [page?.content_json])

  if (error) return <Alert type="error" message={error} showIcon style={{ margin: '24px 0' }} />
  if (!page) return <div style={{ padding: '60px 0', textAlign: 'center' }}><Spin size="large" /></div>

  return (
    <div className="cms-page-root">
      <div className="cms-page-header">
        <Typography.Title level={2} className="cms-page-title">{page.title}</Typography.Title>
        {page.description ? <Typography.Paragraph className="cms-page-desc">{page.description}</Typography.Paragraph> : null}
      </div>
      <CMSContentRenderer sections={sections} />
    </div>
  )
}
