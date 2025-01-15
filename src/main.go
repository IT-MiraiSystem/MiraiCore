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
	"google.golang.org/api/iterator"
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

func TikokuShitade(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header missing", http.StatusUnauthorized)
		return
	}

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
		queryParams := r.URL.Query()
		TikokuZikan := queryParams.Get("latenessTime")
		if latenessSchool(claims["uid"].(string), TikokuZikan) != 200 {
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

func GoSchool(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
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
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type is not application/json", http.StatusUnsupportedMediaType)
		log.Errorf("%s", r.Method+" /signin　"+"415 "+"IP:"+r.RemoteAddr)
		return
	}

	var req SigninRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		log.Errorf("Invalid request payload: %v", err)
		return
	}

	ctx := context.Background()
	sa := option.WithCredentialsFile(AppDir + "/config/FirebaseConfig.json")
	config := &firebase.Config{ProjectID: "it-high-school-app"} // プロジェクトIDを指定
	app, err := firebase.NewApp(ctx, config, sa)
	if err != nil {
		log.Errorf("Failed to initialize Firebase app: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		log.Errorf("Failed to initialize Firestore client: %v", err)
		log.Errorf("%s", r.Method+" /signin　"+"500 "+"IP:"+r.RemoteAddr)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		a := map[string]string{"Status": "Failed", "message": "Internal Server Error"}
		json.NewEncoder(w).Encode(a)
		return
	}
	defer client.Close()

	iter := client.Collection("userPass").Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Errorf("Failed to iterate Firestore documents: %v", err)
			log.Errorf("%s", r.Method+" /signin　"+"500 "+"IP:"+r.RemoteAddr)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			a := map[string]string{"Status": "Failed", "message": "Internal Server Error"}
			json.NewEncoder(w).Encode(a)
			return
		}
		if doc.Ref.ID == req.UID && doc.Data()["pass"] == req.Pass {
			// SQLに登録する+JWTを返す
			if err := SearchUser(req.UID); err == 404 {
				if err := InsertUser(req.UID, req.Email, req.PhotoUrl); err != 200 {
					log.Errorf("%s", r.Method+" /signin　"+"500 "+"IP:"+r.RemoteAddr)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				} else {
					claims := jwt.MapClaims{
						"uid": req.UID,
						"exp": time.Now().Add(time.Hour * 1).Unix(),
					}
					token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
					accessToken, _ := token.SignedString([]byte("ACCESS_SECRET_KEY"))
					a := map[string]string{"token": accessToken}
					log.Info(r.Method + " /signin　" + "200 " + "IP:" + r.RemoteAddr)
					json.NewEncoder(w).Encode(a)
				}
			} else {
				claims := jwt.MapClaims{
					"uid": req.UID,
					"exp": time.Now().Add(time.Hour * 1).Unix(),
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				accessToken, _ := token.SignedString([]byte("ACCESS_SECRET_KEY"))
				a := map[string]string{"token": accessToken}
				log.Info(r.Method + " /signin　" + "200 " + "IP:" + r.RemoteAddr)
				json.NewEncoder(w).Encode(a)
			}
			return
		}
	}
	// 該当するユーザがいなかった場合
	// Firestoreに登録してJWTを返す
	userpass := generateRandomString(30)
	_, err = client.Collection("userPass").Doc(req.UID).Set(ctx, map[string]interface{}{
		"pass": userpass,
	})
	if err != nil {
		log.Errorf("Failed to set Firestore document: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if err := InsertUser(req.UID, req.Email, req.PhotoUrl); err != 200 {
		log.Errorf("%s", r.Method+" /signin　"+"500 "+"IP:"+r.RemoteAddr)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	} else {
		claims := jwt.MapClaims{
			"uid": req.UID,
			"exp": time.Now().Add(time.Hour * 1).Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		accessToken, _ := token.SignedString([]byte("ACCESS_SECRET_KEY"))
		a := map[string]string{"token": accessToken}
		log.Info(r.Method + " /signin　" + "200 " + "IP:" + r.RemoteAddr)
		json.NewEncoder(w).Encode(a)
	}
}

// CustomFormatter is a custom logrus formatter
type CustomFormatter struct {
	TimestampFormat string
}

func (f *CustomFormatter) Format(entry *log.Entry) ([]byte, error) {
	timestamp := entry.Time.Format(f.TimestampFormat)
	logMessage := fmt.Sprintf("%s [%s] %s\n", timestamp, entry.Level.String(), entry.Message)
	return []byte(logMessage), nil
}

func main() {
	var err error
	AppDir, err = filepath.Abs("..")
	if err != nil {
		log.Fatalf("Failed to get absolute path: %v", err)
	}

	// ログの出力先をファイルと標準出力に設定
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
	r.HandleFunc(APIconfig.location+"/ping", ping).Methods("GET")
	r.HandleFunc(APIconfig.location+"/signin", signin).Methods("POST")
	r.HandleFunc(APIconfig.location+"/GoSchool", GoSchool).Methods("GET")
	r.HandleFunc(APIconfig.location+"/TikokuShitade", TikokuShitade).Methods("GET")

	fmt.Println("Server Config is ...")
	fmt.Println("Host:" + APIconfig.host)
	fmt.Println("Port:" + APIconfig.port + "\n")
	log.Info("Server starting...")
	if err := http.ListenAndServe(APIconfig.host+":"+APIconfig.port, r); err != nil {
		log.Errorf("Server failed to start: %v", err)
	}
}
