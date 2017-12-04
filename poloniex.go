package main

import (
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/gorilla/websocket"
	"log"
	"sort"
	"strconv"
	"time"
)

const (
	ws_url     = "wss://api2.poloniex.com/"
	HEART_BEAT = "1010"
	INIT_TYPE  = "i"
	ORDER      = "o"
	HISTORY    = "t"
	REMOVE     = "0.00000000"
	SELLSIDE   = "0"
	BUYSIDE    = "1"
)

type Order struct {
	Price    float64
	Quantity float64
}

type OrderArr []Order

type Orderbook struct {
	Bids      OrderArr
	Asks      OrderArr
	Version   string
	Timestamp int64
	IsValid   bool
}

type Polo struct {
	ws_url  string
	ws      *websocket.Conn
	cmds    []interface{}
	bidsMap map[string]interface{}
	asksMap map[string]interface{}
	channel string
	done    chan struct{}
	Orderbook
}

type WSCommand struct {
	Command string `json:"command"`
	Channel string `json:"channel"`
}

func Interface2String(in interface{}) string {
	return fmt.Sprintf("%s", in)
}

func NewPolo(channel string, done chan struct{}) *Polo {
	polo := &Polo{ws_url: ws_url, channel: channel, done: done}
	polo.Orderbook.IsValid = false
	polo.Connect()
	return polo
}

func (p *Polo) Connect() {
	c, _, err := websocket.DefaultDialer.Dial(p.ws_url, nil)
	if err != nil {
		p.Orderbook.IsValid = false
		p.done <- struct{}{}
		return
	}
	p.ws = c
	p.Handler()
}

func (p *Polo) Subscribe(channel string) {
	subscribeCmd := &WSCommand{"subscribe", channel}
	err := p.ws.WriteJSON(subscribeCmd)
	if err != nil {
		p.Orderbook.IsValid = false
		log.Println("write:", err)
		p.done <- struct{}{}
		return
	}
}

func (p *Polo) Handler() {
	p.Subscribe(p.channel)
	for {
		_, message, err := p.ws.ReadMessage()
		if err != nil {
			p.Orderbook.IsValid = false
			log.Println("read:", err)
			p.done <- struct{}{}
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
		} else if len(arr) == 2 {
			//
		} else {
			for _, book := range arr[2].([]interface{}) {
				t := book.([]interface{})[0].(string)
				switch t {
				case INIT_TYPE:
					// init
					p.Orderbook.Version = Interface2String(arr[1])
					data := book.([]interface{})[1].(map[string]interface{})
					orderbooks := data["orderBook"].([]interface{})
					p.asksMap = orderbooks[0].(map[string]interface{})
					p.bidsMap = orderbooks[1].(map[string]interface{})
					p.updateOrderbook()
				case ORDER:
					// increase
					pre, _ := strconv.ParseInt(p.Orderbook.Version, 10, 64)
					new, _ := strconv.ParseInt(Interface2String(arr[1]), 10, 64)
					if new != pre+1 {
						p.Orderbook.IsValid = false
						p.done <- struct{}{}
						return
					}
					p.Orderbook.Version = Interface2String(arr[1])
					data := book.([]interface{})
					side := Interface2String(data[1])
					price := Interface2String(data[2])
					quantity := Interface2String(data[3])

					if side == BUYSIDE {
						if quantity == REMOVE {
							delete(p.bidsMap, price)
						} else {
							p.bidsMap[price] = data[3]
						}
					} else if side == SELLSIDE {
						if quantity == REMOVE {
							delete(p.asksMap, price)
						} else {
							p.asksMap[price] = data[3]
						}
					}
					p.updateOrderbook()
				case HISTORY:
					// history
				}
			}
		}
	}
}

func (s OrderArr) Len() int {
	return len(s)
}

func (s OrderArr) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s OrderArr) Less(i, j int) bool {
	return s[i].Price < s[j].Price
}

func (s OrderArr) Reverse() OrderArr {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

func (p *Polo) updateOrderbook() {
	tmp := []Order{}
	for p, q := range p.asksMap {
		fp, _ := strconv.ParseFloat(p, 64)
		fq, _ := strconv.ParseFloat(Interface2String(q), 64)
		tmp = append(tmp, Order{fp, fq})
	}
	sort.Sort(OrderArr(tmp))
	p.Orderbook.Asks = tmp
	tmp = []Order{}
	for p, q := range p.bidsMap {
		fp, _ := strconv.ParseFloat(p, 64)
		fq, _ := strconv.ParseFloat(Interface2String(q), 64)
		tmp = append(tmp, Order{fp, fq})
	}
	sort.Sort(OrderArr(tmp))
	p.Orderbook.Bids = OrderArr(tmp).Reverse()
	p.Orderbook.Timestamp = time.Now().Unix()
	p.Orderbook.IsValid = true
	fmt.Printf("%+v\n", p.Orderbook)
}

func main() {
	done := make(chan struct{})
	go NewPolo("BTC_BCH", done)
	for {
		select {
		case <-done:
			time.Sleep(time.Second * 2)
			go NewPolo("BTC_BCH", done)
		}
	}
}
