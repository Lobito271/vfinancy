package treasury

import (
	"time"

	"vfinancy/backend/internal/domain/valueobjects"
)

// NextBillingCycle projects the statement cycle that captures a charge
// placed on at. The cycle starts the day after the previous cut-off,
// closes on the next cut-off (clamped to the month's last day: a 31st
// cut-off falls on Feb 28/29 or the 30th of short months) and the
// payment-due date is the due day of the following month with the same
// clamp. A charge made on the cut-off day itself belongs to that
// statement.
func NextBillingCycle(card *CreditCard, at time.Time) (cycleStart, cycleEnd, paymentDue time.Time) {
	cutOff := monthDay(at, card.CutOffDay)
	if dateAfter(at, cutOff) {
		cutOff = monthDay(addMonth(cutOff), card.CutOffDay)
	}
	start := monthDay(time.Date(cutOff.Year(), cutOff.Month()-1, 1, 0, 0, 0, 0, cutOff.Location()), card.CutOffDay)
	start = time.Date(start.Year(), start.Month(), start.Day()+1, 0, 0, 0, 0, time.UTC)
	due := monthDay(addMonth(cutOff), card.PaymentDueDay)
	return start, cutOff, due
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
// credit card in its current billing cycle.
type CardPaymentProjection struct {
	Card       *CreditCard
	CycleStart time.Time
	CycleEnd   time.Time
	PaymentDue time.Time
	// TotalUSD sums cost_usd of purchase orders charged to the card
	// inside the cycle (cancelled orders excluded).
	TotalUSD valueobjects.Money
	// RefundsUSD sums refund_amount of cancelled purchase orders in
	// the cycle (saldo a favor).
	RefundsUSD valueobjects.Money
	// Status is "open" or "settled" (settled once the cycle end has
	// passed).
	Status string
}

// CurrentCycleStart returns the start date of the current billing cycle
// for the given card relative to now. The cycle begins the day after
// the most recent past cut-off date; a cut-off falling today means the
// next cycle has already started.
func (c *CreditCard) CurrentCycleStart(now time.Time) time.Time {
	cutOff := monthDay(now, c.CutOffDay)
	if dateAfter(now, cutOff) || now.Equal(cutOff) {
		return time.Date(cutOff.Year(), cutOff.Month(), cutOff.Day()+1, 0, 0, 0, 0, time.UTC)
	}
	prevMonth := time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, time.UTC)
	prevCutOff := monthDay(prevMonth, c.CutOffDay)
	return time.Date(prevCutOff.Year(), prevCutOff.Month(), prevCutOff.Day()+1, 0, 0, 0, 0, time.UTC)
}

// cycleEnd projects the cut-off closing the cycle opened at start,
// clamped to its month.
func (c *CreditCard) cycleEnd(start time.Time) time.Time {
	end := monthDay(start, c.CutOffDay)
	if !dateAfter(end, start) {
		end = monthDay(addMonth(end), c.CutOffDay)
	}
	return end
}
