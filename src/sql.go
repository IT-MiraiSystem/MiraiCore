package main

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"strconv"
	"time"

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

type User struct {
	UID           string `json:"uid"`
	Name          string `json:"name"`
	PhotoURL      string `json:"photoURL"`
	GradeInSchool int    `json:"gradeInSchool"`
	ClassInSchool string `json:"classInSchool"`
	Email         string `json:"email"`
	SchoolClub    string `json:"schoolClub"`
	Number        int    `json:"number"`
	Permission    int    `json:"permission"`
}

type Lesson struct {
	ClassID      string `json:"ClassID"`
	DayOfTheWeek string `json:"DayOfTheWeek"`
	LessonNumber int    `json:"LessonNumber"`
	Lesson       string `json:"Lesson"`
	Room         string `json:"Room"`
	Teacher      string `json:"Teacher"`
	StartTime    string `json:"StartTime"`
	EndTime      string `json:"EndTime"`
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
func UserList() (statuscode int, userList []User) {
	db, err := sql.Open("mysql", SQLconfig.user+":"+SQLconfig.pass+"@tcp("+SQLconfig.host+":"+strconv.Itoa(SQLconfig.port)+")/"+SQLconfig.database)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	rows, err := db.Query("SELECT uid,name,email,photoURL,GradeInSchool,ClassInSchool,Number,SchoolClub,Permission FROM Users")
	if err != nil {
		log.Errorf("Error getting user list: %v", err)
		return 500, nil
	}
	defer rows.Close()
	for rows.Next() {
		var user User
		err = rows.Scan(&user.UID, &user.Name, &user.Email, &user.PhotoURL, &user.GradeInSchool, &user.ClassInSchool, &user.Number, &user.SchoolClub, &user.Permission)
		if err != nil {
			log.Fatal(err)
		}
		userList = append(userList, user)
	}
	return 200, userList
}

func UserInfo(uid string) (user User) {
	db, err := sql.Open("mysql", SQLconfig.user+":"+SQLconfig.pass+"@tcp("+SQLconfig.host+":"+strconv.Itoa(SQLconfig.port)+")/"+SQLconfig.database)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	rows, err := db.Query("SELECT uid,name,email,photoURL,GradeInSchool,ClassInSchool,Number,SchoolClub,Permission FROM Users WHERE uid = ?", uid)
	if err != nil {
		log.Errorf("Error getting user info: %v", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			err = rows.Scan(&user.UID, &user.Name, &user.Email, &user.PhotoURL, &user.GradeInSchool, &user.ClassInSchool, &user.Number, &user.SchoolClub, &user.Permission)
			if err != nil {
				log.Fatal(err)
			}
		}
		return user
	}
	return User{}
}

func UpdateLesson(classid string, DayOfTheWeek string, lessonNumber int, lesson string, room string, teacher string, date string) (statuscode int) {
	db, err := sql.Open("mysql", SQLconfig.user+":"+SQLconfig.pass+"@tcp("+SQLconfig.host+":"+strconv.Itoa(SQLconfig.port)+")/"+SQLconfig.database)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	_, err = db.Exec("INSERT INTO ChangeOfClass(ClassID, LeasonNumber, Lesson, Room, Teacher, Date,DayOfTheWeek) VALUES (?,?,?,?,?,?,?)", classid, lessonNumber, lesson, room, teacher, date, DayOfTheWeek)
	if err != nil {
		log.Errorf("Error updating lesson: %v", err)
		return 500
	}
	return 200
}

func GetLesson(classid string, startDate string, EndDate string) (statuscode int, returnLesson map[string]interface{}) {
	db, err := sql.Open("mysql", SQLconfig.user+":"+SQLconfig.pass+"@tcp("+SQLconfig.host+":"+strconv.Itoa(SQLconfig.port)+")/"+SQLconfig.database)
	if err != nil {
		log.Errorf("Error opening database connection: %v", err)
		return 500, nil
	}
	defer db.Close()

	DayOfWeek := [...]string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday"}
	returnLesson = make(map[string]interface{})

	for _, day := range DayOfWeek {
		// クエリ実行
		rows, err := db.Query(`
			SELECT 
				ClassTimetable.ClassID, 
				ClassTimetable.DayOfTheWeek, 
				ClassTimetable.LeasonNumber, 
				COALESCE(ChangeOfClass.Lesson, ClassTimetable.Lesson) AS Lesson, 
				COALESCE(ChangeOfClass.Room, ClassTimetable.Room) AS Room, 
				COALESCE(ChangeOfClass.Teacher, ClassTimetable.Teacher) AS Teacher, 
				ClassTimetable.StartTime, 
				ClassTimetable.EndTime 
			FROM 
				ClassTimetable 
			LEFT JOIN 
				ChangeOfClass 
			ON 
				ClassTimetable.ClassID = ChangeOfClass.ClassID 
				AND ClassTimetable.LeasonNumber = ChangeOfClass.LeasonNumber 
				AND ChangeOfClass.Date BETWEEN ? AND ? 
			WHERE 
				ClassTimetable.ClassID = ? 
				AND ClassTimetable.DayOfTheWeek = ?
			ORDER BY
				ClassTimetable.LeasonNumber;
		`, startDate, EndDate, classid, day)
		if err != nil {
			log.Errorf("Error querying database for day %s: %v", day, err)
			return 500, nil
		}

		// rows を defer でクローズする前に適切に処理
		var lessons []Lesson
		for rows.Next() {
			var lesson Lesson
			if err := rows.Scan(
				&lesson.ClassID,
				&lesson.DayOfTheWeek,
				&lesson.LessonNumber,
				&lesson.Lesson,
				&lesson.Room,
				&lesson.Teacher,
				&lesson.StartTime,
				&lesson.EndTime,
			); err != nil {
				log.Errorf("Error scanning row for day %s: %v", day, err)
				rows.Close() // 明示的にリソースを解放
				return 500, nil
			}
			lessons = append(lessons, lesson)
		}
		rows.Close() // 明示的にリソースを解放

		// その日のデータを結果マップに追加
		returnLesson[day] = lessons
	}

	// 正常終了
	return 200, returnLesson
}

func AttendSchool(uid string) (statuscode int) {
	db, err := sql.Open("mysql", SQLconfig.user+":"+SQLconfig.pass+"@tcp("+SQLconfig.host+":"+strconv.Itoa(SQLconfig.port)+")/"+SQLconfig.database)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	var user User = UserInfo(uid)
	currentDate := time.Now().Format("2006-01-02")
	CommuteTime := time.Now().Format("15:04:05")
	_, err = db.Exec("INSERT INTO GoSchool (uid, name, email, photoURL, GradeInSchool, ClassInSchool, Number, Date, CommuteTime) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)", uid, user.Name, user.Email, user.PhotoURL, user.GradeInSchool, user.ClassInSchool, user.Number, currentDate, CommuteTime)
	if err != nil {
		log.Errorf("Error going to school: %v", err)
		return 500
	}
	return 200
}

func SearchUser(uid string) (statuscode int) {
	db, err := sql.Open("mysql", SQLconfig.user+":"+SQLconfig.pass+"@tcp("+SQLconfig.host+":"+strconv.Itoa(SQLconfig.port)+")/"+SQLconfig.database)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	rows, err := db.Query("SELECT * FROM Users WHERE uid = ?", uid)
	if err != nil {
		log.Errorf("Error searching user: %v", err)
		return 500
	}
	defer rows.Close()
	if rows.Next() {
		return 200
	} else {
		return 404
	}
}

func InsertUser(uid string, email string, photoUrl string) (Permission int, statuscode int) {
	db, err := sql.Open("mysql", SQLconfig.user+":"+SQLconfig.pass+"@tcp("+SQLconfig.host+":"+strconv.Itoa(SQLconfig.port)+")/"+SQLconfig.database)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	// Excelから生徒情報を読み込む
	f, err := excelize.OpenFile(AppDir + "/db/IT未来在学生.xlsx")
	if err != nil {
		log.Errorf("Error: %v", err)
		return 0, 500
	}
	var Numbers []string
	var Names []string
	var Emails []string
	var Clubs []string
	var ClassList []string
	var GradeList []string
	var Permissions []int
	for _, sheetName := range f.GetSheetMap() {
		rows := f.GetRows(sheetName)
		for _, row := range rows {
			GradeList = append(GradeList, row[0])
			ClassList = append(ClassList, row[1])
			Numbers = append(Numbers, row[2])
			Names = append(Names, row[3])
			Emails = append(Emails, row[4])
			Clubs = append(Clubs, row[5])
			if sheetName == "教員" {
				Permissions = append(Permissions, 2)
			} else if sheetName != "教員" {
				Permissions = append(Permissions, 1)
			} else {
				Permissions = append(Permissions, 3)
			}
		}
	}
	for i, e := range Emails {
		if e == email {
			number, err := strconv.Atoi(Numbers[i])
			if err != nil {
				log.Errorf("Error converting number: %v", err)
				return 0, 500
			}
			_, err = db.Exec("INSERT INTO Users(uid, name, photoURL, GradeInSchool, ClassInSchool, email, SchoolClub, Number, Permission) VALUES (?,?,?,?,?,?,?,?,?)", uid, Names[i], photoUrl, GradeList[i], ClassList[i], email, Clubs[i], number, Permissions[i])
			if err != nil {
				log.Errorf("Error inserting user: %v", err)
				return 0, 500
			} else {
				return Permissions[i], 200
			}
		}
	}
	return 0, 404
}
