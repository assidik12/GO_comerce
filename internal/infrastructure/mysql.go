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

	DB, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True",
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

	err = DB.Ping()

	if err != nil {
		log.Fatal(err)
		panic(errors.New("connection to database failed"))
	}

	fmt.Println("connection to database success...")

	DB.SetMaxIdleConns(5)
	DB.SetMaxOpenConns(10)
	DB.SetConnMaxIdleTime(10 * time.Minute)
	DB.SetConnMaxLifetime(60 * time.Minute)

	return DB
}
