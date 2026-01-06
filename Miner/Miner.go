package Miner

import (
	"context"
	"fmt"
	"time"
)

type Miner interface {
	Run(ctx context.Context, OreChan chan<- int64)
}
type BasicMiner struct {
	Speed     time.Duration
	Amount    int64
	Energy    int
	MinerType string
}
type StrongMiner struct {
	Speed  time.Duration
	Amount int64
	Energy int
}

func (t *BasicMiner) Run(ctx context.Context, OreChan chan<- int64) {
	ticker := time.NewTicker(t.Speed)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if t.Energy <= 0 {
				fmt.Printf("%s mainer is out of energy ", t.MinerType)
				return
			}
			OreChan <- t.Amount
			t.Energy--
		}
	}
}

func (t *StrongMiner) Run(ctx context.Context, OreChan chan<- int64) {
	ticker := time.NewTicker(t.Speed)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if t.Energy <= 0 {
				fmt.Println("Strong miner is RIP")
				return
			}
			OreChan <- t.Amount
			t.Energy--
			t.Amount += 3
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
