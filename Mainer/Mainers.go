package Mainer

import (
	"MainerCoal/BayCatalog"
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
	OreChan           chan int
	BuyChan           chan BayCatalog.BuyRequest
	Quit              chan struct{}
	CheckWinCondition func() bool
}

func NewGame() *Game {
	return &Game{
		Balance:           0,
		Inventory:         make(map[string]bool),
		OreChan:           make(chan int),
		BuyChan:           make(chan BayCatalog.BuyRequest),
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
func (g *Game) Run() {
	for {
		select {
		case amount := <-g.OreChan:
			g.Balance += int64(amount)
			fmt.Printf("Balance: %d\n", g.Balance)
		case req := <-g.BuyChan:
			if g.Inventory[req.Item] {
				fmt.Println("Already have this item", req.Item)
				req.Response <- true
				fmt.Println(Error.ErrAlreadyHaveItem)

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
		case <-g.Quit:

			fmt.Println("Game over balance", g.Balance)
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

	price, ok := BayCatalog.Equipments[productStrClear]
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
	req := BayCatalog.BuyRequest{
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
