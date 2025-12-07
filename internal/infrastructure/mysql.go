package infrastructure

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/assidik12/go-restfull-api/config"
)

func DatabaseConnection(c config.Config) *sql.DB {

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True",
		c.DBUser,
		c.DBPassword,
		c.DBHost,
		c.DBPort,
		c.DBName,
	))

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("connecting to database...")

	err = db.Ping()

	if err != nil {
		log.Fatal(err)
		panic(errors.New("connection to database failed"))
	}

	fmt.Println("connection to database success...")

	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(10)
	db.SetConnMaxIdleTime(10 * time.Minute)
	db.SetConnMaxLifetime(60 * time.Minute)

	return db
}
