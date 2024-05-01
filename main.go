package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"

	_ "github.com/lib/pq"
)

type Product struct {
	Name      string
	Price     float64
	Available bool
}

func main() {
	conStr := "postgres://postgres:secret@localhost:5432/gopgtest?sslmode=disable"

	db, err := sql.Open("postgres", conStr)

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	createProductTable(db)

	product := Product{"Book", 19.99, true}

	pk := insertProduct(db, product)

	fmt.Printf("ID:=%v\n", pk)

	queryRow(db, pk)

	queryTable(db)
}

func createProductTable(db *sql.DB) {
	query := `CREATE TABLE IF NOT EXISTS product (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		price NUMERIC(6, 2) NOT NULL,
		available BOOLEAN,
		created timestamp DEFAULT NOW()
	)`

	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

func insertProduct(db *sql.DB, product Product) int {
	query := `INSERT INTO product (name, price, available)
	VALUES ($1, $2, $3) RETURNING id`

	var pk int
	err := db.QueryRow(query, product.Name, product.Price, product.Available).Scan(&pk)
	if err != nil {
		log.Fatal(err)
	}
	return pk
}

func queryRow(db *sql.DB, pk int) {
	var name string
	var price float64
	var available bool

	query := `SELECT name, price, available FROM product WHERE id = $1`

	err := db.QueryRow(query, pk).Scan(&name, &price, &available)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Fatal("No rows found with the id " + strconv.Itoa(pk))
		}
		log.Fatal(err)
	}
	fmt.Printf("Name:= %v, Price:= %v, Available:= %v\n", name, price, available)
}

func queryTable(db *sql.DB) {
	data := []Product{}

	query := `SELECT name, price, available FROM product`
	rows, err := db.Query(query)

	if err != nil {
		log.Fatal(err)
	}

	var name string
	var price float64
	var available bool

	for rows.Next() {
		err := rows.Scan(&name, &price, &available)
		if err != nil {
			log.Fatal(err)
		}
		data = append(data, Product{name, price, available})
	}
	fmt.Println(data, "Data")
}
