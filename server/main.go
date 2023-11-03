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
	_, err := InitializeDatabase()
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/cotacao", Handler)
	fmt.Println("Http Server listening on port 8080")

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func Handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	e, err := SearchCurrency(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
	}

	err = PersistCurrency(ctx, e)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(e.USDBRL.Bid)
}

func PersistCurrency(ctx context.Context, data *USDBRL) error {
	log.Println("Persisting currency")
	db, err := sql.Open("sqlite3", "sqlite3.db")
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	stmt := `INSERT INTO USDBRL(code, codein, name, high, low, var_bid, pct_change, bid, ask, timestamp, create_date) VALUES(?,?,?,?,?,?,?,?,?,?,?)`
	_, err = db.ExecContext(ctx, stmt, data.USDBRL.Code, data.USDBRL.Codein, data.USDBRL.Name, data.USDBRL.High, data.USDBRL.Low, data.USDBRL.VarBid, data.USDBRL.PctChange, data.USDBRL.Bid, data.USDBRL.Ask, data.USDBRL.Timestamp, data.USDBRL.CreateDate)
	if err != nil {
		return err
	}
	log.Println("Inserted data")

	defer db.Close()

	return nil
}

func SearchCurrency(ctx context.Context) (*USDBRL, error) {
	log.Println("Searching currency")

	ctx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var e USDBRL
	err = json.Unmarshal(body, &e)
	if err != nil {
		return nil, err
	}

	log.Println("Found data")

	return &e, nil
}

func InitializeDatabase() (*sql.DB, error) {
	database, err := sql.Open("sqlite3", "sqlite3.db")
	defer database.Close()

	if err != nil {
		return nil, err
	}

	err = Migrate(database)
	if err != nil {
		return nil, err
	}

	log.Println("Created table")

	return database, nil
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
		timestamp varchar(80),
		create_date varchar(80)
	)
	`
	_, err := db.Exec(sqlStmt)
	if err != nil {
		return err
	}

	return nil
}
