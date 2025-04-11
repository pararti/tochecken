package main

import (
	"fmt"
	"gopkg.in/telebot.v4"
	"time"
	"tochecken/app"
	"tochecken/checkers"
	"tochecken/db"
	"tochecken/tgbot"
)

//TIP To run your code, right-click the code and select <b>Run</b>. Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.

func main() {
	ru, err := checkers.NewRubUsd()
	if err != nil {
		fmt.Printf("Error creating RubUsd: %v\n", err)
		return
	}

	binance := checkers.NewBinanceChecker()
	err = binance.FetchTokens(ru, true)

	if err != nil {
		fmt.Printf("Error fetching tokens: %v\n", err)

		return
	}

	kraken := checkers.NewKrakenChecker()
	err = kraken.FetchTokens(ru, true)

	if err != nil {
		fmt.Printf("Error fetching tokens: %v\n", err)

		return
	}

	okx := checkers.NewOkxChecker()
	err = okx.FetchTokens(ru, true)
	if err != nil {
		fmt.Printf("Error fetching tokens: %v\n", err)

		return
	}

	checkersPool := make(map[string]checkers.Checker, 2)
	checkersPool[binance.Type] = binance
	checkersPool[kraken.Type] = kraken
	checkersPool[okx.Type] = okx

	database := db.NewDb()

	a := &app.App{CheckersPoll: checkersPool, DB: database}

	b := tgbot.CreateBot(a)

	//обновляем курс
	go func(ru *checkers.RubUsd) {
		for {
			err := ru.FetchCurs()
			if err != nil {
				fmt.Printf("Error fetching curs: %v\n", err)
			}
			time.Sleep(5 * time.Minute)
		}
	}(ru)

	//обновляем токены
	go func(ru *checkers.RubUsd, cPoll map[string]checkers.Checker) {
		for {
			for t, c := range cPoll {
				err := c.FetchTokens(ru, false)
				if err != nil {
					fmt.Printf("Error fetching tokens for %s: %v\n", t, err)
				}
			}
			time.Sleep(time.Minute)
		}
	}(ru, a.CheckersPoll)

	//проверка на появление новых токенов
	go func(a *app.App, b *tgbot.Bot) {
		for {
			for t, c := range a.CheckersPoll {
				if news := c.GetNews(); len(news) > 0 {
					for _, n := range news {
						resultMsg := "*Новый Токен на " + t + "\\!*\n\n" + tgbot.FormatNewTokenMsg(n)
						for _, u := range a.DB.GetAllUsers() {
							rec := &telebot.User{ID: int64(u.Id)}
							_, err := b.Bot.Send(rec, resultMsg, telebot.ModeMarkdownV2, b.Kb)
							if err != nil {
								fmt.Printf("Error sending message to user %d: %v\n", u.Id, err)
							}
						}
					}
				}
			}
			time.Sleep(time.Minute)
		}

	}(a, b)

	fmt.Println("Запуск бота v1.1.2...")
	b.Bot.Start()
}

//TIP See GoLand help at <a href="https://www.jetbrains.com/help/go/">jetbrains.com/help/go/</a>.
// Also, you can try interactive lessons for GoLand by selecting 'Help | Learn IDE Features' from the main menu.
