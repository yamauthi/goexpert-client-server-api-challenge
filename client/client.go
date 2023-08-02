package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type CurrencyRate struct {
	ExchangeRate float64 `json:"valor"`
}

func main() {
	var rate CurrencyRate
	getExchangeRate(&rate)

	if rate.ExchangeRate != 0.0 {
		saveExchangeRateToFile(&rate)
	}
}

func getExchangeRate(rate *CurrencyRate) {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
	err = json.Unmarshal(body, rate)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao fazer parse da resposta: %v\n", err)
	}

	fmt.Fprintf(os.Stdout, "Valor do dólar: U$ %f\n", rate.ExchangeRate)
}

func saveExchangeRateToFile(rate *CurrencyRate) {
	file, err := os.Create("cotacao.txt")
	if err != nil {
		fmt.Println(err)
		return
	}

	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("Dólar: %f", rate.ExchangeRate))
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Cotação salva com sucesso")
}
