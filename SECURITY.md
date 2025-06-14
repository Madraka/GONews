# 🔒 Security Policy

## 🛡️ Supported Versions

We release patches for security vulnerabilities. Which versions are eligible for receiving such patches depends on the CVSS v3.0 Rating:

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |
| < 1.0   | :x:                |

## 🚨 Reporting a Vulnerability

We take the security of GONews seriously. If you believe you have found a security vulnerability, please report it to us as described below.

### 📧 Where to Report

**Please do NOT report security vulnerabilities through public GitHub issues.**

Instead, please report them via:

1. **GitHub Security Advisories**: Use the [Security Advisories](https://github.com/Madraka/GONews/security/advisories) feature
2. **Email**: Send a detailed report to [security@yourdomain.com] (Replace with actual email)
3. **Private Issue**: Contact the maintainers directly through GitHub

### 📋 What to Include

Please include the following information in your report:

- **Description** of the vulnerability
- **Steps to reproduce** the vulnerability
- **Potential impact** of the vulnerability
- **Possible mitigations** or workarounds
- **Your contact information** for follow-up questions

### 🔍 Example Report

```
Subject: [SECURITY] SQL Injection vulnerability in article search

Description:
The article search functionality is vulnerable to SQL injection attacks
through the 'query' parameter.

Steps to Reproduce:
1. Navigate to /api/search
2. Send POST request with payload: {"query": "'; DROP TABLE articles; --"}
3. Observe that the query is executed without sanitization

Impact:
This could allow attackers to:
- Extract sensitive data from the database
- Modify or delete data
- Potentially gain system access

Suggested Fix:
Use parameterized queries instead of string concatenation
```

## ⏱️ Response Timeline

- **Acknowledgment**: We will acknowledge receipt of your vulnerability report within 48 hours
- **Initial Assessment**: We will provide an initial assessment within 5 business days
- **Regular Updates**: We will provide regular updates on our progress
- **Resolution**: We aim to resolve critical vulnerabilities within 30 days

## 🎯 Vulnerability Assessment

We use the following criteria to assess vulnerabilities:

### 🔴 Critical (CVSS 9.0-10.0)
- Remote code execution
- SQL injection with data access
- Authentication bypass

### 🟠 High (CVSS 7.0-8.9)
- Cross-site scripting (XSS)
- Local file inclusion
- Privilege escalation

### 🟡 Medium (CVSS 4.0-6.9)
- Information disclosure
- Cross-site request forgery (CSRF)
- Insecure direct object references

### 🟢 Low (CVSS 0.1-3.9)
- Missing security headers
- Information leakage
- Minor configuration issues

## 🏆 Recognition

We appreciate security researchers who help improve our security. Eligible reporters may receive:

- **Public acknowledgment** in our security advisories (if desired)
- **Hall of Fame** mention in our documentation
- **Swag** for significant findings (when available)

## 🛠️ Security Best Practices

### For Users
- Keep your GONews installation up to date
- Use strong passwords and enable 2FA when available
- Regularly review access logs
- Use HTTPS in production
- Keep dependencies updated

### For Developers
- Follow secure coding practices
- Use parameterized queries
- Validate all input
- Implement proper authentication and authorization
- Use HTTPS for all communications
- Regularly update dependencies
- Enable security headers
- Use environment variables for secrets

## 🔐 Security Features

GONews includes several security features:

- **JWT Authentication** with configurable expiration
- **Rate Limiting** to prevent abuse
- **Input Validation** on all endpoints
- **SQL Injection Protection** through ORM
- **CORS Configuration** for cross-origin requests
- **Security Headers** (HSTS, CSP, etc.)
- **Environment-based Configuration** for secrets

## 📚 Security Resources

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Go Security Checklist](https://github.com/securecodewarrior/go-security-checklist)
- [Docker Security Best Practices](https://docs.docker.com/engine/security/)
- [PostgreSQL Security](https://www.postgresql.org/docs/current/security.html)

## 📞 Contact

For any questions about this security policy, please contact:

- **Project Maintainer**: [GitHub Profile](https://github.com/Madraka)
- **Security Email**: security@yourdomain.com (Replace with actual email)
- **General Questions**: [GitHub Issues](https://github.com/Madraka/GONews/issues)

---

**Thank you for helping keep GONews and our users safe!** 🙏
