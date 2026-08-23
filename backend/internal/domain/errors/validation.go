package errors

// Validation errors. These are produced by value object constructors and
// by domain-level validators. The application layer maps them to caller-
// facing messages.

var (
	// ErrRequired — a required field was missing or empty.
	ErrRequired = New("REQUIRED", "field is required")

	// ErrInvalidFormat — a string did not match the expected format
	// (e.g. email, phone, document number).
	ErrInvalidFormat = New("INVALID_FORMAT", "value does not match the expected format")

	// ErrOutOfRange — a numeric value is outside the allowed bounds.
	ErrOutOfRange = New("OUT_OF_RANGE", "value is out of allowed range")

	// ErrInvalidEnum — a value is not a recognized enum member.
	ErrInvalidEnum = New("INVALID_ENUM", "value is not a valid enum member")

	// ErrInvalidSKU — a SKU does not match the SKU format rules.
	ErrInvalidSKU = New("INVALID_SKU", "SKU is malformed")

	// ErrInvalidBarcode — a barcode does not match EAN-13/UPC rules.
	ErrInvalidBarcode = New("INVALID_BARCODE", "barcode is malformed")

	// ErrInvalidEmail — value is not a valid email.
	ErrInvalidEmail = New("INVALID_EMAIL", "email is malformed")

	// ErrInvalidPhone — value is not a valid phone.
	ErrInvalidPhone = New("INVALID_PHONE", "phone is malformed")

	// ErrInvalidDocumentNumber — value is not a valid document
	// number (DNI/RUC/CE/etc.).
	ErrInvalidDocumentNumber = New("INVALID_DOCUMENT_NUMBER", "document number is malformed")
)
