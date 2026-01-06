package Mainer

import (
	"MainerCoal/BayCatalog"
	"MainerCoal/Error"
	"context"
	"fmt"
	"net/http"
	"time"
)

type Game struct {
	Balance   int64
	Inventory map[string]bool
	OreChan   chan int
	BuyChan   chan BayCatalog.BuyRequest
	Quit      chan struct{}
}

func NewGame() *Game {
	return &Game{
		Balance:   0,
		Inventory: make(map[string]bool),
		OreChan:   make(chan int),
		BuyChan:   make(chan BayCatalog.BuyRequest),
		Quit:      make(chan struct{}),
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
func (g *Game) Run() error {
	for {
		select {
		case amount := <-g.OreChan:
			g.Balance += int64(amount)
			fmt.Printf("Balance: %d\n", g.Balance)
		case req := <-g.BuyChan:
			if g.Inventory[req.Item] {
				fmt.Println("Already have this item", req.Item)
				req.Response <- true
				return Error.ErrAlreadyHaveItem
			}
			if g.Balance >= req.Cost {
				g.Inventory[req.Item] = true
				g.Balance -= req.Cost
				fmt.Println("Buy successfully")
				req.Response <- true
			} else {
				fmt.Println("Not enough")
				req.Response <- false
				return Error.ErrNotEnoughMoney
			}
		case <-g.Quit:

			fmt.Println("Game over balance", g.Balance)
			return nil
		}
	}
}
func (g *Game) BuyEquipmentHandler(w http.ResponseWriter, r *http.Request) error {
	cost := int64(5)
	answerChan := make(chan bool)
	req := BayCatalog.BuyRequest{
		Cost:     cost,
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
