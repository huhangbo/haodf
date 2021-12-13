package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var db *sqlx.DB

func Init() (err error) {
	// "user:password@tcp(host:port)/dbname"
	//dsn := "root:QWER!@#$qwer1234@tcp(127.0.0.1:3306)/cjy?parseTime=true&loc=Local"
	dsn := "root:root@tcp(127.0.0.1:3306)/go_test?parseTime=true&loc=Local"
	db, err = sqlx.Connect("mysql", dsn)
	if err != nil {
		fmt.Printf("%v", err)
		return
	}
	db.SetMaxOpenConns(200)
	db.SetMaxIdleConns(50)
	return
}

func Close() {
	_ = db.Close()
}
