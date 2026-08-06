package postgres

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"vfinancy/backend/internal/domain/enums"
	"vfinancy/backend/internal/domain/valueobjects"
)

// --- Money ---

func masterdataParseMoney(s string) (valueobjects.Money, error) {
	if s == "" {
		return valueobjects.Zero(), nil
	}
	return valueobjects.MoneyFromString(s)
}

// --- Document number ---

func masterdataParseDocument(docType, number string) (valueobjects.DocumentNumber, error) {
	dt := enums.DocumentType(strings.ToUpper(docType))
	return valueobjects.NewDocumentNumber(dt, number)
}

// --- FullName ---

func masterdataParseFullName(s string) valueobjects.FullName {
	if s == "" {
		return valueobjects.FullName{}
	}
	n, _ := valueobjects.NewFullName(s)
	return n
}

// --- Email ---

func masterdataParseEmail(s string) valueobjects.Email {
	if s == "" {
		return valueobjects.Email{}
	}
	e, _ := valueobjects.NewEmail(s)
	return e
}

// --- Phone ---

func masterdataParsePhone(s string) valueobjects.Phone {
	if s == "" {
		return valueobjects.Phone{}
	}
	p, _ := valueobjects.NewPhone(s)
	return p
}

// --- Address ---

func masterdataParseAddress(s string) valueobjects.Address {
	if s == "" {
		return valueobjects.Address{}
	}
	a, _ := valueobjects.NewAddress(s)
	return a
}

// --- UUID ---

func masterdataParseUUID(s string) uuid.UUID {
	id, _ := uuid.Parse(s)
	return id
}

// --- TaxCategory ---

func masterdataParseTaxCategory(s string) enums.TaxCategory {
	tc := enums.TaxCategory(s)
	if !tc.Valid() {
		return enums.TaxCategory("")
	}
	return tc
}

// --- CustomerStatus ---

func masterdataParseCustomerStatus(s string) enums.CustomerStatus {
	st := enums.CustomerStatus(s)
	if !st.Valid() {
		return enums.CustomerStatus("")
	}
	return st
}

// --- SupplierStatus ---

func masterdataParseSupplierStatus(s string) enums.SupplierStatus {
	st := enums.SupplierStatus(s)
	if !st.Valid() {
		return enums.SupplierStatus("")
	}
	return st
}

// --- SaleStatus ---

func masterdataParseSaleStatus(s string) enums.SaleStatus {
	st := enums.SaleStatus(s)
	if !st.Valid() {
		return enums.SaleStatus("")
	}
	return st
}

// --- PurchaseStatus ---

func masterdataParsePurchaseStatus(s string) enums.PurchaseStatus {
	st := enums.PurchaseStatus(s)
	if !st.Valid() {
		return enums.PurchaseStatus("")
	}
	return st
}

// --- UserStatus ---

func masterdataParseUserStatus(s string) enums.UserStatus {
	st := enums.UserStatus(s)
	if !st.Valid() {
		return enums.UserStatus("")
	}
	return st
}

// --- RoleType ---

func masterdataParseRoleType(s string) enums.RoleType {
	rt := enums.RoleType(s)
	if !rt.Valid() {
		return enums.RoleType("")
	}
	return rt
}

// --- AccountType ---

func masterdataParseAccountType(s string) enums.AccountType {
	at := enums.AccountType(s)
	if !at.Valid() {
		return enums.AccountType("")
	}
	return at
}

// --- JournalStatus ---

func masterdataParseJournalStatus(s string) enums.JournalStatus {
	js := enums.JournalStatus(s)
	if !js.Valid() {
		return enums.JournalStatus("")
	}
	return js
}

// --- JournalType ---

func masterdataParseJournalType(s string) enums.JournalType {
	jt := enums.JournalType(s)
	if !jt.Valid() {
		return enums.JournalType("")
	}
	return jt
}

// --- InventoryMovementType ---

func masterdataParseMovementType(s string) enums.InventoryMovementType {
	mt := enums.InventoryMovementType(s)
	if !mt.Valid() {
		return enums.InventoryMovementType("")
	}
	return mt
}

// --- PaymentMethod ---

func masterdataParsePaymentMethod(s string) enums.PaymentMethod {
	pm := enums.PaymentMethod(s)
	if !pm.Valid() {
		return enums.PaymentMethod("")
	}
	return pm
}

// --- SQL Null helpers ---

func nullIfEmptyFullName(n valueobjects.FullName) any {
	if n.String() == "" {
		return nil
	}
	return n.String()
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func nullIfZeroTime(t *time.Time) any {
	if t == nil || t.IsZero() {
		return nil
	}
	return *t
}

// --- Date helpers ---

func dateToString(d valueobjects.Date) string {
	if d.IsZero() {
		return ""
	}
	return d.Format("2006-01-02")
}

// --- Decimal helpers ---

func decimalFromString(s string) decimal.Decimal {
	d, _ := decimal.NewFromString(s)
	return d
}
