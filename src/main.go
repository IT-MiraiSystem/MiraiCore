package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"
)

// Portは整数型のほうが最適だとは思うが、うまく読み込めないため断念
// まぁ、必要だったら変換を挟んでくれ給え
type ApiConfig struct {
	host string
	port string
}

var APIconfig ApiConfig
var SQLconfig DBconfig

func ping(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	a := map[string]string{"Status": "Success", "message": "pong"}
	json.NewEncoder(w).Encode(a)
}

func signin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	a := map[string]string{"Status": "Success", "message": "pong"}
	json.NewEncoder(w).Encode(a)
}

func main() {
	AppDir, _ := filepath.Abs("..")
	byteArray, _ := ioutil.ReadFile(AppDir + "/config/config.json")
	var config interface{}
	_ = json.Unmarshal(byteArray, &config)
	APIconfig.host, _ = config.(map[string]interface{})["api"].(map[string]interface{})["host"].(string)
	APIconfig.port, _ = config.(map[string]interface{})["api"].(map[string]interface{})["port"].(string)
	SQLconfig, err := DBinit(AppDir + "/config/config.json")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	r := mux.NewRouter()
	r.HandleFunc("/ping", ping).Methods("GET")
	r.HandleFunc("/signin", signin).Methods("POST")

	log.Println("Server starting...")
	log.Println("Server Config is ...")
	log.Println("	Host:" + APIconfig.host)
	log.Println("	Port:" + APIconfig.port)
	if err := http.ListenAndServe(APIconfig.host+":"+APIconfig.port, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
