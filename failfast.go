package btce

// Ticker is a single-return, panicking-on-error wrapper for GetTicker
func (c *Client) Ticker(pairs []string) map[string]TickerInfo {
	result, err := c.GetTicker(pairs)
	if err != nil {
		panic(err)
	}
	return result
}

// Depth is a single-return, panicking-on-error wrapper for GetDepth
func (c *Client) Depth(pairs []string, limit uint) map[string]DepthInfo {
	result, err := c.GetDepth(pairs, limit)
	if err != nil {
		panic(err)
	}
	return result
}

// PublicInfo is a single-return, panicking-on-error wrapper for
// GetPublicInfo (calling GetInfo method of public V3 API, caching
// result)
func (c *Client) PublicInfo() *PublicInfo {
	result, err := c.GetPublicInfo()
	if err != nil {
		panic(err)
	}
	return result
}

// PrivateInfo is a single-return, panicking-on-error wrapper for
// getInfo - see https://btc-e.nz/tapi/docs#getInfo
func (c *Client) PrivateInfo() GetInfoResult {
	result := GetInfoResult{}
	err := c.Call(GetInfoParameters{}, &result)
	if err != nil {
		panic(err)
	}
	return result
}

// ActiveOrders is a single-return, panicking-on-error wrapper for private
// API method ActiveOrders - see https://btc-e.nz/tapi/docs#ActiveOrders
func (c *Client) ActiveOrders(p ActiveOrdersParameters) ActiveOrdersResult {
	result := ActiveOrdersResult{}
	err := c.Call(p, &result)
	if err != nil {
		panic(err)
	}
	return result
}

// Trade is a single-return, panicking-on-error wrapper for private
// API method Trade - see https://btc-e.nz/tapi/docs#Trade
func (c *Client) Trade(p TradeParameters) TradeResult {
	result := TradeResult{}
	err := c.Call(p, &result)
	if err != nil {
		panic(err)
	}
	return result
}

// OrderInfo is a single-return, panicking-on-error wrapper for private
// API method OrderInfo - see https://btc-e.nz/tapi/docs#OrderInfo
func (c *Client) OrderInfo(p OrderInfoParameters) OrderInfoResult {
	result := OrderInfoResult{}
	err := c.Call(p, &result)
	if err != nil {
		panic(err)
	}
	return result
}

// CancelOrder is a single-return, panicking-on-error wrapper for private
// API method CancelOrder - see https://btc-e.nz/tapi/docs#CancelOrder
func (c *Client) CancelOrder(p CancelOrderParameters) CancelOrderResult {
	result := CancelOrderResult{}
	err := c.Call(p, &result)
	if err != nil {
		panic(err)
	}
	return result
}

// TradeHistory is a single-return, panicking-on-error wrapper for private
// API method TradeHistory - see https://btc-e.nz/tapi/docs#TradeHistory
func (c *Client) TradeHistory(p TradeHistoryParameters) TradeHistoryResult {
	result := TradeHistoryResult{}
	err := c.Call(p, &result)
	if err != nil {
		panic(err)
	}
	return result
}

// TransHistory is a single-return, panicking-on-error wrapper for private
// API method TransHistory - see https://btc-e.nz/tapi/docs#TransHistory
func (c *Client) TransHistory(p TransHistoryParameters) TransHistoryResult {
	result := TransHistoryResult{}
	err := c.Call(p, &result)
	if err != nil {
		panic(err)
	}
	return result
}

// CoinDepositAddress is a single-return, panicking-on-error wrapper for private
// API method CoinDepositAddress - see https://btc-e.nz/tapi/docs#CoinDepositAddress
func (c *Client) CoinDepositAddress(p CoinDepositAddressParameters) CoinDepositAddressResult {
	result := CoinDepositAddressResult{}
	err := c.Call(p, &result)
	if err != nil {
		panic(err)
	}
	return result
}

// WithdrawCoin is a single-return, panicking-on-error wrapper for private
// API method WithdrawCoin - see https://btc-e.nz/tapi/docs#WithdrawCoin
func (c *Client) WithdrawCoin(p WithdrawCoinParameters) WithdrawCoinResult {
	result := WithdrawCoinResult{}
	err := c.Call(p, &result)
	if err != nil {
		panic(err)
	}
	return result
}

// CreateCoupon is a single-return, panicking-on-error wrapper for private
// API method CreateCoupon - see https://btc-e.nz/tapi/docs#CreateCoupon
func (c *Client) CreateCoupon(p CreateCouponParameters) CreateCouponResult {
	result := CreateCouponResult{}
	err := c.Call(p, &result)
	if err != nil {
		panic(err)
	}
	return result
}

// RedeemCoupon is a single-return, panicking-on-error wrapper for private
// API method RedeemCoupon - see https://btc-e.nz/tapi/docs#RedeemCoupon
func (c *Client) RedeemCoupon(p RedeemCouponParameters) RedeemCouponResult {
	result := RedeemCouponResult{}
	err := c.Call(p, &result)
	if err != nil {
		panic(err)
	}
	return result
}
