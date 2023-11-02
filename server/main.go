package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
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
	db, err := sql.Open("sqlite3", "sqlite3.db")
	if err != nil {
		log.Fatal(err)
	}
	Migrate(db)
	db.Close()

	http.HandleFunc("/cotacao", Handler)
	fmt.Println("Http Server listening on port 8080")

	err = http.ListenAndServe(":8080", nil)
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

	err = PersistCurrency(db_ctx, e)
	if err != nil {
		log.Fatal(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(e.USDBRL.Bid)
}
func PersistCurrency(ctx context.Context, data *USDBRL) error {
	stmt := `INSERT INTO USDBRL(code, codein, name, high, low, var_bid, pct_change, bid, ask, timestamp, create_date) VALUES(?,?,?,?,?,?,?,?,?,?,?)`
	db, err := sql.Open("sqlite3", "sqlite3.db")
	if err != nil {
		return err
	}

	_, err = db.Exec(stmt, data.USDBRL.Code, data.USDBRL.Codein, data.USDBRL.Name, data.USDBRL.High, data.USDBRL.Low, data.USDBRL.VarBid, data.USDBRL.PctChange, data.USDBRL.Bid, data.USDBRL.Ask, data.USDBRL.Timestamp, data.USDBRL.CreateDate)

	defer db.Close()

	return nil
}

func Migrate(db *sql.DB) error {
	sqlStmt := `
	CREATE TABLE IF NOT EXISTS USDBRL (
		id integer primary key autoincrement,
		code varchar(80),
		codein varchar(80),
		name varchar(80),
		high varchar(80),
		low varchar(80),
		var_bid varchar(80),
		pct_change varchar(80),
		bid varchar(80),
		ask varchar(80),
		timestamp timestamp,
		create_date timestamp
	)
	`
	_, err := db.Exec(sqlStmt)
	if err != nil {
		return err
	}

	return nil
}

func SearchCurrency(ctx context.Context, CurrencyCode string) (*USDBRL, error) {
	if CurrencyCode == "" {
		CurrencyCode = "USD-BRL"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://economia.awesomeapi.com.br/json/last/"+CurrencyCode, nil)
	if err != nil {
		log.Fatal(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	var e USDBRL
	err = json.Unmarshal(body, &e)
	if err != nil {
		log.Fatal(err)
	}

	return &e, nil
}
