package btce

import (
	"encoding/json"
	"github.com/toorop/go-pusher"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"
)

const BTCE_APP_ID = "c354d4d129ee0faa5c92"

type pairData struct {
	Asks map[string]string
	Bids map[string]string
}

type request struct {
	Pair   string
	Result chan *DepthInfo
}

type watcher struct {
	cache  map[string]pairData
	convCache map[string]*DepthInfo
	btce   Client
	info   *PublicInfo
	pusher *pusher.Client
	pchan  chan *pusher.Event
}

func convertDepth(dict map[string]string) []Offer {
	result := make([]Offer, len(dict))
	i := 0
	for k, v := range dict {
		rate, err := strconv.ParseFloat(k, 64)
		if err != nil {
			log.Panic(err)
		}
		amount, err := strconv.ParseFloat(v, 64)
		if err != nil {
			log.Panic(err)
		}
		result[i][0] = rate
		result[i][1] = amount
		i++
	}
	return result
}

func (w *watcher) depth(pair string) *DepthInfo {
	_, ok := w.info.Pairs[pair]
	if !ok {
		return nil
	}
	depth, ready := w.convCache[pair]
	if ready {
		return depth
	}
	pd, found := w.cache[pair]
	if !found {
		pd = pairData{
			Bids: map[string]string{},
			Asks: map[string]string{}}
		w.cache[pair] = pd
		err := w.pusher.Subscribe(pair + ".depth")
		if err != nil {
			log.Panic(err)
		}
		time.Sleep(2 * time.Second)
		di := w.btce.Depth([]string{pair}, 100)[pair]
		for _, o := range di.Asks {
			pd.Asks[fmt.Sprint(o.Rate())] = fmt.Sprint(o.Amount())
		}
		for _, o := range di.Bids {
			pd.Bids[fmt.Sprint(o.Rate())] = fmt.Sprint(o.Amount())
		}

		for {
			select {
			case depthEvent := <-w.pchan:
				w.notify(depthEvent)
			default:
				goto uptodate
			}
		}
	uptodate:
	}
	a, b := convertDepth(pd.Asks), convertDepth(pd.Bids)
	sort.Slice(a, func(i, j int) bool { return a[i].Rate() < a[j].Rate() })
	sort.Slice(b, func(i, j int) bool { return b[i].Rate() > b[j].Rate() })
	depth = &DepthInfo{Asks: a, Bids: b}
	w.convCache[pair] = depth
	return depth
}

func updateCache(dict map[string]string,deltas[][2]interface{}) {
	for _, pair := range deltas {
		key, value := fmt.Sprint(pair[0]), fmt.Sprint(pair[1])
		if value == "0" {
			delete(dict, key)
		} else {
			dict[key] = value
		}
	}
}

func (w *watcher) notify(event *pusher.Event) {
	pair := strings.Split(event.Channel, ".")[0]
	decoded := struct {
		Ask [][2]interface{}
		Bid [][2]interface{}
	}{}
	err := json.Unmarshal([]byte(event.Data), &decoded)
	if err != nil {
		log.Panic(err)
	}
	asks, bids := w.cache[pair].Asks, w.cache[pair].Bids
	updateCache(asks,decoded.Ask)
	updateCache(bids,decoded.Bid)
	delete(w.convCache,pair)
}

func (w *watcher) watch(q chan request) {
	// don't do anything until a  query
	request := <-q
	p, err := pusher.NewClient(BTCE_APP_ID)
	if err != nil {
		log.Panic(err)
	}
	w.pchan, err = p.Bind("depth")
	if err != nil {
		log.Panic(err)
	}
	w.pusher = p
	w.info = w.btce.PublicInfo()
	w.cache = map[string]pairData{}
	w.convCache = map[string]*DepthInfo{}

	request.Result <- w.depth(request.Pair)
	for {
		select {
		case depthEvent := <-w.pchan:
			w.notify(depthEvent)
		case request := <-q:
			request.Result <- w.depth(request.Pair)
		}
	}
}

var queue = make(chan request)

func init() {
	w := watcher{}
	go w.watch(queue)
}

// FastDepth returns DepthInfo (asks and bids) for a pair, subscribing
// for future depth updates with Pusher when the pair is used for the
// first time.
//
// Advantage: new calls for the same pair are fast, and the result is
// kept up to date in background with Pusher events
//
// Disadvantage: if internet connection fails, stale depth info may be
// returned for some time; no real error reporting on problems, just a
// nil return on invalid pairs and panic (in a separate goroutine) on
// other errors. Also, hard-coded limit of 100 bids & asks may make it
// harder to get a full picture of the market state.
func FastDepth(pair string) *DepthInfo {
	r := make(chan *DepthInfo,1)
	queue <- request{Pair: pair, Result: r}
	return <-r
}
