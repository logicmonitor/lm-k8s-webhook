package sync

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/pkg/logger"
	"go.uber.org/zap"
)

func TestTriggerReload(t *testing.T) {
	if err := logger.Init("DEBUG"); err != nil {
		t.Error("error occured while initializing the logger", zap.Error(err))
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/reload", func(rw http.ResponseWriter, r *http.Request) {
		if _, err := rw.Write([]byte("Reload success")); err != nil {
			t.Error("error in writing a response", zap.Error(err))
		}
	})
	mux.HandleFunc("/reload-error", func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusInternalServerError)
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	tests := []struct {
		name string
		args struct {
			reloadUrl  string
			httpClient *http.Client
		}
		wantErr bool
	}{
		{
			name: "test triggerReload",
			args: struct {
				reloadUrl  string
				httpClient *http.Client
			}{
				reloadUrl:  ts.URL + "/reload",
				httpClient: ts.Client(),
			},
			wantErr: false,
		},
		{
			name: "error returned by the target server",
			args: struct {
				reloadUrl  string
				httpClient *http.Client
			}{
				reloadUrl:  ts.URL + "/reload-error",
				httpClient: ts.Client(),
			},
			wantErr: true,
		},
	}
	for _, test := range tests {
		err := triggerReload(test.args.reloadUrl, test.args.httpClient)
		if err == nil && test.wantErr {
			t.Errorf("triggerReload() returned nil, instead of error")
		}
		if err != nil && !test.wantErr {
			t.Errorf("triggerReload() returned an unexpected error: %+v", err)
		}
	}
}
