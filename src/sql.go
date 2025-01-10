package main

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

type DBconfig struct {
	host     string
	port     int
	database string
	user     string
	pass     string
}

func DBinit(configPath string) {
	byteArray, _ := ioutil.ReadFile(configPath)
	var config interface{}
	_ = json.Unmarshal(byteArray, &config)
	var dbconfig DBconfig
	dbconfig.host = config.(map[string]interface{})["db"].(map[string]interface{})["host"].(string)
	dbconfig.port = config.(map[string]interface{})["db"].(map[string]interface{})["port"].(int)
	dbconfig.database = config.(map[string]interface{})["db"].(map[string]interface{})["database"].(string)
	dbconfig.user = config.(map[string]interface{})["db"].(map[string]interface{})["user"].(string)
	dbconfig.pass = config.(map[string]interface{})["db"].(map[string]interface{})["pass"].(string)
	return
}

func SearchUser(uid int) {
	db, err := sql.Open("mysql", SQLconfig.user+":"+SQLconfig.pass+"@tcp("+SQLconfig.host+":"+strconv.Itoa(SQLconfig.port)+")/"+SQLconfig.database)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	db.Query("SELECT *")
}
