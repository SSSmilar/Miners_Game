package Miner

import (
	"MainerCoal/BuyCatalog"
	"MainerCoal/Error"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Game struct {
	Balance           int64
	Inventory         map[string]bool
	OreChan           chan int64
	BuyChan           chan BuyCatalog.BuyRequest
	HireChan          chan BuyCatalog.HireRequest
	Quit              chan struct{}
	CheckWinCondition func() bool
}

func NewGame() *Game {
	return &Game{
		Balance:           0,
		Inventory:         make(map[string]bool),
		OreChan:           make(chan int64),
		BuyChan:           make(chan BuyCatalog.BuyRequest),
		HireChan:          make(chan BuyCatalog.HireRequest),
		Quit:              make(chan struct{}),
		CheckWinCondition: func() bool { return false },
	}
}
func (g *Game) StartPassiveIncome(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				g.OreChan <- 1
			}
		}
	}()
}
func (g *Game) Run(ctx context.Context) {
	startTime := time.Now()
	for {
		select {
		case amount := <-g.OreChan:
			finalBalance := amount
			if g.Inventory["pickaxe"] {
				finalBalance *= 2
			}
			if g.Inventory["cart"] {
				finalBalance *= 3
			}
			g.Balance += finalBalance
			fmt.Printf("Balance: %d , profit : %d , profit without boosts( %d )", g.Balance, finalBalance, amount)
		case req := <-g.BuyChan:
			if g.Inventory[req.Item] {
				fmt.Println("Already have this item", req.Item)
				req.Response <- true
				fmt.Println(Error.ErrAlreadyHaveItem)
				continue
			}
			if g.Balance >= req.Cost {
				g.Balance -= req.Cost
				g.Inventory[req.Item] = true
				fmt.Printf("Buy  %s!  successfully    remainder %d\n ", req.Item, g.Balance)
				if g.CheckWinCondition() {
					fmt.Println("Win Condition")
					return
				}
				req.Response <- true
			} else {
				fmt.Println("Not enough")
				req.Response <- false
				fmt.Println(Error.ErrNotEnoughMoney)

			}
		case req := <-g.HireChan:
			if g.Balance >= req.Cost {
				switch req.MinerType {
				case "tiny":
					go NewTinyMiner().Run(ctx, g.OreChan)
				case "medium":
					go NewMediumMiner().Run(ctx, g.OreChan)
				case "strong":
					go NewStrongMiner().Run(ctx, g.OreChan)
				}
				g.Balance -= req.Cost
				fmt.Printf("Hire %s  successfully", req.MinerType)
				req.Response <- true
			} else {
				fmt.Println("Not enough money to hire")
				req.Response <- false
			}

		case <-g.Quit:
			duration := time.Since(startTime)
			fmt.Printf("Game over  : balance : %d  , play time : %s\n ", g.Balance, duration)
			return
		}
	}
}
func (g *Game) BuyEquipmentHandler(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		return Error.ErrWrongMethod
	}
	prefix := "/buy/"
	productStrOriginal := strings.TrimPrefix(r.URL.Path, prefix)
	if productStrOriginal == "" {
		w.WriteHeader(http.StatusNotFound)
		return Error.ErrInvalidParameters
	}
	if strings.Contains(productStrOriginal, "/") {
		w.WriteHeader(http.StatusBadRequest)
		return Error.ErrInvalidParameters
	}
	productStrClear := strings.ToLower(productStrOriginal)

	price, ok := BuyCatalog.Equipments[productStrClear]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		text, err := w.Write([]byte("Not found " + productStrOriginal))
		if err != nil {
			fmt.Println(err)
			return Error.ErrInvalidVariable
		}
		fmt.Println(text)
		return Error.ErrInvalidParameters
	}

	answerChan := make(chan bool)
	req := BuyCatalog.BuyRequest{
		Item:     productStrClear,
		Cost:     price,
		Response: answerChan,
	}
	g.BuyChan <- req
	isSuccess := <-answerChan
	if isSuccess {
		request, err := w.Write([]byte("Success"))
		if err != nil {
			fmt.Println(err)
			return Error.ErrInvalidVariable
		}
		fmt.Println(request)
	} else {
		request, err := w.Write([]byte("Not enough money"))
		if err != nil {
			fmt.Print(err)
			return Error.ErrNotEnoughMoney
		}
		fmt.Println(request)
	}
	return nil
}

func (g *Game) HireHandler(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		return Error.ErrWrongMethod
	}
	prefix := "/hire/"
	productStrOriginal := strings.TrimPrefix(r.URL.Path, prefix)
	if productStrOriginal == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println(Error.ErrInvalidParameters)
	}
	if strings.Contains(productStrOriginal, "/") {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println(Error.ErrInvalidParameters)
	}
	productStrClear := strings.ToLower(productStrOriginal)

	price, ok := BuyCatalog.WorkForce[productStrClear]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return Error.ErrInvalidVariable
	}
	answerChan := make(chan bool)
	req := BuyCatalog.HireRequest{
		MinerType: productStrClear,
		Cost:      price,
		Response:  answerChan,
	}
	g.HireChan <- req
	isSuccess := <-answerChan
	if isSuccess {
		text, err := w.Write([]byte("Successful hired " + productStrClear))
		if err != nil {
			return Error.ErrBadRequest
		}
		fmt.Println(text)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		text, err := w.Write([]byte("Not enough money"))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return Error.ErrNotEnoughMoney
		}
		fmt.Println(text)
	}
	return nil
}
