package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"strconv"
	"strings"
	"time"
)

var stock StockDetails
var trans StockResponse
var stockTickr string
var percentStock string
var qtyStr string
var portfolioResponse PortfolioResponse

type StockDetails struct {
	List struct {
		Meta struct {
			Count int    `json:"count"`
			Start int    `json:"start"`
			Type  string `json:"type"`
		} `json:"meta"`
		Resources []struct {
			Resource struct {
				Classname string `json:"classname"`
				Fields    struct {
					Name    string `json:"name"`
					Price   string `json:"price"`
					Symbol  string `json:"symbol"`
					Ts      string `json:"ts"`
					Type    string `json:"type"`
					UTCtime string `json:"utctime"`
					Volume  string `json:"volume"`
				} `json:"fields"`
			} `json:"resource"`
		} `json:"resources"`
	} `json:"list"`
}

type StockResponse struct {
	TradeId        uint32  `json:"tradeid"`
	Stocks         string  `json:"stocks"`
	UnvestedAmount float64 `json:"unvestedAmount"`
}

type StockRequest struct {
	Budget                   float32 `json:"budget"`
	StockSymbolAndPercentage string  `json:"stockSymbolAndPercentage"`
}

type PortfolioRequest struct {
	Tradeid uint32 `json:"tradeid"`
}

type PortfolioResponse struct {
	Stocks             string  `json:"stocks"`
	CurrentMarketValue float64 `json:"currentMarketValue"`
	UnvestedAmount     float64 `json:"unvestedAmount"`
}

type StockMarket int

func (t *StockMarket) CheckPortfolio(portfolioRequest *PortfolioRequest, reply *PortfolioResponse) error {
	fmt.Print(portfolioRequest.Tradeid)
	fmt.Print(trans.TradeId)

	var stockLegacy StockDetails
	if portfolioRequest.Tradeid == trans.TradeId {
		//fmt.Print(trans.Stocks)
		fmt.Print(stockTickr)
		fmt.Print(percentStock)
		stockLegacy = stock
		getQuotes(stockTickr)
		fmt.Print(stockLegacy)
		countStocks := stock.List.Meta.Count
		fmt.Print("Count of No. of Stocks is %d ", countStocks)
		sum := 0.00
		var price float64
		stockQty := strings.Split(qtyStr, ",")
		stockResponse := ""
		for i := 0; i < countStocks; i++ {

			if stock.List.Resources[i].Resource.Fields.Price > stockLegacy.List.Resources[i].Resource.Fields.Price {
				price, _ = strconv.ParseFloat(stock.List.Resources[i].Resource.Fields.Price, 64)

				temp, _ := strconv.ParseFloat(stockQty[i], 64)
				sum = sum + price*temp

				stock.List.Resources[i].Resource.Fields.Price = "+" + stock.List.Resources[i].Resource.Fields.Price
				fmt.Print(stock.List.Resources[i].Resource.Fields.Price)
				stockResponse = stockResponse + stock.List.Resources[i].Resource.Fields.Symbol + ":" + stockQty[i] + ":$" + stock.List.Resources[i].Resource.Fields.Price + ","
				portfolioResponse.CurrentMarketValue = sum
				portfolioResponse.UnvestedAmount = trans.UnvestedAmount
				portfolioResponse.Stocks = stockResponse
				*reply = portfolioResponse

			} else if stock.List.Resources[i].Resource.Fields.Price < stockLegacy.List.Resources[i].Resource.Fields.Price {
				price, _ = strconv.ParseFloat(stock.List.Resources[i].Resource.Fields.Price, 64)

				temp, _ := strconv.ParseFloat(stockQty[i], 64)
				sum = sum + price*temp

				stock.List.Resources[i].Resource.Fields.Price = "-" + stock.List.Resources[i].Resource.Fields.Price
				fmt.Print(stock.List.Resources[i].Resource.Fields.Price)
				stockResponse = stockResponse + stock.List.Resources[i].Resource.Fields.Symbol + ":" + stockQty[i] + ":$" + stock.List.Resources[i].Resource.Fields.Price + ","
				portfolioResponse.CurrentMarketValue = sum
				portfolioResponse.UnvestedAmount = trans.UnvestedAmount
				portfolioResponse.Stocks = stockResponse
				*reply = portfolioResponse

			} else {
				price, _ = strconv.ParseFloat(stock.List.Resources[i].Resource.Fields.Price, 64)

				temp, _ := strconv.ParseFloat(stockQty[i], 64)
				sum = sum + price*temp

				fmt.Print(stock.List.Resources[i].Resource.Fields.Price)
				stockResponse = stockResponse + stock.List.Resources[i].Resource.Fields.Symbol + ":" + stockQty[i] + ":$" + stock.List.Resources[i].Resource.Fields.Price + ","
				portfolioResponse.CurrentMarketValue = sum
				portfolioResponse.UnvestedAmount = trans.UnvestedAmount
				portfolioResponse.Stocks = stockResponse
				*reply = portfolioResponse
			}
		}

	} else {
		return errors.New("Trade ID not found")
	}

	return nil
}

func (t *StockMarket) BuyStock(stockRequest *StockRequest, reply *StockResponse) error {
	fmt.Println(stockRequest.StockSymbolAndPercentage)
	stockRequest.StockSymbolAndPercentage = strings.Replace(stockRequest.StockSymbolAndPercentage, ":", ",", strings.Count(stockRequest.StockSymbolAndPercentage, ":"))
	stockRequest.StockSymbolAndPercentage = strings.Replace(stockRequest.StockSymbolAndPercentage, "%", "", strings.Count(stockRequest.StockSymbolAndPercentage, "%"))
	list := strings.Split(stockRequest.StockSymbolAndPercentage, ",")

	stockTickr = ""
	percentStock = ""

	for i := 0; i < len(list); i++ {
		if i%2 == 0 {
			stockTickr = stockTickr + list[i] + ","
		}
		if i%2 != 0 {
			percentStock = percentStock + list[i] + ","
		}
	}

	fmt.Println(stockTickr)
	fmt.Println(percentStock)
	errorStr := processStocks(stockTickr, percentStock, stockRequest.Budget)
	*reply = trans
	if errorStr == "0000" {
		return errors.New("Quote Parsing Error due to wrong symbols")
	} else if errorStr == "0001" {
		return errors.New("Percent greater than 100")
	}
	return nil
}

//timeout constant
const (
	timeout = time.Duration(time.Second * 100)
)

func getQuotes(str string) {
	client := http.Client{Timeout: timeout}
	url := fmt.Sprintf("http://finance.yahoo.com/webservice/v1/symbols/%s/quote?format=json", str)

	res, err := client.Get(url)
	if err != nil {
		fmt.Errorf("Stocks cannot access yahoo finance API: %v", err)
	}
	defer res.Body.Close()
	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Errorf("Stocks cannot read json body: %v", err)
	}
	err = json.Unmarshal(content, &stock)
	if err != nil {
		fmt.Errorf("Stocks cannot parse json data: %v", err)
	}

}

func parseQuotes() (prices string) {

	var priceStr = ""
	count := stock.List.Meta.Count

	for i := 0; i < count; i++ {
		priceStr = priceStr + stock.List.Resources[i].Resource.Fields.Price + ","
	}

	if count != 0 {
		return priceStr
	} else {
		return "0000"
	}
}

func processStocks(stockStr string, percntStr string, balance float64) (error string) {

	getQuotes(stockStr)
	priceStr := parseQuotes()
	if priceStr == "0000" {
		return "0000"
	}

	prices := strings.Split(priceStr, ",")
	percnts := strings.Split(percntStr, ",")
	stocks := strings.Split(stockStr, ",")

	var percentTotal float64
	var prc float64
	var prcnt float64
	var qty int
	qtyStr = ""
	var total float64
	var unvestedAmt float64
	stockCount := len(percnts)

	for i := 0; i < stockCount; i++ {
		prc, _ = strconv.ParseFloat(prices[i], 64)
		prcnt, _ = strconv.ParseFloat(percnts[i], 64)
		qty = int((balance * prcnt) / (100.00 * prc))
		total = total + (float64(qty) * prc)
		qtyStr = qtyStr + strconv.Itoa(qty) + ","
		percentTotal += prcnt
	}

	if percentTotal > 100.00 {
		return "0001"
	}

	stockQty := strings.Split(qtyStr, ",")
	transDetails := ""
	//displaying the response
	if total < balance {
		for i := 0; i < stockCount-1; i++ {
			prc, _ = strconv.ParseFloat(prices[i], 64)
			transDetails = transDetails + stocks[i] + ":" + stockQty[i] + ":$" + prices[i]
			fmt.Print(transDetails)
			if i != stockCount-2 {
				fmt.Printf(",")
				transDetails += ","
			}
		}
		unvestedAmt = balance - total
		fmt.Printf("\nUnvested Amount: $%.2f\n", unvestedAmt)
		t := time.Unix(0, 11/11/2011)
		trans.TradeId = uint32(time.Since(t))
		trans.Stocks = transDetails
		trans.UnvestedAmount = unvestedAmt

	} else {
		fmt.Print("Buying amount exceeds balance")
	}
	return "1111"
}

func main() {

	stockMarket := new(StockMarket)
	rpc.Register(stockMarket)

	tcpAddr, err := net.ResolveTCPAddr("tcp", ":1234")
	checkError(err)

	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		jsonrpc.ServeConn(conn)
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
}
