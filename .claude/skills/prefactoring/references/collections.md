# Domain Collections & Method Placement

**Purpose**: Apply prefactoring principles for collections with behavior and proper method placement.

## When to Use

- Working with arrays/lists that have domain meaning
- Deciding where methods should live
- Encapsulating aggregate operations
- Preventing feature envy anti-pattern

## Core Principles

### Collections with Domain Behavior

Wrap collections when they have domain-specific operations.

```typescript
// Bad: Raw array with scattered logic
const items: OrderItem[] = [];
const total = items.reduce((sum, item) => sum + item.price, 0);
const hasDiscount = items.length > 10;

// Good: Domain collection with behavior
class OrderItems {
  constructor(private readonly items: OrderItem[]) {}

  total(): Money {
    return this.items.reduce((sum, item) => sum.add(item.price), Money.zero());
  }

  hasDiscount(): boolean {
    return this.items.length > BULK_DISCOUNT_THRESHOLD;
  }

  isEmpty(): boolean {
    return this.items.length === 0;
  }
}
```

### Place Methods by Need

Methods belong where their data lives. Avoid feature envy.

```typescript
// Bad: Feature envy - method uses another object's data
class OrderPrinter {
  print(order: Order): string {
    return `${order.id}: ${order.items.length} items, $${order.total}`;
  }
}

class OrderValidator {
  isValid(order: Order): boolean {
    return order.items.length > 0 && order.total > 0;
  }
}

// Good: Methods on the object with the data
class Order {
  toString(): string {
    return `${this.id}: ${this.items.length} items, $${this.total}`;
  }

  isValid(): boolean {
    return this.items.length > 0 && this.total > 0;
  }
}
```

### Static Methods for Non-Instance Operations

If a method doesn't need instance data, it shouldn't be a member.

```typescript
// Bad: Instance method that doesn't use instance data
class DateUtils {
  formatDate(date: Date): string {
    return date.toISOString().split('T')[0];
  }
}

// Good: Static or module-level function
function formatDate(date: Date): string {
  return date.toISOString().split('T')[0];
}

// Or as a static method
class DateUtils {
  static formatDate(date: Date): string {
    return date.toISOString().split('T')[0];
  }
}
```

## Decision Matrix

| Situation                       | Apply                  | Action                         |
| ------------------------------- | ---------------------- | ------------------------------ |
| Array with operations           | Collection Wrapper     | Create domain collection class |
| Method uses other object's data | Place Methods by Need  | Move to data owner             |
| Method doesn't use `this`       | Static/Module Function | Extract to static or function  |
| Multiple methods on same data   | Domain Object          | Group into cohesive class      |

## Collection Wrapper Checklist

When creating a domain collection, consider:

- [ ] Does it encapsulate the underlying array/map?
- [ ] Does it provide domain-specific query methods?
- [ ] Does it enforce invariants (e.g., non-empty)?
- [ ] Does it hide implementation details?

```typescript
class UserGroup {
  private constructor(private readonly users: User[]) {
    if (users.length === 0) throw new EmptyGroupError();
  }

  static create(users: User[]): UserGroup {
    return new UserGroup(users);
  }

  findByEmail(email: Email): User | undefined {
    return this.users.find((u) => u.email.equals(email));
  }

  activeUsers(): UserGroup {
    return new UserGroup(this.users.filter((u) => u.isActive));
  }
}
```

## Related References

- [value-objects.md](./value-objects.md): Creating domain types
- [separation.md](./separation.md): Separating concerns in code
