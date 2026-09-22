# Security Policy

## Reporting a vulnerability

Please report security issues **privately** rather than opening a public issue.

Because this is a self-hosted panel that manages network infrastructure and
handles authentication, take the usual precautions with your own deployment:

- Change every default credential before exposing the panel to a network.
- Don't reuse passwords between dashGO and other services.
- Put it behind a reverse proxy with TLS.
- Keep it off the public internet unless you specifically need it there.

## Hardening checklist for self-hosters

- [ ] Admin password changed from the default
- [ ] JWT secret regenerated (not the shipped default)
- [ ] Database not world-readable
- [ ] Deployed behind TLS
