package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const apiTimeout = 2000
const bdTimeout = 1000

type AwesomeAPIResponse struct {
	Exchange CurrencyRate `json:"USDBRL"`
}

type CurrencyRate struct {
	ID    int    `gorm:"primaryKey" json:"-"`
	Code  string `json:"code"`
	Value string `json:"bid"`
}

func main() {
	http.HandleFunc("/", ExchangeRateHandler)
	http.ListenAndServe(":8080", nil)
}

func ExchangeRateHandler(rw http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/cotacao" {
		log.Println("404 Not Found")
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	currencyRate, err := getExchangeRate()
	if err != nil {
		log.Println(err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = saveCurrencyRate(*currencyRate)
	if err != nil {
		log.Println(err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)

	rw.Write([]byte("{\"valor\":" + currencyRate.Value + "}"))
}

func getExchangeRate() (*CurrencyRate, error) {
	apiContext, apiCancel := context.WithTimeout(context.Background(), apiTimeout*time.Millisecond)
	defer apiCancel()

	req, err := http.NewRequestWithContext(apiContext, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var awesomeapi AwesomeAPIResponse
	json.Unmarshal(body, &awesomeapi)
	return &awesomeapi.Exchange, nil
}

func saveCurrencyRate(currencyRate CurrencyRate) error {
	db, err := gorm.Open(sqlite.Open("currency.db"), &gorm.Config{})
	if err != nil {
		return err
	}

	db.AutoMigrate(&CurrencyRate{})

	ctx, cancel := context.WithTimeout(context.Background(), bdTimeout*time.Millisecond)
	defer cancel()

	err = db.WithContext(ctx).Create(&currencyRate).Error
	if err != nil {
		return err
	}

	return nil
}
