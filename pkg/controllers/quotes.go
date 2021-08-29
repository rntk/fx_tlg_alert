package controllers

import (
	"context"
	"fmt"
	"log"
	"time"

	"fx_alert/pkg/db"
	"fx_alert/pkg/quoter"
	"fx_alert/pkg/telegram"
)

func ProcessQuotes(ctx context.Context, dbH *db.DB, qHolder *quoter.Holder, tlg *telegram.Telegram) {
	ticker := time.NewTicker(time.Minute)
	log.Printf("Quotes controller started")
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			qHolder.Update(ctx, 2)
			checkUsersLevelAlerts(ctx, dbH, qHolder, tlg)
			checkUsersRangeAlerts(ctx, dbH, qHolder, tlg)
		}
	}
}

func checkUsersLevelAlerts(ctx context.Context, dbH *db.DB, qHolder *quoter.Holder, tlg *telegram.Telegram) {
	ids := dbH.Users()
	for _, ID := range ids {
		select {
		case <-ctx.Done():
			return
		default:
			break
		}
		values := dbH.ListLevels(ID)
		for _, val := range values {
			select {
			case <-ctx.Done():
				return
			default:
				break
			}
			q, err := qHolder.GetQuote(val.Key)
			if err != nil {
				log.Printf("Can't get quotes for: %d. %q. %v", ID, val.Key, err)
				continue
			}
			if !val.IsAlert(q.Close) {
				continue
			}
			go func(ID int64, val db.LevelValue, p float64) {
				msg := fmt.Sprintf("Level alert: %s. \n. Current: %.5f", val.String(), p)
				if err := tlg.SendMessage(ID, 0, telegram.Answer{Text: msg}); err != nil {
					log.Printf("Can't send alert: %d. %q. %v", ID, msg, err)
					return
				}
				log.Printf("Sent alert: %d. %q", ID, msg)
				if err := dbH.DeleteLevelValue(ID, val.Key, val.Value); err != nil {
					log.Printf("Can't delete: %d. %q. %v", ID, val.String(), err)
					return
				}
				log.Printf("Deleted: %d. %q", ID, val.String())
			}(ID, val, q.Close)
		}
	}
}

func checkUsersRangeAlerts(ctx context.Context, dbH *db.DB, qHolder *quoter.Holder, tlg *telegram.Telegram) {
	ids := dbH.Users()
	for _, ID := range ids {
		select {
		case <-ctx.Done():
			return
		default:
			break
		}
		values := dbH.ListDeltas(ID)
		for _, val := range values {
			select {
			case <-ctx.Done():
				return
			default:
				break
			}
			q, err := qHolder.GetQuote(val.Key)
			if err != nil {
				log.Printf("Can't get quotes for: %d. %q. %v", ID, val.Key, err)
				continue
			}
			if !val.IsAlert(q.Close) {
				continue
			}
			go func(ID int64, val db.DeltaValue, p float64) {
				msg := fmt.Sprintf("Delta alert: %s. \n. Current: %.5f", val.String(), p)
				if err := tlg.SendMessage(ID, 0, telegram.Answer{Text: msg}); err != nil {
					log.Printf("Can't send alert: %d. %q. %v", ID, msg, err)
					return
				}
				log.Printf("Sent alert: %d. %q", ID, msg)
				if err := dbH.DeleteDeltaValue(ID, val.Key, val.Value); err != nil {
					log.Printf("Can't delete: %d. %q. %v", ID, val.String(), err)
					return
				}
				log.Printf("Deleted: %d. %q", ID, val.String())
			}(ID, val, q.Close)
		}
	}
}
