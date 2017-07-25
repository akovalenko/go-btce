// A command-line tool for typical btc-e operations (mostly for
// demonstrative purposes, though it can be useful)

package main

import (
	"flag"
	"fmt"
	"github.com/akovalenko/go-btce"
	"sort"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type whichOrders struct {
	id        uint64
	pair      string
	minRate   float64
	maxRate   float64
	minAmount float64
	maxAmount float64
}

func (filter whichOrders) satisfies(id uint64, order btce.ActiveOrder) bool {
	if filter.id != 0 && filter.id != id {
		return false
	}
	if filter.pair != "" && filter.pair != order.Pair {
		return false
	}
	if filter.minRate != 0 && filter.minRate > order.Rate {
		return false
	}
	if filter.maxRate != 0 && filter.maxRate < order.Rate {
		return false
	}
	if filter.minAmount != 0 && filter.minAmount > order.Amount {
		return false
	}
	if filter.maxAmount != 0 && filter.maxAmount < order.Amount {
		return false
	}
	return true
}

func (filter whichOrders) matchActiveOrders(set btce.ActiveOrdersResult) []uint64 {
	result := []uint64{}
	for id, description := range set {
		if filter.satisfies(id, description) {
			result = append(result, id)
		}
	}
	return result
}

func (filter *whichOrders) flagSet() *flag.FlagSet {
	f := flag.NewFlagSet("order filtering parameters", flag.ExitOnError)
	f.Uint64Var(&filter.id, "id", 0, "Order identifier (0 = no filter)")
	f.StringVar(&filter.pair, "pair", "", "Currency pair")
	f.Float64Var(&filter.minRate, "min-rate", 0, "Minimum rate")
	f.Float64Var(&filter.maxRate, "max-rate", 0, "Maximum rate (0 = no limit)")
	f.Float64Var(&filter.minAmount, "min-amount", 0, "Minimum amount")
	f.Float64Var(&filter.maxAmount, "max-amount", 0, "Maximum amount (0 = no limit)")
	return f
}

var keyFile string

func init() {
	flag.StringVar(&keyFile, "key", "key.json", `API token JSON file; "{key:.. secret:..}"`)
}

// printFunds prints available funds (returned from private getInfo
// and trade operations)
func printFunds(m map[string]float64) {
	keys := make([]string, len(m))
	i := 0
	for key := range m {
		keys[i] = key
		i++
	}
	sort.Strings(keys)
	fmt.Println("Available funds:")
	for _,key := range keys {
		if m[key]>0 {
			fmt.Printf("%v %.8f\n", key, m[key])
		}
	}
}

// getClient initializes BTC-e client with -key
func getClient() *btce.Client {
	c := &btce.Client{}
	err := c.ReadKey(keyFile)
	if err != nil {
		log.Fatal(err)
	}
	return c
}

// listOrders lists orders matching criteria
func listOrders(w whichOrders) {
	c := getClient()
	orders := c.ActiveOrders(btce.ActiveOrdersParameters{Pair: w.pair})
	selected := w.matchActiveOrders(orders)
	for _, id := range selected {
		order := orders[id]
		fmt.Println("Order #", id, order.Pair, order.Type,
			"rate:", order.Rate, "amount:", order.Amount,
			"status: ", order.Status)
	}
}

func cancelOrders(w whichOrders) {
	c := getClient()
	orders := c.ActiveOrders(btce.ActiveOrdersParameters{Pair: w.pair})
	selected := w.matchActiveOrders(orders)
	var funds map[string]float64
	for _, id := range selected {
		order := orders[id]
		fmt.Println("Cancelling order #", id, order.Pair, order.Type,
			"rate:", order.Rate, "amount:", order.Amount)
		r := c.CancelOrder(btce.CancelOrderParameters{OrderId: id})
		funds = r.Funds
	}
	if funds != nil {
		printFunds(funds)
	}
}

func placeOrder(t string, amount string, pair string, rate string) {
	c := getClient()
	info := c.PublicInfo()
	_, foundPair := info.Pairs[pair]
	if !foundPair {
		pairs := make([]string, len(info.Pairs))
		i := 0
		for k := range info.Pairs {
			pairs[i] = k
			i++
		}
		log.Fatal("Specify pair: one of ", strings.Join(pairs, ", "))
	}
	if t != "buy" && t != "sell" {
		log.Fatal("Order type: buy or sell expected, got", t)
	}
	a, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		log.Fatal(err)
	}
	r := 0e0
	if rate != "" {
		r, err = strconv.ParseFloat(rate, 64)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		if t != "sell" {
			log.Fatal("Buy on market rate not supported -- sell only")
		}
		r = info.Pairs[pair].MinAmount
	}
	tr := c.Trade(btce.TradeParameters{Pair: pair, Type: t, Amount: a, Rate: r})
	if tr.OrderId == 0 {
		fmt.Println("Order fully executed.")
	} else {
		fmt.Printf("New order #%v, received %.8f, remains %.8f\n",
			tr.OrderId, tr.Received, tr.Remains)
	}
	printFunds(tr.Funds)
}

func monitorDepth(pair string) {
	ptr := btce.FastDepth(pair)
	for {
		fmt.Println("Depth for ",pair)
		for i:=0; i<20; i++ {
			fmt.Printf("ask: %.8f %.8f  bid: %.8f %.8f\n",
				ptr.Asks[i].Rate(),ptr.Asks[i].Amount(),
				ptr.Bids[i].Rate(),ptr.Bids[i].Amount())
		}
		for  {
			time.Sleep(time.Second/2)
			new := btce.FastDepth(pair)
			// hack: FastDepth returns the same pointer if
			// there were no updates
			if new!=ptr {
				ptr = new
				break
			}
		}
	}
}


func main() {
	flag.Parse()
	switch flag.Arg(0) {
	case "orders":
		w := whichOrders{}
		w.flagSet().Parse(flag.Args()[1:])
		listOrders(w)
	case "cancel":
		w := whichOrders{}
		w.flagSet().Parse(flag.Args()[1:])
		cancelOrders(w)
	case "place":
		placeOrder(flag.Arg(1), flag.Arg(2), flag.Arg(3), flag.Arg(4))
	case "fastdepth":
		monitorDepth(flag.Arg(1))
	default:
		fmt.Printf(`Usage: %v [-key key.json] subcommand
Subcommands:
 orders -- list orders (all or matching, try orders -h for usage)
 cancel -- cancel orders (all or matching, -h for help)
 place <sell/buy> <amount> <pair> <rate> -- place order
 fastdepth <pair> -- monitor depth instantly, update as orders change
`, os.Args[0])
	}
}
