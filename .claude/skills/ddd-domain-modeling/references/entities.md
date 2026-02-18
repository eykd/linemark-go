# Entities

Entities have identity that persists through state changes.

## Table of Contents

- [Structure](#structure)
- [Factory Methods](#factory-methods)
- [Invariant Enforcement](#invariant-enforcement)
- [Aggregate Roots](#aggregate-roots)
- [Testing Entities](#testing-entities)

## Structure

```typescript
// Props interface for internal state
interface TaskProps {
  id: string;
  userId: string;
  title: string;
  status: TaskStatus; // Value object
  createdAt: Date;
}

export class Task {
  // Private constructor enforces factory usage
  private constructor(private readonly props: TaskProps) {}

  // Factory for new entities - validates business rules
  static create(input: { userId: string; title: string }): Task {
    if (!input.title || input.title.trim().length < 3) {
      throw new Error('Title must be at least 3 characters');
    }
    return new Task({
      id: crypto.randomUUID(),
      userId: input.userId,
      title: input.title.trim(),
      status: TaskStatus.pending(),
      createdAt: new Date(),
    });
  }

  // Factory for reconstitution from persistence - no validation
  static reconstitute(props: TaskProps): Task {
    return new Task(props);
  }

  // Getters expose state (never setters)
  get id(): string {
    return this.props.id;
  }
  get title(): string {
    return this.props.title;
  }
  get isCompleted(): boolean {
    return this.props.status.isCompleted;
  }

  // Behavior methods enforce business rules
  complete(): void {
    if (this.props.status.isCompleted) {
      throw new Error('Task already completed');
    }
    this.props.status = TaskStatus.completed();
  }

  rename(newTitle: string): void {
    if (!newTitle || newTitle.trim().length < 3) {
      throw new Error('Title must be at least 3 characters');
    }
    this.props.title = newTitle.trim();
  }
}
```

## Factory Methods

Two factory methods serve different purposes:

| Method           | Purpose                      | Validates | Generates ID |
| ---------------- | ---------------------------- | --------- | ------------ |
| `create()`       | New entities from user input | Yes       | Yes          |
| `reconstitute()` | Rebuild from persistence     | No        | No           |

```typescript
// Application layer uses create()
const task = Task.create({ userId: "user-1", title: "Buy milk" });

// Repository uses reconstitute()
static reconstitute(row: TaskRow): Task {
  return Task.reconstitute({
    id: row.id,
    userId: row.user_id,
    title: row.title,
    status: row.completed ? TaskStatus.completed() : TaskStatus.pending(),
    createdAt: new Date(row.created_at)
  });
}
```

## Invariant Enforcement

Invariants are rules that must **always** be true:

```typescript
export class Order {
  private constructor(private props: OrderProps) {
    // Constructor invariant: orders must have at least one item
    if (props.items.length === 0) {
      throw new Error('Order must have at least one item');
    }
  }

  addItem(item: LineItem): void {
    // Method invariant: no duplicates
    if (this.props.items.some((i) => i.productId === item.productId)) {
      throw new Error('Product already in order');
    }
    this.props.items.push(item);
  }

  removeItem(productId: string): void {
    // Invariant: can't remove last item
    if (this.props.items.length === 1) {
      throw new Error('Cannot remove last item from order');
    }
    this.props.items = this.props.items.filter((i) => i.productId !== productId);
  }
}
```

## Aggregate Roots

Aggregate roots control access to child entities:

```typescript
export class Order {
  private props: OrderProps;

  // Child entities accessed through root
  get items(): ReadonlyArray<LineItem> {
    return [...this.props.items]; // Defensive copy
  }

  // Modifications go through root's methods
  addItem(productId: string, quantity: number, price: Money): void {
    const item = LineItem.create({ productId, quantity, price });
    this.props.items.push(item);
    this.recalculateTotal();
  }

  private recalculateTotal(): void {
    this.props.total = this.props.items.reduce(
      (sum, item) => sum.add(item.subtotal),
      Money.zero(this.props.currency)
    );
  }
}
```

## Testing Entities

Domain entities are pure—test without mocks:

```typescript
describe('Task', () => {
  describe('create', () => {
    it('creates task with valid title', () => {
      const task = Task.create({ userId: 'user-1', title: 'Buy milk' });

      expect(task.title).toBe('Buy milk');
      expect(task.isCompleted).toBe(false);
      expect(task.id).toBeDefined();
    });

    it('rejects short titles', () => {
      expect(() => Task.create({ userId: 'user-1', title: 'ab' })).toThrow(
        'Title must be at least 3 characters'
      );
    });

    it('trims whitespace', () => {
      const task = Task.create({ userId: 'user-1', title: '  Buy milk  ' });
      expect(task.title).toBe('Buy milk');
    });
  });

  describe('complete', () => {
    it('marks pending task as completed', () => {
      const task = Task.create({ userId: 'user-1', title: 'Test' });
      task.complete();
      expect(task.isCompleted).toBe(true);
    });

    it('throws when completing twice', () => {
      const task = Task.create({ userId: 'user-1', title: 'Test' });
      task.complete();
      expect(() => task.complete()).toThrow('Task already completed');
    });
  });
});
```
