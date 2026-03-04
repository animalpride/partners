# core-auth

Placeholder for the authentication/RBAC service. Intended stack:

- Go + Gin for HTTP routing
- JWT for auth tokens (HS256 or RS256)
- SQL-first migrations via golang-migrate (see migrations/)
- Configuration pulled from a shared main-config (plus secrets override) rather than per-service files

Suggested layout (to be created when implementing):

- cmd/core-auth/main.go
- internal/config
- internal/routes
- internal/middleware (JWT/RBAC)
- internal/repository (GORM or sqlx)
- internal/models

Read only the portions you need from main-config (mounted or injected) and prefer secrets over config values. Startup should fail if required DB/JWT settings are absent.

## User Invitation Flow

DenOps uses an invite-only registration flow. Admins can invite users from the UI (Settings → Users) or via the auth API. Invitations are single-use, expire after 48 hours, and apply a role at registration time. The invite link opens `/accept-invitation` in the UI for token validation and account creation.

Email delivery uses the shared [development/main-config.yml](development/main-config.yml) SMTP relay settings with optional TLS/auth flags. Update `email.links.invite_base_url` and `email.links.reset_base_url` to point at your UI host so invitation and reset links resolve correctly.
