# Project btcego #

Library and tools for [WEX](https://wex.nz) cryptocurrency exchange (former BTC-e, hence naming).

## Installation ##

    go get github.com/akovalenko/go-btce

## Library: btce ##

A wrapper around public and private (trading) API remote calls.

Disclaimer: *[WEX](https://wex.nz) is not affiliated with the
project.* It's implemented using public documentation:

  * [Public API](https://wex.nz/api/3/documentation): accessible to anyone, no API key required
  * [Private API](https://wex.nz/tapi/docs): requires an API key
  * [Push API](https://wex.nz/pushAPI/docs)

Some ideas were borrowed from
the
[CodeReclaimers/btce-api](https://github.com/CodeReclaimers/btce-api)
Python library.

See [documentation](https://godoc.org/github.com/akovalenko/go-btce) for details.

Donate bitcoin to [3GfDvGGd6fqxwR118L7X6XfVFuwni9F33m](bitcoin:3GfDvGGd6fqxwR118L7X6XfVFuwni9F33m) if you find it useful:

![bitcoin](https://www.freeformatter.com/qr-code?w=350&h=350&e=Q&c=bitcoin%3A3GfDvGGd6fqxwR118L7X6XfVFuwni9F33m)

## Data ##

For private methods, API keys are expected to be provided in the
following format (JSON):

~~~ json
{
"key": "ZZZZZZZZ-ZZZZZZZZ-ZZZZZZZZ-ZZZZZZZZ-ZZZZZZZZ",
"secret": "8888888888888888888888888888888888888888888888888888888888888888"
}
~~~


## Tool: btce ##

Meant mostly as an API usage example.

    btce -key otherkey.json orders -pair ltc_btc
	btce place sell 0.001 btc_usd 9999
	btce cancel -pair btc_usd -min-rate 9000
	# Fast depth updates using Push API
    btce fastdepth btc_usd

## Bot: simplexchange ##

A trading bot with a predefined (but tunable) strategy.
See [its own README](examples/simplexchange) for details.
