package sync

import (
	"testing"

	"github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/pkg/config"
)

func TestValidateResourceConfig(t *testing.T) {
	tests := []struct {
		name string
		args struct {
			resource config.Resource
		}
		wantErr bool
	}{
		{
			name: "ValidateResourceConfig for ConfigMap resource",
			args: struct {
				resource config.Resource
			}{
				resource: config.Resource{
					"kind":     ConfigMapResourceKind,
					"name":     "test-configmap",
					"fileName": "test-file",
				},
			},
			wantErr: false,
		},
		{
			name: "ValidateResourceConfig for MutatingWebhookConfiguration resource",
			args: struct {
				resource config.Resource
			}{
				resource: config.Resource{
					"kind": MutatingWebhookConfigurationKind,
					"name": "test-mutatingWebhookConfiguration",
				},
			},
			wantErr: false,
		},
		{
			name: "ValidateResourceConfig for invalid resource kind",
			args: struct {
				resource config.Resource
			}{
				resource: config.Resource{
					"kind": "invalid-resource",
					"name": "test-invalid-resource",
				},
			},
			wantErr: true,
		},
	}
	for _, test := range tests {
		err := ValidateResourceConfig(test.args.resource)
		if err == nil && test.wantErr {
			t.Errorf("ValidateResourceConfig() returned nil, instead of error")
		}
		if err != nil && !test.wantErr {
			t.Errorf("ValidateResourceConfig() returned an unexpected error: %+v", err)
		}
	}
}

func TestBuildAndValidateConfigmapResource(t *testing.T) {
	tests := []struct {
		name string
		args struct {
			resource config.Resource
		}
		wantErr bool
	}{
		{
			name: "ValidateResourceConfig for ConfigMap resource",
			args: struct {
				resource config.Resource
			}{
				resource: config.Resource{
					"kind":     ConfigMapResourceKind,
					"name":     "test-configmap",
					"fileName": "test-file",
				},
			},
			wantErr: false,
		},
	}
	for _, test := range tests {
		err := buildAndValidateConfigmapResource(test.args.resource)
		if err == nil && test.wantErr {
			t.Errorf("buildAndValidateConfigmapResource() returned nil, instead of error")
		}
		if err != nil && !test.wantErr {
			t.Errorf("buildAndValidateConfigmapResource() returned an unexpected error: %+v", err)
		}
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name   string
		fields struct {
			configMapResource K8sResource
		}
		wantErr bool
	}{
		{
			name: "Validate for ConfigMap resource",
			fields: struct {
				configMapResource K8sResource
			}{
				configMapResource: ConfigMapResource{
					Name:     "test-configmap",
					FileName: "test-file-name",
				},
			},
			wantErr: false,
		},
		{
			name: "Validate for ConfigMap resource with validation error in name field",
			fields: struct {
				configMapResource K8sResource
			}{
				configMapResource: ConfigMapResource{
					Name:     "",
					FileName: "test-file-name",
				},
			},
			wantErr: true,
		},
		{
			name: "Validate for ConfigMap resource with validation error in fileName field",
			fields: struct {
				configMapResource K8sResource
			}{
				configMapResource: ConfigMapResource{
					Name:     "test-configmap",
					FileName: "",
				},
			},
			wantErr: true,
		},
		{
			name: "Validate for MutatingWebhookConfiguration resource",
			fields: struct {
				configMapResource K8sResource
			}{
				configMapResource: MutatingWebhookConfigurationResource{
					Name: "test-mutatingWebhookConfiguration",
				},
			},
			wantErr: false,
		},
		{
			name: "Validate for MutatingWebhookConfiguration resource with validation error in name field",
			fields: struct {
				configMapResource K8sResource
			}{
				configMapResource: MutatingWebhookConfigurationResource{
					Name: "",
				},
			},
			wantErr: true,
		},
	}
	for _, test := range tests {
		err := test.fields.configMapResource.validate()
		if err == nil && test.wantErr {
			t.Errorf("validate() returned nil, instead of error")
		}
		if err != nil && !test.wantErr {
			t.Errorf("validate() returned an unexpected error: %+v", err)
		}
	}
}

func TestValidateReloadEndpoint(t *testing.T) {
	tests := []struct {
		name string
		args struct {
			endpoint string
		}
		wantErr bool
	}{
		{
			name: "Validate reload endpoint",
			args: struct {
				endpoint string
			}{
				endpoint: "http://localhost:3030",
			},
			wantErr: false,
		},
		{
			name: "Validate reload endpoint with invalid endpoint",
			args: struct {
				endpoint string
			}{
				endpoint: "1234",
			},
			wantErr: true,
		},
	}
	for _, test := range tests {
		err := ValidateReloadEndpoint(test.args.endpoint)
		if err == nil && test.wantErr {
			t.Errorf("ValidateReloadEndpoint() returned nil, instead of error")
		}
		if err != nil && !test.wantErr {
			t.Errorf("ValidateReloadEndpoint() returned an unexpected error: %+v", err)
		}
	}
}
