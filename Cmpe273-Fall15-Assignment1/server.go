package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"strconv"
	"strings"
)

type Call struct {
	List flist `json:"list"`
}

type flist struct {
	Meta      fmeta        `json:"-"`
	Resources []fresources `json:"resources"`
}

type fmeta struct {
	Type  string `json:"-"`
	Start int32  `json:"-"`
	Count int32  `json:"-"`
}

type fresources struct {
	Resource fresource `json:"resource"`
}

type fresource struct {
	Classname string  `json:"classname"`
	Fields    ffields `json:"fields"`
}

type ffields struct {
	Price  string `json:"price"`
	Symbol string `json:"symbol"`
}

type Args struct {
	StockSymbolAndPercentage string
	UserBudget               float64
}

type Quote struct {
	Stocks         string  `json:"stocksymbol"`
	UnvestedAmount float64 `json:"stockprice"`

	TradeId int `json:"id"`
}

type Id struct {
	TradeId int `json:"id"`
}

type UpdQuote struct {
	Stocks         string  `json:"stocksymbol"`
	UnvestedAmount float64 `json:"stockprice"`
}

type StockCalc int

var M map[int]Quote

func (t *StockCalc) StockPrice(args *Args, quote *Quote) error {

	a := string(args.StockSymbolAndPercentage[:])

	a = strings.Replace(a, ":", ",", -1)
	a = strings.Replace(a, "%", ",", -1)
	a = strings.Replace(a, ",,", ",", -1)
	a = strings.Trim(a, " ")
	a = strings.Replace(a, "\"", "", -1)
	a = strings.TrimSpace(a)
	a = strings.TrimSuffix(a, ",")
	Stockstmp := strings.Split(a, ",")

	Total := 0.0
	var ReqUrl string

	for i := 0; i < len(Stockstmp); i++ {
		i = i + 1

		temp, _ := strconv.ParseFloat(Stockstmp[i], 64)
		Total = (temp * args.UserBudget * 0.01)
		fmt.Println(Stockstmp[i-1], Total)
		ReqUrl = ReqUrl + (Stockstmp[i-1] + ",")

	}
	ReqUrl = strings.TrimSuffix(ReqUrl, ",")

	UrlStr := "http://finance.yahoo.com/webservice/v1/symbols/" + ReqUrl + "/quote?format=json"

	client := &http.Client{}

	resp, _ := client.Get(UrlStr)
	req, _ := http.NewRequest("GET", UrlStr, nil)

	req.Header.Add("If-None-Match", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, _ = client.Do(req)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		var C Call
		body, _ := ioutil.ReadAll(resp.Body)

		err := json.Unmarshal(body, &C)

		n := len(Stockstmp)

		Quo := make([]float64, n, n)

		for i := 0; i < n; i++ {
			i = i + 1
			TempFloat, _ := strconv.ParseFloat(Stockstmp[i], 64)
			Quo[i] = (TempFloat * args.UserBudget * 0.01)

		}

		var buffer bytes.Buffer
		q := 0
		for _, Sample := range C.List.Resources {

			temp1 := Sample.Resource.Fields.Symbol
			temp2, _ := strconv.ParseFloat(Sample.Resource.Fields.Price, 64)
			temp3 := (int)(Quo[q+1] / temp2)
			temp4 := math.Mod(Quo[q+1], temp2)
			q = q + 2

			quote.Stocks = fmt.Sprintf("%s:%d:$%g", temp1, temp3, temp2)

			quote.UnvestedAmount = quote.UnvestedAmount + temp4

			buffer.WriteString(quote.Stocks)
			buffer.WriteString(",")
		}
		quote.TradeId = quote.TradeId + 1
		quote.Stocks = (buffer.String())

		quote.Stocks = strings.TrimSuffix(quote.Stocks, ",")

		M = map[int]Quote{
			quote.TradeId: {quote.Stocks, quote.UnvestedAmount, quote.TradeId},
		}

		if err == nil {

		}
	} else {
		fmt.Println(resp.Status)

	}
	return nil
}

func main() {
	red := new(StockCalc)

	server := rpc.NewServer()
	server.Register(red)
	server.HandleHTTP(rpc.DefaultRPCPath, rpc.DefaultDebugPath)
	listener, e := net.Listen("tcp", ":1234")
	if e != nil {

		log.Fatal("listen error:", e)
	}
	for {
		if conn, err := listener.Accept(); err != nil {
			log.Fatal("accept error: " + err.Error())
		} else {
			log.Printf("new connection established\n")
			go server.ServeCodec(jsonrpc.NewServerCodec(conn))
		}
	}
}
