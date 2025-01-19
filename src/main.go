package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	firebase "firebase.google.com/go"
	"github.com/gorilla/mux"
	"github.com/natefinch/lumberjack"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/option"

	"github.com/dgrijalva/jwt-go"
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

func ping(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	a := map[string]string{"Status": "Success", "message": "pong"}
	log.Info(r.Method + " /ping　" + "200 " + "IP:" + r.RemoteAddr)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(a)
}

func LessonChange(w http.ResponseWriter, r *http.Request) {

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
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte("ACCESS_SECRET_KEY"), nil
	})
	if err != nil {
		log.Errorf("JWT parsing error: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"Status": "Failed", "message": "Unauthorized"})
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

	if classID == "" || startDate == "" || endDate == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"Status":  "Failed",
			"message": "Missing required parameters (ClassID, StartDate, EndDate)",
		})
		return
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

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte("ACCESS_SECRET_KEY"), nil
	})
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		log.Errorf("%s", r.Method+" /GoSchool　"+"401 "+"IP:"+r.RemoteAddr)
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
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
	fmt.Println(authHeader)
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
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"uid":        user.uid,
			"Permission": user.Permission,
			"exp":        time.Now().Add(time.Hour * 24).Unix(),
		})
		tokenString, err := token.SignedString([]byte("ACCESS_SECRET_KEY"))
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
		permission, statuscode := InsertUser(uid, email, photoURL)
		if statuscode == 200 {
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
				"uid":        uid,
				"Permission": permission,
				"exp":        time.Now().Add(time.Hour * 24).Unix(),
			})
			tokenString, err := token.SignedString([]byte("ACCESS_SECRET_KEY"))
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

	// ルーティング設定
	r := mux.NewRouter()
	r.HandleFunc(APIconfig.location+"/ping", ping).Methods("GET", "POST")
	r.HandleFunc(APIconfig.location+"/signin", signin).Methods("GET")
	r.HandleFunc(APIconfig.location+"/GoSchool", GoSchool).Methods("GET")
	r.HandleFunc(APIconfig.location+"/LessonDetails", LessonDetails).Methods("GET")

	fmt.Println("Server Config is ...")
	fmt.Println("Host:" + APIconfig.host)
	fmt.Println("Port:" + APIconfig.port + "\n")
	log.Info("Server starting...")
	if err := http.ListenAndServe(APIconfig.host+":"+APIconfig.port, r); err != nil {
		log.Errorf("Server failed to start: %v", err)
	}
}
