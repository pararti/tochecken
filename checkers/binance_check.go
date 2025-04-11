package checkers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"tochecken/models"
	"tochecken/tools"
)

type BinanceData struct {
	Data []struct {
		Name        string   `json:"b"`    //TRUMP
		Type        string   `json:"q"`    //USDT
		Description string   `json:"an"`   //OFFICIAL TRUMP its description
		Qn          string   `json:"qn"`   //TetherUS
		Price       string   `json:"c"`    //Price
		Volume      string   `json:"v"`    //объем токена
		Cap         string   `json:"qv"`   //капитализация
		Tags        []string `json:"tags"` //теги нас интересует newListing
	} `json:"data"`
	Success bool `json:"success"`
}

type BinanceChecker struct {
	Type   string
	News   []string
	Tokens map[string]*models.CommonModel
}

func NewBinanceChecker() *BinanceChecker {

	return &BinanceChecker{Type: "Binance", Tokens: make(map[string]*models.CommonModel, 32), News: make([]string, 0, 1)}
}

func (c *BinanceChecker) FetchTokens(ru *RubUsd, firstStart bool) error {
	url := "https://www.binance.com/bapi/asset/v2/public/asset-service/product/get-products?includeEtf=true"
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error fetching data: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %v", err)
	}

	var response BinanceData
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("error decoding JSON: %v", err)
	}

	if response.Success != true {
		return errors.New("failed fetch data param success does not true")
	}

	needTag := "newListing"
	for _, e := range response.Data {
		if e.Type != "USDT" {
			continue
		}

		if slices.Contains(e.Tags, needTag) == false {
			continue
		}
		var m models.CommonModel
		m.Name = e.Name

		parsedPrice, _ := strconv.ParseFloat(e.Price, 64)
		m.PriceRub = fmt.Sprintf("%.2f₽", parsedPrice/ru.Curs)
		if parsedPrice < 0.01 {
			m.Price = fmt.Sprintf("%.3f$", parsedPrice)
		} else {
			m.Price = fmt.Sprintf("%.2f$", parsedPrice)
		}
		m.Cap = tools.FormatNumberUSD(e.Cap)
		m.Link = "https://www.binance.com/en/trade/" + strings.ToUpper(e.Name) + "_USDT"

		if firstStart == false {
			if _, ok := c.Tokens[m.Name]; !ok {
				c.News = append(c.News, m.Name)
			}
		}

		c.Tokens[m.Name] = &m

	}

	return nil
}

func (c *BinanceChecker) GetNews() []*models.CommonModel {
	ms := make([]*models.CommonModel, 0, len(c.News))

	for _, n := range c.News {
		ms = append(ms, c.Tokens[n])
	}

	c.News = make([]string, 0)

	return ms
}

func (c *BinanceChecker) GetTokens() map[string]*models.CommonModel {
	return c.Tokens
}

func (c *BinanceChecker) GetType() string {
	return c.Type
}
