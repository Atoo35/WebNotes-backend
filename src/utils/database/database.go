package database

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	zlog "github.com/Atoo35/WebNotes-backend/src/utils/logger"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

type SQLClientI interface {
	NewDB()
	Get(email string) (string, error)
	Set(user *User) error
}

type sqlClient struct {
	DB *sql.DB
}

var SQLClient SQLClientI = &sqlClient{}
var dbOnce sync.Once

func (sc *sqlClient) NewDB() {
	dbOnce.Do(func() {
		if sc.DB != nil {
			return
		}
		db, err := sql.Open("sqlite3", "./tokens.db")
		if err != nil {
			// log.Fatalf exits the process
			log.Fatalf("Failed to open sqlite DB: %v", err)
			return
		}

		// Configure connection pooling. For sqlite the maximum open
		// connections should usually be 1; it doesn't support multiple
		// simultaneous connections well when using the file-based driver.
		db.SetMaxOpenConns(1)
		db.SetMaxIdleConns(1)
		db.SetConnMaxLifetime(time.Hour)

		// verify the connection
		if err := db.Ping(); err != nil {
			if zlog.Logger != nil {
				zlog.Logger.Error("Failed to ping DB", zap.Error(err))
			}
			// Close the db if ping failed
			_ = db.Close()
			return
		}

		sqlStmt := `
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			email TEXT,
			access_token TEXT,
			refresh_token TEXT
		);
		`
		_, err = db.Exec(sqlStmt)
		if err != nil {
			if zlog.Logger != nil {
				zlog.Logger.Error("Failed to create table 'users':", zap.Error(err))
			} else {
				log.Printf("Failed to create table 'users': %v\n", err)
			}
			_ = db.Close()
			return
		}

		if zlog.Logger != nil {
			zlog.Logger.Info("Database initialized and 'users' table created (if not exists)")
		}
		sc.DB = db
	})
}

// Close closes the database connection. It's safe to call multiple times.
func (sc *sqlClient) Close() error {
	if sc.DB == nil {
		return nil
	}
	err := sc.DB.Close()
	if err == nil {
		sc.DB = nil
	}
	return err
}

func (sc *sqlClient) Get(email string) (string, error) {
	if sc.DB == nil {
		fmt.Println("DB not initialised")
		return "", fmt.Errorf("DB not initialised")
	}
	var accessToken string
	err := sc.DB.QueryRow("SELECT access_token FROM users WHERE email = ?", email).Scan(&accessToken)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("no user found with email: %s", email)
		}
		zlog.Logger.Error("Failed to query 'users' table:", zap.Error(err))
		return "", err
	}
	return accessToken, nil
}

func (sc *sqlClient) Set(user *User) error {
	if sc.DB == nil {
		fmt.Println("DB not initialised")
		return fmt.Errorf("DB not initialised")
	}

	_, err := sc.DB.Exec("INSERT INTO users (email, access_token, refresh_token) VALUES (?, ?, ?)", user.Email, user.AccessToken, user.RefreshToken)
	if err != nil {
		zlog.Logger.Error("Failed to insert into 'users' table:", zap.Error(err))
		return err
	}
	return nil
}
