package btce

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"
)

// HttpClient is an http.Client used for requests to btc-e.
var HttpClient = &http.Client{Timeout: 5 * time.Second}

// NewClient creates a BTC-e client, given a base URL which should
// normally be either https://btc-e.com or https://btc-e.nz (for
// victims of the Russian firewall).
//
// Public API methods can be called immediately on newly-created
// client. For private methods, use ReadKey to load API authentication
// data or initialize client.Auth as you want.
func NewClient(rawurl string) (*Client, error) {
	return &Client{URL: rawurl}, nil
}

// ReadKey loads an API key in JSON format from a file.  The value of
// "nonce" may be present, but it's not necessary: server's "invalid
// nonce" message provides corrected nonce value and are used for
// retrying private API calls immediately.
func (c *Client) ReadKey(fileName string) error {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	return decoder.Decode(&c.Auth)
}

// ResolveReference resolves (supposedly-)relative URL according to
// the base URL of a client.
func (c *Client) ResolveReference(path string) string {
	baseStr := c.URL
	if baseStr == "" {
		baseStr = DefaultBaseURL
	}
	baseURL, err := url.Parse(baseStr)
	if err != nil {
		log.Panic(err)
	}
	refURL, err := url.Parse(path)
	if err != nil {
		log.Panic(err)
	}
	return baseURL.ResolveReference(refURL).String()
}

// CallPublicAPIv3 calls a public API method, giving the list of pairs
// and optional extra GET parameters in url.Values, parsing the result
// into v on success. In addition to HTTP and decode errors,
// server-side call failure is checked and returned in the same way.
func (c *Client) CallPublicAPIv3(method string, pairs []string, v interface{}, values *url.Values) error {
	path := "/api/3/" + method + "/" + strings.Join(pairs, "-")
	if values != nil {
		url := &url.URL{Path: path, RawQuery: values.Encode()}
		path = url.String()
	}
	req, err := http.NewRequest("GET", c.ResolveReference(path), nil)
	data, err := c.doHttp(req, c.retries().GeneralError)
	if err != nil {
		return err
	}
	errDecoder := json.NewDecoder(bytes.NewReader(data))
	okDecoder := json.NewDecoder(bytes.NewReader(data))
	result := &RemoteResult{Success: 1}
	err = errDecoder.Decode(result)
	if err == nil && result.Success == 0 {
		return errors.New(result.Error)
	}
	return okDecoder.Decode(v)
}

// SignQuery signs a query string (including nonce) with a secret
// using SHA512 HMAC
func (a Auth) SignQuery(query string) string {
	hmac := hmac.New(sha512.New, []byte(a.Secret))
	hmac.Write([]byte(query))
	return hex.EncodeToString(hmac.Sum(nil))
}

func (c *Client) makeRemoteRequest(url string, v url.Values) (*http.Request, error) {
	query := v.Encode()
	req, err := http.NewRequest("POST", url, strings.NewReader(query))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Key", c.Auth.Key)
	req.Header.Add("Sign", c.Auth.SignQuery(query))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	return req, nil
}

func (c *Client) doHttp(req *http.Request, retries uint) ([]byte, error) {
	resp, err := HttpClient.Do(req)
	if err == nil {
		if resp.StatusCode == 200 {
			defer resp.Body.Close()
			data, err := ioutil.ReadAll(resp.Body)
			if err == nil {
				return data, nil
			}
		} else {
			resp.Body.Close()
		}
	}
	if retries == 0 {
		return nil, err
	}
	if req.Body != nil {
		if req.GetBody == nil {
			return nil, err
		}
		req.Body, err = req.GetBody()
		if err != nil {
			return nil, err
		}
	}
	return c.doHttp(req, retries-1)
}

// remoteCall calls a remote method param["method"] with given
// parameters, setting param["nonce"] from c.Auth.Nonce and
// incrementing the latter. Error returns represent failures of HTTP
// and decoding, but not remote-call failure, which is a normal
// RemoteResult with Success==0.
func (c *Client) remoteCall(param map[string]string) (*RemoteResult, error) {
	v := url.Values{}
	for name, value := range param {
		v.Set(name, value)
	}
	v.Set("nonce", fmt.Sprint(c.Auth.Nonce))
	c.Auth.Nonce++

	req, err := c.makeRemoteRequest(c.ResolveReference("/tapi"), v)
	if err != nil {
		return nil, err
	}
	data, err := c.doHttp(req, c.retries().GeneralError)
	if err != nil {
		return nil, err
	}
	result := &RemoteResult{}
	err = json.Unmarshal(data, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// remoteCallRetryNonce wraps remoteCall, ensuring c.Auth.Nonce
// correction when needed
func (c *Client) remoteCallRetry(param map[string]string) (*RemoteResult, error) {
	retries := c.retries()
	for {
		result, err := c.remoteCall(param)
		if err == nil {
			if result.Success == 0 &&
				strings.HasPrefix(result.Error, "invalid nonce parameter;") {
				if retries.NonceCorrection == 0 {
					return result, err
				}
				retries.NonceCorrection--
				newNonceString := result.Error[strings.LastIndex(result.Error, ":")+1:]
				_, err = fmt.Sscan(newNonceString, &c.Auth.Nonce)
				if err != nil {
					return nil, err
				}
				if traceRpc {
					log.Println("Nonce replaced:", c.Auth.Nonce)
				}
			} else {
				return result, err
			}
		} else {
			return nil, err
		}
	}
}

// isBlank checks if parameter should be omitted. For now it's zero
// values of integer fields, supposed to be a filtering parameters
// (defaulting to zero, or infinite -- with literal zero not making
// sense, or an arbitrary default limit (count) -- with literal zero
// not making sense either).
func isBlank(v reflect.Value) bool {
	kind := v.Kind()
	switch {
	case reflect.Int <= kind && kind <= reflect.Int64:
		return v.Int() == 0
	case reflect.Uint <= kind && kind <= reflect.Uint64:
		return v.Uint() == 0
	default:
		return false
	}
}

// replaceSuffix replaces string suffix (if present) with another one
func replaceSuffix(value string, suffix string, replacement string) string {
	if strings.HasSuffix(value, suffix) {
		return strings.TrimSuffix(value, suffix) + replacement
	} else {
		return value
	}
}

func formatValue(name string, v reflect.Value, pairInfo PairInfo) string {
	switch name {
	case "amount":
		return fmt.Sprintf("%.8f", v.Float())
	case "rate":
		return fmt.Sprintf("%."+fmt.Sprint(pairInfo.DecimalPlaces)+"f",
			v.Float())
	default:
		if isBlank(v) {
			return ""
		} else {
			return fmt.Sprint(v.Interface())
		}
	}
}

func (c *Client) formatParameters(v interface{}) (map[string]string, error) {
	info, err := c.GetPublicInfo()
	if err != nil {
		return nil, err
	}
	param := map[string]string{}
	ti := reflect.TypeOf(v)
	vi := reflect.ValueOf(v)
	name := ti.Name()
	if !strings.HasSuffix(name, "Parameters") {
		return nil, errors.New("Expected struct named ****Parameters")
	}
	methodName := strings.TrimSuffix(name, "Parameters")
	if methodName == "GetInfo" {
		methodName = "getInfo"
	}
	param["method"] = methodName
	vPair := vi.FieldByName("Pair") // concerning pair
	var pairInfo PairInfo
	if vPair.IsValid() {
		var ok bool
		if pairName := vPair.String(); pairName != "" {
			pairInfo, ok = info.Pairs[pairName]
			if !ok {
				return nil, errors.New("Unknown pair: " + pairName)
			}
		}
	}
	for i := 0; i < ti.NumField(); i++ {
		fieldName := ti.Field(i).Name
		paramName := strings.ToLower(replaceSuffix(fieldName, "Id", "_id"))
		if paramName == "coinname" {
			paramName = "coinName"
		}
		paramValue := vi.Field(i)
		stringValue := formatValue(paramName, paramValue, pairInfo)
		if stringValue != "" {
			param[paramName] = stringValue
		}
	}
	return param, nil
}

var traceRpc bool

func init() {
	flag.BoolVar(&traceRpc, "traceRpc", false, "Trace BTC-e RPC calls")
}

// Call invokes a private API method, with pstruct representing
// parameters and dst representing a return value. Type of pstruct
// should be a struct type with name ending with "Parameters" and
// started with method name (exception is a private getInfo method
// whose name is capitalized in GetInfoParameters).
//
// Destination structure is decoded as JSON from the "return" field of
// the answer. Predefined structures provided by this library (see
// private.go) have names starting with method name (umm, getInfo
// again) and ending with "Result", by convention.
//
// For each known method, there's also a convenience wrapper that
// returns appropriate result structure as a single value, panicking
// on errors (see ActiveOrders, Trade, OrderInfo...).
func (c *Client) Call(pstruct interface{}, dst interface{}) error {
	param, err := c.formatParameters(pstruct)
	if err != nil {
		return err
	}
	if traceRpc {
		log.Println("RPC param:", param)
	}
	result, err := c.remoteCallRetry(param)
	if traceRpc && result != nil {
		if result.Return != nil {
			log.Println("RPC result/ success:", result.Success,
				"error:", result.Error,
				"return(json):", string(*result.Return))
		} else {
			log.Println("RPC result/ success:", result.Success,
				"error:", result.Error,
				"no return(json)")
		}
	}

	if err != nil {
		return err
	}
	if result.Success == 0 {
		if param["method"] == "ActiveOrders" &&
			strings.HasPrefix(result.Error, "no orders") {
			return nil
		}
		if param["method"] == "TradeHistory" && strings.HasPrefix(result.Error,
			"no trades") {
			return nil
		}
		return errors.New(result.Error)
	}
	return json.Unmarshal(*result.Return, dst)
}

func (c *Client) retries() Retries {
	retries := DefaultRetries
	if c.Retries != nil {
		retries = *c.Retries
	}
	return retries
}
