package Miner

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Miner interface {
	Run(ctx context.Context, OreChan chan<- int64)
}
type BasicMiner struct {
	ID        int64
	Speed     time.Duration
	Amount    int64
	Energy    int64
	MinerType string
}
type StrongMiner struct {
	ID     int64
	Speed  time.Duration
	Amount int64
	Energy int64
}

func (t *BasicMiner) Run(ctx context.Context, OreChan chan<- int64, db *pgxpool.Pool) {
	workTicker := time.NewTicker(t.Speed)
	defer workTicker.Stop()
	saveTicker := time.NewTicker(10 * time.Second)
	defer saveTicker.Stop()

	save := func() {
		_, err := db.Exec(context.Background(), "UPDATE mainers SET energy  = $1  , amount = $2 WHERE  id = $3", t.Energy, t.Amount, t.ID)
		if err != nil {
			fmt.Println("Err save miner for DB  , context kill ", err)
			return
		}
		fmt.Printf("💾 [Miner %d] Saved: Energy=%d, Amount=%d\n", t.ID, t.Energy, t.Amount)
	}
	for {
		select {
		case <-ctx.Done():
			save()
			return
		case <-workTicker.C:
			if t.Energy <= 0 {
				fmt.Printf("%s mainer is out of energy ", t.MinerType)
				if _, err := db.Exec(context.Background(), "UPDATE mainers SET energy = 0 , deleted_at = NOW()  WHERE id  = $1", t.ID); err != nil {
					fmt.Println("Err Save energy for db energy == 0 ", err)
					return
				}
				return
			}
			OreChan <- t.Amount
			t.Energy--
		case <-saveTicker.C:
			go save()
		}
	}
}

func (t *StrongMiner) Run(ctx context.Context, OreChan chan<- int64, db *pgxpool.Pool) {
	workTicker := time.NewTicker(t.Speed)
	defer workTicker.Stop()

	saveTicker := time.NewTicker(10 * time.Second)
	defer saveTicker.Stop()
	save := func() {
		_, err := db.Exec(context.Background(), "UPDATE mainers SET energy  = $1  , amount = $2 WHERE  id = $3", t.Energy, t.Amount, t.ID)
		if err != nil {
			fmt.Println("Err save miner for DB  , context kill ", err)
			return
		}
		fmt.Printf("💾 [Miner %d] Saved: Energy=%d, Amount=%d\n", t.ID, t.Energy, t.Amount)
	}
	for {
		select {
		case <-ctx.Done():
			save()
			return
		case <-workTicker.C:
			if t.Energy <= 0 {
				fmt.Println("Strong miner is RIP")
				if _, err := db.Exec(context.Background(), "UPDATE mainers SET energy = 0  , deleted_at = NOW() WHERE id  = $1", t.ID); err != nil {
					fmt.Println("Err Save energy for db energy == 0 ", err)
					return
				}
				return
			}
			OreChan <- t.Amount
			t.Energy--
			t.Amount += 3
		case <-saveTicker.C:
			go save()
		}
	}
}
func NewTinyMiner() *BasicMiner {
	return &BasicMiner{
		Speed:     3 * time.Second,
		Amount:    1,
		Energy:    30,
		MinerType: "tiny",
	}
}
func NewMediumMiner() *BasicMiner {
	return &BasicMiner{
		Speed:     2 * time.Second,
		Amount:    3,
		Energy:    45,
		MinerType: "medium",
	}
}
func NewStrongMiner() *StrongMiner {
	return &StrongMiner{
		Speed:  1 * time.Second,
		Amount: 10,
		Energy: 60,
	}
}
