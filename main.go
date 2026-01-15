package main

import (
	"MainerCoal/Miner"
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
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
	ctxPostgres, ClosePostgres := context.WithCancel(context.Background())
	ctxRestore, CloseRestore := context.WithCancel(context.Background())
	url := "postgres://postgres:2976@localhost:5432/postgres"
	conn, err := pgxpool.New(ctxPostgres, url)
	if err != nil {
		fmt.Println("Не могу подключиться к базе ", err)
		os.Exit(1)
	}
	defer ClosePostgres()
	fmt.Println("Успешное подклчюение")
	game := Miner.NewGame(conn)
	fmt.Println("Загружаю сохранения")
	if err := game.RestoreMiners(ctxRestore); err != nil {
		fmt.Println("Ошибка востановления! ", err)
		CloseRestore()
	}
	defer CloseRestore()

	ctx, cancel := context.WithCancel(context.Background())
	ctxMiners, cancelMiners := context.WithCancel(ctx)
	defer cancelMiners()
	defer cancel()

	game.StartPassiveIncome(ctx)
	go game.Run(ctxMiners)
	http.HandleFunc("/buy/", TransformerHandler(game.BuyEquipmentHandler))
	http.HandleFunc("/hire/", TransformerHandler(game.HireHandler))
	http.HandleFunc("/dismiss/", TransformerHandler(game.HandlerDismissMiner))
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println(err)
	}
}
