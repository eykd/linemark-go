# Naming & Code Communication

**Purpose**: Apply prefactoring principles for clear naming, self-documenting code, and explicit intent.

## When to Use

- Naming functions, classes, or variables
- Making code self-documenting
- Choosing between implicit and explicit behavior
- Deciding whether to build or reuse

## Core Principles

### A Rose by Any Other Name

Each concept gets one clear, consistent name from the domain language.

```typescript
// Bad: Inconsistent naming
function getUser() {}
function fetchCustomer() {} // Same concept, different name
function retrievePerson() {} // Same concept, different name

// Good: Consistent ubiquitous language
function findUser() {}
function findUserById() {}
function findUserByEmail() {}
```

### Communicate with Your Code

Code should communicate intent without requiring comments.

```typescript
// Bad: Comment explains unclear code
// Check if user can access the resource
if (user.role === 'admin' || user.id === resource.ownerId) {
}

// Good: Self-documenting code
const isAdmin = user.hasRole(Role.ADMIN);
const isOwner = resource.isOwnedBy(user);
if (isAdmin || isOwner) {
}
```

### Explicitness Beats Implicitness

State intent clearly. Avoid magic behavior.

```typescript
// Bad: Implicit behavior
function processOrder(order: Order, options?: object): void {
  const shouldNotify = (options as any)?.notify !== false; // Implicit default
}

// Good: Explicit parameters
function processOrder(order: Order, options: { notify: boolean }): void {
  if (options.notify) {
    /* ... */
  }
}
```

### Don't Reinvent the Wheel

Use existing solutions before creating new ones.

```typescript
// Bad: Custom date formatting
function formatDate(date: Date): string {
  return `${date.getMonth() + 1}/${date.getDate()}/${date.getFullYear()}`;
}

// Good: Use established library
import { format } from 'date-fns';
const formatted = format(date, 'MM/dd/yyyy');
```

## Decision Matrix

| Situation                     | Apply                 | Action                   |
| ----------------------------- | --------------------- | ------------------------ |
| Same concept, different names | Ubiquitous Language   | Standardize naming       |
| Code needs explanation        | Communicate with Code | Rename for clarity       |
| Magic defaults                | Explicitness          | Make parameters explicit |
| Standard problem              | Don't Reinvent        | Use existing library     |

## Naming Checklist

When naming, ask:

- [ ] Does it use domain language? (not technical jargon)
- [ ] Is it consistent with similar concepts?
- [ ] Does it reveal intent? (not implementation)
- [ ] Is it searchable? (avoid abbreviations)

```typescript
// Domain language examples
class Order {} // Not: DataRecord, Entity
class Money {} // Not: NumberWrapper, AmountValue
function placeOrder() {} // Not: processData, handleRequest
```

## Related References

- [separation.md](./separation.md): Code structure and DRY
- [value-objects.md](./value-objects.md): Naming domain types
