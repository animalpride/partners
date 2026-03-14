import { Button, Typography } from "antd";

const DEFAULT_MESSAGE =
  "The partners.animalpride.com experience is being redesigned to better serve our partners. Please allow us a bit of time to put that together.";

export function ComingSoonPage({ message }) {
  return (
    <main className="coming-soon-shell" aria-live="polite">
      <div className="coming-soon-aurora" />
      <div className="coming-soon-grid" />
      <section className="coming-soon-card">
        <img
          src="/Logo-Wordmark-Dog-PartnersPlatform2.png"
          alt="Animal Pride Partners"
          className="coming-soon-logo"
        />
        <p className="coming-soon-kicker">Partners Platform Update</p>
        <Typography.Title className="coming-soon-title" level={1}>
          A New Partner Experience Is Coming Soon
        </Typography.Title>
        <Typography.Paragraph className="coming-soon-message">
          {message || DEFAULT_MESSAGE}
        </Typography.Paragraph>
        <Typography.Paragraph className="coming-soon-meta">
          We are rebuilding key workflows and content to make partner operations
          faster, clearer, and more reliable.
        </Typography.Paragraph>
        <Button
          type="primary"
          size="large"
          href="mailto:partners@animalpride.com"
        >
          Contact Partner Team
        </Button>
      </section>
    </main>
  );
}
