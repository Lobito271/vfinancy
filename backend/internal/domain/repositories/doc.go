// Package repositories is the persistence abstraction. Every concrete
// implementation lives in the infrastructure layer; the application
// layer takes a dependency on the interfaces defined here.
//
// The interfaces are organized by bounded context. Each context has
// its own file:
//
//   * user_repository.go, role_repository.go, permission_repository.go
//   * customer_repository.go, supplier_repository.go, product_repository.go,
//     category_repository.go, brand_repository.go,
//     warehouse_currency_repository.go
//   * inventory_batch_repository.go, inventory_movement_repository.go
//   * purchase_repository.go, supplier_payment_repository.go
//   * sales_repository.go, customer_payment_repository.go
//   * treasury_repository.go
//   * accounting_repository.go
//
// Shared abstractions live in the same package:
//
//   * pagination.go — PageRequest, Page[T], Sort, TimeRange
//   * errors.go     — ErrNotFound, ErrDuplicate, ErrForeignKey, ...
//   * transaction.go — TransactionManager, TxRunner
//   * unit_of_work.go — UnitOfWork, ContextWithUnitOfWork
//
// The package depends only on the standard library, the domain
// entities, and value objects. It must never import any infrastructure
// package.
package repositories
