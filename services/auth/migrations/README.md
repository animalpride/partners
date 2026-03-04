Place SQL migrations here for core-auth (golang-migrate format).

- Use sequential versions: 000001_init.up.sql / 000001_init.down.sql
- Avoid destructive changes without backups; prefer additive changes.
- Ensure DSN for ap_auth is supplied via env when running migrations.
