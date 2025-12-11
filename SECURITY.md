# Security Policy

## Supported Versions

We release patches for security vulnerabilities. Currently supported versions:

| Version | Supported          |
| ------- | ------------------ |
| 0.x.x   | :white_check_mark: |

## Reporting a Vulnerability

We take security seriously. If you discover a security vulnerability in Tempus, please report it privately.

### How to Report

**DO NOT** open a public GitHub issue for security vulnerabilities.

Instead, please report security issues by:

1. **Email**: Send details to the maintainers (create a private security advisory on GitHub)
2. **GitHub Security Advisories**: Use the "Security" tab in the repository

### What to Include

Please include the following information:

- **Type of vulnerability** (e.g., injection, authentication bypass, etc.)
- **Full paths of source file(s)** related to the vulnerability
- **Location of the affected code** (tag/branch/commit or direct URL)
- **Step-by-step instructions** to reproduce the issue
- **Proof-of-concept or exploit code** (if possible)
- **Impact of the vulnerability** - what can an attacker do?
- **Suggested fix** (if you have one)

### Example Report

```
Subject: [SECURITY] Command Injection in ICS filename handling

Type: Command Injection
Location: main.go, lines 1234-1240
Affected versions: All versions <= 0.5.0

Description:
The filename parameter in the 'create' command is not properly sanitized,
allowing shell command injection through specially crafted filenames.

Steps to reproduce:
1. Run: tempus create --output "test;rm -rf /"
2. The system executes 'rm -rf /' after creating the file

Impact:
An attacker who can control the --output parameter can execute
arbitrary commands with the privileges of the user running Tempus.

Suggested fix:
Sanitize filename input using filepath.Clean() and validate
against a whitelist of allowed characters.
```

## Response Timeline

- **Initial Response**: Within 48 hours of report
- **Confirmation**: Within 5 business days
- **Fix Timeline**: Depends on severity
  - Critical: 1-7 days
  - High: 7-30 days
  - Medium: 30-90 days
  - Low: Next planned release

## Security Best Practices for Users

### OAuth Credentials

**NEVER** commit OAuth credentials to version control:

```bash
# Use environment variables
export TEMPUS_CLIENT_ID="your-client-id.apps.googleusercontent.com"
export TEMPUS_CLIENT_SECRET="your-client-secret"

# Or use a .env file (which is .gitignored)
echo "TEMPUS_CLIENT_ID=..." >> .env
echo "TEMPUS_CLIENT_SECRET=..." >> .env
```

### Token Storage

Google Calendar tokens are stored locally in the file specified by `--token-file`. Keep this file secure:

```bash
# Recommended permissions
chmod 600 ~/.tempus/google_token.json

# Never commit to git
echo "*_token.json" >> .gitignore
```

### Input Validation

Be cautious when using untrusted input in ICS files:

- **Filenames**: Avoid special characters that could be interpreted by shells
- **Descriptions**: Sanitize user-provided text to prevent injection
- **URLs**: Validate attachment URLs before including them

### Running with Least Privilege

Run Tempus with the minimum necessary permissions:

```bash
# Good: normal user
./tempus create ...

# Bad: unnecessary root access
sudo ./tempus create ...  # DON'T DO THIS
```

## Known Security Considerations

### OAuth Device Flow

Tempus uses OAuth 2.0 Device Flow for Google Calendar authentication. This is secure for desktop applications but requires:

1. **User verification** - Always verify the authorization URL matches `accounts.google.com`
2. **Token protection** - Store tokens in user-only readable files (`chmod 600`)
3. **Token rotation** - Tokens expire and are automatically refreshed

### ICS File Generation

Generated ICS files may contain sensitive information:

- Event titles and descriptions
- Attendee email addresses
- Location information
- Notes and metadata

**Recommendation**: Be mindful when sharing ICS files publicly.

### Dependencies

Tempus depends on several third-party Go modules. We regularly:

- Monitor security advisories via `go list -m -u all`
- Update dependencies to patch known vulnerabilities
- Use GitHub's Dependabot for automated security updates

## Security Scanning

This project uses automated security scanning:

- **gosec**: Go security checker (runs weekly)
- **Dependabot**: Dependency vulnerability scanning
- **CodeQL**: Semantic code analysis (on PRs)

## Disclosure Policy

When we receive a security report:

1. **Confirm** the vulnerability and determine affected versions
2. **Develop** a fix and prepare release notes
3. **Release** patched versions for all supported releases
4. **Announce** the vulnerability publicly after patches are available

We will credit the reporter in release notes (unless they request anonymity).

## Comments on This Policy

If you have suggestions for improving this policy, please open an issue or submit a pull request.

## Contact

For security-related questions that are not vulnerabilities, you can open a public issue with the `security` label.
