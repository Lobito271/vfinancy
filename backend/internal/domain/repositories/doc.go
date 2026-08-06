// Package repositories is the persistence abstraction. Every concrete
// implementation lives in the infrastructure layer; the application
// layer takes a dependency on the types defined here.
//
// Repository interfaces for each aggregate are defined in the owning
// feature package (internal/features/<context>). This package holds
// only the shared abstractions those interfaces build on:
//
//   * pagination.go — PageRequest, Page[T], Sort, TimeRange, UUIDSlice
//   * errors.go     — ErrNotFound, ErrDuplicate, ErrForeignKey, ...
//   * transaction.go — TransactionManager, TxRunner
//
// The package depends only on the standard library, the domain
// errors, and value objects. It must never import any infrastructure
// package.
package repositories
