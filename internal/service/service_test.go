package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/tomkaith13/dist-kv-store/internal/router"
	"github.com/tomkaith13/dist-kv-store/internal/server"
)

func TestABasicSetAndGet(t *testing.T) {
	config := router.Config{
		RequestTimeout: 60 * time.Second,
	}

	sLeaderConfig := server.Config{
		Address:         "localhost:9999",
		ShutdownTimeout: time.Second * 5,
	}
	serviceLeaderConfig := Config{
		KeyMaxLen:    100,
		ValMaxLen:    200,
		MaxMapSize:   1000,
		RaftNodeID:   "1",
		RaftAddr:     "localhost:23001",
		RaftStoreDir: "./test-raft-dir",
		RaftLeader:   true,
		Debug:        true,
	}
	zlogger := zerolog.New(os.Stderr).
		Level(zerolog.DebugLevel).
		With().
		Timestamp().
		Logger()
	router := router.New(config, zlogger)
	kv_service := New(zlogger, serviceLeaderConfig)
	httpLeaderServer := server.New(zlogger, router.GetRouter(), sLeaderConfig, kv_service)

	httpLeaderServer.AddHandler(server.GET, "/key/{id}", GetHandler)
	httpLeaderServer.AddHandler(server.POST, "/key", SetHandler)

	body := SetRequestBody{
		Key: "a",
		Val: "b",
	}

	b, err := json.Marshal(body)
	if err != nil {
		t.Fatal("failed to marshal Set Request Body")
	}

	url := fmt.Sprintf("http://%s/key", sLeaderConfig.Address)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	httpLeaderServer.GetRouter().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("SET key failed. Expected: %d, got: %d", http.StatusOK, rr.Code)
	}

	reqUrl := fmt.Sprintf("http://%s/key/a", sLeaderConfig.Address)
	getReq, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		t.Fatal(err)
	}

	rr = httptest.NewRecorder()
	httpLeaderServer.GetRouter().ServeHTTP(rr, getReq)
	if rr.Code != http.StatusOK {
		t.Fatalf("SET key failed. Expected: %d, got: %d", http.StatusOK, rr.Code)
	}

	expectedResp := `{ "a" : "b" }`
	if rr.Body.String() != expectedResp {
		t.Fatalf("GET response failed. Expected: %s, got: %s", expectedResp, rr.Body.String())
	}

}
