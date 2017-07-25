// simplexchange uses BTC-e.com API to implement a trading strategy
// making profit from volatility.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/akovalenko/go-btce"
	"io/ioutil"
	"log"
	"math"
	"os"
	"sort"
	"strings"
	"time"
)

type Strategy struct {
	Pair       string
	Base       float64
	Min        float64
	Max        float64
	Step       float64
	KeyFile    string
	URL        string
	Spread     float64
	Capitalize string
	Unit       float64
	Upscale    float64
}

type State struct {
	Orders   []uint64
	Earnings map[string]float64
}

type DynamicData struct {
	Rates     []float64
	Amounts   []float64
	BaseUnit  string
	OtherUnit string
	FixInBase bool
}

func (d DynamicData) FindRate(rate float64, exact bool) int {
	pos := sort.SearchFloat64s(d.Rates, rate)
	if exact && d.Rates[pos] != rate {
		log.Fatal("Exact rate search failed for", rate)
	}
	return pos
}

type Sxcrobot struct {
	strategy         Strategy
	state            State
	data             DynamicData
	client           *btce.Client
	info             *btce.PublicInfo
	loadedKey        bool
	funds            map[string]float64
	writeStateCreate bool
	StrategyFile     string
	StateFile        string
}

func (s *Sxcrobot) LoadStrategy() {
	loadJSON(s.StrategyFile, &s.strategy)
}
func (s *Sxcrobot) LoadState() {
	loadJSON(s.StateFile, &s.state)
	if s.state.Earnings == nil {
		s.state.Earnings = map[string]float64{}
	}
}
func (s *Sxcrobot) SaveState() {
	newOrders := []uint64{}
	for _, id := range s.state.Orders {
		if id != 0 {
			newOrders = append(newOrders, id)
		}
	}
	s.state.Orders = newOrders
	if s.writeStateCreate {
		file, err := os.Open(s.StateFile)
		file.Close()
		if err == nil {
			log.Fatal("Refuse to overwrite with empty state:", s.StateFile)
		}
	}
	saveJSON(s.StateFile, s.state)
}

func (s *Sxcrobot) EnsureClient() *btce.Client {
	if s.client == nil {
		client, err := btce.NewClient(s.strategy.URL)
		failOn(err)
		s.client = client
	}
	return s.client
}

func (s *Sxcrobot) EnsureKeyedClient() *btce.Client {
	s.EnsureClient()
	if !s.loadedKey {
		err := s.client.ReadKey(s.strategy.KeyFile)
		failOn(err)
		s.loadedKey = true
	}
	return s.client
}

func (s *Sxcrobot) EnsurePublicInfo() *btce.PublicInfo {
	s.EnsureClient()
	if s.info == nil {
		s.info = s.client.PublicInfo()
	}
	return s.info
}

func (s *Sxcrobot) EnsureData() {
	s.data = s.strategy.ToDynamicData()
	s.EnsurePublicInfo()
	s.data.FixFormat(s.info.Pairs[s.strategy.Pair])
}

func (s *Sxcrobot) NextRate(rate float64, movement int) float64 {
	pos := s.data.FindRate(rate, true)
	pos += movement
	for s.data.Rates[pos]/rate < s.strategy.Spread && rate/s.data.Rates[pos] < s.strategy.Spread {
		pos += movement
	}
	return s.data.Rates[pos]
}

func (s *Sxcrobot) NextAmount(amount float64, rate float64, movement int) float64 {
	qfee := (100 - s.info.Pairs[s.strategy.Pair].Fee) / 100
	nextRate := s.NextRate(rate, movement)
	if movement == -1 {
		// closed sell
		output := amount * rate * qfee // so much "dollars"
		if s.data.FixInBase {
			// if we fix in "btc", use up all our "dollars" without remainder
			return output / nextRate
		} else {
			// if we fix in bucks, we intend to have our original amount as buy order output:
			// nextAmount*feeq == amount
			earned := output - (nextRate * amount / qfee)
			s.state.Earnings[s.data.OtherUnit] += earned
			log.Printf("Earned %.8f %v, totalling %.8f", earned,
				s.data.OtherUnit, s.state.Earnings[s.data.OtherUnit])
			return amount / qfee
		}
	} else {
		// closed buy
		output := amount * qfee
		if s.data.FixInBase {
			earned := output - (amount / qfee * rate / nextRate)
			s.state.Earnings[s.data.BaseUnit] += earned
			log.Printf("Earned %.8f %v, totalling %.8f", earned,
				s.data.BaseUnit, s.state.Earnings[s.data.BaseUnit])
			return amount / qfee * rate / nextRate
		} else {
			return output
		}
	}
}

func (s *Sxcrobot) UpdateFunds() {
	s.funds = s.client.PrivateInfo().Funds
}

func (s *Sxcrobot) PlaceOrder(dir string, amount float64, rate float64) {
	log.Println("Placing order:", dir, "for", amount, "at", rate)
	param := btce.TradeParameters{Pair: s.strategy.Pair, Type: dir, Rate: rate, Amount: amount}
	result := s.client.Trade(param)
	s.funds = result.Funds
	if result.OrderId == 0 {
		s.RePlaceOrder(dir, amount, rate)
	} else {
		s.state.Orders = append(s.state.Orders, result.OrderId)
		s.SaveState()
	}
}

func (s *Sxcrobot) RePlaceOrder(dir string, amount float64, rate float64) {
	if dir == "sell" {
		s.PlaceOrder("buy", s.NextAmount(amount, rate, -1), s.NextRate(rate, -1))
	} else {
		s.PlaceOrder("sell", s.NextAmount(amount, rate, 1), s.NextRate(rate, 1))
	}
}

func (s *Sxcrobot) CmdInit() {
	s.LoadStrategy()
	s.writeStateCreate = true
	s.state.Earnings = map[string]float64{}
	s.SaveState()
}

func (s *Sxcrobot) CmdPlace() {
	s.LoadStrategy()
	s.LoadState()
	s.EnsurePublicInfo()
	s.EnsureKeyedClient()
	s.EnsureData()
	depths, err := s.client.GetDepth([]string{s.strategy.Pair}, 1)
	failOn(err)
	depth := depths[s.strategy.Pair]
	if len(depth.Asks) < 1 || len(depth.Bids) < 1 {
		log.Fatal("Out of asks or bids entirely")
	}
	high := s.data.FindRate(depth.Asks[0].Rate(), false)
	low := s.data.FindRate(depth.Bids[0].Rate(), false)
	tipping := (low + high) / 2
	log.Println("Would sell at", s.data.Rates[tipping+1], "and buy at", s.data.Rates[tipping-1])
	low = tipping - 1
	high = tipping + 1

	active := s.client.ActiveOrders(btce.ActiveOrdersParameters{Pair: s.strategy.Pair})

	for _, id := range s.state.Orders {
		orderdata, ok := active[id]
		if !ok {
			log.Println("Order #", id, " closed. Next update will peek it up...")
		}
		log.Println("Existing order #", id)
		pos := s.data.FindRate(orderdata.Rate, true)
		switch orderdata.Type {
		case "sell":
			if high < pos+1 {
				high = pos + 1
			}
		case "buy":
			// lower low watermark
			if low > pos-1 {
				low = pos - 1
			}
		}
	}

	log.Println("Sell starts at", s.data.Rates[high], "buy at", s.data.Rates[low])
	s.UpdateFunds()
	log.Println("Placing sell orders")
	for s.funds[s.data.BaseUnit] >= s.data.Amounts[high] {
		s.PlaceOrder("sell", s.data.Amounts[high], s.data.Rates[high])
		high++
	}
	log.Println("Placing buy orders")
	for {
		rate := s.data.Rates[low]
		futureRate := s.NextRate(rate, 1)
		futureAmount := s.data.Amounts[s.data.FindRate(futureRate, true)]
		thisAmount := s.NextAmount(futureAmount, futureRate, -1)
		if s.funds[s.data.OtherUnit] < rate*thisAmount {
			break
		}
		s.PlaceOrder("buy", thisAmount, rate)
		low--
	}
}

func (s *Sxcrobot) CmdUpdate() {
	log.Println("Loading strategy...")
	s.LoadStrategy()
	log.Println("Loading state...")
	s.LoadState()
	log.Println("Getting public info...")
	s.EnsurePublicInfo()
	log.Println("Reading key...")
	s.EnsureKeyedClient()
	log.Println("Calculating data...")
	s.EnsureData()
	log.Println("Calling ActiveOrders..")
	active := s.client.ActiveOrders(btce.ActiveOrdersParameters{Pair: s.strategy.Pair})
	log.Println("Found active orders:", len(active), "/ known:", len(s.state.Orders))
	var changed bool
	for index, id := range s.state.Orders {
		log.Println("Checking order #", id)
		_, ok := active[id]
		if !ok {
			log.Println("Order went away:", id)
			orderinfo := s.client.OrderInfo(btce.OrderInfoParameters{id})[id]
			s.state.Orders[index] = 0 // wipe
			changed = true
			if orderinfo.Status == 1 {
				s.RePlaceOrder(orderinfo.Type, orderinfo.StartAmount, orderinfo.Rate)
			} else {
				log.Println("-- Cancelled, not replaced, just forgotten")
			}
		}
	}
	if changed {
		log.Println("Saving state..")
		s.SaveState()
	}
}

func (s *Sxcrobot) CmdCancel() {
	s.LoadStrategy()
	s.LoadState()
	s.EnsurePublicInfo()
	s.EnsureKeyedClient()
	s.EnsureData()
	active := s.client.ActiveOrders(btce.ActiveOrdersParameters{Pair: s.strategy.Pair})
	var changed bool
	for index, id := range s.state.Orders {
		_, ok := active[id]
		if ok {
			result := btce.CancelOrderResult{}
			err := s.client.Call(btce.CancelOrderParameters{id}, &result)
			if err != nil {
				log.Println("Skipping order #", id, ":", err)
			}
			s.state.Orders[index] = 0
			changed = true
		}
	}
	if changed {
		s.SaveState()
	}
}

var strategyFile string
var stateFile string

func init() {
	flag.StringVar(&strategyFile, "strategy", "strategy.json", "Strategy file")
	flag.StringVar(&stateFile, "state", "state.json", "State file")
}

func failOn(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func loadJSON(fileName string, v interface{}) {
	data, err := ioutil.ReadFile(fileName)
	failOn(err)
	err = json.Unmarshal(data, v)
	failOn(err)
}

func (s Strategy) ToDynamicData() DynamicData {
	d := DynamicData{}

	cur := strings.Split(s.Pair, "_")
	d.BaseUnit, d.OtherUnit = cur[0], cur[1]
	d.FixInBase = (d.BaseUnit == s.Capitalize)
	if !d.FixInBase && d.OtherUnit != s.Capitalize {
		log.Fatal("Unknown currency for capitalization:", s.Capitalize)
	}
	if s.Base < s.Min || s.Base > s.Max || s.Base <= 0 ||
		s.Min <= 0 || s.Max <= 0 || s.Step <= 0 || s.Unit <= 0 || s.Upscale <= 0 {
		log.Fatal("Contradictive or invalid rate boundaries")
	}
	stepsDown := 0
	stepsUp := 0
	for rate := s.Base; rate >= s.Min; rate /= s.Step {
		stepsDown++
	}
	for rate := s.Base; rate <= s.Max; rate *= s.Step {
		stepsUp++
	}
	count := stepsUp + stepsDown - 1
	if count < 1 || count > 100000 {
		log.Fatal("Something went wrong with levels")
	}
	d.Amounts = make([]float64, count)
	d.Rates = make([]float64, count)
	for i, rate, amount := stepsDown, s.Base, s.Unit; i >= 0; i-- {
		d.Rates[i], d.Amounts[i] = rate, amount
		rate /= s.Step
		amount /= s.Upscale
	}
	for i, rate, amount := stepsDown, s.Base, s.Unit; i < count; i++ {
		d.Rates[i], d.Amounts[i] = rate, amount
		rate *= s.Step
		amount *= s.Upscale
	}

	return d
}

func (d *DynamicData) FixFormat(info btce.PairInfo) {
	rc := math.Pow(10, float64(info.DecimalPlaces))
	ac := math.Pow(10, 8)
	for i, rate := range d.Rates {
		d.Rates[i] = float64(int64(rate*rc+0.5)) / rc
		d.Amounts[i] = float64(int64(d.Amounts[i]*ac+0.5)) / ac
	}
}

func saveJSON(fileName string, v interface{}) {
	data, err := json.Marshal(v)
	failOn(err)
	tempFile := fileName + ".tmpnew"
	err = ioutil.WriteFile(tempFile, data, 0644)
	failOn(err)
	err = os.Rename(tempFile, fileName)
	failOn(err)
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func main() {
	flag.Parse()
	bot := Sxcrobot{StrategyFile: strategyFile, StateFile: stateFile}
	switch flag.Arg(0) {
	case "init":
		log.Println("Writing initial (empty) state")
		bot.CmdInit()
	case "place":
		log.Println("Placing freed capital to orders")
		bot.CmdPlace()
	case "update":
		log.Println("Updating orders by closure")
		bot.CmdUpdate()
	case "cancel":
		log.Println("Updating orders by closure")
		bot.CmdCancel()
	case "monitor":
		for {
			log.Println("Updating...")
			bot.CmdUpdate()
			log.Println("Sleeping...")
			time.Sleep(2 * time.Second)

		}
	case "example":
		haveStrategy, _ := exists(strategyFile)
		if haveStrategy {
			fmt.Println("File ", strategyFile, " already exists, not writing example")
		} else {
			ioutil.WriteFile(strategyFile, []byte(exampleStrategy), 0644)
			fmt.Println("Created strategy file example: ", strategyFile)
		}
		keyFile := "key.json"
		haveKey, _ := exists(keyFile)
		if haveKey {
			fmt.Println("File ", keyFile, " already exists, not writing example")
		} else {
			ioutil.WriteFile(keyFile, []byte(exampleKey), 0644)
			fmt.Println("Created key file example: ", keyFile)
		}
	default:
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Println(
			` subcommands:
    init (write empty state)
    place (place orders according to strategy)
    update (replace corresponding orders for closed orders)
    monitor (constantly update while running)
    cancel (cancel every order managed by simplexchange)

    example (create same strategy.json and key.json when they don't exist)

  flags (precede a subcommand):`)
		flag.PrintDefaults()
	}
}
