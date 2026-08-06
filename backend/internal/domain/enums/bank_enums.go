package enums

// BankAccountType categorizes a bank account.
type BankAccountType string

const (
	BankAccountTypeChecking BankAccountType = "checking"
	BankAccountTypeSavings  BankAccountType = "savings"
)

func (b BankAccountType) Valid() bool {
	return b == BankAccountTypeChecking || b == BankAccountTypeSavings
}

func (b BankAccountType) String() string { return string(b) }

// BankTransactionType classifies a single bank transaction.
type BankTransactionType string

const (
	BankTxTypeDeposit      BankTransactionType = "deposit"
	BankTxTypeWithdrawal   BankTransactionType = "withdrawal"
	BankTxTypeFee          BankTransactionType = "fee"
	BankTxTypeInterest     BankTransactionType = "interest"
	BankTxTypeTransfer     BankTransactionType = "transfer"
	BankTxTypeOther        BankTransactionType = "other"
)

func (b BankTransactionType) Valid() bool {
	switch b {
	case BankTxTypeDeposit, BankTxTypeWithdrawal, BankTxTypeFee,
		BankTxTypeInterest, BankTxTypeTransfer, BankTxTypeOther:
		return true
	}
	return false
}

func (b BankTransactionType) String() string { return string(b) }
