package main

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	firebase "firebase.google.com/go"
	"github.com/gorilla/mux"
	"github.com/natefinch/lumberjack"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/option"

	"github.com/golang-jwt/jwt"
)

type ApiConfig struct {
	host     string
	port     string
	location string
}

type SigninRequest struct {
	UID      string `json:"uid"`
	Email    string `json:"email"`
	Pass     string `json:"pass"`
	PhotoUrl string `json:"photoUrl"`
}

type LessonChangeRequest struct {
	ClassID      string `json:"ClassID"`
	Lesson       string `json:"Lesson"`
	Room         string `json:"Room"`
	Teacher      string `json:"Teacher"`
	Date         string `json:"Date"`
	DayOfTheWeek string `json:"DayOfTheWeek"`
	LessonNumber int    `json:"LessonNumber"`
}

type IssueRegisterRequest struct {
	ClassID string `json:"ClassID"`
	Issue   string `json:"Issue"`
	Term    string `json:"term"`
	Lesson  string `json:"Lesson"`
}

type Subject struct {
	UID     string `json:"UID"`
	Subject string `json:"Subject"`
}

type CustomFormatter struct {
	TimestampFormat string
}

func (f *CustomFormatter) Format(entry *log.Entry) ([]byte, error) {
	timestamp := entry.Time.Format(f.TimestampFormat)
	logMessage := fmt.Sprintf("%s [%s] %s\n", timestamp, entry.Level.String(), entry.Message)
	return []byte(logMessage), nil
}

var AppDir string
var APIconfig ApiConfig
var SQLconfig DBconfig

var secretKey *rsa.PrivateKey
var publicKey *rsa.PublicKey

func ping(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	a := map[string]string{"Status": "Success", "message": "pong"}
	log.Info(r.Method + " /ping　" + "200 " + "IP:" + r.RemoteAddr)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(a)
}

// Admin権限は授業変更を調べるためのJWTを発行するだけ
func AdminGet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		log.Errorf("%s", r.Method+" /AdminGet　"+"405 "+"IP:"+r.RemoteAddr)
		a := map[string]string{"Status": "Failed", "message": "Method Not Allowed"}
		json.NewEncoder(w).Encode(a)
		return
	}
	var body map[string]string
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		log.Errorf("%s", r.Method+" /AdminGet　"+"400 "+"IP:"+r.RemoteAddr)
		a := map[string]string{"Status": "Failed", "message": "Bad Request"}
		json.NewEncoder(w).Encode(a)
		return
	}
	if body["username"] == "" || body["password"] == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		log.Errorf("%s", r.Method+" /AdminGet　"+"400 "+"IP:"+r.RemoteAddr)
		a := map[string]string{"Status": "Failed", "message": "Missing required parameters"}
		json.NewEncoder(w).Encode(a)
		return
	} else {
		if body["username"] == "admin" && body["password"] == "dswx{W_qOu^g~@Ik%6n8RoSPlp%A:p-F?v!r}^eRfrbxjhat&#" {
			tokenString, err := GenerateJWT("admin", 3, secretKey)
			if err != nil {
				log.Errorf("Failed to sign token: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				a := map[string]string{"Status": "Failed", "message": "Internal Server Error"}
				json.NewEncoder(w).Encode(a)
				return
			}
			w.WriteHeader(http.StatusOK)
			a := map[string]string{"token": tokenString}
			log.Infof("%s", r.Method+" /AdminGet　"+"200 "+"IP:"+r.RemoteAddr)
			json.NewEncoder(w).Encode(a)
		}

	}
}

func IssueRegister(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		log.Errorf("%s", r.Method+" /IssueRegister　"+"405 "+"IP:"+r.RemoteAddr)
		a := map[string]string{"Status": "Failed", "message": "Method Not Allowed"}
		json.NewEncoder(w).Encode(a)
		return
	}
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header missing", http.StatusUnauthorized)
		log.Errorf("%s", r.Method+" /IssueRegister　"+"401 "+"IP:"+r.RemoteAddr)
		a := map[string]string{"Status": "Failed", "message": "Authorization header missing"}
		json.NewEncoder(w).Encode(a)
		return
	}
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
		a := map[string]string{"Status": "Failed", "message": "Invalid Authorization header format"}
		json.NewEncoder(w).Encode(a)
		return
	}
	token, err := VerifyToken(tokenString, publicKey)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		log.Errorf("%s", r.Method+" /IssueRegister　"+"401 "+"IP:"+r.RemoteAddr)
		a := map[string]string{"Status": "Failed", "message": "Unauthorized"}
		json.NewEncoder(w).Encode(a)
		return
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if claims["permission"].(float64) < 2 {
			http.Error(w, "Forbidden", http.StatusForbidden)
			log.Errorf("%s", r.Method+" /IssueRegister　"+"403 "+"IP:"+r.RemoteAddr)
			a := map[string]string{"Status": "Failed", "message": "Forbidden"}
			json.NewEncoder(w).Encode(a)
			return
		} else {
			var issueRegister IssueRegisterRequest
			err := json.NewDecoder(r.Body).Decode(&issueRegister)
			if err != nil {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				log.Errorf("%s", r.Method+" /IssueRegister　"+"400 "+"IP:"+r.RemoteAddr)
				a := map[string]string{"Status": "Failed", "message": "Bad Request"}
				json.NewEncoder(w).Encode(a)
				return
			}
			if issueRegister.ClassID == "" || issueRegister.Issue == "" {
				http.Error(w, "Missing required parameters", http.StatusBadRequest)
				log.Errorf("%s", r.Method+" /IssueRegister　"+"400 "+"IP:"+r.RemoteAddr)
				a := map[string]string{"Status": "Failed", "message": "Missing required parameters"}
				json.NewEncoder(w).Encode(a)
				return
			}
			if InsertIssue(issueRegister.ClassID, issueRegister.Issue, issueRegister.Lesson, issueRegister.Term) != 200 {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				log.Errorf("%s", r.Method+" /IssueRegister　"+"500 "+"IP:"+r.RemoteAddr)
				a := map[string]string{"Status": "Failed", "message": "Internal Server Error"}
				json.NewEncoder(w).Encode(a)
				return
			}
			log.Infof("%s", r.Method+" /IssueRegister　"+"200 "+"IP:"+r.RemoteAddr)
			w.WriteHeader(http.StatusOK)
			a := map[string]string{"Status": "Success", "message": "Issue registered"}
			json.NewEncoder(w).Encode(a)
		}
	}
}

func InsertSubject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		log.Errorf("%s", r.Method+" /InsertSubject　"+"405 "+"IP:"+r.RemoteAddr)
		a := map[string]string{"Status": "Failed", "message": "Method Not Allowed"}
		json.NewEncoder(w).Encode(a)
		return
	}
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header missing", http.StatusUnauthorized)
		log.Errorf("%s", r.Method+" /InsertSubject　"+"401 "+"IP:"+r.RemoteAddr)
		a := map[string]string{"Status": "Failed", "message": "Authorization header missing"}
		json.NewEncoder(w).Encode(a)
		return
	}
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
		a := map[string]string{"Status": "Failed", "message": "Invalid Authorization header format"}
		json.NewEncoder(w).Encode(a)
		return
	}
	token, err := VerifyToken(tokenString, publicKey)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		log.Errorf("%s", r.Method+" /InsertSubject　"+"401 "+"IP:"+r.RemoteAddr)
		a := map[string]string{"Status": "Failed", "message": "Unauthorized"}
		json.NewEncoder(w).Encode(a)
		return
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if permission, ok := claims["permission"].(float64); ok && permission < 2 {
			http.Error(w, "Forbidden", http.StatusForbidden)
			log.Errorf("%s", r.Method+" /InsertSubject　"+"403 "+"IP:"+r.RemoteAddr)
			a := map[string]string{"Status": "Failed", "message": "Forbidden"}
			json.NewEncoder(w).Encode(a)
			return
		} else {
			var subjects []Subject
			err := json.NewDecoder(r.Body).Decode(&subjects)
			if err != nil {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				log.Errorf("%s", r.Method+" /InsertSubject　"+"400 "+"IP:"+r.RemoteAddr)
				a := map[string]string{"Status": "Failed", "message": "Bad Request"}
				json.NewEncoder(w).Encode(a)
				return
			}
			for _, subject := range subjects {
				if subject.UID == "" || subject.Subject == "" {
					http.Error(w, "Missing required parameters", http.StatusBadRequest)
					log.Errorf("%s", r.Method+" /InsertSubject　"+"400 "+"IP:"+r.RemoteAddr)
					a := map[string]string{"Status": "Failed", "message": "Missing required parameters"}
					json.NewEncoder(w).Encode(a)
					return
				}
				if InsertSubjectData(subject.UID, subject.Subject) != 200 {
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					log.Errorf("%s", r.Method+" /InsertSubject　"+"500 "+"IP:"+r.RemoteAddr)
					a := map[string]string{"Status": "Failed", "message": "Internal Server Error"}
					json.NewEncoder(w).Encode(a)
					return
				}
				log.Infof("%s", r.Method+" /InsertSubject　"+"200 "+"IP:"+r.RemoteAddr)
				w.WriteHeader(http.StatusOK)
				a := map[string]string{"Status": "Success", "message": "Subject inserted"}
				json.NewEncoder(w).Encode(a)
			}
		}
	} else {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		log.Errorf("%s", r.Method+" /InsertSubject　"+"401 "+"IP:"+r.RemoteAddr)
		a := map[string]string{"Status": "Failed", "message": "Unauthorized"}
		json.NewEncoder(w).Encode(a)
	}
}

func userList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		log.Errorf("%s", r.Method+" /userList　"+"405 "+"IP:"+r.RemoteAddr)
		a := map[string]string{"Status": "Failed", "message": "Method Not Allowed"}
		json.NewEncoder(w).Encode(a)
		return
	}
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header missing", http.StatusUnauthorized)
		log.Errorf("%s", r.Method+" /userList　"+"401 "+"IP:"+r.RemoteAddr)
		a := map[string]string{"Status": "Failed", "message": "Authorization header missing"}
		json.NewEncoder(w).Encode(a)
		return
	}
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
		a := map[string]string{"Status": "Failed", "message": "Invalid Authorization header format"}
		json.NewEncoder(w).Encode(a)
		return
	}
	token, err := VerifyToken(tokenString, publicKey)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		log.Errorf("%s", r.Method+" /userList　"+"401 "+"IP:"+r.RemoteAddr)
		a := map[string]string{"Status": "Failed", "message": "Unauthorized"}
		json.NewEncoder(w).Encode(a)
		return
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if permission, ok := claims["permission"].(float64); ok && permission > 2 {
			http.Error(w, "Forbidden", http.StatusForbidden)
			log.Errorf("%s", r.Method+" /userList　"+"403 "+"IP:"+r.RemoteAddr)
			a := map[string]string{"Status": "Failed", "message": "Forbidden"}
			json.NewEncoder(w).Encode(a)
			return
		}
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			log.Errorf("%s", r.Method+" /userList　"+"401 "+"IP:"+r.RemoteAddr)
			a := map[string]string{"Status": "Failed", "message": "Unauthorized"}
			json.NewEncoder(w).Encode(a)
			return
		} else {
			err, users := UserList()
			if err != 200 {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				log.Errorf("%s", r.Method+" /userList　"+"500 "+"IP:"+r.RemoteAddr)
				a := map[string]string{"Status": "Failed", "message": "Internal Server Error"}
				json.NewEncoder(w).Encode(a)
				return
			}
			log.Infof("%s", r.Method+" /userList　"+"200 "+"IP:"+r.RemoteAddr)
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(users)
		}
	} else {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		log.Errorf("%s", r.Method+" /userList　"+"401 "+"IP:"+r.RemoteAddr)
		a := map[string]string{"Status": "Failed", "message": "Unauthorized"}
		json.NewEncoder(w).Encode(a)
	}
}

func LessonChange(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		log.Errorf("%s", r.Method+" /LessonChange　"+"405 "+"IP:"+r.RemoteAddr)
		a := map[string]string{"Status": "Failed", "message": "Method Not Allowed"}
		json.NewEncoder(w).Encode(a)
		return
	}
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header missing", http.StatusUnauthorized)
		log.Errorf("%s", r.Method+" /LessonChange　"+"401 "+"IP:"+r.RemoteAddr)
		a := map[string]string{"Status": "Failed", "message": "Authorization header missing"}
		json.NewEncoder(w).Encode(a)
		return
	}
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
		a := map[string]string{"Status": "Failed", "message": "Invalid Authorization header format"}
		json.NewEncoder(w).Encode(a)
		return
	}
	token, err := VerifyToken(tokenString, publicKey)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		log.Errorf("%s", r.Method+" /LessonChange　"+"401 "+"IP:"+r.RemoteAddr)
		a := map[string]string{"Status": "Failed", "message": "Unauthorized"}
		json.NewEncoder(w).Encode(a)
		return
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if permission, ok := claims["permission"].(float64); ok && permission > 2 {
			http.Error(w, "Forbidden", http.StatusForbidden)
			log.Errorf("%s", r.Method+" /LessonChange　"+"403 "+"IP:"+r.RemoteAddr)
			a := map[string]string{"Status": "Failed", "message": "Forbidden"}
			json.NewEncoder(w).Encode(a)
			return
		}
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			log.Errorf("%s", r.Method+" /LessonChange　"+"401 "+"IP:"+r.RemoteAddr)
			a := map[string]string{"Status": "Failed", "message": "Unauthorized"}
			json.NewEncoder(w).Encode(a)
			return
		} else {
			var lessonChanges []LessonChangeRequest
			err := json.NewDecoder(r.Body).Decode(&lessonChanges)
			if err != nil {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				log.Errorf("%s", r.Method+" /LessonChange　"+"400 "+"IP:"+r.RemoteAddr)
				a := map[string]string{"Status": "Failed", "message": "Bad Request"}
				json.NewEncoder(w).Encode(a)
				return
			}
			for _, lessonChange := range lessonChanges {
				if lessonChange.ClassID == "" || lessonChange.Lesson == "" || lessonChange.Room == "" || lessonChange.Teacher == "" || lessonChange.Date == "" || lessonChange.DayOfTheWeek == "" || lessonChange.LessonNumber == 0 {
					http.Error(w, "Missing required parameters", http.StatusBadRequest)
					log.Errorf("%s", r.Method+" /LessonChange　"+"400 "+"IP:"+r.RemoteAddr)
					a := map[string]string{"Status": "Failed", "message": "Missing required parameters"}
					json.NewEncoder(w).Encode(a)
					return
				}
				lessonNumber := lessonChange.LessonNumber
				if UpdateLesson(lessonChange.ClassID, lessonChange.DayOfTheWeek, lessonNumber, lessonChange.Lesson, lessonChange.Room, lessonChange.Teacher, lessonChange.Date) != 200 {
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					log.Errorf("%s", r.Method+" /LessonChange　"+"500 "+"IP:"+r.RemoteAddr)
					a := map[string]string{"Status": "Failed", "message": "Internal Server Error"}
					json.NewEncoder(w).Encode(a)
				} else {
					log.Infof("%s", r.Method+" /LessonChange　"+"200 "+"IP:"+r.RemoteAddr)
					a := map[string]string{"Status": "Success", "message": "Lesson changes processed"}
					json.NewEncoder(w).Encode(a)
					return
				}
			}
			w.WriteHeader(http.StatusOK)
			a := map[string]string{"Status": "Success", "message": "Lesson changes processed"}
			json.NewEncoder(w).Encode(a)
		}

	} else {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		log.Errorf("%s", r.Method+" /LessonChange　"+"401 "+"IP:"+r.RemoteAddr)
		a := map[string]string{"Status": "Failed", "message": "Unauthorized"}
		json.NewEncoder(w).Encode(a)
	}
}

func LessonDetails(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		log.Errorf("%s", r.Method+" /LessonDetails　"+"405 "+"IP:"+r.RemoteAddr)
		a := map[string]string{"Status": "Failed", "message": "Method Not Allowed"}
		json.NewEncoder(w).Encode(a)
		return
	}
	// Authorizationヘッダーの検証
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header missing", http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"Status": "Failed", "message": "Authorization header missing"})
		return
	}
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"Status": "Failed", "message": "Invalid Authorization header format"})
		return
	}

	// JWTトークンの検証
	token, err := VerifyToken(tokenString, publicKey)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"Status": "Failed", "message": "Unauthorized"})
		a := map[string]string{"Status": "Failed", "message": "Unauthorized"}
		json.NewEncoder(w).Encode(a)
		return
	}

	// トークンのクレームを検証
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"Status": "Failed", "message": "Unauthorized"})
		return
	}

	// クエリパラメータの検証
	Param := r.URL.Query()
	classID := Param.Get("ClassID")
	startDate := Param.Get("StartDate")
	endDate := Param.Get("EndDate")

	if startDate == "" || endDate == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"Status":  "Failed",
			"message": "Missing required parameters (ClassID, StartDate, EndDate)",
		})
		return
	}
	if classID == "" {
		classID = claims["uid"].(string)
	}

	// レッスンデータの取得
	log.Infof("Access FROM %s", claims["uid"].(string))
	statusCode, lesson := GetLesson(classID, startDate, endDate)
	if statusCode != http.StatusOK {
		log.Errorf("Error retrieving lesson data: %v", lesson)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"Status": "Failed", "message": "Internal Server Error"})
		return
	}

	// レスポンスの生成
	if len(lesson) == 0 {
		http.Error(w, "Not Found", http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"Status": "Failed", "message": "Not Found"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(lesson)
}

func GoSchool(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		log.Errorf("%s", r.Method+" /GoSchool　"+"405 "+"IP:"+r.RemoteAddr)
		a := map[string]string{"Status": "Failed", "message": "Method Not Allowed"}
		json.NewEncoder(w).Encode(a)
		return
	}
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header missing", http.StatusUnauthorized)
		return
	}

	// "Bearer " プレフィックスを削除
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
		return
	}

	token, err := VerifyToken(tokenString, publicKey)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		log.Errorf("%s", r.Method+" /GoSchool　"+"401 "+"IP:"+r.RemoteAddr)
		a := map[string]string{"Status": "Failed", "message": "Unauthorized"}
		json.NewEncoder(w).Encode(a)
		return
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if AttendSchool(claims["uid"].(string)) != 200 {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Errorf("%s", r.Method+" /GoSchool　"+"500 "+"IP:"+r.RemoteAddr)
			return
		}
		log.Info(r.Method + " /GoSchool　" + "200 " + "IP:" + r.RemoteAddr)
		w.WriteHeader(http.StatusOK)
		a := map[string]string{"Status": "Success", "message": "GoSchool"}
		json.NewEncoder(w).Encode(a)
	} else {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		log.Errorf("%s", r.Method+" /GoSchool　"+"401 "+"IP:"+r.RemoteAddr)
	}
}

func signin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		log.Errorf("%s", r.Method+" /signin　"+"405 "+"IP:"+r.RemoteAddr)
		a := map[string]string{"Status": "Failed", "message": "Method Not Allowed"}
		json.NewEncoder(w).Encode(a)
		return
	}
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header missing", http.StatusUnauthorized)
		log.Errorf("%s", r.Method+" /signin　"+"401 "+"IP:"+r.RemoteAddr)
		a := map[string]string{"Status": "Failed", "message": "Authorization header missing"}
		json.NewEncoder(w).Encode(a)
		return
	}
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
		a := map[string]string{"Status": "Failed", "message": "Invalid Authorization header format"}
		json.NewEncoder(w).Encode(a)
		return
	}
	opt := option.WithCredentialsFile(AppDir + "/config/FirebaseConfig.json")
	config := &firebase.Config{ProjectID: "it-high-school-app"}
	app, err := firebase.NewApp(context.Background(), config, opt)
	if err != nil {
		log.Errorf("Failed to create Firebase app: %v", err)
		log.Errorf("%s", r.Method+" /signin　"+"500 "+"IP:"+r.RemoteAddr)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		a := map[string]string{"Status": "Failed", "message": "Internal Server Error"}
		json.NewEncoder(w).Encode(a)
		return
	}
	client, err := app.Auth(context.Background())
	if err != nil {
		log.Errorf("Failed to create Firebase client: %v", err)
		log.Errorf("%s", r.Method+" /signin　"+"500 "+"IP:"+r.RemoteAddr)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		a := map[string]string{"Status": "Failed", "message": "Internal Server Error"}
		json.NewEncoder(w).Encode(a)
		return
	}
	token, err := client.VerifyIDToken(context.Background(), tokenString)
	if err != nil {
		log.Errorf("Failed to verify ID token: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		a := map[string]string{"Status": "Failed", "message": "Unauthorized"}
		json.NewEncoder(w).Encode(a)
		return
	}
	uid := token.UID

	userRecord, err := client.GetUser(context.Background(), uid)
	if err != nil {
		log.Errorf("Failed to get user data: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		a := map[string]string{"Status": "Failed", "message": "Internal Server Error"}
		json.NewEncoder(w).Encode(a)
		return
	}

	email := userRecord.Email
	photoURL := userRecord.PhotoURL

	user := UserInfo(uid)
	if user != (User{}) {
		tokenString, err := GenerateJWT(user.UID, user.Permission, secretKey)
		if err != nil {
			log.Errorf("Failed to sign token: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			a := map[string]string{"Status": "Failed", "message": "Internal Server Error"}
			json.NewEncoder(w).Encode(a)
			return
		} else {
			w.WriteHeader(http.StatusOK)
			a := map[string]string{"token": tokenString}
			json.NewEncoder(w).Encode(a)
			return
		}
	} else {
		permission, statuscode := InsertUser(uid, email, photoURL)
		if statuscode == 200 {
			tokenString, err := GenerateJWT(uid, permission, secretKey)
			if err != nil {
				log.Errorf("Failed to sign token: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				a := map[string]string{"Status": "Failed", "message": "Internal Server Error"}
				json.NewEncoder(w).Encode(a)
				return
			}
			w.WriteHeader(http.StatusOK)
			a := map[string]string{"token": tokenString}
			json.NewEncoder(w).Encode(a)
		} else {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Errorf("%s /signin 500 IP:%s", r.Method, r.RemoteAddr)
			a := map[string]string{"Status": "Failed", "message": "Internal Server Error"}
			json.NewEncoder(w).Encode(a)
		}
	}
}

func main() {
	var err error
	AppDir, err = filepath.Abs("..")
	if err != nil {
		log.Fatalf("Failed to get absolute path: %v", err)
	}

	logFile := &lumberjack.Logger{
		Filename:   AppDir + "/log/MiraiCore-API.log",
		MaxSize:    500,
		MaxBackups: 20,
		MaxAge:     0,
	}
	if err := os.Chmod(AppDir+"/log/MiraiCore-API.log", 0777); err != nil {
		log.Errorf("Failed to set permissions on log file: %v", err)
	}
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))
	log.SetFormatter(&CustomFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})

	byteArray, _ := ioutil.ReadFile(AppDir + "/config/config.json")
	var config interface{}
	_ = json.Unmarshal(byteArray, &config)
	APIconfig.host, _ = config.(map[string]interface{})["api"].(map[string]interface{})["host"].(string)
	APIconfig.port, _ = config.(map[string]interface{})["api"].(map[string]interface{})["port"].(string)
	APIconfig.location, _ = config.(map[string]interface{})["api"].(map[string]interface{})["location"].(string)
	SQLconfig, err = DBinit(AppDir + "/config/config.json")
	if err != nil {
		log.Errorf("Failed to initialize database: %v", err)
	}
	// 鍵の設定
	rawSercretKey, err := ioutil.ReadFile(AppDir + "/certification/secret.key")
	if err != nil {
		log.Errorf("Failed to read secret key: %v", err)
		panic(err)
	}
	rawPublicKey, err := ioutil.ReadFile(AppDir + "/certification/publickey.pem")
	if err != nil {
		log.Errorf("Failed to read public key: %v", err)
		panic(err)
	}

	secretKey, publicKey, err = ParseKeys(rawSercretKey, rawPublicKey)
	if err != nil {
		log.Errorf("Failed to parse keys: %v", err)
		panic(err)
	}

	// ルーティング設定
	r := mux.NewRouter()
	r.HandleFunc(APIconfig.location+"/ping", ping).Methods("GET", "POST")
	r.HandleFunc(APIconfig.location+"/signin", signin).Methods("GET")
	r.HandleFunc(APIconfig.location+"/GoSchool", GoSchool).Methods("GET")
	r.HandleFunc(APIconfig.location+"/LessonDetails", LessonDetails).Methods("GET")
	r.HandleFunc(APIconfig.location+"/userList", userList).Methods("GET")
	r.HandleFunc(APIconfig.location+"/LessonChange", LessonChange).Methods("POST")
	r.HandleFunc(APIconfig.location+"/AdminGet", AdminGet).Methods("POST")
	r.HandleFunc(APIconfig.location+"/InsertSubject", InsertSubject).Methods("POST")
	fmt.Println("Server Config is ...")
	fmt.Println("Host:" + APIconfig.host)
	fmt.Println("Port:" + APIconfig.port + "\n")
	log.Info("Server starting...")
	if err := http.ListenAndServe(APIconfig.host+":"+APIconfig.port, r); err != nil {
		log.Errorf("Server failed to start: %v", err)
	}
}
