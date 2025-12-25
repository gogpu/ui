# Security Policy

## Supported Versions

gogpu/ui is currently in early development (v0.x.x).

| Version | Supported          |
| ------- | ------------------ |
| 0.1.x   | :white_check_mark: |
| < 0.1.0 | :x:                |

## Reporting a Vulnerability

**DO NOT** open a public GitHub issue for security vulnerabilities.

Instead, please report security issues via:

1. **Private Security Advisory** (preferred):
   https://github.com/gogpu/ui/security/advisories/new

2. **GitHub Discussions** (for less critical issues):
   https://github.com/gogpu/ui/discussions

### What to Include

- Description of the vulnerability
- Steps to reproduce
- Affected versions
- Potential impact

### Response Timeline

- **Initial Response**: Within 72 hours
- **Fix & Disclosure**: Coordinated with reporter

## Security Considerations

gogpu/ui is a GUI library that uses gogpu ecosystem for rendering. Users should be aware of:

1. **User Input** - Widgets handle user input; ensure proper sanitization in your application
2. **Accessibility** - Screen readers receive widget content; avoid exposing sensitive data
3. **Theme Styling** - Custom themes may affect visual security indicators
4. **Dependencies** - gogpu/gg and gogpu/gogpu use native GPU libraries

## Security Contact

- **GitHub Security Advisory**: https://github.com/gogpu/ui/security/advisories/new
- **Public Issues**: https://github.com/gogpu/ui/issues

---

**Thank you for helping keep gogpu/ui secure!**
