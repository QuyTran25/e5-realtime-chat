package main

import (
	"encoding/json"
	"net/http"
)

// Cấu trúc dữ liệu bạn bè
type Friend struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Handler trả về danh sách bạn bè giả lập
func friendsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	friends := []map[string]interface{}{
        {"name": "Nguyễn Văn A", "avatar": "", "online": true},
        {"name": "Trần Thị B", "avatar": "", "online": false},
        {"name": "Lê Minh C", "avatar": "", "online": true},
	}
    json.NewEncoder(w).Encode(friends)
}
