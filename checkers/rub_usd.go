package checkers

import (
	"encoding/json"
	"io"
	"net/http"
)

type RubUsdRaw struct {
	Rates struct {
		USD float64 `json:"USD"`
	} `json:"rates"`
}

type RubUsd struct {
	Curs float64 `json:"USD"`
}

func NewRubUsd() (*RubUsd, error) {
	ru := RubUsd{}

	err := ru.FetchCurs()
	if err != nil {
		return nil, err
	}

	return &ru, nil
}

func (r *RubUsd) FetchCurs() error {
	url := "https://www.cbr-xml-daily.ru/latest.js"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "*/*")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:133.0) Gecko/20100101 Firefox/133.0")
	req.Header.Set("Accept-Language", "ru-RU,ru;q=0.8,en-US;q=0.5,en;q=0.3")
	req.Header.Set("Connection", "keep-alive")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	result := RubUsdRaw{}

	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	r.Curs = result.Rates.USD

	return nil
}
