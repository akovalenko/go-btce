package btce

import (
	"encoding/json"
)


// DefaultBaseURL is a base URL used by default if Client is not
// created with NewClient (where URL parameter is mandatory).
//
// Alternative domain in .nz is used for the sake of Russian firewall
// victims, having btc-e.com blocked.
const DefaultBaseURL = "https://wex.nz"

// DefaultClient is an application-wide instance of btce.Client, just
// to avoid passing it around when you only need one.
var DefaultClient = &Client{}

// Auth represents BTC-e API authentication data: key, secret and
// the current nonce value.
type Auth struct {
	Key    string `json:"key"`
	Secret string `json:"secret"`
	Nonce  uint64 `json:"nonce"`
}

// Retries describe how many times to retry a call on "invalid nonce"
// error (with nonce correction) and on other HTTP, server-side or
// decoding errors.
type Retries struct {
	NonceCorrection uint
	GeneralError    uint
}

// DefaultRetries are used when client.Retries is nil.
//
// Having non-zero NonceCorrection is essential when there's no
// guarantee of accurate Auth.Nonce tracking (and there isn't, unless
// you take care of saving Auth.Nonce <em>synchronously</em> after
// each remote call). Increasing NonceCorrection is useful when the
// same key can be occasionally used by several applications (like, a
// bot doing its long-term work and a user doing a one-off operation
// from the shell). Separate keys should be used in this case, so
// NonceCorrection==1 will always be enough, but we default to 10 to
// lower a chance of breaking in a non-standard situation when you
// just can't do it.
//
// Larger GeneralError retries can be useful, if not socially
// responsible, to get through during a high load or DDoS attack on
// btc-e.
var DefaultRetries = Retries{NonceCorrection: 10, GeneralError: 5}

// Client represents btc-e.com client settings
type Client struct {
	URL     string
	Info    *PublicInfo // queried and stored when first needed
	Auth    Auth
	Retries *Retries
}

// RemoteResult represents a result of a private API call (always
// having that format) or a result of a FAILED public API call (having
// that format on errors only)
type RemoteResult struct {
	Success uint             `json:"success"`
	Error   string           `json:"error"`
	Return  *json.RawMessage `json:"return"`
}
