# Contributing to CD-Gun

Thank you for your interest in contributing! We appreciate your help.

## How to Contribute

### Reporting Bugs

- Check existing [issues](https://github.com/omnorm/cd-gun/issues) first
- Provide a clear description of the bug
- Include reproduction steps and expected vs. actual behavior
- Mention your OS, Go version, and CD-Gun version

### Suggesting Features

- Open a discussion [issue](https://github.com/omnorm/cd-gun/issues)
- Explain the use case and why it would be useful
- Include examples if relevant

### Submitting Code

1. **Fork** the repository
2. **Create a branch** for your feature/fix:
   ```bash
   git checkout -b feature/my-feature
   ```
3. **Make your changes** and ensure:
   - Code is formatted: `make fmt`
   - Tests pass: `make test`
   - Linter is happy: `make lint`
4. **Commit** with clear messages:
   ```bash
   git commit -m "feat: add support for X" -m "Description of what and why"
   ```
5. **Push** to your fork and **create a Pull Request**

## Code Standards

- Write clear, maintainable Go code
- Add tests for new functionality
- Update documentation if you change behavior
- Follow Go conventions (CamelCase for exported symbols, etc.)
- Use the existing code style as reference

## Development Setup

```bash
# Clone and build
git clone https://github.com/omnorm/cd-gun.git
cd cd-gun

make build
make test

# Run locally
./bin/cd-gun-agent -config examples/simple-deploy.yaml -log-level debug
```

## Pull Request Process

- Link to related issues
- Describe what you changed and why
- Keep commits clean and logical
- Be responsive to review feedback

## Questions?

Open an issue with the `question` label or start a discussion.

---

Thank you for making CD-Gun better! 🚀
