# Value Objects & Type Design

**Purpose**: Apply prefactoring principles when wrapping primitives, grouping data, and creating meaningful types.

## When to Use

- Creating new types for domain concepts
- Wrapping primitive values (strings, numbers)
- Grouping related parameters into objects
- Extracting magic numbers/strings to constants

## Core Principles

### Be Abstract All the Way

Never use primitives for domain concepts. Wrap them in meaningful types.

```typescript
// Bad: Primitives lose domain meaning
function createUser(email: string, age: number): void {}

// Good: Domain types with validation
class Email {
  constructor(private readonly value: string) {
    if (!value.includes('@')) throw new InvalidEmailError(value);
  }
  toString(): string {
    return this.value;
  }
}

class Age {
  constructor(private readonly years: number) {
    if (years < 0 || years > 150) throw new InvalidAgeError(years);
  }
}

function createUser(email: Email, age: Age): void {}
```

### Clump Data

Group related values into cohesive objects to reduce cognitive load.

```typescript
// Bad: Parameter explosion
function placeOrder(
  street: string,
  city: string,
  zip: string,
  country: string,
  cardNumber: string,
  expiry: string,
  cvv: string
): void {}

// Good: Cohesive value objects
class Address {
  constructor(
    readonly street: string,
    readonly city: string,
    readonly zip: string,
    readonly country: string
  ) {}
}

class PaymentDetails {
  constructor(
    readonly cardNumber: string,
    readonly expiry: string,
    readonly cvv: string
  ) {}
}

function placeOrder(address: Address, payment: PaymentDetails): void {}
```

### Never Let a Constant Slip

Use named constants for all meaningful values.

```typescript
// Bad: Magic numbers
if (retryCount > 3) {
  /* give up */
}
if (order.total > 1000) {
  /* apply discount */
}

// Good: Named constants
const MAX_RETRY_ATTEMPTS = 3;
const BULK_ORDER_THRESHOLD = 1000;

if (retryCount > MAX_RETRY_ATTEMPTS) {
  /* give up */
}
if (order.total > BULK_ORDER_THRESHOLD) {
  /* apply discount */
}
```

### Splitters Can Be Lumped

Start with fine-grained abstractions. It's easier to combine than to split.

```typescript
// Start specific, generalize later
class EmailNotification {
  /* email-specific */
}
class SmsNotification {
  /* sms-specific */
}

// Later, if needed, create common abstraction
interface Notification {
  send(recipient: Recipient, message: Message): Promise<void>;
}
```

## Decision Matrix

| Situation             | Apply                   | Example                          |
| --------------------- | ----------------------- | -------------------------------- |
| String with format    | Be Abstract All the Way | `Email`, `Url`, `PhoneNumber`    |
| Number with units     | Be Abstract All the Way | `Money`, `Duration`, `Distance`  |
| Related parameters    | Clump Data              | `Address`, `DateRange`           |
| Literal value         | Named Constants         | `MAX_RETRIES`, `DEFAULT_TIMEOUT` |
| Uncertain granularity | Splitters Can Be Lumped | Start specific                   |

## Related References

- [collections.md](./collections.md): Domain collections with behavior
- [naming.md](./naming.md): Naming conventions for types
