# Interface Contracts & Validation

**Purpose**: Apply prefactoring principles when defining APIs, validating inputs, and designing for testability.

## When to Use

- Defining public APIs or contracts
- Implementing input validation at boundaries
- Writing tests that verify behavior
- Designing testable components

## Core Principles

### Create Interface Contracts

Document and enforce preconditions, postconditions, and invariants.

```typescript
/**
 * Transfers money between accounts.
 * @precondition Both accounts must exist and be active
 * @precondition source account must have sufficient balance
 * @postcondition source.balance = source.balance - amount
 * @postcondition target.balance = target.balance + amount
 * @throws InsufficientFundsError if balance too low
 * @throws AccountNotFoundError if account doesn't exist
 */
interface TransferService {
  transfer(source: AccountId, target: AccountId, amount: Money): Promise<void>;
}
```

### Validate, Validate, Validate

Validate at every system boundary. Fail fast with clear error messages.

```typescript
class CreateUserHandler {
  async handle(request: unknown): Promise<Response> {
    // Validate at entry point
    const validated = this.validate(request);
    if (!validated.success) {
      return Response.json({ errors: validated.errors }, { status: 400 });
    }

    // Domain types are already valid
    const email = new Email(validated.data.email);
    const user = await this.useCase.execute(email);
    return Response.json(user);
  }
}
```

### Test the Interface, Not the Implementation

Test contracts, not internal details. Enable refactoring without breaking tests.

```typescript
// Bad: Tests implementation details
it('should call repository.save with correct SQL', () => {
  expect(mockDb.query).toHaveBeenCalledWith('INSERT INTO...');
});

// Good: Tests contract
it('should persist user and return with ID', async () => {
  const user = User.create({ email: new Email('test@example.com') });
  const saved = await repository.save(user);

  expect(saved.id).toBeDefined();
  const retrieved = await repository.findById(saved.id);
  expect(retrieved?.email).toEqual(user.email);
});
```

### Build Flexibility for Testing

Design with dependency injection and clear interfaces.

```typescript
// Testable design: dependencies injected
class OrderService {
  constructor(
    private readonly repository: OrderRepository,
    private readonly paymentGateway: PaymentGateway,
    private readonly notifier: OrderNotifier
  ) {}
}

// In tests: inject test doubles
const service = new OrderService(
  new InMemoryOrderRepository(),
  new MockPaymentGateway(),
  new SpyOrderNotifier()
);
```

## Decision Matrix

| Situation            | Apply               | Example                        |
| -------------------- | ------------------- | ------------------------------ |
| Public API method    | Interface Contracts | Document pre/postconditions    |
| External input       | Validate            | Check at system boundary       |
| Writing tests        | Test Interface      | Verify contract, not internals |
| Complex dependencies | Build Flexibility   | Use dependency injection       |

## Related References

- [error-handling.md](./error-handling.md): Error strategies and messaging
- [architecture.md](./architecture.md): Interface design at module level
