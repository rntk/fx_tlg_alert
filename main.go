package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"

	"fx_alert/pkg/controllers"
	"fx_alert/pkg/db"
	"fx_alert/pkg/quoter"
	"fx_alert/pkg/telegram"
)

func main() {
	dbPath := flag.String("db", "db.json", "path to database")
	dbH, err := db.New(*dbPath, true)
	if err != nil {
		log.Panicf("Can't create database: %q. %v", *dbPath, err)
	}

	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Panicf("BOT_TOKEN not set")
	}
	tlg := telegram.New(token)
	defer tlg.Stop()
	ctx, cancelFn := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	wg.Add(1)
	quotesHolder := quoter.NewHolder(quoter.GetAllowedSymbols())
	go func() {
		defer wg.Done()
		controllers.ProcessBotCommands(ctx, dbH, quotesHolder, tlg)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		controllers.ProcessQuotes(ctx, dbH, quotesHolder, tlg)
	}()
	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, os.Interrupt)
	<-stopCh
	log.Print("Stopping...")
	cancelFn()
	wg.Wait()
	log.Print("Done")
}
