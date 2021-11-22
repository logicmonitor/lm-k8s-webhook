package handler

import (
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/logicmonitor/lm-k8s-webhook/pkg/config"
	"github.com/logicmonitor/lm-k8s-webhook/pkg/mutation"

	"gomodules.xyz/jsonpatch/v2"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	testclient "k8s.io/client-go/kubernetes/fake"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var logger = logf.Log.WithName("unit-tests")

// getFakeK8sClient returns the dummy kubernetes client object for testing
func getFakeK8sClient() (*config.K8sClient, error) {
	return config.NewK8sClient(nil, func(r *rest.Config) (kubernetes.Interface, error) {
		return testclient.NewSimpleClientset(), nil
	})
}

func TestHandle(t *testing.T) {

	k8sClient, err := getFakeK8sClient()
	if err != nil {
		t.Errorf("Error occured in getting fake k8s client: %v", err)
		return
	}
	os.Setenv("CLUSTER_NAME", "default")
	defer os.Unsetenv("CLUSTER_NAME")

	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	decoder, err := admission.NewDecoder(scheme)

	if err != nil {
		t.Errorf("Error occured in getting decoder: %v", err)
		return
	}

	tests := []struct {
		name                 string
		LMPodMutationHandler *LMPodMutationHandler
		args                 struct {
			ctx context.Context
			req admission.Request
		}
		wantPayload struct {
			Response admission.Response
		}
	}{
		{
			name: "Handle the admission request",
			LMPodMutationHandler: &LMPodMutationHandler{
				Client: k8sClient,
				// LMConfig: nil,
				Log:     logger,
				decoder: decoder,
			},
			args: struct {
				ctx context.Context
				req admission.Request
			}{
				context.Background(),
				admission.Request{
					AdmissionRequest: admissionv1.AdmissionRequest{
						UID:                "78e13294-bb55-41e4-8b01-8ef459f496f7",
						Kind:               v1.GroupVersionKind{Version: "v1", Kind: "Pod"},
						Resource:           v1.GroupVersionResource{Version: "v1", Resource: "pods"},
						SubResource:        "",
						RequestKind:        &v1.GroupVersionKind{Version: "v1", Kind: "Pod"},
						RequestResource:    &v1.GroupVersionResource{Version: "v1", Resource: "pods"},
						RequestSubResource: "",
						Name:               "",
						Namespace:          "default",
						Operation:          admissionv1.Create,
						Object: runtime.RawExtension{Raw: []byte(
							`{
								"apiVersion": "v1",
								"kind": "Pod",
								"metadata": {
									"name": "foo",
									"namespace": "default"
								},
								"spec": {
									"containers": [
										{
											"image": "bar:v2",
											"name": "bar"
										}
									]
								}
							}`)},
					},
				},
			},
			wantPayload: struct {
				Response admission.Response
			}{
				Response: admission.Response{
					Patches:           []jsonpatch.Operation{{Operation: "add"}},
					AdmissionResponse: admissionv1.AdmissionResponse{Allowed: true},
				},
			},
		},
	}

	for _, tt := range tests {
		resp := tt.LMPodMutationHandler.Handle(context.Background(), tt.args.req)

		if resp.AdmissionResponse.Allowed != tt.wantPayload.Response.Allowed {
			t.Errorf("Handle() returned AdmissionResponse.Allowed = %v, but expected AdmissionResponse.Allowed = %v", resp.AdmissionResponse.Allowed, tt.wantPayload.Response.Allowed)
			return
		}

		if len(resp.Patches) != 0 && (resp.Patches[0].Operation != tt.wantPayload.Response.Patches[0].Operation) {
			t.Errorf("Handle() returned Patch with operation = %v, but expected operation = %v", resp.Patches[0].Operation, tt.wantPayload.Response.Patches[0].Operation)
			return
		}
	}
}

func TestInjectDecoder(t *testing.T) {

	k8sClient, err := getFakeK8sClient()
	if err != nil {
		t.Errorf("Error occured in getting fake k8s client: %v", err)
		return
	}

	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	decoder, err := admission.NewDecoder(scheme)

	if err != nil {
		t.Errorf("Error occured in getting decoder: %v", err)
		return
	}

	tests := []struct {
		name                 string
		LMPodMutationHandler *LMPodMutationHandler
		args                 struct {
			decoder *admission.Decoder
		}
		wantErr bool
	}{
		{
			name: "Inject Decoder",
			LMPodMutationHandler: &LMPodMutationHandler{
				Client: k8sClient,
				// LMConfig: nil,
				Log:     logger,
				decoder: decoder,
			},
			args: struct{ decoder *admission.Decoder }{
				decoder: decoder,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		err := tt.LMPodMutationHandler.InjectDecoder(tt.args.decoder)

		if (err != nil) != tt.wantErr {
			t.Errorf("InjectDecoder() error = %v, wantErr %v", err, tt.wantErr)
			return
		}
	}
}

func TestNewParams(t *testing.T) {
	test := struct {
		name string
		args struct {
			pod             corev1.Pod
			mutationHandler *LMPodMutationHandler
			namespace       string
		}
		expectedPayload *mutation.Params
	}{
		name: "New Params",
		args: struct {
			pod             corev1.Pod
			mutationHandler *LMPodMutationHandler
			namespace       string
		}{
			pod: corev1.Pod{ObjectMeta: v1.ObjectMeta{Name: "demo"}},
			mutationHandler: &LMPodMutationHandler{
				Client: &config.K8sClient{},
				Log:    logger,
				// LMConfig: &config.Config{},
			},
			namespace: "default",
		},
		expectedPayload: &mutation.Params{
			Client:    &config.K8sClient{},
			Pod:       &corev1.Pod{ObjectMeta: v1.ObjectMeta{Name: "demo"}},
			LMConfig:  config.GetConfig(),
			Mutations: mutation.Mutations,
			Namespace: "default",
			Log:       logger,
		},
	}

	muatationParams := NewParams(&test.args.pod, test.args.mutationHandler, test.args.namespace)

	if !reflect.DeepEqual(muatationParams, test.expectedPayload) {
		t.Errorf("NewParams() returned = %v, expected %v", muatationParams, test.expectedPayload)
		return
	}
}
