# Layer Rules Reference

## Table of Contents

1. [Layer Overview](#layer-overview)
2. [Dependency Matrix](#dependency-matrix)
3. [Layer Contents](#layer-contents)
4. [Directory Mapping](#directory-mapping)
5. [Refactoring Patterns](#refactoring-patterns)

---

## Layer Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    Presentation Layer                        │
│   (Handlers, Controllers, Templates, CLI)                   │
├─────────────────────────────────────────────────────────────┤
│                    Infrastructure Layer                      │
│   (Repositories, External APIs, Caches, File System)        │
├─────────────────────────────────────────────────────────────┤
│                    Application Layer                         │
│   (Use Cases, DTOs, Application Services)                   │
├─────────────────────────────────────────────────────────────┤
│                      Domain Layer                            │
│   (Entities, Value Objects, Domain Services, Interfaces)    │
└─────────────────────────────────────────────────────────────┘
                    Dependencies point DOWN ↓
```

---

## Dependency Matrix

| From \ To          | Domain | Application | Infrastructure | Presentation |
| ------------------ | ------ | ----------- | -------------- | ------------ |
| **Domain**         | ✓      | ❌          | ❌             | ❌           |
| **Application**    | ✓      | ✓           | ❌             | ❌           |
| **Infrastructure** | ✓      | ✓           | ✓              | ❌           |
| **Presentation**   | ✓      | ✓           | ✓              | ✓            |

**Key rule:** ✓ = allowed, ❌ = violation

---

## Layer Contents

### Domain Layer

**Purpose:** Core business logic and rules, independent of technical concerns.

**Contains:**

- Entities (aggregate roots with identity)
- Value Objects (immutable, identity-less)
- Domain Services (operations spanning entities)
- Repository Interfaces (ports for data access)
- Domain Events
- Domain Exceptions

**Allowed imports:**

- Other domain types only
- Language built-ins (Date, Map, Set, etc.)

**Forbidden:**

- Framework packages
- Database clients
- HTTP types
- File system APIs
- External service clients

### Application Layer

**Purpose:** Orchestrate domain objects to fulfill use cases.

**Contains:**

- Use Cases / Application Services
- DTOs (Data Transfer Objects)
- Input/Output boundaries
- Application Events
- Validation schemas (for DTOs)

**Allowed imports:**

- Domain types (entities, value objects, interfaces)
- Application types (DTOs, other use cases)

**Forbidden:**

- Infrastructure implementations
- Presentation types (Request, Response)
- Framework-specific decorators

### Infrastructure Layer

**Purpose:** Implement technical concerns and integrate with external systems.

**Contains:**

- Repository Implementations
- External API Clients
- Cache Implementations
- Message Queue Adapters
- File System Access
- Database Migrations

**Allowed imports:**

- Domain interfaces (to implement)
- Application types (to use DTOs)
- Framework packages
- Database clients
- External SDKs

### Presentation Layer

**Purpose:** Handle HTTP/CLI/UI and translate to application calls.

**Contains:**

- HTTP Handlers / Controllers
- Route Definitions
- Middleware
- HTML Templates
- CLI Commands
- Request/Response Mapping

**Allowed imports:**

- Application use cases
- Domain types (for response mapping)
- Infrastructure (for dependency injection setup)
- Framework packages

---

## Directory Mapping

### Standard Structure

```
src/
├── domain/
│   ├── entities/
│   ├── value-objects/
│   ├── services/
│   └── interfaces/        # Repository/port interfaces HERE
│
├── application/
│   ├── use-cases/
│   ├── dto/
│   └── services/
│
├── infrastructure/
│   ├── repositories/      # Implements domain/interfaces
│   ├── cache/
│   ├── external/
│   └── persistence/
│
└── presentation/
    ├── handlers/
    ├── middleware/
    ├── templates/
    └── routes/
```

### Alternative Names

Some projects use different names:

| Standard       | Alternatives                    |
| -------------- | ------------------------------- |
| domain         | core, model, business           |
| application    | use-cases, services, app        |
| infrastructure | adapters, data, external, infra |
| presentation   | web, api, http, ui, cli         |

---

## Refactoring Patterns

### Extract Interface to Domain

**Before:**

```
infrastructure/repositories/UserRepository.ts  # interface + impl
```

**After:**

```
domain/interfaces/UserRepository.ts            # interface only
infrastructure/repositories/D1UserRepository.ts # implementation
```

### Remove Infrastructure from Domain

**Before:**

```typescript
// domain/entities/User.ts
import { db } from '@infrastructure/database';

export class User {
  async save() {
    await db.insert('users', this);
  }
}
```

**After:**

```typescript
// domain/entities/User.ts
export class User {
  // Pure entity, no persistence
}

// domain/interfaces/UserRepository.ts
export interface UserRepository {
  save(user: User): Promise<void>;
}

// infrastructure/repositories/D1UserRepository.ts
export class D1UserRepository implements UserRepository {
  async save(user: User) {
    await this.db.insert('users', user);
  }
}
```

### Remove Framework Types from Use Cases

**Before:**

```typescript
// application/use-cases/CreateUser.ts
async execute(request: Request): Promise<Response>
```

**After:**

```typescript
// application/dto/CreateUserRequest.ts
export interface CreateUserRequest {
  email: string;
  name: string;
}

// application/use-cases/CreateUser.ts
async execute(dto: CreateUserRequest): Promise<UserResponse>

// presentation/handlers/UserHandler.ts
async handleCreate(request: Request): Promise<Response> {
  const dto = await this.parseRequest(request);
  const result = await this.createUser.execute(dto);
  return this.toResponse(result);
}
```

### Dependency Injection Setup

Wire dependencies in composition root (entry point):

```typescript
// src/index.ts (or composition-root.ts)
export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    // Create infrastructure implementations
    const userRepo = new D1UserRepository(env.DB);
    const cache = new KVCacheService(env.CACHE);

    // Create use cases with injected dependencies
    const createUser = new CreateUser(userRepo);
    const getUser = new GetUser(userRepo, cache);

    // Create handlers with use cases
    const handlers = new UserHandlers(createUser, getUser);

    // Route request
    return router.handle(request, handlers);
  },
};
```
