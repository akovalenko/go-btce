package main

const exampleStrategy string =`
{
    "DESCRIPTION": "Example task file for btc-e bot (simplexchange)",
    "pair": "btc_usd",
    "_1_": "-- Data for possible rate evaluation",
    "base": 2518.838,
    "step": 1.0095,
    "min": 1000.0,
    "max": 9000.0,
    "_2_": "-- Connection configuration",
    "keyfile": "key.json",
    "url": "https://wex.nz",
    "_3_": "-- Order placement configuration",
    "spread": 1.006,
    "capitalize": "btc",
    "unit": 0.003,
    "upscale": 1
}`

const exampleKey string = `
{
    "key": "ZZZZZZZZ-ZZZZZZZZ-ZZZZZZZZ-ZZZZZZZZ-ZZZZZZZZ",
    "secret": "8888888888888888888888888888888888888888888888888888888888888888"
}`
