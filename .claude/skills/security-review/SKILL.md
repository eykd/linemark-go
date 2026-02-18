---
name: security-review
description: Review code for security vulnerabilities and best practices. Use when (1) reviewing code for security issues, (2) auditing authentication/session handling, (3) checking for XSS/CSRF/SQL injection vulnerabilities, (4) evaluating security headers and CSP, (5) validating input handling and output encoding, (6) assessing password hashing and secrets management, (7) reviewing rate limiting and brute force protection, or (8) general security hardening of web applications.
---

# Security Review Skill

Systematic security review following OWASP guidelines and defense-in-depth principles.

## Audience: Non-Technical Managers

**CRITICAL**: Write the ENTIRE review for non-technical managers, not developers.

- Technical jargon is OK when explained briefly and concisely
- Always follow technical terms with plain English explanation (e.g., "XSS (cross-site scripting) - injecting malicious code into web pages")
- Explain security risks in business terms (data breach, financial loss, reputation damage)
- Focus on impacts and consequences, not implementation details
- Keep explanations concise - don't over-explain
- Keep all sections accessible to non-technical readers

## Review Process

1. **Identify security surface**: Authentication, data handling, user input, external APIs
2. **Check each domain** using references below
3. **Prioritize findings**: Critical > High > Medium > Low
4. **Provide actionable fixes** with code examples

## Security Domains

### Authentication & Sessions

See [references/auth-security.md](references/auth-security.md) for password hashing (Argon2id), session management (`__Host-` cookies), account lockout, constant-time comparisons.

### Web Security (XSS/CSRF/Headers)

See [references/web-security.md](references/web-security.md) for XSS prevention, CSRF tokens, Content Security Policy, security headers.

### Data Security (Injection/Validation)

See [references/data-security.md](references/data-security.md) for SQL injection prevention, input validation, parameterized queries.

### Quick Checklist

See [references/checklist.md](references/checklist.md) for rapid security audit.

## Critical Patterns to Flag

### Always Critical

```typescript
// SQL Injection - string interpolation
db.prepare(`SELECT * FROM users WHERE id = '${userId}'`) // ❌
// Missing output encoding
`<div>${userInput}</div>`; // ❌ (without safe template)

// Weak password hashing
bcrypt.hash(password, 10); // ❌ Use Argon2id
md5(password); // ❌
sha256(password); // ❌ (no salt)

// Hardcoded secrets
const API_KEY = 'sk-abc123...'; // ❌
```

### Always High

```typescript
// Missing CSRF on state-changing endpoints
app.post('/api/transfer', handler)  // ❌ (no CSRF)

// Insecure cookies
res.cookie('session', id)  // ❌ (missing flags)

// Missing rate limiting on auth
app.post('/login', handler)  // ❌

// Timing-vulnerable comparisons
if (token === storedToken)  // ❌ (use constant-time)
```

## Secure Patterns

### Safe HTML Templating

```typescript
function html(strings: TemplateStringsArray, ...values: unknown[]): string {
  return strings.reduce((result, str, i) => {
    const value = values[i - 1];
    return result + encodeHtml(value) + str;
  });
}
html`<div>${user.name}</div>`; // ✅ Auto-encoded
```

### Parameterized Queries

```typescript
await db.prepare('SELECT * FROM users WHERE id = ?').bind(userId).first(); // ✅
```

### Secure Cookies

```typescript
`__Host-session=${sessionId}; HttpOnly; Secure; SameSite=Lax; Path=/`; // ✅
```

### Constant-Time Compare

```typescript
function constantTimeEqual(a: string, b: string): boolean {
  if (a.length !== b.length) return false;
  let result = 0;
  for (let i = 0; i < a.length; i++) {
    result |= a.charCodeAt(i) ^ b.charCodeAt(i);
  }
  return result === 0;
} // ✅
```

## Review Output Format

```markdown
## Security Review: [Component]

### Critical Issues

1. **[Issue]** - [File:Line]
   - Problem: [Description in plain English]
   - Risk: [Business impact - data breach, financial loss, etc.]
   - Fix: [What needs to be done]

### High Priority

...

### Recommendations

...

---

## Copy-Paste Prompt for Claude Code

**REQUIRED when findings exist**: Provide a ready-to-use prompt in a code block.
```

[Specific, actionable prompt with file paths and line numbers that addresses all Critical and High Priority items]

```

```

## Framework-Specific

### HTMX Security

```html
<meta
  name="htmx-config"
  content='{
  "selfRequestsOnly": true,
  "allowScriptTags": false,
  "allowEval": false
}'
/>
<div hx-disable>${userContent}</div>
<!-- Safe zone -->
```

### Cloudflare Workers

- Use `wrangler secret` for sensitive values
- Verify KV TTL for sessions
- Check D1 uses parameterized queries

### Node.js/Express

- Verify `helmet` middleware
- Check `express-rate-limit`
- Validate CSRF protection
