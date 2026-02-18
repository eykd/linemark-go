---
name: ddd-domain-modeling
description: 'Use when: (1) building domain entities with validation, (2) creating value objects, (3) defining repository interfaces, (4) implementing domain services, (5) questions about DDD patterns or business logic encapsulation.'
---

# DDD Domain Modeling

Create domain models in pure TypeScript with zero external dependencies.

## Core Principle

The domain layer contains **pure business logic only**:

- No framework imports
- No async operations in entities/value objects
- No database or HTTP concerns
- Validate invariants in constructors
- Dependencies point inward (infrastructure → application → domain)

## Directory Structure

```
src/domain/
├── entities/           # Aggregate roots and entities
│   ├── Order.ts
│   └── Order.spec.ts
├── value-objects/      # Immutable value types
│   ├── Money.ts
│   └── Email.ts
├── services/           # Stateless domain logic
│   └── PricingService.ts
└── interfaces/         # Repository ports (interfaces only)
    └── OrderRepository.ts
```

## Quick Reference

### Entity (3 parts)

```typescript
export class Order {
  private constructor(private props: OrderProps) {
    // Validate invariants here
  }

  static create(input: CreateOrderInput): Order {
    /* validate + new */
  }
  static reconstitute(props: OrderProps): Order {
    /* from persistence */
  }

  // Getters + behavior methods that enforce rules
}
```

### Value Object

```typescript
export class Money {
  private constructor(
    public readonly amount: number,
    public readonly currency: string
  ) {}

  static of(amount: number, currency: string): Money {
    if (amount < 0) throw new Error('Amount cannot be negative');
    return new Money(amount, currency);
  }

  add(other: Money): Money {
    if (this.currency !== other.currency) throw new Error('Currency mismatch');
    return Money.of(this.amount + other.amount, this.currency);
  }

  equals(other: Money): boolean {
    return this.amount === other.amount && this.currency === other.currency;
  }
}
```

### Repository Interface

```typescript
// Domain layer - interface only
export interface OrderRepository {
  findById(id: string): Promise<Order | null>;
  save(order: Order): Promise<void>;
}
```

### Domain Service

```typescript
// Stateless, operates on multiple entities/value objects
export class PricingService {
  calculateTotal(items: LineItem[], discounts: Discount[]): Money {
    // Complex calculation logic here
  }
}
```

## Workflow

1. **Identify the concept**: Entity (has identity) or Value Object (defined by attributes)?
2. **Define invariants**: What rules must always be true?
3. **Choose pattern**: See detailed references below
4. **Write tests first**: Domain code is pure—test without mocks

## Detailed References

- **Entities with identity and lifecycle**: See [references/entities.md](references/entities.md)
- **Value Objects with validation**: See [references/value-objects.md](references/value-objects.md)
- **Repository interfaces (ports)**: See [references/repositories.md](references/repositories.md)
- **Domain services for complex logic**: See [references/domain-services.md](references/domain-services.md)

## Anti-Patterns to Avoid

| Anti-Pattern                      | Instead                                                            |
| --------------------------------- | ------------------------------------------------------------------ |
| `import { D1Database }` in domain | Define interface in domain, implement in infrastructure            |
| `async` in entity methods         | Keep entities synchronous; async belongs in repositories           |
| Exposing setters                  | Provide behavior methods: `order.addItem()` not `order.items = []` |
| Validation in application layer   | Validate in entity/value object constructors                       |
| Anemic domain model               | Put behavior with the data it operates on                          |
