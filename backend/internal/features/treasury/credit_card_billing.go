package treasury

import (
	"time"

	"vfinancy/backend/internal/domain/valueobjects"
)

// CardBillingCycle is the projected statement window for a card charge.
type CardBillingCycle struct {
	// CutOffDate is the statement cut-off that captures the charge.
	CutOffDate valueobjects.Date
	// DueDate is the day the issuer must be paid.
	DueDate valueobjects.Date
}

// NextBillingCycle projects the first statement that would include a
// charge placed on anchor and the corresponding payment-due date. The
// cut-off and payment-due days come from the card's billing cycle. A
// charge made on the cut-off day itself belongs to that statement.
func (c *CreditCard) NextBillingCycle(anchor time.Time) CardBillingCycle {
	cutOff := monthDay(anchor, c.CutOffDay)
	if dateAfter(anchor, cutOff) {
		cutOff = monthDay(addMonth(cutOff), c.CutOffDay)
	}
	due := monthDay(addMonth(cutOff), c.PaymentDueDay)
	return CardBillingCycle{
		CutOffDate: valueobjects.NewDateFromTime(cutOff),
		DueDate:    valueobjects.NewDateFromTime(due),
	}
}

// monthDay returns midnight of the given month at day, clamping the day
// to the last day of the month (e.g. a 31st cut-off in February).
func monthDay(anchor time.Time, day int) time.Time {
	y, m, _ := anchor.Date()
	if last := lastDay(y, m); day > last {
		day = last
	}
	return time.Date(y, m, day, 0, 0, 0, 0, anchor.Location())
}

// addMonth returns the same day one month later, keeping the result
// inside the following month (clamped to its last day).
func addMonth(t time.Time) time.Time {
	y, m, d := t.Date()
	if last := lastDay(y, m+1); d > last {
		d = last
	}
	return time.Date(y, m+1, d, 0, 0, 0, 0, t.Location())
}

// lastDay returns the number of days in a given month (months are
// 1-based; month 13 rolls into January of the next year).
func lastDay(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// dateAfter reports whether anchor (a real timestamp) is strictly after
// midnight of target in calendar terms.
func dateAfter(anchor, target time.Time) bool {
	ay, am, ad := anchor.Date()
	ty, tm, td := target.Date()
	if ay != ty {
		return ay > ty
	}
	if am != tm {
		return am > tm
	}
	return ad > td
}

// CardPaymentProjection is the projected payment summary for a single
// credit card in the current billing cycle.
type CardPaymentProjection struct {
	CardID         string  `json:"cardId"`
	Issuer         string  `json:"issuer"`
	LastFour       string  `json:"lastFour"`
	CardHolder     string  `json:"cardHolder"`
	ProjectedUSD   float64 `json:"projectedUSD"`
	CycleStart     string  `json:"cycleStart"`
	NextCutOffDate string  `json:"nextCutOffDate"`
	NextPaymentDate string `json:"nextPaymentDate"`
}

// CurrentCycleStart returns the start date of the current billing cycle
// for the given card relative to today. The cycle begins the day after
// the most recent past cut-off date.
func (c *CreditCard) CurrentCycleStart(now time.Time) time.Time {
	cutOff := monthDay(now, c.CutOffDay)
	if dateAfter(now, cutOff) || now.Equal(cutOff) {
		return time.Date(cutOff.Year(), cutOff.Month(), cutOff.Day()+1, 0, 0, 0, 0, time.UTC)
	}
	prevMonth := time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, time.UTC)
	prevCutOff := monthDay(prevMonth, c.CutOffDay)
	return time.Date(prevCutOff.Year(), prevCutOff.Month(), prevCutOff.Day()+1, 0, 0, 0, 0, time.UTC)
}
