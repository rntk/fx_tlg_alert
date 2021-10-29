package patterns

import (
	"math"

	"fx_alert/pkg/quoter"
)

const PinBarPercent = 70.0

func pinBar(q quoter.Quote) *Pattern {
	h := q.High - q.Low

	t := q.High - math.Max(q.Open, q.Close)
	p := (t * 100) / h
	if p >= PinBarPercent {
		return &Pattern{Name: PinBar, Sentiment: Bear}
	}

	t = math.Min(q.Open, q.Close) - q.Low
	p = (t * 100) / h
	if p >= PinBarPercent {
		return &Pattern{Name: PinBar, Sentiment: Bull}
	}

	return nil
}

func FindPattern(q quoter.Quote) *Pattern {
	if p := pinBar(q); p != nil {
		return p
	}

	return nil
}
