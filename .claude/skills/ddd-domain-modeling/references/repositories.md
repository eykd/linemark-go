# Repository Interfaces

Repository interfaces define the contract for persistence in the domain layer. Implementation lives in infrastructure.

## Table of Contents

- [The Ports and Adapters Pattern](#the-ports-and-adapters-pattern)
- [Interface Design](#interface-design)
- [Common Repository Methods](#common-repository-methods)
- [Specification Pattern](#specification-pattern)
- [Implementation Guidelines](#implementation-guidelines)

## The Ports and Adapters Pattern

```
┌─────────────────────────────────────────────────────┐
│                    Domain Layer                      │
│  ┌─────────────────────────────────────────────┐    │
│  │  interface TaskRepository (PORT)            │    │
│  │    findById(id: string): Promise<Task>      │    │
│  │    save(task: Task): Promise<void>          │    │
│  └─────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────┘
                         ▲
                         │ implements
┌─────────────────────────────────────────────────────┐
│                Infrastructure Layer                  │
│  ┌─────────────────────────────────────────────┐    │
│  │  class D1TaskRepository (ADAPTER)           │    │
│  │    constructor(db: D1Database)              │    │
│  │    async findById(id): Promise<Task>        │    │
│  │    async save(task): Promise<void>          │    │
│  └─────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────┘
```

**Key Rule**: Domain layer defines the interface. Infrastructure layer provides the implementation.

## Interface Design

### Basic Repository Interface

```typescript
// src/domain/interfaces/TaskRepository.ts
import type { Task } from '../entities/Task';

export interface TaskRepository {
  findById(id: string): Promise<Task | null>;
  findAll(): Promise<Task[]>;
  save(task: Task): Promise<void>;
  delete(id: string): Promise<void>;
}
```

### Repository with Query Methods

```typescript
// src/domain/interfaces/OrderRepository.ts
import type { Order } from '../entities/Order';
import type { OrderStatus } from '../value-objects/OrderStatus';
import type { DateRange } from '../value-objects/DateRange';

export interface OrderRepository {
  // Core CRUD
  findById(id: string): Promise<Order | null>;
  save(order: Order): Promise<void>;
  delete(id: string): Promise<void>;

  // Domain-specific queries
  findByCustomerId(customerId: string): Promise<Order[]>;
  findByStatus(status: OrderStatus): Promise<Order[]>;
  findByDateRange(range: DateRange): Promise<Order[]>;

  // Aggregate queries
  countByStatus(status: OrderStatus): Promise<number>;
  existsById(id: string): Promise<boolean>;
}
```

### Repository with Pagination

```typescript
export interface Page<T> {
  items: T[];
  total: number;
  page: number;
  pageSize: number;
  hasNext: boolean;
  hasPrevious: boolean;
}

export interface PageRequest {
  page: number;
  pageSize: number;
  sortBy?: string;
  sortOrder?: 'asc' | 'desc';
}

export interface ProductRepository {
  findById(id: string): Promise<Product | null>;
  save(product: Product): Promise<void>;

  // Paginated queries
  findAll(pageRequest: PageRequest): Promise<Page<Product>>;
  findByCategory(categoryId: string, pageRequest: PageRequest): Promise<Page<Product>>;

  search(query: string, pageRequest: PageRequest): Promise<Page<Product>>;
}
```

## Common Repository Methods

| Method           | Purpose           | Returns          |
| ---------------- | ----------------- | ---------------- |
| `findById(id)`   | Get single entity | `Entity \| null` |
| `findAll()`      | Get all entities  | `Entity[]`       |
| `save(entity)`   | Insert or update  | `void`           |
| `delete(id)`     | Remove entity     | `void`           |
| `existsById(id)` | Check existence   | `boolean`        |
| `count()`        | Total count       | `number`         |

### Naming Conventions

```typescript
// Query methods start with "find"
findById(id: string): Promise<Task | null>;
findByUserId(userId: string): Promise<Task[]>;
findByStatus(status: TaskStatus): Promise<Task[]>;
findCompletedBefore(date: Date): Promise<Task[]>;

// Boolean checks use "exists" or "is"
existsById(id: string): Promise<boolean>;
existsByEmail(email: Email): Promise<boolean>;

// Counts use "count"
count(): Promise<number>;
countByStatus(status: TaskStatus): Promise<number>;
```

## Specification Pattern

For complex queries, define specifications in the domain:

```typescript
// src/domain/interfaces/TaskRepository.ts
export interface TaskSpecification {
  isSatisfiedBy(task: Task): boolean;
  toSql?(): { where: string; params: unknown[] }; // Optional optimization
}

export interface TaskRepository {
  findById(id: string): Promise<Task | null>;
  findAll(): Promise<Task[]>;
  findMatching(spec: TaskSpecification): Promise<Task[]>;
  save(task: Task): Promise<void>;
}

// Usage in domain
export class OverdueTaskSpecification implements TaskSpecification {
  constructor(private now: Date = new Date()) {}

  isSatisfiedBy(task: Task): boolean {
    return task.dueDate < this.now && !task.isCompleted;
  }

  toSql() {
    return {
      where: 'due_date < ? AND completed = 0',
      params: [this.now.toISOString()],
    };
  }
}
```

## Implementation Guidelines

### Infrastructure Implementation

```typescript
// src/infrastructure/repositories/D1TaskRepository.ts
import type { D1Database } from '@cloudflare/workers-types';
import type { TaskRepository } from '@domain/interfaces/TaskRepository';
import { Task } from '@domain/entities/Task';
import { TaskStatus } from '@domain/value-objects/TaskStatus';

interface TaskRow {
  id: string;
  user_id: string;
  title: string;
  completed: number; // SQLite boolean
  created_at: string;
}

export class D1TaskRepository implements TaskRepository {
  constructor(private readonly db: D1Database) {}

  async findById(id: string): Promise<Task | null> {
    const row = await this.db.prepare('SELECT * FROM tasks WHERE id = ?').bind(id).first<TaskRow>();

    return row ? this.toDomain(row) : null;
  }

  async findAll(): Promise<Task[]> {
    const { results } = await this.db
      .prepare('SELECT * FROM tasks ORDER BY created_at DESC')
      .all<TaskRow>();

    return results.map((row) => this.toDomain(row));
  }

  async save(task: Task): Promise<void> {
    await this.db
      .prepare(
        `
        INSERT INTO tasks (id, user_id, title, completed, created_at)
        VALUES (?, ?, ?, ?, ?)
        ON CONFLICT(id) DO UPDATE SET
          title = excluded.title,
          completed = excluded.completed
      `
      )
      .bind(
        task.id,
        task.userId,
        task.title,
        task.isCompleted ? 1 : 0,
        task.createdAt.toISOString()
      )
      .run();
  }

  async delete(id: string): Promise<void> {
    await this.db.prepare('DELETE FROM tasks WHERE id = ?').bind(id).run();
  }

  // Map database row to domain entity
  private toDomain(row: TaskRow): Task {
    return Task.reconstitute({
      id: row.id,
      userId: row.user_id,
      title: row.title,
      completed: row.completed === 1,
      createdAt: new Date(row.created_at),
    });
  }
}
```

### Unit of Work (Optional)

For transactional consistency across multiple repositories:

```typescript
// src/domain/interfaces/UnitOfWork.ts
export interface UnitOfWork {
  tasks: TaskRepository;
  users: UserRepository;
  commit(): Promise<void>;
  rollback(): Promise<void>;
}

// Usage in application layer
async execute(request: CreateOrderRequest): Promise<void> {
  const uow = this.unitOfWorkFactory.create();
  try {
    const user = await uow.users.findById(request.userId);
    const order = Order.create({ ... });
    await uow.orders.save(order);
    await uow.commit();
  } catch (error) {
    await uow.rollback();
    throw error;
  }
}
```

### Testing with Repository Interface

```typescript
// In-memory implementation for unit tests
class InMemoryTaskRepository implements TaskRepository {
  private tasks = new Map<string, Task>();

  async findById(id: string): Promise<Task | null> {
    return this.tasks.get(id) ?? null;
  }

  async findAll(): Promise<Task[]> {
    return [...this.tasks.values()];
  }

  async save(task: Task): Promise<void> {
    this.tasks.set(task.id, task);
  }

  async delete(id: string): Promise<void> {
    this.tasks.delete(id);
  }

  // Test helper
  clear(): void {
    this.tasks.clear();
  }
}

// Use case test
describe('CreateTask', () => {
  let repository: InMemoryTaskRepository;
  let useCase: CreateTask;

  beforeEach(() => {
    repository = new InMemoryTaskRepository();
    useCase = new CreateTask(repository);
  });

  it('creates and persists task', async () => {
    const result = await useCase.execute({ userId: 'u1', title: 'Test' });

    const saved = await repository.findById(result.id);
    expect(saved).not.toBeNull();
    expect(saved!.title).toBe('Test');
  });
});
```
