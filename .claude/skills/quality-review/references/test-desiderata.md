# Test Quality Evaluation Reference

Guidance for evaluating test quality in pull request reviews.

---

## When to Apply

Evaluate test quality when the PR includes test files:

- `*.spec.ts`, `*.test.ts` (TypeScript/JavaScript)
- `test_*.py`, `*_test.py` (Python)
- `*_test.go` (Go)
- Files in `tests/`, `__tests__/`, `test/` directories

---

## Kent Beck's Test Desiderata

Evaluate tests against these 12 properties:

### Core Properties

| Property          | Question                          | Signs of Violation                             |
| ----------------- | --------------------------------- | ---------------------------------------------- |
| **Isolated**      | Can tests run independently?      | Shared state, test order dependencies          |
| **Composable**    | Can tests run in any combination? | Global setup/teardown affecting others         |
| **Deterministic** | Same result every run?            | Time-based, random data, external dependencies |
| **Fast**          | Quick feedback?                   | I/O operations, network calls, large datasets  |

### Structural Properties

| Property                  | Question                            | Signs of Violation                            |
| ------------------------- | ----------------------------------- | --------------------------------------------- |
| **Writable**              | Easy to add new tests?              | Copy-paste boilerplate, complex setup         |
| **Readable**              | Clear what's being tested?          | Magic numbers, unclear assertions, long tests |
| **Behavioral**            | Tests outcomes, not implementation? | Testing private methods, mocking internals    |
| **Structure-insensitive** | Survive refactoring?                | Tightly coupled to implementation details     |

### Meta Properties

| Property       | Question               | Signs of Violation                           |
| -------------- | ---------------------- | -------------------------------------------- |
| **Automated**  | No manual steps?       | Manual assertions, visual inspection needed  |
| **Specific**   | Clear failure message? | Generic assertions, unclear failure location |
| **Predictive** | Catches real bugs?     | Tests only happy path, missing edge cases    |
| **Inspiring**  | Confidence to deploy?  | Low coverage, missing critical paths         |

---

## GOOS Principles

From "Growing Object-Oriented Software, Guided by Tests":

### Test Classification

| Type            | Purpose                   | Characteristics        |
| --------------- | ------------------------- | ---------------------- |
| **Unit**        | Single component behavior | Fast, isolated, no I/O |
| **Integration** | Component collaboration   | Tests real boundaries  |
| **Acceptance**  | User-visible behavior     | End-to-end, slower     |

### Mocking Guidelines

**Mock Roles, Not Objects**

```typescript
// Good: Mock the role (interface)
const mockNotifier = mock<NotificationService>();
await service.processOrder(mockNotifier);
verify(mockNotifier.notify(order)).called();

// Bad: Mock the implementation
const mockEmail = mock(EmailClient); // Too specific
```

**Don't Mock Values**

```typescript
// Good: Use real value objects
const money = new Money(100, 'USD');
expect(cart.total()).toEqual(money);

// Bad: Mock simple values
const mockMoney = mock<Money>(); // Unnecessary
```

**Verify Interactions, Not State**

```typescript
// Good: Verify the right call was made
verify(repository.save(entity)).called();

// Less ideal: Check internal state
expect(repository.entities).toContain(entity);
```

---

## Test Anti-Patterns to Flag

### Critical Anti-Patterns

| Anti-Pattern               | Description                         | Fix                                |
| -------------------------- | ----------------------------------- | ---------------------------------- |
| **Missing Assertions**     | Test runs but verifies nothing      | Add meaningful assertions          |
| **Testing Framework Code** | Testing library/framework internals | Test your code, not theirs         |
| **Shared Mutable State**   | Tests affect each other             | Isolate setup, use fresh instances |
| **Production Data**        | Tests depend on real data           | Use fixtures or factories          |

### High Priority Anti-Patterns

| Anti-Pattern          | Description                          | Fix                           |
| --------------------- | ------------------------------------ | ----------------------------- |
| **Test-Per-Method**   | Mapping 1:1 with implementation      | Test behaviors, not methods   |
| **Excessive Mocking** | More mocks than real objects         | Simplify design, use fakes    |
| **Logic in Tests**    | Conditionals, loops in test code     | Keep tests linear and obvious |
| **Obscure Test**      | Can't understand what's being tested | Improve naming, simplify      |

### Medium Priority Anti-Patterns

| Anti-Pattern                | Description                    | Fix                           |
| --------------------------- | ------------------------------ | ----------------------------- |
| **Test Doubles Everywhere** | Never testing real integration | Add focused integration tests |
| **Assertion Roulette**      | Multiple unrelated assertions  | One concept per test          |
| **Long Tests**              | Hard to understand at a glance | Extract setup, split tests    |
| **Eager Test**              | Testing too many things        | Focus on one behavior         |

---

## Test Pain as Design Feedback

When tests are hard to write, it often indicates design problems:

| Test Difficulty             | Possible Design Issue             |
| --------------------------- | --------------------------------- |
| **Complex setup**           | Object has too many dependencies  |
| **Hard to mock**            | Interface too large, violates ISP |
| **Brittle tests**           | Implementation details exposed    |
| **Slow tests**              | Missing abstraction boundaries    |
| **Can't test in isolation** | Tight coupling between components |

---

## Coverage Verification

**CRITICAL**: This project requires 100% test coverage threshold.

When reviewing test changes, ALWAYS verify coverage:

```bash
npx jest --coverage --findRelatedTests [changed-files]
```

**Check all four metrics**:

- Branches: Must be 100%
- Functions: Must be 100%
- Lines: Must be 100%
- Statements: Must be 100%

### Finding Format for Insufficient Coverage

**If coverage is below 100%**, report as HIGH severity:

```markdown
### Finding: Test coverage below 100%

- **Severity**: High
- **Category**: test
- **File**: path/to/file.ts
- **Description**: Coverage is at X% (branches: X%, functions: X%, lines: X%, statements: X%). Project requires 100% coverage threshold.
- **Impact**: Pre-commit hooks will fail. Untested code paths may contain bugs.
- **Fix**: Add tests for uncovered lines. Run `npx jest --coverage` to see gaps. Use istanbul ignore ONLY for truly untestable code (defensive type guards, platform-specific edge cases).
```

**If 95-99% coverage**:

- Flag as HIGH severity finding
- List specific uncovered lines from coverage report
- Suggest specific test cases for each gap
- Only suggest istanbul ignore if truly justified

---

## Evaluation Output Format

For the Test Quality section of the review:

```markdown
### Test Quality

#### Classification Assessment

- **Unit Tests**: [Present/Missing] - [Assessment]
- **Integration Tests**: [Present/Missing] - [Assessment]
- **Acceptance Tests**: [Present/Missing] - [Assessment]

#### Test Desiderata Check

| Property      | Status    | Notes                  |
| ------------- | --------- | ---------------------- |
| Isolated      | Pass/Fail | [Specific observation] |
| Deterministic | Pass/Fail | [Specific observation] |
| Fast          | Pass/Fail | [Specific observation] |
| Readable      | Pass/Fail | [Specific observation] |
| Behavioral    | Pass/Fail | [Specific observation] |
| Specific      | Pass/Fail | [Specific observation] |

#### Coverage Report

- **Branches**: X%
- **Functions**: X%
- **Lines**: X%
- **Statements**: X%
- **Status**: [Pass/FAIL - see findings]

#### Mocking Assessment

- **Pattern**: [Appropriate/Over-mocked/Under-mocked]
- **Observations**: [Specific feedback on mocking usage]

#### Anti-Patterns Detected

[List any anti-patterns found, or "None detected"]

#### Design Feedback

[Any observations about design issues indicated by test difficulty, or "No design concerns"]
```

---

## Quick Checklist

For rapid evaluation, verify these essentials:

- [ ] Tests are behavioral (test what, not how)
- [ ] Tests are readable and self-documenting
- [ ] Tests are fast (no unnecessary I/O)
- [ ] Tests are deterministic (no flaky tests)
- [ ] Tests are isolated (no shared state)
- [ ] Tests are specific (clear failure messages)
- [ ] Mocking is appropriate (roles, not values)
- [ ] No critical anti-patterns present
- [ ] Coverage is 100% for all metrics
- [ ] Test pain signals are addressed or acknowledged

---

## Examples

### Good Test Pattern

```typescript
describe('OrderService', () => {
  it('notifies customer when order ships', async () => {
    // Arrange
    const notifier = mock<NotificationService>();
    const order = createOrder({ status: 'packed' });
    const service = new OrderService(notifier);

    // Act
    await service.shipOrder(order);

    // Assert
    verify(notifier.notifyShipped(order)).called();
    expect(order.status).toBe('shipped');
  });
});
```

**Why it's good**:

- Clear behavior being tested (in the name)
- Minimal setup
- Tests role interaction (notifier)
- Verifiable outcome
- Follows AAA pattern (Arrange-Act-Assert)

### Bad Test Pattern

```typescript
describe('OrderService', () => {
  let db: Database;
  let cache: Cache;

  beforeAll(() => {
    db = new RealDatabase(); // Shared state
    cache = new RealCache();
  });

  it('works', async () => {
    // Vague name
    const service = new OrderService(db, cache);
    const result = await service.process({}); // What's being tested?
    expect(result).toBeTruthy(); // Weak assertion
  });
});
```

**Why it's bad**:

- Shared mutable state (beforeAll with real instances)
- Unclear what behavior is tested
- No isolation (real DB, real cache)
- Weak assertion (truthy)
- Vague test name ("works")
