package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/rpc/jsonrpc"
	"os"
)

type StockRequest struct {
	StockSymbolAndPercentage string  `json:"stockSymbolAndPercentage"`
	Budget                   float32 `json:"budget"`
}

type StockResponse struct {
	TradeId        uint32  `json:"tradeid"`
	Stocks         string  `json:"stocks"`
	UnvestedAmount float64 `json:"unvestedAmount"`
}

type PortfolioRequest struct {
	Tradeid uint32 `json:"tradeid"`
}

type PortfolioResponse struct {
	Stocks             string  `json:"stocks"`
	CurrentMarketValue float64 `json:"currentMarketValue"`
	UnvestedAmount     float64 `json:"unvestedAmount"`
}

var stockRequest StockRequest
var portfolioRequest PortfolioRequest

func main() {
	if len(os.Args) != 4 {
		fmt.Println("Usage: ", os.Args[0], "server:port")
		log.Fatal(1)
	}
	service := os.Args[1]

	client, err := jsonrpc.Dial("tcp", service)
	if err != nil {
		log.Fatal("dialing:", err)
	}

	if os.Args[2] == "buy" {
		fmt.Printf("Buying Stocks..\n ")
		content := []byte(os.Args[3])
		err = json.Unmarshal(content, &stockRequest)
		var reply StockResponse
		err = client.Call("StockMarket.BuyStock", stockRequest, &reply)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("\n%+v\n", reply)

	} else if os.Args[2] == "checkPortfolio" {
		fmt.Printf("Checking Portfolio.. \n")
		content := []byte(os.Args[3])
		err = json.Unmarshal(content, &portfolioRequest)
		var reply PortfolioResponse
		//tradeId,_ := strconv.ParseInt(os.Args[3],10,64)
		err = client.Call("StockMarket.CheckPortfolio", portfolioRequest, &reply)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("\n%+v\n", reply)

	} else {

		fmt.Printf("Invalid Input")
	}
}
