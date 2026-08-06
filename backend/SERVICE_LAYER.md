# Service Layer (Phase 1.4)

`backend/internal/application/services/` is the **business service layer**. It is the only place in the system that is allowed to make business decisions. Repositories are pure persistence; the application use case layer (Phase 1.5) will compose services; the UI is a thin shell over use cases.

## 1. Layout

```
application/services/
├── errors.go         # service-level sentinel errors with stable codes
├── tx.go             # TxManager interface + adapter
├── common/
│   └── logger.go     # thin slog wrapper used by every service
├── customer/         # customer lifecycle, credit, debt
├── supplier/         # supplier lifecycle, debt
├── product/          # product catalog, cost/price changes, margin
├── inventory/        # batches, movements, 25-day clearance rule
├── purchasing/       # purchase orders, approval, AP, supplier payments
├── sales/            # sales lifecycle, payments, advances
├── customerpayments/ # customer payments, advances, allocations
├── treasury/         # bank accounts, credit cards, exchange rates
├── accounting/       # journal entries, posting, ledger queries
└── reporting/        # AR/AP/Profit summaries
```

## 2. Conventions

### 2.1 Constructor injection
Every service receives its dependencies through the constructor. **No `init()` functions, no global state, no hidden singletons.**

```go
type CustomerService struct {
    repo repositories.CustomerRepository
    txm  services.TxManager
    log  *common.Logger
}

func New(repo repositories.CustomerRepository, txm services.TxManager, log *common.Logger) *CustomerService {
    if repo == nil || txm == nil || log == nil {
        panic("customer: nil dependency")
    }
    return &CustomerService{repo: repo, txm: txm, log: log}
}
```

### 2.2 Transactions
Every multi-step operation runs inside `txm.WithinTransaction(ctx, fn)`. The UoW stamped on the context is the only way the service reads/writes through repositories. Direct repo access is reserved for single-query reads that do not need a transaction.

```go
err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
    uow := repositories.UnitOfWorkFromContext(ctx)
    customer, err := uow.Customers().GetByID(ctx, id)
    // ...
    return uow.Customers().Update(ctx, customer)
})
```

### 2.3 Errors
The service returns sentinel errors with stable codes. Callers match by `derrors.IsCode(err, "INSUFFICIENT_STOCK")` (the underlying `derrors` package). The service-level sentinels live in `application/services/errors.go`.

### 2.4 Logging
Every business outcome logs a structured event. Examples:

- `"customer created"` with `customer_id`, `company_id`, `document_number`
- `"sale created"` with `sale_id`, `number`, `customer_id`, `total`, `profit`
- `"journal entry posted"` with `entry_id`, `by`
- `"purchase approved"` with `po_id`

No PII (no full names, no phone numbers) and no full document numbers in logs.

### 2.5 Method naming
- Public method names are verbs in the present tense: `Create`, `Update`, `Cancel`, `Block`, `MarkAsPaid`, `Post`, `Reverse`, `Receive`, `Issue`, `Adjust`, `Transfer`.
- Each method is a single business operation. Multi-step work is split into one method per step, all orchestrated by the application use case layer.

### 2.6 Inputs
Every public method takes a typed `XxxInput` struct, not a long list of positional parameters. The struct groups the input logically and is easy to extend without breaking callers.

## 3. Per-service summary

### CustomerService (`customer/`)
- `Create`, `Update`, `Block`, `Unblock`, `Deactivate`
- `UpdateCreditLimit`, `ChangePaymentTerms` (indirectly through Update)
- `RecordSale`, `RecordPayment` (debt updates)
- `AvailableCredit`, `IsOverLimit`, `CanPlaceSale` (over-limit guard)
- `OutstandingBalance`, `GetByID`, `GetByDocument`
- Validates blocked/inactive status before mutating.

### SupplierService (`supplier/`)
- `Create`, `Update`, `Deactivate`
- `RecordPurchase`, `RecordPayment`
- `OutstandingBalance`, `CanPlacePurchase`
- `GetByID`

### ProductService (`product/`)
- `Create`, `UpdateCost`, `UpdateSalePrice`, `UpdateStockLimits`
- `Activate`, `Deactivate`
- `CalculateMargin(price, cost)` static helper for the UI
- All updates log the new cost/price/margin for visibility.

### InventoryService (`inventory/`)
- `Receive(ReceiveInput)` — initial receipt, creates a batch + a movement.
- `Issue(IssueInput)` and `Adjust(AdjustInput)` share the `consumeOrAdjust` helper which applies the change, blocks negative stock, and writes a movement row.
- `Transfer(TransferInput)` — same-product only, uses the UoW for both sides.
- `GenerateClearanceCandidates(companyID, at)` returns all batches past `max_sale_date` with `current_quantity > 0`.
- `NeedsClearanceSoon(companyID)` — within 3 days of clearance.
- `StockFor(batchID)` and `AgingReport(companyID)`.
- The 25-day rule is implemented at the domain layer (`InventoryBatch.MaximumSaleDate`, `IsClearance`, `NeedsClearanceSoon`); the service just surfaces it.

### PurchasingService (`purchasing/`)
- `Create(CreateInput)` with multiple line items.
- `Approve(id)`, `MarkAsReceived(id, at)`, `Cancel(id, reason)`, `Reconcile(id)`.
- `RegisterSupplierPayment(PayInput, purchaseID)` allocates to a PO.
- `GetByID(id)`.
- Does NOT generate journal entries — the application use case calls `AccountingService` in the same transaction.

### SalesService (`sales/`)
- `Create(CreateInput)` validates blocked customers, over-limit sales, and duplicate products, then constructs the sale and persists it. The returned `CreateResult` carries outbound-movement metadata that the use case passes to `InventoryService.Issue` in the same transaction.
- `Cancel`, `ApplyPayment`, `MarkAsPaid`, `MarkAsPartiallyPaid`.
- `OutstandingBalance`, `GetByID`.

### CustomerPaymentService (`customerpayments/`)
- `Register(PayInput)` — a customer payment.
- `RegisterAdvance(AdvanceInput)` — a customer advance (prepayment).
- `ApplyToSale(paymentID, saleID, amount)` — allocates part of a payment to a sale, updating both the payment's allocation list and the sale's paid amount, in the same transaction.
- `ApplyAdvanceToSale(advanceID, saleID, amount)` — same for advances.
- `OutstandingByCustomer(customerID)`.

### TreasuryService (`treasury/`)
- `OpenAccount`, `IssueCard`.
- `ChargeCard`, `PayCard` — credit-card lifecycle.
- `RegisterTransaction(RegisterTransactionInput)` — bank movement.
- `ReconcileTransaction(id, by)`.
- `UpsertExchangeRate(UpsertExchangeRateInput)`, `LatestExchangeRate(from, to)`.

### AccountingService (`accounting/`)
- `CreateEntry(EntryInput)` builds a draft entry, calls `entry.Validate()` BEFORE saving (unbalanced entries are rejected), persists.
- `Post(id, by)` transitions draft → posted (immutable thereafter).
- `Reverse(originalID, reversingID)` — marks the original as reversed.
- `AccountBalance`, `TrialBalance` (delegates to `LedgerRepository`).
- `CreateChartOfAccounts`, `ListChartOfAccounts`.

### ReportingService (`reporting/`)
- Read-only aggregations over repos.
- `ReceivableSummaryByCustomer`, `PayableSummaryBySupplier`, `ProfitInRange`.

## 4. Anti-patterns (do not)

- ❌ Business logic in repositories.
- ❌ Cross-service calls inside a single service. Use the use case layer to compose.
- ❌ Cross-aggregate updates without a transaction.
- ❌ Logging PII or full document numbers.
- ❌ Returning `*pgconn.PgError` or `*sql.Error` from services. Translate at the persistence boundary.
- ❌ Implicit `time.Now()` inside service methods. Accept `now` as a parameter for testability.
- ❌ `panic()` outside of constructor nil-checks.

## 5. Tests

`backend/internal/application/services/*/*_test.go` use:

- An in-memory fake `UnitOfWork` that exposes only the repos under test.
- A `fakeTxManager` that exposes the UoW via context and tracks commits/rollbacks.
- A silent logger (`*slog.Logger` writing to `io.Discard`).

Current coverage by service:

| Service       | Tests |
|---------------|-------|
| Customer      | 9 (create happy/sad, block, unblock, record sale, record payment, available credit, can-place-sale guard, deactivate) |
| Inventory     | 5 (receive, 25-day clearance rule, negative-stock guard, transfer same-product only, clearance candidates) |
| Sales         | 5 (create empty, create duplicate item, cancel, apply payment auto-status, outstanding balance) |
| Accounting    | 3 (create rejects unbalanced, balanced persists, post + double-post) |
| Supplier, Product, Purchasing, CustomerPayment, Treasury, Reporting | stubs (compile-tested only) |

All tests use mocked repositories. Integration tests that hit the real PostgreSQL live under `infrastructure/persistence/postgres/` (Phase 1.3) and cover the data layer. Service tests are pure unit tests.

Run with: `go test ./backend/internal/application/services/...`.
