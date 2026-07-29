package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"wxcloudrun-golang/db"
	"wxcloudrun-golang/service"
)

func main() {
	// 先注册路由并立即监听端口，避免云托管冷启动时因为数据库连接阻塞而触发 102002。
	http.HandleFunc("/", service.IndexHandler)
	http.HandleFunc("/api/count", service.CounterHandler)
	http.HandleFunc("/wechat/login", service.WechatLoginHandler)
	http.HandleFunc("/healthz", healthHandler)

	go initDatabaseWithRetry()

	server := &http.Server{
		Addr:              ":80",
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Println("http server listening on :80")
	log.Fatal(server.ListenAndServe())
}

func initDatabaseWithRetry() {
	delay := time.Second

	for attempt := 1; ; attempt++ {
		startedAt := time.Now()
		err := db.Init()
		if err == nil {
			log.Printf("database ready, attempt=%d, cost=%s", attempt, time.Since(startedAt))
			return
		}

		log.Printf(
			"database init failed, attempt=%d, cost=%s, retryIn=%s, error=%v",
			attempt,
			time.Since(startedAt),
			delay,
			err,
		)

		time.Sleep(delay)
		if delay < 8*time.Second {
			delay *= 2
		}
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")

	if !db.IsReady() {
		w.Header().Set("Retry-After", "1")
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"ready":   false,
			"error":   "SERVICE_STARTING",
		})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"ready":   true,
	})
}
