# Data Security: Injection & Validation

## Table of Contents

- [SQL Injection Prevention](#sql-injection-prevention)
- [Input Validation](#input-validation)
- [Mass Assignment Protection](#mass-assignment-protection)
- [Secrets Management](#secrets-management)

## SQL Injection Prevention

### Parameterized Queries (Required)

```typescript
// ✅ CORRECT - Parameter binding
const user = await db.prepare('SELECT * FROM users WHERE id = ?').bind(userId).first();

// ✅ CORRECT - Multiple parameters
const tasks = await db
  .prepare('SELECT * FROM tasks WHERE user_id = ? AND status = ?')
  .bind(userId, status)
  .all();

// ❌ CRITICAL - String interpolation
const user = await db.prepare(`SELECT * FROM users WHERE id = '${userId}'`).first();

// ❌ CRITICAL - Template literal
const user = await db.prepare(`SELECT * FROM users WHERE email = '${email}'`).first();
```

### Dynamic Query Building

```typescript
// Safe: Dynamic conditions with parameters
function buildQuery(filters: TaskFilters): { sql: string; params: unknown[] } {
  const conditions: string[] = ['user_id = ?'];
  const params: unknown[] = [filters.userId];

  if (filters.status) {
    conditions.push('status = ?');
    params.push(filters.status);
  }

  if (filters.createdAfter) {
    conditions.push('created_at > ?');
    params.push(filters.createdAfter.toISOString());
  }

  return {
    sql: `SELECT * FROM tasks WHERE ${conditions.join(' AND ')} LIMIT ?`,
    params: [...params, filters.limit ?? 50],
  };
}
```

### Safe Dynamic Column Names

```typescript
// Allowlist pattern for column names
const ALLOWED_COLUMNS = new Set(['created_at', 'updated_at', 'title', 'status']);
const ALLOWED_DIRECTIONS = new Set(['ASC', 'DESC']);

function buildOrderClause(column: string, direction: string): string {
  if (!ALLOWED_COLUMNS.has(column)) {
    throw new Error(`Invalid column: ${column}`);
  }
  if (!ALLOWED_DIRECTIONS.has(direction.toUpperCase())) {
    throw new Error(`Invalid direction: ${direction}`);
  }
  // Safe to interpolate after validation
  return `ORDER BY ${column} ${direction.toUpperCase()}`;
}
```

### Flag These as Critical

- Any string interpolation in SQL
- Template literals with user input in queries
- Dynamic table/column names without allowlist
- Raw SQL execution without parameterization

## Input Validation

### Validation Framework

```typescript
type ValidationResult<T> =
  | { success: true; value: T }
  | { success: false; errors: ValidationError[] };

interface ValidationError {
  field: string;
  message: string;
  code: string;
}

class StringValidator {
  constructor(
    private field: string,
    private config: {
      minLength?: number;
      maxLength?: number;
      pattern?: RegExp;
      required?: boolean;
    } = {}
  ) {}

  validate(input: unknown): ValidationResult<string> {
    const errors: ValidationError[] = [];

    if (typeof input !== 'string') {
      return {
        success: false,
        errors: [
          {
            field: this.field,
            message: 'Must be a string',
            code: 'INVALID_TYPE',
          },
        ],
      };
    }

    const trimmed = input.trim();

    if (this.config.required && !trimmed) {
      errors.push({ field: this.field, message: 'Required', code: 'REQUIRED' });
    }
    if (this.config.minLength && trimmed.length < this.config.minLength) {
      errors.push({
        field: this.field,
        message: `Min ${this.config.minLength} chars`,
        code: 'TOO_SHORT',
      });
    }
    if (this.config.maxLength && trimmed.length > this.config.maxLength) {
      errors.push({
        field: this.field,
        message: `Max ${this.config.maxLength} chars`,
        code: 'TOO_LONG',
      });
    }
    if (this.config.pattern && !this.config.pattern.test(trimmed)) {
      errors.push({ field: this.field, message: 'Invalid format', code: 'INVALID_FORMAT' });
    }

    return errors.length ? { success: false, errors } : { success: true, value: trimmed };
  }
}
```

### Allowlist Validation

```typescript
// For fixed-set inputs (select boxes, radio buttons)
class AllowlistValidator<T extends string> {
  constructor(
    private field: string,
    private allowed: readonly T[]
  ) {}

  validate(input: unknown): ValidationResult<T> {
    if (typeof input !== 'string') {
      return {
        success: false,
        errors: [
          {
            field: this.field,
            message: 'Must be a string',
            code: 'INVALID_TYPE',
          },
        ],
      };
    }

    if (!this.allowed.includes(input as T)) {
      // Log as security event - shouldn't happen with valid client
      console.warn('Allowlist violation', { field: this.field, value: input });
      return {
        success: false,
        errors: [
          {
            field: this.field,
            message: 'Invalid selection',
            code: 'NOT_ALLOWED',
          },
        ],
      };
    }

    return { success: true, value: input as T };
  }
}

// Usage
const STATUSES = ['pending', 'active', 'completed'] as const;
const statusValidator = new AllowlistValidator('status', STATUSES);
```

### Email Validation

```typescript
class EmailValidator extends StringValidator {
  private static PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

  constructor(field: string) {
    super(field, { maxLength: 254, pattern: EmailValidator.PATTERN });
  }

  validate(input: unknown): ValidationResult<string> {
    const result = super.validate(input);
    if (!result.success) return result;

    // Normalize for storage (prevent duplicate accounts)
    const normalized = result.value.toLowerCase();
    return { success: true, value: normalized };
  }
}
```

### Validation in Handlers

```typescript
async function handleCreateUser(request: Request): Promise<Response> {
  const body = await request.json();

  const validation = createUserValidator.validate(body);

  if (!validation.success) {
    // Return errors for HTMX to display
    return new Response(renderValidationErrors(validation.errors), {
      status: 422,
      headers: { 'Content-Type': 'text/html' },
    });
  }

  // Proceed with validated, typed data
  const user = await createUser(validation.value);
  return new Response(renderSuccess(user));
}
```

### Flag These as High

- Missing server-side validation
- Client-side only validation
- Type coercion without validation
- Missing length limits on strings

## Mass Assignment Protection

### Explicit Field Selection

```typescript
// ❌ WRONG - Updates any field client sends
async function updateUser(id: string, data: Record<string, unknown>) {
  await db.prepare('UPDATE users SET ...').bind(...Object.values(data));
}

// ✅ CORRECT - Explicit allowed fields
interface UpdateUserRequest {
  name?: string;
  email?: string;
  // Note: no 'role' or 'isAdmin' fields
}

async function updateUser(id: string, data: UpdateUserRequest) {
  const updates: string[] = [];
  const params: unknown[] = [];

  if (data.name !== undefined) {
    updates.push('name = ?');
    params.push(data.name);
  }
  if (data.email !== undefined) {
    updates.push('email = ?');
    params.push(data.email);
  }

  if (updates.length === 0) return;

  params.push(id);
  await db
    .prepare(`UPDATE users SET ${updates.join(', ')} WHERE id = ?`)
    .bind(...params)
    .run();
}
```

### DTO Pattern

```typescript
// Define exactly what's accepted
interface CreateTaskDto {
  title: string;
  description: string;
  status: 'pending' | 'active';
  // Explicitly omit: userId, createdAt, id (set server-side)
}

function createTask(dto: CreateTaskDto, userId: string): Task {
  return {
    id: generateId(), // Server-controlled
    userId, // From session
    createdAt: new Date(), // Server-controlled
    ...dto, // Only allowed fields
  };
}
```

### Flag These as High

- Spreading request body into entities
- Dynamic updates from user input
- Missing field allowlists

## Secrets Management

### Environment Variables

```typescript
// ✅ CORRECT - Access via env binding
export default {
  async fetch(request: Request, env: Env) {
    const apiKey = env.API_SECRET; // From wrangler secret
  },
};

// ❌ CRITICAL - Hardcoded secrets
const API_KEY = 'sk-abc123...';
const DB_PASSWORD = 'password123';
```

### Cloudflare Workers

```bash
# Set secrets via CLI (not in wrangler.toml)
wrangler secret put API_SECRET
wrangler secret put JWT_SECRET
```

### wrangler.toml

```toml
[vars]
# Only non-sensitive configuration
ENVIRONMENT = "production"
LOG_LEVEL = "info"

# Never put secrets here!
```

### Flag These as Critical

- Secrets in source code
- Secrets in config files committed to git
- API keys in client-side code
- Secrets in error messages or logs

### Secrets Checklist

- [ ] No hardcoded secrets
- [ ] Secrets in environment variables
- [ ] Secrets not logged
- [ ] Secrets not in error responses
- [ ] Different secrets per environment
- [ ] Secret rotation capability
