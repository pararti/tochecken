package db

import (
	"database/sql"
	"log"
	"time"
	"tochecken/models"

	_ "github.com/mattn/go-sqlite3"
)

type Db struct {
	DB         *sql.DB
	ValidCache bool
	Users      map[int]*models.User
}

func NewDb() *Db {
	db, err := sql.Open("sqlite3", "./tgbot.db")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY, name TEXT, flags INTEGER, created_at TEXT)`)
	if err != nil {
		log.Fatal(err)
	}

	usersChache := make(map[int]*models.User)

	return &Db{DB: db, ValidCache: false, Users: usersChache}
}

func (db *Db) AddUser(userId int, name string) {
	now := time.Now().Format("01.02.2006 15:04:05")
	sql := "INSERT INTO users (id, name, flags, created_at) VALUES (?, ?, ?, ?)"
	_, err := db.DB.Exec(sql, userId, name, 0, now)
	if err != nil {
		log.Println(err)
		return
	}
	db.ValidCache = false
}

func (db *Db) GetUser(userId int) *models.User {
	if db.ValidCache {
		return db.Users[userId]
	}

	sql := "SELECT id, name, flags FROM users WHERE id = ?"
	row := db.DB.QueryRow(sql, userId)
	if row.Err() != nil {
		log.Println(row.Err())
		return nil
	}
	user := models.User{}
	err := row.Scan(&user.Id, &user.Name, &user.Flags)
	if err != nil {
		log.Println(err)
		return nil
	}

	return &user
}

func (db *Db) GetAllUsers() map[int]*models.User {
	if db.ValidCache {
		return db.Users
	}

	sql := "SELECT id, name, flags FROM users"
	rows, err := db.DB.Query(sql)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer rows.Close()

	for rows.Next() {
		u := models.User{}
		err := rows.Scan(&u.Id, &u.Name, &u.Flags)
		if err != nil {
			log.Println(err)
			continue
		}
		db.Users[u.Id] = &u
	}

	db.ValidCache = true

	return db.Users
}
