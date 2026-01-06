package main

import (
	"MainerCoal/Mainer"
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
	game := Mainer.NewGame()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	game.StartPassiveIncome(ctx)
	go game.Run()
	http.HandleFunc("/buy", TransformerHandler(game.BuyEquipmentHandler))
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println(err)
	}
	select {}
}
