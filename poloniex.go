package main

import (
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/mariuspass/recws"
	"log"
	"time"
)

const (
	ws_url     = "wss://api2.poloniex.com/"
	HEART_BEAT = "1010"
	INIT_TYPE  = "i"
	ORDER      = "o"
	HISTORY    = "t"
)

type Order struct {
	Price    float64
	Quantity float64
}

type Orderbook struct {
	Bids      []Order
	Asks      []Order
	Version   float64
	Timestamp time.Time
	IsValid   bool
}

type Polo struct {
	ws_url  string
	rc      *recws.RecConn
	cmds    []interface{}
	channel string
	Orderbook
}

type WSCommand struct {
	Command string `json:"command"`
	Channel string `json:"channel"`
}

func Interface2String(in interface{}) string {
	return fmt.Sprintf("%s", in)
}

func NewPolo(channel string) *Polo {
	polo := &Polo{ws_url: ws_url, channel: channel}
	polo.Orderbook.IsValid = false
	return polo
}

func (p *Polo) Connect() {
	rc := &recws.RecConn{RecIntvlFactor: 1}
	p.rc = rc
	p.rc.Dial(p.ws_url, nil)
	p.Handler()
}

func (p *Polo) Subscribe(channel string) {
	subscribeCmd := &WSCommand{"subscribe", channel}
	p.rc.WriteJSON(subscribeCmd)
}

func (p *Polo) Handler() {
	p.Subscribe("BTC_BCH")
	for {
		_, message, err := p.rc.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			time.Sleep(p.rc.RecIntvlMin)
			p.Handler()
			return
		}
		log.Printf("recv1: %s", message)
		js, _ := simplejson.NewJson(message)
		arr, err := js.Array()
		if err != nil {
			return
		}
		code := Interface2String(arr[0])
		if code == HEART_BEAT {
			fmt.Println("heartbeat")
		} else {
			for _, book := range arr[2].([]interface{}) {
				t := book.([]interface{})[0].(string)
				switch t {
				case INIT_TYPE:
					// init
					data := book.([]interface{})[1].(map[string]interface{})
					fmt.Println(data["currencyPair"])
				case ORDER:
					// increase
				case HISTORY:
					// history
				}
			}
		}
	}
}

func main() {
	polo := NewPolo("BTC_BCH")
	polo.Connect()
}
