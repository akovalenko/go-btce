package btce

// Private API methods, results and parameters, per
// https://wex.nz/tapi/docs

type ActiveOrdersParameters struct {
	Pair string
}

type ActiveOrdersResult map[uint64]ActiveOrder

type ActiveOrder struct {
	Pair             string
	Type             string
	Amount           float64
	Rate             float64
	TimestampCreated int64 `json:"timestamp_created"`
	Status           uint
}

type GetInfoParameters struct{}

type GetInfoResult struct {
	Funds  map[string]float64
	Rights struct {
		Info     uint
		Trade    uint
		Withdraw uint
	}
	TransactionCount uint  `json:"transaction_count"`
	OpenOrders       uint  `json:"open_orders"`
	ServerTime       int64 `json:"server_time"`
}

type TradeParameters struct {
	Pair   string
	Type   string
	Rate   float64
	Amount float64
}

type TradeResult struct {
	Received float64
	Remains  float64
	OrderId  uint64 `json:"order_id"`
	Funds    map[string]float64
}

type OrderInfoParameters struct {
	OrderId uint64
}

type OrderInfo struct {
	Pair             string
	Type             string
	StartAmount      float64 `json:"start_amount"`
	Amount           float64
	Rate             float64
	TimestampCreated int64
	Status           uint
}

type OrderInfoResult map[uint64]OrderInfo

type CancelOrderParameters struct {
	OrderId uint64
}

type CancelOrderResult struct {
	OrderId uint64 `json:"order_id"`
	Funds   map[string]float64
}

type TradeHistoryParameters struct {
	From   uint
	Count  uint
	FromId uint64
	EndId  uint64
	Order  string
	Since  int64
	End    int64
	Pair   string
}

type TradeHistoryResult map[uint64]TradeHistoryItem
type TradeHistoryItem struct {
	Pair        string
	Type        string
	Amount      float64
	Rate        float64
	OrderId     uint64 `json:"order_id"`
	IsYourOrder uint   `json:"is_your_order"`
	Timestamp   int64
}

type TransHistoryParameters struct {
	From   uint
	Count  uint
	FromId uint64
	EndId  uint64
	Order  string
	Since  int64
	End    int64
}
type TransHistoryResult map[uint64]TransHistoryItem

type TransHistoryItem struct {
	Type      string
	Amount    float64
	Currency  string
	Desc      string
	Status    uint
	Timestamp int64
}

type CoinDepositAddressParameters struct{ CoinName string }
type CoinDepositAddressResult struct{ Address string }

type WithdrawCoinParameters struct {
	CoinName string
	Amount   float64
	Address  string
}

type WithdrawCoinResult struct {
	TransId    uint64 `json:"tId"`
	AmountSent float64
	Funds      map[string]float64
}

type CreateCouponParameters struct {
	Currency string
	Amount   float64
	Receiver string
}
type CreateCouponResult struct {
	Coupon  string
	TransId uint64
	Funds   map[string]float64
}

type RedeemCouponParameters struct{ Coupon string }
type RedeemCouponResult struct{
	CouponAmount float64
	CouponCurrency string
	TransId uint64
	Funds map[string]float64
}
