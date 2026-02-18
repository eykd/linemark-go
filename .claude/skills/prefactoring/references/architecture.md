# Architecture & System Design

**Purpose**: Apply prefactoring principles when designing system structure, modules, and boundaries.

## When to Use

- Creating new modules or packages
- Defining system boundaries or layer organization
- Making architectural decisions about component relationships
- Choosing between inheritance and composition

## Core Principles

### Well-Defined Interfaces

Define interfaces before implementation. Document preconditions, postconditions, and side effects.

```typescript
/** Repository interface defines contract, not implementation */
interface UserRepository {
  /** @throws NotFoundError if user doesn't exist */
  findById(id: UserId): Promise<User>;
  /** @returns created user with generated ID */
  save(user: User): Promise<User>;
}
```

### Decomposition & Modularity

Split systems into cohesive, loosely-coupled modules. Each module has one clear responsibility.

```typescript
// Good: Separate modules by domain concern
// src/users/domain/        - User entities, value objects
// src/users/application/   - Use cases
// src/users/infrastructure/- Repository implementations
// src/orders/domain/       - Order entities (separate bounded context)
```

### Separation of Concerns

Each module addresses one concern. Orthogonal concerns live in separate modules.

```typescript
// Separate concerns: validation, persistence, notification
class CreateUserUseCase {
  constructor(
    private readonly validator: UserValidator, // Validation concern
    private readonly repository: UserRepository, // Persistence concern
    private readonly notifier: UserNotifier // Notification concern
  ) {}
}
```

### Hierarchy & Layers

Dependencies flow in one direction: higher layers depend on lower layers.

```typescript
// Layer hierarchy (dependencies flow down)
// Handlers    -> Use Cases    -> Domain      -> (nothing)
// (adapters)  -> (application) -> (entities)

// Use case depends on domain, not vice versa
class ProcessOrderUseCase {
  execute(request: ProcessOrderRequest): Promise<Order> {
    const order = Order.create(request); // Domain has no use case dependency
    return this.repository.save(order);
  }
}
```

### Packaging

Components that change together should be packaged together.

```typescript
// Group by feature/domain, not by type
// Good:
// src/orders/OrderEntity.ts
// src/orders/OrderRepository.ts
// src/orders/CreateOrderUseCase.ts

// Bad:
// src/entities/OrderEntity.ts
// src/repositories/OrderRepository.ts
// src/usecases/CreateOrderUseCase.ts
```

### Think in Interfaces, Not Inheritance

Prefer composition and interfaces over inheritance hierarchies.

```typescript
// Bad: Rigid inheritance
class Animal {
  move(): void {}
}
class Bird extends Animal {
  fly(): void {}
}
class Penguin extends Bird {} // Can't fly!

// Good: Composition via interfaces
interface Movable {
  move(): void;
}
interface Flyable {
  fly(): void;
}

class Penguin implements Movable {
  move(): void {
    /* waddle */
  }
}
```

## Decision Matrix

| Situation             | Apply                    | Example                      |
| --------------------- | ------------------------ | ---------------------------- |
| New module            | Decomposition, Packaging | Group related files together |
| API boundary          | Well-Defined Interfaces  | Document contracts           |
| Cross-cutting concern | Separation of Concerns   | Extract to separate module   |
| Class hierarchy       | Think in Interfaces      | Prefer composition           |
| Dependency direction  | Hierarchy                | Higher depends on lower      |

## Related References

- [abstraction.md](./abstraction.md): Type-level design decisions
- [interfaces.md](./interfaces.md): Contract design and validation
