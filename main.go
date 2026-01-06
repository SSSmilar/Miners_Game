package main

import (
	"MainerCoal/Miner"
	"context"
	"fmt"
	"net/http"
)

type HandlerWitchError func(http.ResponseWriter, *http.Request) error

func TransformerHandler(h HandlerWitchError) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := h(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			fmt.Println(err)
		}
	}
}
func main() {
	game := Miner.NewGame()
	ctx, cancel := context.WithCancel(context.Background())
	ctxMiners, cancelMiners := context.WithCancel(ctx)
	defer cancelMiners()
	defer cancel()

	game.StartPassiveIncome(ctx)
	go game.Run(ctxMiners)
	http.HandleFunc("/buy/", TransformerHandler(game.BuyEquipmentHandler))
	http.HandleFunc("/hire/", TransformerHandler(game.HireHandler))
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println(err)
	}
}
