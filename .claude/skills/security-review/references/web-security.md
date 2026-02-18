# Web Security: XSS, CSRF, CSP, Headers

## Table of Contents

- [XSS Prevention](#xss-prevention)
- [CSRF Protection](#csrf-protection)
- [Content Security Policy](#content-security-policy)
- [Security Headers](#security-headers)

## XSS Prevention

### Output Encoding (Primary Defense)

```typescript
class HtmlEncoder {
  private static readonly ENTITIES: Record<string, string> = {
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#x27;',
  };

  static encode(str: string): string {
    return str.replace(/[&<>"']/g, (char) => this.ENTITIES[char]);
  }
}

// Safe tagged template literal
function html(strings: TemplateStringsArray, ...values: unknown[]): string {
  return strings.reduce((result, str, i) => {
    const value = values[i - 1];
    const encoded =
      value == null
        ? ''
        : typeof value === 'string'
          ? HtmlEncoder.encode(value)
          : HtmlEncoder.encode(String(value));
    return result + encoded + str;
  });
}

// Usage
html`<div class="user">${user.name}</div>`; // Auto-encoded
```

### Context-Specific Encoding

| Context        | Encoding Required                      |
| -------------- | -------------------------------------- |
| HTML body      | `&<>"'` → entities                     |
| HTML attribute | `&<>"'` → entities + quote attribute   |
| JavaScript     | JSON.stringify or escape special chars |
| URL            | encodeURIComponent                     |
| CSS            | Avoid user input; whitelist if needed  |

### HTMX Security Configuration

```html
<meta
  name="htmx-config"
  content='{
  "selfRequestsOnly": true,
  "allowScriptTags": false,
  "allowEval": false,
  "historyCacheSize": 0
}'
/>
```

| Setting                  | Purpose                              |
| ------------------------ | ------------------------------------ |
| `selfRequestsOnly: true` | Prevents SSRF via HTMX               |
| `allowScriptTags: false` | Blocks script execution in responses |
| `allowEval: false`       | Disables eval-based handlers         |
| `historyCacheSize: 0`    | Prevents sensitive data caching      |

### Safe Zones for User Content

```html
<div hx-disable>
  ${renderUserContent(content)}
  <!-- HTMX attrs ignored here -->
</div>
```

### Flag These as Critical

- String interpolation without encoding
- `innerHTML` with user data
- `document.write()` with user data
- `eval()` with any external input
- Missing HTMX security config

## CSRF Protection

### Session-Tied CSRF Tokens

```typescript
// Generate token with session
function createSession(userId: string): Session {
  const csrfToken = generateSecureToken(); // 256 bits
  return {
    sessionId: generateSessionId(),
    userId,
    csrfToken, // Stored in session, not separate
  };
}

// Validate in middleware
function validateCsrf(session: Session, token: string): boolean {
  return session.validateCsrfToken(token); // Constant-time compare
}
```

### Include Token in Requests

```html
<!-- Via body attribute for all HTMX requests -->
<body hx-headers='{"X-CSRF-Token": "${csrfToken}"}'>
  <!-- Or via hidden field for forms -->
  <input type="hidden" name="_csrf" value="${csrfToken}" />
</body>
```

### CSRF Middleware

```typescript
async function csrfMiddleware(request: Request, session: Session | null) {
  const method = request.method.toUpperCase();

  // Skip safe methods
  if (['GET', 'HEAD', 'OPTIONS'].includes(method)) return null;

  if (!session) return new Response('Unauthorized', { status: 401 });

  // Extract from header or form
  const token = request.headers.get('X-CSRF-Token') ?? (await extractFromForm(request));

  if (!token || !session.validateCsrfToken(token)) {
    return new Response('CSRF validation failed', { status: 403 });
  }

  // Also validate Origin header
  if (!validateOrigin(request)) {
    return new Response('Invalid origin', { status: 403 });
  }

  return null; // Continue
}
```

### Origin Validation

```typescript
function validateOrigin(request: Request): boolean {
  const origin = request.headers.get('Origin');
  const host = request.headers.get('Host');

  if (!origin) {
    const referer = request.headers.get('Referer');
    if (!referer) return true; // Allow but log
    try {
      return new URL(referer).host === host;
    } catch {
      return false;
    }
  }

  try {
    return new URL(origin).host === host;
  } catch {
    return false;
  }
}
```

### Flag These as High

- State-changing endpoints without CSRF validation
- CSRF tokens not tied to sessions
- Missing Origin/Referer validation
- GET requests that modify state

## Content Security Policy

### Recommended CSP

```typescript
function buildCsp(nonce?: string): string {
  const directives = {
    'default-src': ["'self'"],
    'script-src': ["'self'", nonce ? `'nonce-${nonce}'` : ''],
    'style-src': ["'self'", "'unsafe-inline'"], // Required for Tailwind
    'img-src': ["'self'", 'data:', 'https:'],
    'font-src': ["'self'"],
    'connect-src': ["'self'"],
    'form-action': ["'self'"],
    'frame-ancestors': ["'none'"],
    'base-uri': ["'self'"],
    'object-src': ["'none'"],
  };

  return Object.entries(directives)
    .map(([key, values]) => `${key} ${values.filter(Boolean).join(' ')}`)
    .join('; ');
}
```

### CSP Directive Reference

| Directive         | Purpose                                  |
| ----------------- | ---------------------------------------- |
| `default-src`     | Fallback for all fetches                 |
| `script-src`      | JavaScript sources                       |
| `style-src`       | CSS sources                              |
| `connect-src`     | XHR, fetch, WebSocket                    |
| `frame-ancestors` | Who can embed (replaces X-Frame-Options) |
| `form-action`     | Form submission targets                  |
| `base-uri`        | Allowed `<base>` URLs                    |

### Alpine.js CSP Compatibility

```html
<!-- Option 1: Allow eval (less secure) -->
<meta http-equiv="Content-Security-Policy" content="script-src 'self' 'unsafe-eval'" />

<!-- Option 2: Use CSP build (recommended) -->
<script src="/js/alpine.csp.min.js"></script>
<!-- Then define all x-data in JS, not inline -->
```

### Flag These as Medium

- Missing CSP header
- `unsafe-inline` for scripts
- `unsafe-eval` without justification
- Overly permissive `default-src`

## Security Headers

### Required Headers

```typescript
function addSecurityHeaders(response: Response): Response {
  const headers = new Headers(response.headers);

  // HSTS - force HTTPS
  headers.set('Strict-Transport-Security', 'max-age=63072000; includeSubDomains; preload');

  // Prevent MIME sniffing
  headers.set('X-Content-Type-Options', 'nosniff');

  // Clickjacking protection
  headers.set('X-Frame-Options', 'DENY');

  // Referrer policy
  headers.set('Referrer-Policy', 'strict-origin-when-cross-origin');

  // Permissions policy
  headers.set('Permissions-Policy', 'camera=(), microphone=(), geolocation=()');

  return new Response(response.body, {
    status: response.status,
    headers,
  });
}
```

### Header Checklist

| Header                      | Value                                 | Purpose               |
| --------------------------- | ------------------------------------- | --------------------- |
| `Strict-Transport-Security` | `max-age=63072000; includeSubDomains` | Force HTTPS           |
| `X-Content-Type-Options`    | `nosniff`                             | Prevent MIME sniffing |
| `X-Frame-Options`           | `DENY`                                | Prevent clickjacking  |
| `Content-Security-Policy`   | See above                             | Resource restrictions |
| `Referrer-Policy`           | `strict-origin-when-cross-origin`     | Control referrer      |
| `Permissions-Policy`        | `camera=(), microphone=()`            | Disable features      |

### Cache Control for Authenticated Content

```typescript
// Prevent caching of sensitive responses
headers.set('Cache-Control', 'no-store, no-cache, must-revalidate');
headers.set('Pragma', 'no-cache');
```

### Flag These as Medium

- Missing HSTS
- Short HSTS max-age (< 1 year)
- Missing X-Content-Type-Options
- X-Frame-Options allowing embedding
- Sensitive content without no-store
