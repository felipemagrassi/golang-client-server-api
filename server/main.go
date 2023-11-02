package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type USDBRL struct {
	USDBRL ExchangeRate `json:"USDBRL"`
}

type ExchangeRate struct {
	Code       string `json:"code"`
	Codein     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"varBid"`
	PctChange  string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Timestamp  string `json:"timestamp"`
	CreateDate string `json:"create_date"`
}

func main() {
	http.HandleFunc("/cotacao", Handler)
	fmt.Println("Http Server listening on port 8080")

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

func Handler(w http.ResponseWriter, r *http.Request) {
	http_ctx, http_cancel := context.WithTimeout(context.Background(), time.Millisecond*200)
	defer http_cancel()

	db_ctx, db_cancel := context.WithTimeout(context.Background(), time.Millisecond*10)
	defer db_cancel()

	e, err := SearchCurrency(http_ctx, "")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(e)
}

func PersistCurrency(ctx context.Context, db *sql.DB) {

}

func SearchCurrency(ctx context.Context, CurrencyCode string) (*USDBRL, error) {
	if CurrencyCode == "" {
		CurrencyCode = "USD-BRL"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://economia.awesomeapi.com.br/json/last/"+CurrencyCode, nil)
	if err != nil {
		panic(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	var e USDBRL
	err = json.Unmarshal(body, &e)
	if err != nil {
		panic(err)
	}

	return &e, nil
}
