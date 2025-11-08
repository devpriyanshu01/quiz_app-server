package sqlconnect

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func ConnectDb() (*sql.DB, error) {

	dbName := os.Getenv("DB_NAME")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbPort := os.Getenv("DB_PORT")
	host := os.Getenv("HOST")

	fmt.Println("Trying to connect to MariaDB")

	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPassword, host, dbPort, dbName)
	
	fmt.Println("databaseurl:", connectionString)

	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		return nil, err
	}

	fmt.Println("Connection established with MariaDB")

	return db, nil
}
