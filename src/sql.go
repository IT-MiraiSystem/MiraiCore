package main

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"strconv"

	"github.com/360EntSecGroup-Skylar/excelize"
	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
)

type DBconfig struct {
	host     string
	port     int
	database string
	user     string
	pass     string
}

func DBinit(configPath string) (DBconfig, error) {
	byteArray, err := ioutil.ReadFile(configPath)
	if err != nil {
		return DBconfig{}, err
	}
	var config map[string]interface{}
	err = json.Unmarshal(byteArray, &config)
	if err != nil {
		return DBconfig{}, err
	}
	var dbconfig DBconfig
	dbconfig.host = config["db"].(map[string]interface{})["host"].(string)
	portFloat := config["db"].(map[string]interface{})["port"].(float64)
	dbconfig.port = int(portFloat)
	dbconfig.database = config["db"].(map[string]interface{})["database"].(string)
	dbconfig.user = config["db"].(map[string]interface{})["user"].(string)
	dbconfig.pass = config["db"].(map[string]interface{})["pass"].(string)
	return dbconfig, nil
}

func SearchUser(uid int) {
	db, err := sql.Open("mysql", SQLconfig.user+":"+SQLconfig.pass+"@tcp("+SQLconfig.host+":"+strconv.Itoa(SQLconfig.port)+")/"+SQLconfig.database)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	db.Query("SELECT * FROM users WHERE id = ?", uid)
}

func InsertUser(uid string, email string, photoUrl string) {
	db, err := sql.Open("mysql", SQLconfig.user+":"+SQLconfig.pass+"@tcp("+SQLconfig.host+":"+strconv.Itoa(SQLconfig.port)+")/"+SQLconfig.database)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	// Excelから生徒情報を読み込む
	f, err := excelize.OpenFile(AppDir + "/db/IT未来在学生.xlsx")
	if err != nil {
		log.Errorf("Error: %v", err)
		return
	}
	var Numbers []string
	var Names []string
	var Emails []string
	var Clubs []string
	var ClassList []string
	var GradeList []string
	for _, sheetName := range f.GetSheetMap() {
		rows := f.GetRows(sheetName)
		for _, row := range rows {
			GradeList = append(GradeList, row[0])
			ClassList = append(ClassList, row[1])
			Numbers = append(Numbers, row[2])
			Names = append(Names, row[3])
			Emails = append(Emails, row[4])
			Clubs = append(Clubs, row[5])
		}
	}
	for i, e := range Emails {
		if e == email {
			number, err := strconv.Atoi(Numbers[i])
			if err != nil {
				log.Errorf("Error converting number: %v", err)
				return
			}
			log.Infof("Inserting user: %v", Names[i])
			_, err = db.Exec("INSERT INTO Users(uid,name,photoURL,GradeInSchool,ClassInSchool,email,SchoolClub,Number) VALUES (?,?,?,?,?,?,?,?)", uid, Names[i], photoUrl, GradeList[i], ClassList[i], email, Clubs[i], number)
			if err != nil {
				log.Errorf("Error inserting user: %v", err)
				return
			}
			break
		}
	}
}
