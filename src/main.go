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

	firebase "firebase.google.com/go"
	"github.com/gorilla/mux"
	"github.com/natefinch/lumberjack"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type ApiConfig struct {
	host string
	port string
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
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
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
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if doc.Ref.ID == req.UID && doc.Data()["pass"] == req.Pass {
			log.Info(r.Method + " /signin　" + "200 " + "IP:" + r.RemoteAddr)
			a := map[string]string{"Status": "Success", "message": "pong"}
			json.NewEncoder(w).Encode(a)
			// SQLに登録する+JWTを返す
			InsertUser(req.UID, req.Email, req.PhotoUrl)
			return
		}
	}

	log.Info(r.Method + " /signin　" + "404 " + "IP:" + r.RemoteAddr)
	http.Error(w, "Account Not Found", http.StatusNotFound)
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
	SQLconfig, err = DBinit(AppDir + "/config/config.json")
	if err != nil {
		log.Errorf("Failed to initialize database: %v", err)
	}
	// ルーティング設定
	r := mux.NewRouter()
	r.HandleFunc("/ping", ping).Methods("GET")
	r.HandleFunc("/signin", signin).Methods("POST")

	fmt.Println("Server Config is ...")
	fmt.Println("Host:" + APIconfig.host)
	fmt.Println("Port:" + APIconfig.port + "\n")
	log.Info("Server starting...")
	if err := http.ListenAndServe(APIconfig.host+":"+APIconfig.port, r); err != nil {
		log.Errorf("Server failed to start: %v", err)
	}
}
