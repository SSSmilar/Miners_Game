package Miner

import (
	"MainerCoal/BuyCatalog"
	"MainerCoal/Error"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Game struct {
	Balance           int64
	Inventory         map[string]bool
	OreChan           chan int64
	BuyChan           chan BuyCatalog.BuyRequest
	HireChan          chan BuyCatalog.HireRequest
	Quit              chan struct{}
	CheckWinCondition func() bool
	DB                *pgxpool.Pool
	PerSecond         int64
}

func NewGame(db *pgxpool.Pool) *Game {
	return &Game{
		Balance:           5000,
		Inventory:         make(map[string]bool),
		OreChan:           make(chan int64),
		BuyChan:           make(chan BuyCatalog.BuyRequest),
		HireChan:          make(chan BuyCatalog.HireRequest),
		Quit:              make(chan struct{}),
		CheckWinCondition: func() bool { return false },
		DB:                db,
		PerSecond:         0,
	}
}
func (g *Game) SaveIncome(ctx context.Context, amount int64) error {
	if amount == 0 {
		return nil
	}
	query := "UPDATE  users SET balance = balance + $1 WHERE id = 1 "
	if _, err := g.DB.Exec(ctx, query, amount); err != nil {
		return fmt.Errorf("ошибка сохранения дохода %w ", err)
	}
	fmt.Printf("💾 [AUTO-SAVE] Сохранено +%d монет в базу\n", amount)
	return nil
}
func (g *Game) HandlerDismissMiner(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "DELETE" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return Error.ErrWrongMethod
	}
	prefix := "/dismiss/"
	minerName := strings.TrimPrefix(r.URL.Path, prefix)
	if minerName == "" {
		w.WriteHeader(http.StatusBadRequest)
		return Error.ErrInvalidParameters
	}
	if strings.Contains(minerName, "/") {
		w.WriteHeader(http.StatusBadRequest)
		return Error.ErrInvalidParameters
	}
	err := g.DissmissMainer(r.Context(), minerName)
	if err != nil {
		w.Write([]byte("ошибка " + err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return nil
	}
	w.WriteHeader(http.StatusOK)

	return nil
}
func (g *Game) DissmissMainer(ctx context.Context, name string) error {
	query := `
			UPDATE mainers 
			SET deleted_at = NOW()
			WHERE name = $1 AND deleted_at IS NULL 
`
	tag, err := g.DB.Exec(ctx, query, name)
	if err != nil {
		return Error.ErrDataBase
	}
	if tag.RowsAffected() == 0 {
		return Error.ErrMinersDB
	}
	fmt.Printf("Шахтертёр %s был успешно уволен (Soft Delete)\n", name)
	return nil
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
	var unsavedIncome int64 = 0
	saveTicker := time.NewTicker(3 * time.Second)
	defer saveTicker.Stop()
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
			unsavedIncome += finalBalance
			g.Balance += finalBalance
			fmt.Printf("Balance: %d , profit : %d , profit without boosts( %d )\n", g.Balance, finalBalance, amount)
		case <-saveTicker.C:
			if unsavedIncome > 0 {
				if err := g.SaveIncome(ctx, unsavedIncome); err != nil {
					fmt.Println("Ошибка автосейва", err)
				}
				unsavedIncome = 0
			}
		case req := <-g.BuyChan:
			if unsavedIncome > 0 {
				if err := g.SaveIncome(ctx, unsavedIncome); err != nil {
					fmt.Println("Ошибка сейва при покупке", err)
				}
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
			}
		case req := <-g.HireChan:
			if unsavedIncome > 0 {
				if err := g.SaveIncome(ctx, unsavedIncome); err != nil {
					fmt.Println("Ошибка сейва при покупке", err)
					req.Response <- false
					continue
				}
				unsavedIncome = 0
			}
			var initialEnergy int64
			switch req.MinerType {
			case "tiny":
				initialEnergy = 30
			case "medium":
				initialEnergy = 60
			case "strong":
				initialEnergy = 100
			}
			minerID, uniqueName, err := g.BuyMinerTx(ctx, req.MinerType, req.Cost, initialEnergy)
			if err != nil {
				fmt.Println("Ошибка транзакции", err)
				req.Response <- false
				continue
			}
			g.Balance -= req.Cost
			fmt.Printf("✅ Куплен %s (ID: %d). Баланс: %d\n", uniqueName, minerID, g.Balance)

			switch req.MinerType {
			case "tiny":
				worker := NewTinyMiner()
				worker.ID = minerID
				go worker.Run(ctx, g.OreChan, g.DB)
			case "medium":
				worker := NewMediumMiner()
				worker.ID = minerID
				go worker.Run(ctx, g.OreChan, g.DB)
			case "strong":
				worker := NewStrongMiner()
				worker.ID = minerID
				go worker.Run(ctx, g.OreChan, g.DB)
			}
			req.Response <- true
		case <-g.Quit:
			if unsavedIncome > 0 {
				if err := g.SaveIncome(ctx, unsavedIncome); err != nil {
					fmt.Println("Ошибка при сохронении при выходе ", err)
				}
			}
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
		w.WriteHeader(http.StatusBadRequest)
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
func (g *Game) BuyMinerTx(ctx context.Context, mType string, cost int64, energy int64) (int64, string, error) {
	tx, err := g.DB.Begin(ctx)
	if err != nil {
		return 0, "", Error.ErrTransaction
	}
	defer tx.Rollback(ctx)
	var currentBalance int64
	if err := tx.QueryRow(ctx, "SELECT balance FROM users WHERE id = 1 FOR UPDATE").Scan(&currentBalance); err != nil {
		return 0, "", Error.ErrRegiments
	}
	if currentBalance < cost {
		return 0, "", Error.ErrNotEnoughMoney
	}
	if _, err := tx.Exec(ctx, "UPDATE users SET balance = balance - $1 WHERE id = 1 ", cost); err != nil {
		return 0, "", fmt.Errorf("База ругается %w", err)
	}
	suffix := GenerateRandomString(5)
	uniqueName := fmt.Sprintf("%s_%s", mType, suffix)
	var minerID int64
	var startAmount int64
	switch mType {
	case "tiny":
		startAmount = 1
	case "medium":
		startAmount = 3
	case "strong":
		startAmount = 5
	}
	insertQuery := `
        INSERT INTO mainers (name, type, energy, amount) 
        VALUES ($1, $2, $3, $4) 
        RETURNING id
    `
	err = tx.QueryRow(ctx, insertQuery, uniqueName, mType, energy, startAmount).Scan(&minerID)
	if err != nil {
		return 0, "", fmt.Errorf("База ругается %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, "", Error.ErrSaveChange
	}
	return minerID, uniqueName, nil
}
func (g *Game) RestoreMiners(ctx context.Context) error {
	query := "SELECT  id , type , energy , amount FROM mainers WHERE deleted_at IS NULL AND energy > 0 "
	rows, err := g.DB.Query(ctx, query)
	if err != nil {
		fmt.Println(err)
		return Error.ErrDataBase
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		var id int64
		var mType string
		var energy int64
		var amount int64
		if err := rows.Scan(&id, &mType, &energy, &amount); err != nil {
			fmt.Println(err)
			return Error.ErrDataBase
		}
		switch mType {
		case "tiny":
			g.PerSecond += amount
			miner := NewTinyMiner()
			miner.ID = id
			miner.Amount = amount
			miner.Energy = energy
			go miner.Run(ctx, g.OreChan, g.DB)
		case "medium":
			g.PerSecond += amount
			miner := NewMediumMiner()
			miner.ID = id
			miner.Amount = amount
			miner.Energy = energy
			go miner.Run(ctx, g.OreChan, g.DB)
		case "strong":
			g.PerSecond += amount
			miner := NewStrongMiner()
			miner.ID = id
			miner.Amount = amount
			miner.Energy = energy
			go miner.Run(ctx, g.OreChan, g.DB)
		}
		count++
		if count > 0 {
			fmt.Printf("♻️  Восстановлено %d шахтеров из базы. Доход: %d/сек\n", count, g.PerSecond)
		}
	}
	return nil
}
