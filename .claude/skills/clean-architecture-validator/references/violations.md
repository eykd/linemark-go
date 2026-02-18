# Violation Patterns Reference

## Table of Contents

1. [Domain Layer Violations](#domain-layer-violations)
2. [Application Layer Violations](#application-layer-violations)
3. [Interface Misplacement](#interface-misplacement)
4. [Common Framework Violations](#common-framework-violations)

---

## Domain Layer Violations

Domain code must be pure—no external dependencies.

### Infrastructure Import

**Bad:**

```typescript
// src/domain/entities/User.ts
import { D1Database } from '@cloudflare/workers-types'; // ❌
import { PrismaClient } from '@prisma/client'; // ❌
```

**Fix:** Remove all infrastructure imports. Domain entities should only import other domain types.

### Direct Database Access

**Bad:**

```typescript
// src/domain/services/UserService.ts
export class UserService {
  async findUser(db: D1Database, id: string) {
    // ❌
    return db.prepare('SELECT * FROM users WHERE id = ?').bind(id).first();
  }
}
```

**Fix:** Define repository interface in domain, inject implementation:

```typescript
// src/domain/interfaces/UserRepository.ts
export interface UserRepository {
  findById(id: string): Promise<User | null>;
}

// src/domain/services/UserService.ts
export class UserService {
  constructor(private userRepo: UserRepository) {}
  async findUser(id: string) {
    return this.userRepo.findById(id);
  }
}
```

### HTTP/Request Objects in Domain

**Bad:**

```typescript
// src/domain/entities/Task.ts
import { Request } from '@cloudflare/workers-types'; // ❌

export class Task {
  static fromRequest(req: Request) { ... } // ❌
}
```

**Fix:** Use plain DTOs. Parse requests in presentation layer.

### Async Operations for Non-Essential Logic

**Bad:**

```typescript
// src/domain/entities/Order.ts
export class Order {
  async calculateTotal() {
    // ❌ - fetches exchange rates
    const rate = await fetch('https://api.exchange.com/rate');
    return this.amount * rate;
  }
}
```

**Fix:** Pass rates as parameters or use domain services with injected dependencies.

---

## Application Layer Violations

Application layer orchestrates domain objects but must not know infrastructure details.

### Concrete Infrastructure Instantiation

**Bad:**

```typescript
// src/application/use-cases/CreateUser.ts
import { D1UserRepository } from '@infrastructure/repositories/D1UserRepository'; // ❌

export class CreateUser {
  async execute(data: CreateUserRequest) {
    const repo = new D1UserRepository(db); // ❌
    // ...
  }
}
```

**Fix:** Accept interface via constructor:

```typescript
import type { UserRepository } from '@domain/interfaces/UserRepository';

export class CreateUser {
  constructor(private userRepo: UserRepository) {} // ✓
  async execute(data: CreateUserRequest) {
    // use this.userRepo
  }
}
```

### Framework Types in Use Case Signatures

**Bad:**

```typescript
// src/application/use-cases/GetTasks.ts
import { Request, Response } from '@cloudflare/workers-types'; // ❌

export class GetTasks {
  async execute(request: Request): Promise<Response> {
    // ❌
    // ...
  }
}
```

**Fix:** Use plain DTOs:

```typescript
export class GetTasks {
  async execute(query: GetTasksQuery): Promise<TaskResponse[]> {
    // ✓
    // ...
  }
}
```

### Direct External API Calls

**Bad:**

```typescript
// src/application/use-cases/SendNotification.ts
export class SendNotification {
  async execute(userId: string) {
    await fetch('https://api.sendgrid.com/send', { ... }); // ❌
  }
}
```

**Fix:** Define port interface, inject adapter:

```typescript
// src/domain/interfaces/NotificationService.ts
export interface NotificationService {
  send(to: string, message: string): Promise<void>;
}

// src/application/use-cases/SendNotification.ts
export class SendNotification {
  constructor(private notifier: NotificationService) {}
  async execute(userId: string) {
    await this.notifier.send(userId, 'Hello');
  }
}
```

---

## Interface Misplacement

Repository interfaces are ports—they belong in the domain layer.

### Interface in Infrastructure

**Bad:**

```
src/infrastructure/repositories/
├── UserRepository.ts      # Interface defined here ❌
└── D1UserRepository.ts    # Implementation
```

**Fix:**

```
src/domain/interfaces/
└── UserRepository.ts      # Interface here ✓

src/infrastructure/repositories/
└── D1UserRepository.ts    # Implementation imports interface
```

### Interface Importing Implementation Types

**Bad:**

```typescript
// src/domain/interfaces/CacheService.ts
import { KVNamespace } from '@cloudflare/workers-types'; // ❌

export interface CacheService {
  get(key: string, kv: KVNamespace): Promise<string | null>; // ❌
}
```

**Fix:** Keep interface pure:

```typescript
export interface CacheService {
  get(key: string): Promise<string | null>; // ✓
}
```

---

## Common Framework Violations

### Cloudflare Workers

Violation indicators:

- `@cloudflare/workers-types` in domain/application
- `D1Database`, `KVNamespace`, `DurableObject` in domain
- `Env` type in domain entities

### Express/Fastify/Hono

Violation indicators:

- `Request`, `Response`, `Context` in domain/application
- Middleware references in use cases
- Route handlers in application layer

### Database ORMs

Violation indicators:

- Prisma/TypeORM/Drizzle decorators on domain entities
- `@Entity()`, `@Column()` in domain
- Database client types in domain interfaces

### Validation Libraries

**Acceptable:** Using Zod/Yup in application layer DTOs
**Violation:** Decorators/validators directly on domain entities

```typescript
// Bad - domain entity
import { z } from 'zod'; // ❌ in domain
export class User {
  @IsEmail() email: string; // ❌
}

// Good - application DTO
import { z } from 'zod';
export const CreateUserSchema = z.object({ ... }); // ✓ in application/dto
```
