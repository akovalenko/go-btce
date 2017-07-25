package btce

import (
	"testing"
	"time"
)

func TestFastDepth(t *testing.T) {
	d := FastDepth("btc_usd")
	t.Log("First run: would sell 1 BTC for ",d.Evaluate(1).Sell, " usd")
	for i:=0; i<5; i++ {
		time.Sleep(time.Second/5)
		d = FastDepth("btc_usd")
		t.Log("Would sell 1 BTC for ",d.Evaluate(1).Sell, " usd")
	}
}
