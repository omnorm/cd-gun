# Security Policy

## Reporting Security Issues

**Do not open public issues for security vulnerabilities.** Instead, please report them privately.

### Vulnerability Disclosure Process

If you discover a security vulnerability in CD-Gun:

1. **Do not** disclose it publicly or in GitHub issues
2. **Email** the maintainers at: oh.my.ext@gmail.com
3. Include:
   - Description of the vulnerability
   - Steps to reproduce (if applicable)
   - Potential impact and severity
   - Suggested fix (if you have one)

We will:
- Acknowledge receipt within 48 hours
- Work with you to verify and assess the issue
- Develop and test a fix
- Prepare a security release
- Coordinate the public disclosure

### Response Timeline

- **Critical (CVSS 9.0-10.0)**: Fix and release within 7 days
- **High (CVSS 7.0-8.9)**: Fix and release within 30 days  
- **Medium (CVSS 4.0-6.9)**: Fix and release within 90 days
- **Low (CVSS 0.1-3.9)**: Fix in next release

## Security Considerations

### For Users

When using CD-Gun, keep in mind:

- **Secrets Management**: Never hardcode secrets in config files. Use environment variables or secrets management systems (HashiCorp Vault, cloud secret managers, etc.)
- **Script Security**: Review any deployment scripts carefully before adding them to CD-Gun config. Only run scripts from trusted sources
- **Permissions**: Run CD-Gun with minimal necessary privileges. Use systemd service configuration to enforce security boundaries (see `deployments/cd-gun.sudoers`)
- **Git Credentials**: Store git credentials securely:
  - Use SSH keys instead of HTTPS with passwords
  - Use personal access tokens with minimal scopes for HTTPS
  - Never commit credentials to repositories
- **Network Security**: If polling private repositories, ensure your network is properly secured
- **Log Inspection**: Review logs regularly for suspicious activity; enable debug logging only when needed
- **Updates**: Keep CD-Gun updated to receive security patches promptly

### For Contributors

Security best practices in code:

- **Input Validation**: Sanitize all user inputs from config files and environment
- **Code Execution**: Avoid executing arbitrary code without validation; use allowlists for actions
- **Secret Handling**: Don't log secrets or expose them in error messages
- **Dependency Management**: Keep dependencies up to date, run `go mod tidy` regularly
- **Error Messages**: Avoid revealing system details or paths in error messages
- **File Permissions**: Verify permissions before creating state/cache files
- **Dependency Auditing**: Report security issues in dependencies to maintainers immediately

### Known Security Boundaries

- CD-Gun executes shell scripts with the privileges of the systemd service user
- Scripts have access to all environment variables provided in the config
- Scripts can access git repositories through SSH keys available to the service user
- No sandboxing or containerization is performed; use systemd security features instead

## Supported Versions

| Version | Released | Status | Support Until |
|---------|----------|--------|--|
| 0.1.x | 2025-12 | Early Development | TBD |
| 0.2.x+ | Planned | TBD | TBD |

We recommend always using the latest version for security updates. Early development versions (0.1.x) may have limited security support — use with caution in production.

## Dependencies

CD-Gun minimizes external dependencies to reduce attack surface:

```
github.com/omnorm/cd-gun
└── gopkg.in/yaml.v3 (YAML parsing)
```

### Dependency Security

- Dependencies are vendored and reviewed
- We monitor for security updates via `go list -u -m all`
- Critical dependency updates are released immediately
- See [go.sum](go.sum) for the exact versions and hashes

## Security Features in Deployments

When installed via `make install`, CD-Gun runs with security hardening:

- **Non-root user**: Runs as `cd-gun` user (unprivileged)
- **systemd security options** (from `deployments/cd-gun.service`):
  - `NoNewPrivileges=yes` — prevents privilege escalation
  - `PrivateTmp=yes` — private `/tmp` namespace
  - `ProtectSystem=strict` — read-only filesystem
  - `ProtectHome=yes` — no access to home directories
  - `ReadWritePaths=/var/lib/cd-gun` — minimal write access
  - `RestrictRealtime=yes` — no real-time scheduling
  - `RestrictNamespaces=yes` — no namespace manipulation

## Contact

- **Report a vulnerability**: Use GitHub private security advisory
- **General security questions**: Feel free to open an issue with `[SECURITY]` tag
- **More info**: See [CONTRIBUTING.md](CONTRIBUTING.md) for community guidelines
- `gopkg.in/yaml.v3` — YAML parsing

We regularly audit dependencies for security vulnerabilities. Check `go.mod` for the complete list.

---

Thank you for helping keep CD-Gun secure!
