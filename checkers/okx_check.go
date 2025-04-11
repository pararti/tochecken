package checkers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"tochecken/models"
	"tochecken/tools"
)

const minPrice = 1e-3

type OkxData struct {
	Data struct {
		List []struct {
			FullNameSeo string  `json:"fullNameSeo"` //path
			Last        float64 `json:"last"`        //price
			MarketCap   float64 `json:"marketCap"`
			Project     string  `json:"project"` //name
		} `json:"list"`
	} `json:"data"`
}

type OkxChecker struct {
	Type   string
	News   []string
	Tokens map[string]*models.CommonModel
}

func NewOkxChecker() *OkxChecker {

	return &OkxChecker{Type: "OKX", Tokens: make(map[string]*models.CommonModel, 32), News: make([]string, 0, 1)}
}

func (c *OkxChecker) FetchTokens(ru *RubUsd, firstStart bool) error {
	url := "https://www.okx.com/v2/support/info/announce/listProject?pageNum=1&pageSize=30&typeId=41&sortField=marketCap&sortType=desc&countryFilter=0"
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error fetching data: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %v", err)
	}

	var response OkxData
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("error decoding JSON: %v", err)
	}

	for _, e := range response.Data.List {
		if e.Last < minPrice {
			continue
		}

		var m models.CommonModel
		m.Name = e.Project

		m.PriceRub = fmt.Sprintf("%.2f₽", e.Last/ru.Curs)
		if e.Last < 0.01 {
			m.Price = fmt.Sprintf("%.3f$", e.Last)
		} else {
			m.Price = fmt.Sprintf("%.2f$", e.Last)
		}
		m.Cap = tools.FormatNumberUSDFloat(e.MarketCap)

		m.Link = "https://www.okx.com/ru/price/" + e.FullNameSeo + "-" + strings.ToLower(e.Project)

		if firstStart == false {
			if _, ok := c.Tokens[m.Name]; !ok {
				c.News = append(c.News, m.Name)
			}
		}

		c.Tokens[m.Name] = &m

	}

	return nil
}

func (c *OkxChecker) GetNews() []*models.CommonModel {
	ms := make([]*models.CommonModel, 0, len(c.News))

	for _, n := range c.News {
		ms = append(ms, c.Tokens[n])
	}

	c.News = make([]string, 0)

	return ms
}

func (c *OkxChecker) GetTokens() map[string]*models.CommonModel {
	return c.Tokens
}

func (c *OkxChecker) GetType() string {
	return c.Type
}
