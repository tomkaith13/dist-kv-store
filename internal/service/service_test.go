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
	if rr.Code != http.StatusCreated {
		t.Fatalf("SET key failed. Expected: %d, got: %d", http.StatusCreated, rr.Code)
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
	if rr.Code != http.StatusCreated {
		t.Fatalf("SET key failed. Expected: %d, got: %d", http.StatusCreated, rr.Code)
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

func BenchmarkNoRaftWithOneNodesSetAndGetTake2(b *testing.B) {
	// Save the current log output
	originalLogOutput := log.Writer()

	// Redirect logs to ioutil.Discard to suppress them
	log.SetOutput(io.Discard)

	// Redirect stdout and stderr to ioutil.Discard
	devNull, _ := os.Open(os.DevNull)
	defer devNull.Close()
	os.Stdout = devNull
	os.Stderr = devNull

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
		MaxMapSize:   100,
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
	httpLeaderServer.Run()

	time.Sleep(2 * time.Second)

	b.ResetTimer()

	for i := 0; i < serviceLeaderConfig.MaxMapSize-1; i++ {
		body := SetRequestBody{
			Key: "a" + strconv.Itoa(i),
			Val: "b" + strconv.Itoa(i),
		}

		b, _ := json.Marshal(body)

		url := fmt.Sprintf("http://%s/key", sLeaderConfig.Address)
		// req, _ := http.NewRequest("POST", url, bytes.NewBuffer(b))

		http.Post(url, "application/json", bytes.NewBuffer([]byte(b)))
		// if err != nil {
		// 	fmt.Println("Error:", err)
		// 	return
		// }

		// rr := httptest.NewRecorder()
		// httpLeaderServer.GetRouter().ServeHTTP(rr, req)

		reqUrl := fmt.Sprintf("http://%s/key/a"+strconv.Itoa(i), sLeaderConfig.Address)
		http.Get(reqUrl)
		// getReq, _ := http.NewRequest("GET", reqUrl, nil)

		// rr = httptest.NewRecorder()
		// httpLeaderServer.GetRouter().ServeHTTP(rr, getReq)
	}
	b.StopTimer()
	log.SetOutput(originalLogOutput)

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
		KeyMaxLen:    1000000,
		ValMaxLen:    2000000,
		MaxMapSize:   100000,
		RaftNodeID:   "1",
		RaftAddr:     "localhost:25001",
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
