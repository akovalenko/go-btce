package btce

import (
	"fmt"
	"net/url"
)

// Public API methods, results and parameters, per
// https://btc-e.com/api/3/documentation

// PublicInfo represents a result returned by GetInfo method of the public API
// (don't confuse with getInfo method of the private API)
type PublicInfo struct {
	ServerTime int64               `json:"server_time"`
	Pairs      map[string]PairInfo `json:"pairs"`
}

// PairInfo represents a currency pair data in PublicInfo
type PairInfo struct {
	DecimalPlaces uint    `json:"decimal_places"`
	MinPrice      float64 `json:"min_price"`
	MaxPrice      float64 `json:"max_price"`
	MinAmount     float64 `json:"min_amount"`
	Hidden        uint    `json:"hidden"`
	Fee           float64 `json:"fee"`
}

// TickerInfo represents a result of the "ticker" method of public API
type TickerInfo struct {
	High          float64 `json:"high"`
	Low           float64 `json:"low"`
	Average       float64 `json:"avg"`
	Volume        float64 `json:"vol"`
	CurrentVolume float64 `json:"vol_cur"`
	Buy           float64 `json:"buy"`
	Sell          float64 `json:"sell"`
	Updated       int64   `json:"updated"`
}

// Offer represents an ask or bid item in DepthInfo. It has to be
// decoded into an array; convenient accessors for Rate and Amount are
// added to avoid explicit indices.
type Offer [2]float64

func (o Offer) Rate() float64   { return o[0] }
func (o Offer) Amount() float64 { return o[1] }

// DepthInfo represents a result of the "depth" method of public API
type DepthInfo struct {
	Asks []Offer `json:"asks"`
	Bids []Offer `json:"bids"`
}

// GetTicker retrieves public ticker information on currency pairs
func (c Client) GetTicker(pairs []string) (map[string]TickerInfo, error) {
	tickers := map[string]TickerInfo{}
	err := c.CallPublicAPIv3("ticker", pairs, &tickers, nil)
	if err != nil {
		return nil, err
	}
	return tickers, nil
}

// GetDepth retrieves market depth information on currency pairs, up
// to limit items in both directions.
func (c Client) GetDepth(pairs []string, limit uint) (map[string]DepthInfo, error) {
	depth := map[string]DepthInfo{}
	err := c.CallPublicAPIv3("depth", pairs, &depth,
		&url.Values{"limit": []string{fmt.Sprint(limit)}})
	if err != nil {
		return nil, err
	}
	return depth, nil
}

// GetPublicInfo retrieves public API information on all available
// currency pairs, caching it for a given client once and for all.
func (c Client) GetPublicInfo() (*PublicInfo, error) {
	if c.Info == nil {
		info := &PublicInfo{}
		err := c.CallPublicAPIv3("info", nil, info, nil)
		if err != nil {
			return nil, err
		}
		c.Info = info
	}
	return c.Info, nil
}
