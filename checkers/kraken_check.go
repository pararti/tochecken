package checkers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"tochecken/models"
	"tochecken/tools"
)

type KrakenData struct {
	Result struct {
		Data []struct {
			Symbol      string   `json:"symbol"` //token name MELANIA
			Name        string   `json:"name"`
			Price       string   `json:"price"`      //price usd
			Volume24H   string   `json:"volume_24h"` //кол-во токенов
			MarketCap   string   `json:"market_cap"` //объём рынка капитал
			ListingDate int      `json:"listing_date"`
			Categories  []string `json:"categories"`
		} `json:"data"`
	} `json:"result"`
	Errors []interface{} `json:"errors"`
}

type KrakenChecker struct {
	Type   string
	News   []string
	Tokens map[string]*models.CommonModel
}

func NewKrakenChecker() *KrakenChecker {

	return &KrakenChecker{Type: "Kraken", Tokens: make(map[string]*models.CommonModel, 32), News: make([]string, 0, 1)}
}

func (c *KrakenChecker) FetchTokens(ru *RubUsd, firstStart bool) error {
	url := "https://iapi.kraken.com/api/internal/markets/all/assets?sort_by=listing_date&page=0&sort_order=descending&quote_symbol=usd&tradable=true&page_size=16"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:133.0) Gecko/20100101 Firefox/133.0")
	req.Header.Set("Accept-Language", "ru-RU,ru;q=0.8,en-US;q=0.5,en;q=0.3")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("X-Kraken-Asset-Name", "new")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "no-cors")
	req.Header.Set("Sec-Fetch-Site", "same-site")
	req.Header.Set("Sec-GPC", "1")
	req.Header.Set("Referer", "https://www.kraken.com/")

	// Создаем клиент
	client := &http.Client{}

	// Выполняем запрос
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %v", err)
	}

	var response KrakenData
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("error decoding JSON: %v", err)
	}

	for _, e := range response.Result.Data {

		var m models.CommonModel
		m.Name = e.Symbol

		parsedPrice, _ := strconv.ParseFloat(e.Price, 64)
		m.PriceRub = fmt.Sprintf("%.2f₽", parsedPrice/ru.Curs)
		if parsedPrice < 0.01 {
			m.Price = fmt.Sprintf("%.3f$", parsedPrice)
		} else {
			m.Price = fmt.Sprintf("%.2f$", parsedPrice)
		}
		m.Cap = tools.FormatNumberUSD(e.MarketCap)

		m.Link = "https://www.kraken.com/ru-ru/prices/" + strings.ReplaceAll(strings.ToLower(e.Name), " ", "-")

		if firstStart == false {
			if _, ok := c.Tokens[m.Name]; !ok {
				c.News = append(c.News, m.Name)
			}
		}

		c.Tokens[m.Name] = &m

	}

	return nil
}

func (c *KrakenChecker) GetNews() []*models.CommonModel {
	ms := make([]*models.CommonModel, 0, len(c.News))

	for _, n := range c.News {
		ms = append(ms, c.Tokens[n])
	}

	c.News = make([]string, 0)

	return ms
}

func (c *KrakenChecker) GetTokens() map[string]*models.CommonModel {
	return c.Tokens
}

func (c *KrakenChecker) GetType() string {
	return c.Type
}
