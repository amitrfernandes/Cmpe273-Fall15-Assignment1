package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/rpc/jsonrpc"
	"os"
	"strconv"
)

type Args struct {
	StockSymbolAndPercentage string
	UserBudget               float64
}

type Quote struct {
	Stocks string `json:"stocksymbol"`

	UnvestedAmount float64 `json:"stockprice"`
	TradeId        int     `json:"id"`
}

type Id struct {
	TradeId int `json:"id"`
}

type UpdQuote struct {
	Stocks         string  `json:"stocksymbol"`
	UnvestedAmount float64 `json:"stockprice"`
}

func main() {

	client, err := net.Dial("tcp", "127.0.0.1:1234")
	if err != nil {
		log.Fatal("dialing:", err)
	}
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter Stock Symbol: ")
	StockSymbolAndPercentage, _ := reader.ReadString('\n')
	fmt.Println(StockSymbolAndPercentage)
	fmt.Print("Enter budget: ")
	var UserBudget float64
	fmt.Scan(&UserBudget)

	args := Args{StockSymbolAndPercentage, UserBudget}

	var quo Quote
	c := jsonrpc.NewClient(client)
	err = c.Call("StockCalc.StockPrice", args, &quo)
	if err != nil {
		log.Fatal("arith error:", err)
	}
	f, _ := strconv.ParseFloat("quo.Stocks", 64)

	fmt.Println("TradeID:", quo.TradeId)
	fmt.Println("Stocks", quo.Stocks)
	fmt.Println("Remaining amount", quo.UnvestedAmount)

	client, err = net.Dial("tcp", "127.0.0.1:1234")
	if err != nil {
		log.Fatal("dialing:", err)
	}
	fmt.Println("\nEnter Trade Id")
	var s int
	fmt.Scan(&s)
	if s == 1 {

		args := Args{StockSymbolAndPercentage, UserBudget}

		var quo Quote

		c := jsonrpc.NewClient(client)
		err = c.Call("StockCalc.StockPrice", args, &quo)
		if err != nil {
			log.Fatal("error:", err)
		}

		x, _ := strconv.ParseFloat("quo.Stocks", 64)
		y := x - f
		fmt.Println("Stocks:", quo.Stocks)
		fmt.Println("Profit/Loss per stock:", y)
		fmt.Println("Uninvested amount:", quo.UnvestedAmount)

	}

}
