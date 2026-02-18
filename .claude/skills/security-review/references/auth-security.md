# Authentication & Session Security

## Table of Contents

- [Password Hashing](#password-hashing)
- [Session Management](#session-management)
- [Brute Force Protection](#brute-force-protection)
- [Timing Attack Prevention](#timing-attack-prevention)

## Password Hashing

### Required: Argon2id

OWASP 2025 recommendation. Protects against GPU and side-channel attacks.

```typescript
import { argon2id } from '@noble/hashes/argon2';
import { randomBytes } from '@noble/hashes/utils';

const config = {
  memoryCost: 19456, // 19 MiB minimum
  timeCost: 2,
  parallelism: 1,
  hashLength: 32,
  saltLength: 16,
};

async function hashPassword(password: string): Promise<string> {
  const salt = randomBytes(config.saltLength);
  const hash = argon2id(password, salt, {
    m: config.memoryCost,
    t: config.timeCost,
    p: config.parallelism,
    dkLen: config.hashLength,
  });
  const saltB64 = btoa(String.fromCharCode(...salt));
  const hashB64 = btoa(String.fromCharCode(...hash));
  return `$argon2id$v=19$m=${config.memoryCost},t=${config.timeCost},p=${config.parallelism}$${saltB64}$${hashB64}`;
}
```

### Flag These as Critical

- `md5()`, `sha1()`, `sha256()` without salt
- `bcrypt` (use Argon2id instead)
- Plain text passwords
- Reversible encryption for passwords

### Password Requirements

- Minimum 12 characters (NIST 800-63B)
- Maximum 128 characters
- Check against common password list
- No complexity requirements (per NIST)

## Session Management

### Secure Cookie Pattern

```typescript
const cookie = [
  `__Host-session=${sessionId}`, // __Host- prefix required
  'HttpOnly', // No JavaScript access
  'Secure', // HTTPS only
  'SameSite=Lax', // CSRF protection
  'Path=/', // Required for __Host-
  `Max-Age=${SESSION_TTL}`,
].join('; ');
```

### Session ID Requirements

- 256+ bits cryptographic randomness
- Generated server-side only
- Regenerate on privilege changes (login, role change)

```typescript
function generateSessionId(): string {
  const bytes = crypto.getRandomValues(new Uint8Array(32));
  return Array.from(bytes, (b) => b.toString(16).padStart(2, '0')).join('');
}
```

### Session Storage (KV Example)

```typescript
// Store with TTL for automatic expiration
await env.SESSIONS.put(
  `session:${sessionId}`,
  JSON.stringify({ userId, csrfToken, createdAt }),
  { expirationTtl: 86400 } // 24 hours
);
```

### Flag These as High

- Session IDs in URLs
- Missing `HttpOnly` flag
- Missing `Secure` flag
- `SameSite=None` without justification
- Long session lifetimes without refresh

## Brute Force Protection

### Account Lockout

```typescript
class User {
  static readonly MAX_FAILED_ATTEMPTS = 5;
  static readonly LOCK_DURATION_MS = 15 * 60 * 1000;

  isLocked(): boolean {
    if (!this.lockedUntil) return false;
    return new Date() < new Date(this.lockedUntil);
  }

  recordFailedLogin(): User {
    const attempts = this.failedLoginAttempts + 1;
    const shouldLock = attempts >= User.MAX_FAILED_ATTEMPTS;
    return new User({
      ...this.data,
      failedLoginAttempts: attempts,
      lockedUntil: shouldLock ? new Date(Date.now() + User.LOCK_DURATION_MS).toISOString() : null,
    });
  }
}
```

### Rate Limiting (Sliding Window)

```typescript
interface RateLimitResult {
  allowed: boolean;
  remaining: number;
  resetAt: Date;
}

async function checkRateLimit(
  key: string,
  limit: number,
  windowMs: number
): Promise<RateLimitResult> {
  const now = Date.now();
  const windowStart = now - windowMs;
  // Sliding window implementation...
}

// Apply to auth endpoints
const ipLimit = await checkRateLimit(`ip:${clientIp}`, 10, 60000);
const accountLimit = await checkRateLimit(`account:${email}`, 5, 300000);
```

### Flag These as High

- Auth endpoints without rate limiting
- No account lockout after failed attempts
- Missing IP-based rate limiting
- Lockout bypass via password reset

## Timing Attack Prevention

### Constant-Time Comparison

```typescript
function constantTimeEqual(a: string, b: string): boolean {
  if (a.length !== b.length) return false;
  let result = 0;
  for (let i = 0; i < a.length; i++) {
    result |= a.charCodeAt(i) ^ b.charCodeAt(i);
  }
  return result === 0;
}

// For Uint8Array
function constantTimeEqualBytes(a: Uint8Array, b: Uint8Array): boolean {
  if (a.length !== b.length) return false;
  let result = 0;
  for (let i = 0; i < a.length; i++) {
    result |= a[i] ^ b[i];
  }
  return result === 0;
}
```

### Use For

- Password verification
- CSRF token validation
- API key comparison
- Session ID comparison
- Any security-sensitive string comparison

### Flag These as High

- `===` or `==` for tokens/secrets
- Early return on mismatch
- `String.prototype.localeCompare()`

## Account Enumeration Prevention

### Generic Error Messages

```typescript
// ❌ Wrong - reveals valid accounts
if (!user) return error('User not found');
if (!validPassword) return error('Invalid password');

// ✅ Correct - generic message
if (!user || !validPassword) {
  return error('Invalid email or password');
}
```

### Consistent Timing

```typescript
async function login(email: string, password: string) {
  const user = await findUser(email);

  // Always hash even if user not found (prevents timing leak)
  const dummyHash = '$argon2id$v=19$m=19456,t=2,p=1$...';
  const hashToVerify = user?.passwordHash ?? dummyHash;

  const valid = await verifyPassword(password, hashToVerify);

  if (!user || !valid) {
    return error('Invalid email or password');
  }
}
```
