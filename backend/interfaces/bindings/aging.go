package bindings

import "github.com/google/uuid"

type AgingDTO struct {
	Open          string `json:"open"`
	Overdue0to30  string `json:"overdue0to30"`
	Overdue31to60 string `json:"overdue31to60"`
	Overdue61to90 string `json:"overdue61to90"`
	Overdue90plus string `json:"overdue90plus"`
}

func agingDTO(open string, buckets map[string]string) *AgingDTO {
	return &AgingDTO{Open: open, Overdue0to30: buckets["0-30"], Overdue31to60: buckets["31-60"], Overdue61to90: buckets["61-90"], Overdue90plus: buckets["90+"]}
}

func (a *App) GetCustomerAging(customerID string) (*AgingDTO, error) {
	id, err := uuid.Parse(customerID)
	if err != nil {
		return nil, err
	}
	open, err := a.accountsReceivable.GetOpenBalanceForCustomer(a.Context(), id)
	if err != nil {
		return nil, err
	}
	buckets, err := a.accountsReceivable.ListAgingBucket(a.Context(), id)
	if err != nil {
		return nil, err
	}
	return agingDTO(open, buckets), nil
}

func (a *App) GetSupplierAging(supplierID string) (*AgingDTO, error) {
	id, err := uuid.Parse(supplierID)
	if err != nil {
		return nil, err
	}
	open, err := a.accountsPayable.GetOpenBalanceForSupplier(a.Context(), id)
	if err != nil {
		return nil, err
	}
	buckets, err := a.accountsPayable.ListAgingBucket(a.Context(), id)
	if err != nil {
		return nil, err
	}
	return agingDTO(open, buckets), nil
}
