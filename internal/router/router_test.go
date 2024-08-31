package router

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/tomkaith13/dist-kv-store/internal/service"
)

func TestRouter(t *testing.T) {
	config := Config{
		RequestTimeout: 1 * time.Second,
	}
	zlogger := zerolog.New(os.Stderr).
		Level(zerolog.DebugLevel).
		With().
		Timestamp().
		Logger()
	router := New(config, zlogger)
	router.AddHandler(GET, "/hello-long", service.HelloHandlerLong)
	// Create a test server using httptest
	ts := httptest.NewServer(router.GetRouter())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/hello-long")
	if err != nil {
		t.Fatal("unable to call hello-long")
	}
	if resp.StatusCode != http.StatusGatewayTimeout {
		t.Fatal("status code failed to match")
	}

}
