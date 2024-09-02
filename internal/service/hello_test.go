package service

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/tomkaith13/dist-kv-store/internal/router"
	"github.com/tomkaith13/dist-kv-store/internal/server"
)

func TestHelloHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/hello", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	config := router.Config{
		RequestTimeout: 60 * time.Second,
	}
	sConfig := server.Config{
		Address:         "9999",
		ShutdownTimeout: time.Second * 5,
	}
	serviceConfig := Config{
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
	kv_service := New(zlogger, serviceConfig)
	httpServer := server.New(zlogger, router.GetRouter(), sConfig, kv_service)

	httpServer.AddHandler(server.GET, "/hello", HelloHandler)

	httpServer.GetRouter().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	// Check the response body
	expected := "world!"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestHelloLongHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/hello-long", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	config := router.Config{
		RequestTimeout: 2 * time.Second,
	}
	sConfig := server.Config{
		Address:         "9999",
		ShutdownTimeout: time.Second * 5,
	}
	serviceConfig := Config{
		RaftNodeID:   "1",
		RaftAddr:     "localhost:24001",
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
	kv_service := New(zlogger, serviceConfig)
	httpServer := server.New(zlogger, router.GetRouter(), sConfig, kv_service)

	httpServer.AddHandler(server.GET, "/hello-long", HelloHandlerLong)

	httpServer.GetRouter().ServeHTTP(rr, req)

	if rr.Code != http.StatusGatewayTimeout {
		t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusGatewayTimeout)
	}

}
