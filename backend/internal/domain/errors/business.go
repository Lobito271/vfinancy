package errors

// Business rule violations. These errors mean the caller asked the domain
// to do something that contradicts a business invariant.

var (
	// ErrInsufficientStock — the requested quantity exceeds the
	// available stock for a (product, warehouse) pair.
	ErrInsufficientStock = New("INSUFFICIENT_STOCK", "stock is not enough for the requested movement")

	// ErrInvalidPayment — a payment amount is invalid (zero, negative,
	// or not aligned to the document currency rules).
	ErrInvalidPayment = New("INVALID_PAYMENT", "payment amount is invalid")

	// ErrNegativeQuantity — a quantity argument is negative or zero
	// when the operation requires a positive value.
	ErrNegativeQuantity = New("NEGATIVE_QUANTITY", "quantity must be greater than zero")

	// ErrNegativeMoney — a monetary argument is negative when the
	// operation requires a non-negative value.
	ErrNegativeMoney = New("NEGATIVE_MONEY", "monetary value must be non-negative")

	// ErrCustomerInactive — an operation is attempted on a customer
	// that is not active (e.g. blocked or inactive).
	ErrCustomerInactive = New("CUSTOMER_INACTIVE", "customer is not active")

	// ErrSupplierInactive — an operation is attempted on a supplier
	// that is not active.
	ErrSupplierInactive = New("SUPPLIER_INACTIVE", "supplier is not active")

	// ErrInvalidCurrency — currencies in an operation do not match
	// (e.g. mixing PEN and USD on the same line item).
	ErrInvalidCurrency = New("INVALID_CURRENCY", "currencies do not match")

	// ErrSaleAlreadyPaid — a payment is being applied to a sale that
	// is already fully paid.
	ErrSaleAlreadyPaid = New("SALE_ALREADY_PAID", "sale is already fully paid")

	// ErrPurchaseCancelled — an operation is attempted on a purchase
	// that is already cancelled.
	ErrPurchaseCancelled = New("PURCHASE_CANCELLED", "purchase is cancelled")

	// ErrInvalidStateTransition — the requested state transition is
	// not allowed from the current state.
	ErrInvalidStateTransition = New("INVALID_STATE_TRANSITION", "state transition is not allowed")

	// ErrEmptyDocument — a sale or purchase is being persisted with
	// no line items.
	ErrEmptyDocument = New("EMPTY_DOCUMENT", "document must have at least one line item")

	// ErrPaymentExceedsBalance — a payment (or advance application)
	// exceeds the outstanding balance of the document.
	ErrPaymentExceedsBalance = New("PAYMENT_EXCEEDS_BALANCE", "payment exceeds outstanding balance")

	// ErrAdvanceNegative — a customer advance would become negative
	// after an application.
	ErrAdvanceNegative = New("ADVANCE_NEGATIVE", "advance cannot become negative")

	// ErrUnbalancedJournalEntry — a journal entry does not satisfy
	// SUM(debit) == SUM(credit).
	ErrUnbalancedJournalEntry = New("UNBALANCED_JOURNAL_ENTRY", "journal entry must balance debits and credits")

	// ErrJournalEntryAlreadyPosted — attempting to mutate or post a
	// journal entry that is already posted.
	ErrJournalEntryAlreadyPosted = New("JOURNAL_ENTRY_POSTED", "journal entry is already posted")

	// ErrJournalEntryReversed — attempting to mutate a journal entry
	// that has been reversed.
	ErrJournalEntryReversed = New("JOURNAL_ENTRY_REVERSED", "journal entry has been reversed")

	// ErrClosedPeriod — attempting to post into a closed fiscal period.
	ErrClosedPeriod = New("CLOSED_PERIOD", "fiscal period is closed")

	// ErrDuplicateLine — a document line with the same product already
	// exists; the caller should add to the existing line or remove it first.
	ErrDuplicateLine = New("DUPLICATE_LINE", "document already contains a line for this product")
)
