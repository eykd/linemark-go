# Separation & DRY Principles

**Purpose**: Apply prefactoring principles for eliminating duplication and separating concerns.

## When to Use

- Deciding between duplication and abstraction
- Separating business rules from technical logic
- Implementing varying algorithms
- Structuring complex logic

## Core Principles

### Don't Repeat Yourself

Every piece of knowledge has a single, authoritative representation.

```typescript
// Bad: Duplicated validation logic
function createUser(email: string): void {
  if (!email.includes('@')) throw new Error('Invalid');
}
function updateEmail(email: string): void {
  if (!email.includes('@')) throw new Error('Invalid');
}

// Good: Single source of truth
class Email {
  constructor(value: string) {
    if (!value.includes('@')) throw new InvalidEmailError(value);
  }
}

function createUser(email: Email): void {}
function updateEmail(email: Email): void {}
```

### Separate Policy from Implementation

Keep the "what" separate from the "how" for flexibility and clarity.

```typescript
// Bad: Policy mixed with implementation
function calculateDiscount(order: Order): number {
  if (order.total > 1000) return order.total * 0.1;
  if (order.items.length > 10) return order.total * 0.05;
  return 0;
}

// Good: Policy separate from implementation
interface DiscountPolicy {
  applies(order: Order): boolean;
  calculate(order: Order): Money;
}

class BulkOrderDiscount implements DiscountPolicy {
  applies(order: Order): boolean {
    return order.total.exceeds(BULK_THRESHOLD);
  }
  calculate(order: Order): Money {
    return order.total.multiply(0.1);
  }
}

class DiscountCalculator {
  constructor(private policies: DiscountPolicy[]) {}

  calculate(order: Order): Money {
    const applicable = this.policies.find((p) => p.applies(order));
    return applicable?.calculate(order) ?? Money.zero();
  }
}
```

### Avoid Premature Generalization

Solve the specific problem first. Generalize when patterns emerge.

```typescript
// Bad: Premature abstraction before second use case
interface DataProcessor<T, R> {
  process(input: T): R;
  validate(input: T): boolean;
  transform(input: T): T;
}

// Good: Solve specific problem first
function processUserRegistration(data: RegistrationData): User {
  // Implement specifically for this use case
}

// Later, when you have 2-3 similar cases, THEN abstract
function processOrderSubmission(data: OrderData): Order {
  // If pattern emerges, consider shared abstraction
}
```

### Adapt a Prefactoring Attitude

Eliminate duplication before it occurs. Look for patterns during design.

```typescript
// Before implementing feature #2, check if it shares logic with feature #1
// If so, extract shared logic BEFORE implementing #2

// Example: Before adding SMS notifications alongside email
interface NotificationChannel {
  send(recipient: Recipient, message: Message): Promise<void>;
}

// Now both implementations follow same pattern
class EmailChannel implements NotificationChannel {
  /* ... */
}
class SmsChannel implements NotificationChannel {
  /* ... */
}
```

## Decision Matrix

| Situation                          | Apply                          | Action                     |
| ---------------------------------- | ------------------------------ | -------------------------- |
| Logic in multiple places           | DRY                            | Extract to single location |
| Algorithm can vary                 | Policy/Implementation          | Use strategy pattern       |
| Building first implementation      | Avoid Premature Generalization | Keep specific              |
| About to implement similar feature | Prefactoring Attitude          | Extract pattern first      |

## When to Extract vs. Duplicate

```
Rule of Three:
1st occurrence: Just write it
2nd occurrence: Note the duplication
3rd occurrence: Extract the abstraction
```

## Related References

- [naming.md](./naming.md): Naming extracted abstractions
- [architecture.md](./architecture.md): Module-level separation
