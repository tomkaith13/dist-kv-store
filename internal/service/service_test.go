package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
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

func TestRaftWithOneNodesSetAndGet(t *testing.T) {
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
		Debug:        false,
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

func BenchmarkNoRaftWithOneNodesSetAndGet(b *testing.B) {

	log.SetOutput(io.Discard)
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

	b.ResetTimer()

	for i := 0; i < 10; i++ {
		body := SetRequestBody{
			Key: "a" + strconv.Itoa(i),
			Val: "b" + strconv.Itoa(i),
		}

		b, _ := json.Marshal(body)

		url := fmt.Sprintf("http://%s/key", sLeaderConfig.Address)
		req, _ := http.NewRequest("POST", url, bytes.NewBuffer(b))

		rr := httptest.NewRecorder()
		httpLeaderServer.GetRouter().ServeHTTP(rr, req)

		reqUrl := fmt.Sprintf("http://%s/key/a"+strconv.Itoa(i), sLeaderConfig.Address)
		getReq, _ := http.NewRequest("GET", reqUrl, nil)

		rr = httptest.NewRecorder()
		httpLeaderServer.GetRouter().ServeHTTP(rr, getReq)
	}
	b.StopTimer()

}
