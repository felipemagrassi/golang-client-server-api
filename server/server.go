package main

import (
	"context"
	"database/sql"
	"encoding/json"
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

type ExchangeHandler struct {
	DB *sql.DB
}

func main() {
	db, err := InitializeDatabase()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	mux := http.NewServeMux()
	mux.Handle("/cotacao", &ExchangeHandler{DB: db})

	log.Fatal(http.ListenAndServe(":8080", mux))
}

func (h *ExchangeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data, err := SearchCurrency()
	if err != nil {
		log.Fatal(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
	}

	err = PersistCurrency(h.DB, &data.USDBRL)
	if err != nil {
		log.Fatal(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data.USDBRL.Bid)
}

func PersistCurrency(db *sql.DB, data *ExchangeRate) error {
	log.Println("Persisting currency")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	stmt, err := db.PrepareContext(ctx, "INSERT INTO exchange(code, codein, name, high, low, var_bid, pct_change, bid, ask, timestamp, create_date) VALUES(?,?,?,?,?,?,?,?,?,?,?);")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(data.Code, data.Codein, data.Name, data.High, data.Low, data.VarBid, data.PctChange, data.Bid, data.Ask, data.Timestamp, data.CreateDate)
	if err != nil {
		return err
	}

	log.Println("Currency Persisted")

	return nil
}

func SearchCurrency() (*USDBRL, error) {
	log.Println("Searching currency")

	ctx, cancel := context.WithTimeout(context.Background(), 2000*time.Millisecond)
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

	log.Println("Currency found")

	return &e, nil
}

func InitializeDatabase() (*sql.DB, error) {
	database, err := sql.Open("sqlite3", "sqlite3.db")

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
	stmt, err := db.Prepare(`
	CREATE TABLE IF NOT EXISTS exchange (
		id integer primary key autoincrement,
		code TEXT,
		codein TEXT,
		name TEXT,
		high TEXT,
		low TEXT,
		var_bid TEXT,
		pct_change TEXT,
		bid TEXT,
		ask TEXT,
		timestamp TEXT,
		create_date TEXT
	);
	`)
	if err != nil {
		return err
	}

	_, err = stmt.Exec()
	if err != nil {
		return err
	}

	return nil
}
