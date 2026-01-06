package Miner

import (
	"context"
	"fmt"
	"time"
)

type Miner interface {
	Run(ctx context.Context, OreChan chan<- int64)
}
type BasicMainer struct {
	Speed      time.Duration
	Amount     int64
	Energy     int
	typeMainer string
}
type StrongMiner struct {
	Speed  time.Duration
	Amount int64
	Energy int
}

func (t *BasicMainer) Run(ctx context.Context, OreChan chan<- int64) {
	ticker := time.NewTicker(t.Speed)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if t.Energy <= 0 {
				fmt.Printf("%s mainer is out of energy ", t.typeMainer)
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
				fmt.Println("String miner is RIP")
				return
			}
			OreChan <- t.Amount
			t.Energy--
			t.Amount += 3
		}
	}
}
func NewTinyMiner() *BasicMainer {
	return &BasicMainer{
		Speed:      3 * time.Second,
		Amount:     1,
		Energy:     30,
		typeMainer: "tiny",
	}
}
func NewMediumMiner() *BasicMainer {
	return &BasicMainer{
		Speed:      2 * time.Second,
		Amount:     3,
		Energy:     45,
		typeMainer: "medium",
	}
}
func NewStrongMiner() *StrongMiner {
	return &StrongMiner{
		Speed:  1 * time.Second,
		Amount: 10,
		Energy: 60,
	}
}
