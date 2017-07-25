package btce

import (
//	"errors"
	"math"
)

func makeOffer(rate float64, amount float64) Offer {
	return Offer{rate, amount}
}

func sumDepth(amount float64, offers []Offer) float64 {
	if math.IsNaN(amount) {
		return math.NaN()
	}
	cost := 0.0
	for amount > 0 {
		if len(offers) == 0 {
			return math.NaN()
		}
		if amount < offers[0].Amount() {
			cost += amount * offers[0].Rate()
			return cost
		}
		cost += offers[0].Amount() * offers[0].Rate()
		amount -= offers[0].Amount()
		offers = offers[1:]
	}
	return cost
}

// Invert returns a DepthInfo as if the base currency was swapped with
// the secondary one.
func (source DepthInfo) Invert() DepthInfo {
	new := DepthInfo{
		Asks: make([]Offer, len(source.Bids)),
		Bids: make([]Offer, len(source.Asks)),
	}

	for i, offer := range source.Bids {
		new.Asks[i] = makeOffer(1/offer.Rate(),
			offer.Amount()*offer.Rate())
	}
	for i, offer := range source.Asks {
		new.Bids[i] = makeOffer(1/offer.Rate(),
			offer.Amount()*offer.Rate())
	}

	return new
}

type Evaluation struct {
	Buy  float64
	Sell float64
}

// Evaluate returns potential result of buying or selling the
// specified amount of the pair's base currency. NaN is used if there
// are not enough asks or bids for buying or selling the amount.
func (source DepthInfo) Evaluate(amount float64) Evaluation {
	return Evaluation{Buy: sumDepth(amount, source.Asks),
		Sell: sumDepth(amount, source.Bids)}
}
