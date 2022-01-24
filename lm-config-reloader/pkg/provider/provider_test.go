package provider

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestGetParsedPullInterval(t *testing.T) {
	tests := []struct {
		name   string
		fields struct {
			DefaultRemoteProvider
		}
		wantErr     bool
		wantPayload struct {
			parsedDuration time.Duration
			err            error
		}
	}{
		{
			name: "test GetParsedPullInterval",
			fields: struct {
				DefaultRemoteProvider
			}{
				DefaultRemoteProvider: DefaultRemoteProvider{
					Provider:     "test",
					PullInterval: "5s",
				},
			},
			wantErr: false,
			wantPayload: struct {
				parsedDuration time.Duration
				err            error
			}{
				parsedDuration: 5 * time.Second,
				err:            nil,
			},
		},
		{
			name: "test GetParsedPullInterval for empty interval",
			fields: struct {
				DefaultRemoteProvider
			}{
				DefaultRemoteProvider: DefaultRemoteProvider{
					Provider:     "test",
					PullInterval: "",
				},
			},
			wantErr: false,
			wantPayload: struct {
				parsedDuration time.Duration
				err            error
			}{
				parsedDuration: 0,
				err:            nil,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parsedDuration, err := test.fields.DefaultRemoteProvider.GetParsedPullInterval()
			if err == nil && test.wantErr {
				t.Errorf("GetParsedPullInterval() returned nil, instead of error")
			}
			if err != nil && !test.wantErr {
				t.Errorf("GetParsedPullInterval() returned an unexpected error: %+v", err)
			}
			if !cmp.Equal(test.wantPayload.parsedDuration, parsedDuration) || !cmp.Equal(test.wantPayload.err, err) {
				t.Errorf("expected response: parsedDuration= %+v and err= %+v, but got: parsedDuration= %+v and err= %v", test.wantPayload.parsedDuration, test.wantPayload.err, parsedDuration, err)
			}
		})
	}
}

func TestSetPullInterval(t *testing.T) {
	tests := []struct {
		name string
		args struct {
			pullInterval string
		}
		fields struct {
			DefaultRemoteProvider
		}
		wantPayload struct {
			DefaultRemoteProvider
		}
	}{
		{
			name: "test SetPullInterval",
			args: struct {
				pullInterval string
			}{
				pullInterval: "5s",
			},
			fields: struct {
				DefaultRemoteProvider
			}{
				DefaultRemoteProvider: DefaultRemoteProvider{
					Provider: "test",
				},
			},
			wantPayload: struct {
				DefaultRemoteProvider
			}{
				DefaultRemoteProvider: DefaultRemoteProvider{
					Provider:     "test",
					PullInterval: "5s",
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.fields.DefaultRemoteProvider.SetPullInterval(test.args.pullInterval)
			if !cmp.Equal(test.wantPayload.DefaultRemoteProvider, test.fields.DefaultRemoteProvider) {
				t.Errorf("expected response: DefaultRemoteProvider= %+v, but got: DefaultRemoteProvider= %+v", test.wantPayload.DefaultRemoteProvider, test.fields.DefaultRemoteProvider)
			}
		})
	}
}

func TestGetPullInterval(t *testing.T) {
	tests := []struct {
		name   string
		fields struct {
			DefaultRemoteProvider
		}
		wantPayload struct {
			pullInterval string
		}
	}{
		{
			name: "test GetPullInterval",
			fields: struct {
				DefaultRemoteProvider
			}{
				DefaultRemoteProvider: DefaultRemoteProvider{
					Provider:     "test",
					PullInterval: "5s",
				},
			},
			wantPayload: struct {
				pullInterval string
			}{
				pullInterval: "5s",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pullInterval := test.fields.DefaultRemoteProvider.GetPullInterval()
			if !cmp.Equal(test.wantPayload.pullInterval, pullInterval) {
				t.Errorf("expected response: pullInterval= %+v, but got: pullInterval= %+v", test.wantPayload.pullInterval, pullInterval)
			}
		})
	}
}
