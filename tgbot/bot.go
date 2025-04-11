package tgbot

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
	"tochecken/app"
	"tochecken/models"
	"tochecken/tools"

	"github.com/joho/godotenv"
	"gopkg.in/telebot.v4"
)

var emojis = []string{
	"😊", "😍", "🤔", "🙌", "🎉", "👍", "🔥", "💡", "😎",
	"😅", "😇", "🐱", "🦄", "✨", "🚀", "🌟",
	"🥳", "😺", "🍀", "💪", "😏",
	"🧠", "📚", "🌍", "🍕", "☕", "🏆", "🤝",
}

var token = "" //prod
var logChatId = 0
var owner = ""

type WaitCommand struct {
	ChatId int64
	UserId int64
	Type   string
	Params []string
}

type Bot struct {
	Bot *telebot.Bot
	Wc  *WaitCommand
	Kb  *telebot.ReplyMarkup
}

func CreateBot(a *app.App) *Bot {
	initEnvVars()
	pref := telebot.Settings{
		Token:  token,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := telebot.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	// Клавиатура выбора биржи
	exchangeMenu := &telebot.ReplyMarkup{}
	btnBinance := exchangeMenu.Data("Binance", "binance")
	btnKraken := exchangeMenu.Data("Kraken", "kraken")
	btnOkx := exchangeMenu.Data("OKX", "okx")
	btnHelp := exchangeMenu.Data("О боте", "aboutBot")

	exchangeMenu.Inline(
		exchangeMenu.Row(btnBinance),
		exchangeMenu.Row(btnOkx),
		exchangeMenu.Row(btnKraken),
		exchangeMenu.Row(btnHelp),
	)

	// Обработчик кнопки Binance
	b.Handle(&btnBinance, func(c telebot.Context) error {
		bc, ok := a.CheckersPoll["Binance"]
		if !ok {
			return errors.New("binance not found in checkerPool")
		}
		err := c.Respond()
		if err != nil {
			log.Println(err)
		}
		txt := formatCryptoMsg(bc.GetTokens())

		return c.Send(fmt.Sprintf("*%s:*\n\n%s", bc.GetType(), txt), exchangeMenu, telebot.ModeMarkdownV2)
	})

	// Обработчик кнопки Kraken
	b.Handle(&btnKraken, func(c telebot.Context) error {
		kc, ok := a.CheckersPoll["Kraken"]
		if !ok {
			return errors.New("kraken not found in checkerPool")
		}
		err := c.Respond()
		if err != nil {
			log.Println(err)
		}
		txt := formatCryptoMsg(kc.GetTokens())

		return c.Send(fmt.Sprintf("*%s:*\n\n%s", kc.GetType(), txt), exchangeMenu, telebot.ModeMarkdownV2)
	})

	// Обработчик кнопки OKX
	b.Handle(&btnOkx, func(c telebot.Context) error {
		kc, ok := a.CheckersPoll["OKX"]
		if !ok {
			return errors.New("okx not found in checkerPool")
		}
		err := c.Respond()
		if err != nil {
			log.Println(err)
		}
		txt := formatCryptoMsg(kc.GetTokens())

		return c.Send(fmt.Sprintf("*%s:*\n\n%s", kc.GetType(), txt), exchangeMenu, telebot.ModeMarkdownV2)
	})

	// Обработчик команды /start
	b.Handle("/start", func(c telebot.Context) error {
		// Отправляем главное сообщение с кнопкой
		a.DB.AddUser(int(c.Sender().ID), c.Sender().Username)
		txt := "*Что это?*\nБот, который уведомляет о листинге новых токенов на криптобиржах\nВ данный момент мониторинг идёт по двум крупнейшим криптобиржам *Binance*, *OKX*, *Kraken*\n\n*Что\\-то ещё?*\nДа, можно получить список новых токенов по каждой бирже\n"

		return c.Send(txt, exchangeMenu, telebot.ModeMarkdownV2)
	})

	b.Handle(&btnHelp, func(c telebot.Context) error {
		err := c.Respond()
		if err != nil {
			log.Println(err)
		}
		return c.Send("*Бот создан с целью мониторинга появляения новых токенов на криптобиржах\n\n*"+
			"*Частота обновления данных*\n*Binance* \\- каждую минуту\n*OKX* \\- каждую минуту\n*Kraken* \\- каждую минуту\n*Курс рубля к доллару* \\- каждые 5 минут\n\n"+
			"В будущем будет добавлен мониторинг новых криптобирж\n"+
			"По вопросам и предложениям пишите @"+owner, exchangeMenu, telebot.ModeMarkdownV2)
	})

	b.Handle("/users", func(c telebot.Context) error {
		if c.Sender().Username != owner {
			return errors.New("Try get users by " + c.Sender().Username)
		}

		msg := formatUsersMsg(a.DB.GetAllUsers())
		msg = strings.ReplaceAll(msg, ".", "\\.")

		return c.Send("*Список пользователей*:\n\n"+msg, telebot.ModeMarkdownV2)
	})

	wc := &WaitCommand{}
	b.Handle("/sendAll", func(c telebot.Context) error {
		if c.Sender().Username != owner {
			return errors.New("Try sendAll by " + c.Sender().Username)
		}

		wc.UserId = c.Sender().ID
		wc.ChatId = c.Chat().ID
		wc.Type = "sendAll"

		return c.Send("Капитан, введите сообщение, которое будет отправлено ВСЕМ пользователям:")
	})

	b.Handle(telebot.OnText, func(c telebot.Context) error {
		if wc.Type != "" && wc.UserId == c.Sender().ID {
			switch wc.Type {
			case "sendAll":
				go func(msg string) {
					for _, u := range a.DB.GetAllUsers() {
						rec := &telebot.User{ID: int64(u.Id)}
						b.Send(rec, msg, telebot.ModeMarkdownV2)
					}
				}(c.Message().Text)

			case "send":
				go func(msg string, userIds []string) {
					for _, u := range userIds {
						id, _ := strconv.ParseInt(u, 10, 64)
						rec := &telebot.User{ID: id}
						b.Send(rec, msg, telebot.ModeMarkdownV2)
					}
				}(c.Message().Text, wc.Params)

			default:
				wc.Type = ""
				return c.Send("Неизвестная команда")
			}

			wc.Type = ""
		}

		return c.Send(handleMsg(c, b), exchangeMenu)
	})

	b.Handle("/send", func(c telebot.Context) error {
		if c.Sender().Username != owner {
			return errors.New("Try send by " + c.Sender().Username)
		}

		wc.UserId = c.Sender().ID
		wc.ChatId = c.Chat().ID
		wc.Type = "send"
		wc.Params = c.Args()

		return c.Send("Капитан, введите сообщение, которое будет отправлено пользователям: " + strings.Join(wc.Params, ", "))
	})

	b.Handle(telebot.OnPhoto, func(c telebot.Context) error {
		return c.Send(handleMsg(c, b), exchangeMenu)
	})

	b.Handle(telebot.OnVideo, func(c telebot.Context) error {
		return c.Send(handleMsg(c, b), exchangeMenu)
	})

	b.Handle(telebot.OnAudio, func(c telebot.Context) error {
		return c.Send(handleMsg(c, b), exchangeMenu)
	})

	b.Handle(telebot.OnForward, func(c telebot.Context) error {
		return c.Send(handleMsg(c, b), exchangeMenu)
	})

	b.Handle(telebot.OnDocument, func(c telebot.Context) error {
		return c.Send(handleMsg(c, b), exchangeMenu)
	})

	b.Handle(telebot.OnSticker, func(c telebot.Context) error {
		return c.Send(handleMsg(c, b), exchangeMenu)
	})

	b.Handle(telebot.OnAnimation, func(c telebot.Context) error {
		return c.Send(handleMsg(c, b), exchangeMenu)
	})

	b.Handle(telebot.OnAddedToGroup, func(c telebot.Context) error {
		log.Println("AG", c.Chat().ID)

		return nil
	})

	return &Bot{Bot: b, Wc: wc, Kb: exchangeMenu}
}

func FormatNewTokenMsg(m *models.CommonModel) string {
	return fmt.Sprintf("*[%s](%s)*:  %s \\| %s \\~\\(%s\\)\n\n", m.Name, m.Link, tools.PointsToSlashPoints(m.Price), tools.PointsToSlashPoints(m.PriceRub), tools.PointsToSlashPoints(m.Cap))

}

func formatCryptoMsg(m map[string]*models.CommonModel) string {
	var txt string
	for _, m := range m {
		txt = txt + fmt.Sprintf("*[%s](%s)*:  %s \\| %s \\~\\(%s\\)\n\n", m.Name, m.Link, tools.PointsToSlashPoints(m.Price), tools.PointsToSlashPoints(m.PriceRub), tools.PointsToSlashPoints(m.Cap))
	}

	return txt
}

func formatUsersMsg(us map[int]*models.User) string {
	var txt string
	for _, u := range us {
		txt = txt + fmt.Sprintf("`%d` \\- @%s\n", u.Id, u.Name)
	}

	return txt
}

func handleMsg(c telebot.Context, b *telebot.Bot) string {
	_, err := b.Forward(telebot.ChatID(logChatId), c.Message())
	log.Println(c.Sender().ID, c.Sender().Username)
	if err != nil {
		log.Printf("Ошибка пересылки сообщения: %v", err)
	}

	return "Ваше сообщение получено " + emojis[rand.Intn(len(emojis))]
}

func initEnvVars() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	token = os.Getenv("BOT_TOKEN")
	logChatId, err = strconv.Atoi(os.Getenv("LOG_GHAT_ID"))
	if err != nil {
		log.Println(err)
	}
	owner = os.Getenv("OWNER_NICKNAME")
}
