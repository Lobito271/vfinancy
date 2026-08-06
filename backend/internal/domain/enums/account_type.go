package enums

// AccountType is the broad classification of a chart-of-accounts entry.
// Each account falls into exactly one of these categories, which drives
// the normal balance (debit vs credit) and the financial statement where
// it appears.
type AccountType string

const (
	AccountTypeAsset     AccountType = "asset"
	AccountTypeLiability AccountType = "liability"
	AccountTypeEquity    AccountType = "equity"
	AccountTypeIncome    AccountType = "income"
	AccountTypeExpense   AccountType = "expense"
)

func (a AccountType) Valid() bool {
	switch a {
	case AccountTypeAsset, AccountTypeLiability, AccountTypeEquity,
		AccountTypeIncome, AccountTypeExpense:
		return true
	}
	return false
}

// NormalBalance is the side that increases the account's balance:
//   * assets, expenses → debit
//   * liabilities, equity, income → credit
//
// This is used when posting a journal entry: the system can infer
// whether to debit or credit an account based on its type and the
// economic intent of the transaction.
func (a AccountType) NormalBalance() NormalBalance {
	switch a {
	case AccountTypeAsset, AccountTypeExpense:
		return DebitNormal
	}
	return CreditNormal
}

func (a AccountType) String() string { return string(a) }

// NormalBalance is the side that increases an account.
type NormalBalance string

const (
	DebitNormal  NormalBalance = "debit"
	CreditNormal NormalBalance = "credit"
)

func (n NormalBalance) String() string { return string(n) }
