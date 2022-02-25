package sync

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/pkg/config"
	"github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/pkg/fetcher"
	"github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/pkg/logger"
	"go.uber.org/zap"

	"github.com/google/go-cmp/cmp"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	testclient "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

// getFakeK8sClient returns the dummy kubernetes client object for testing
func getFakeK8sClientForSync() (*config.K8sClient, error) {
	return config.NewK8sClient(nil, func(r *rest.Config) (kubernetes.Interface, error) {
		return testclient.NewSimpleClientset(
			&corev1.ConfigMap{ObjectMeta: v1.ObjectMeta{Name: "test-configmap", Namespace: "test"}, Data: map[string]string{"test-file": "test-data"}},
			&corev1.Pod{ObjectMeta: v1.ObjectMeta{Name: "test-pod", Namespace: "test"}},
			&admissionregistrationv1.MutatingWebhookConfiguration{
				ObjectMeta: v1.ObjectMeta{Name: "test-mutatingWebhookConfiguration"},
				Webhooks: []admissionregistrationv1.MutatingWebhook{
					{
						Name: "test-webhook",
						ObjectSelector: &v1.LabelSelector{
							MatchLabels: map[string]string{"foo": "bar"},
						},
						NamespaceSelector: &v1.LabelSelector{
							MatchExpressions: []v1.LabelSelectorRequirement{
								{Key: "environment", Operator: "In", Values: []string{"dev", "staging"}},
							},
						},
					},
				},
			}), nil
	})
}

func TestSync(t *testing.T) {
	if err := logger.Init("DEBUG"); err != nil {
		t.Error("error occured while initializing the logger", zap.Error(err))
	}
	os.Setenv("POD_NAMESPACE", "test")
	os.Setenv("POD_NAME", "test-pod")
	defer os.Unsetenv("POD_NAMESPACE")
	defer os.Unsetenv("POD_NAME")
	k8sClient, err := getFakeK8sClientForSync()
	if err != nil {
		t.Errorf("Error occured in getting fake k8s client: %v", err)
		return
	}

	// start test server
	mux := http.NewServeMux()
	mux.HandleFunc("/reload", func(rw http.ResponseWriter, r *http.Request) {
		if _, err := rw.Write([]byte("Reload success")); err != nil {
			t.Error("error in writing a response", zap.Error(err))
		}
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	tests := []struct {
		name string
		args struct {
			fetcherResponse *fetcher.Response
		}
		fields struct {
			configSyncer ConfigSyncer
		}
		wantErr     bool
		wantPayload struct {
			configMap                    corev1.ConfigMap
			mutatingWebhookConfiguration admissionregistrationv1.MutatingWebhookConfiguration
			pod                          corev1.Pod
		}
	}{
		{
			name: "Sync for configmap with no config change found",
			args: struct {
				fetcherResponse *fetcher.Response
			}{
				fetcherResponse: &fetcher.Response{
					FileName: "test-file",
					FileData: []byte("test-data"),
				},
			},
			fields: struct {
				configSyncer ConfigSyncer
			}{
				configSyncer: configMapConfigSyncer{
					Resource: ConfigMapResource{
						Name:     "test-configmap",
						FileName: "test-file",
					},
					k8sClient: k8sClient,
				},
			},
			wantErr: false,
			wantPayload: struct {
				configMap                    corev1.ConfigMap
				mutatingWebhookConfiguration admissionregistrationv1.MutatingWebhookConfiguration
				pod                          corev1.Pod
			}{
				configMap: corev1.ConfigMap{
					ObjectMeta: v1.ObjectMeta{Name: "test-configmap", Namespace: "test"},
					Data:       map[string]string{"test-file": "test-data"},
				},
				pod: corev1.Pod{
					ObjectMeta: v1.ObjectMeta{
						Name: "test-pod", Namespace: "test",
					},
				},
			},
		},
		{
			name: "Sync for configmap with change found",
			args: struct {
				fetcherResponse *fetcher.Response
			}{
				fetcherResponse: &fetcher.Response{
					FileName: "test-file",
					FileData: []byte("updated-test-data"),
				},
			},
			fields: struct {
				configSyncer ConfigSyncer
			}{
				configSyncer: configMapConfigSyncer{
					Resource: ConfigMapResource{
						Name:     "test-configmap",
						FileName: "test-file",
					},
					k8sClient: k8sClient,
				},
			},
			wantErr: false,
			wantPayload: struct {
				configMap                    corev1.ConfigMap
				mutatingWebhookConfiguration admissionregistrationv1.MutatingWebhookConfiguration
				pod                          corev1.Pod
			}{
				configMap: corev1.ConfigMap{
					ObjectMeta: v1.ObjectMeta{Name: "test-configmap", Namespace: "test"},
					Data:       map[string]string{"test-file": "updated-test-data"},
				},
				pod: corev1.Pod{
					ObjectMeta: v1.ObjectMeta{
						Name: "test-pod", Namespace: "test",
						Annotations: map[string]string{"lm-config-reloader/test-file-configHash": "8d5e2a59220ca4ac504cf569c7ba7a0e7b5d266fe61cb6cd0af3912d1cf1f108"}},
				},
			},
		},
		{
			name: "Sync for mutatingWebhookConfigurations with no config change found",
			args: struct {
				fetcherResponse *fetcher.Response
			}{
				fetcherResponse: &fetcher.Response{
					FileName: "test-file",
					FileData: []byte(
						"apiVersion: admissionregistration.k8s.io/v1\nkind: MutatingWebhookConfiguration\nmetadata:\n  name: test-mutatingWebhookConfiguration\nwebhooks:\n- name: test-webhook\n  objectSelector:\n    matchLabels:\n      foo: bar\n  namespaceSelector:\n    matchExpressions:\n    - key: environment\n      operator: In\n      values:\n      - dev\n      - staging\n",
					),
				},
			},
			fields: struct {
				configSyncer ConfigSyncer
			}{
				configSyncer: mutatingWebhookConfigSyncer{
					Resource: MutatingWebhookConfigurationResource{
						Name: "test-mutatingWebhookConfiguration",
					},
					k8sClient: k8sClient,
				},
			},
			wantErr: false,
			wantPayload: struct {
				configMap                    corev1.ConfigMap
				mutatingWebhookConfiguration admissionregistrationv1.MutatingWebhookConfiguration
				pod                          corev1.Pod
			}{
				mutatingWebhookConfiguration: admissionregistrationv1.MutatingWebhookConfiguration{
					ObjectMeta: v1.ObjectMeta{Name: "test-mutatingWebhookConfiguration"},
					Webhooks: []admissionregistrationv1.MutatingWebhook{
						{
							Name:           "test-webhook",
							ObjectSelector: &v1.LabelSelector{MatchLabels: map[string]string{"foo": "bar"}},
							NamespaceSelector: &v1.LabelSelector{
								MatchExpressions: []v1.LabelSelectorRequirement{
									{Key: "environment", Operator: "In", Values: []string{"dev", "staging"}},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Sync for mutatingWebhookConfigurations with config change found",
			args: struct {
				fetcherResponse *fetcher.Response
			}{
				fetcherResponse: &fetcher.Response{
					FileName: "test-file",
					FileData: []byte(
						"apiVersion: admissionregistration.k8s.io/v1\nkind: MutatingWebhookConfiguration\nmetadata:\n  name: test-mutatingWebhookConfiguration\nwebhooks:\n- name: test-webhook\n  objectSelector:\n    matchLabels:\n      app: test\n  namespaceSelector:\n    matchExpressions:\n    - key: environment\n      operator: In\n      values:\n      - dev\n",
					),
				},
			},
			fields: struct {
				configSyncer ConfigSyncer
			}{
				configSyncer: mutatingWebhookConfigSyncer{
					Resource: MutatingWebhookConfigurationResource{
						Name: "test-mutatingWebhookConfiguration",
					},
					k8sClient: k8sClient,
				},
			},
			wantErr: false,
			wantPayload: struct {
				configMap                    corev1.ConfigMap
				mutatingWebhookConfiguration admissionregistrationv1.MutatingWebhookConfiguration
				pod                          corev1.Pod
			}{
				mutatingWebhookConfiguration: admissionregistrationv1.MutatingWebhookConfiguration{
					ObjectMeta: v1.ObjectMeta{Name: "test-mutatingWebhookConfiguration"},
					Webhooks: []admissionregistrationv1.MutatingWebhook{
						{
							Name:           "test-webhook",
							ObjectSelector: &v1.LabelSelector{MatchLabels: map[string]string{"app": "test"}},
							NamespaceSelector: &v1.LabelSelector{
								MatchExpressions: []v1.LabelSelectorRequirement{
									{Key: "environment", Operator: "In", Values: []string{"dev"}},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Sync for configmap and send reload request",
			args: struct {
				fetcherResponse *fetcher.Response
			}{
				fetcherResponse: &fetcher.Response{
					FileName: "test-file",
					FileData: []byte("updated-test-data-1"),
				},
			},
			fields: struct {
				configSyncer ConfigSyncer
			}{
				configSyncer: configMapConfigSyncer{
					Resource: ConfigMapResource{
						Name:     "test-configmap",
						FileName: "test-file",
					},
					k8sClient:        k8sClient,
					ReloaderEndpoint: ts.URL + "/reload",
					HttpClient:       ts.Client(),
				},
			},
			wantErr: false,
			wantPayload: struct {
				configMap                    corev1.ConfigMap
				mutatingWebhookConfiguration admissionregistrationv1.MutatingWebhookConfiguration
				pod                          corev1.Pod
			}{
				configMap: corev1.ConfigMap{
					ObjectMeta: v1.ObjectMeta{Name: "test-configmap", Namespace: "test"},
					Data:       map[string]string{"test-file": "updated-test-data-1"},
				},
				pod: corev1.Pod{
					ObjectMeta: v1.ObjectMeta{
						Name: "test-pod", Namespace: "test",
						Annotations: map[string]string{"lm-config-reloader/test-file-configHash": "71028f43f33acebfdcc1cc039a69d75b0da26e63b586f24bde4bd87f44975a2f"}},
				},
			},
		},
		{
			name: "Sync for configmap with send reload request returning error",
			args: struct {
				fetcherResponse *fetcher.Response
			}{
				fetcherResponse: &fetcher.Response{
					FileName: "test-file",
					FileData: []byte("updated-test-data-2"),
				},
			},
			fields: struct {
				configSyncer ConfigSyncer
			}{
				configSyncer: configMapConfigSyncer{
					Resource: ConfigMapResource{
						Name:     "test-configmap",
						FileName: "test-file",
					},
					k8sClient:        k8sClient,
					ReloaderEndpoint: ts.URL + "/invalid-path",
					HttpClient:       ts.Client(),
				},
			},
			wantErr: true,
			wantPayload: struct {
				configMap                    corev1.ConfigMap
				mutatingWebhookConfiguration admissionregistrationv1.MutatingWebhookConfiguration
				pod                          corev1.Pod
			}{
				configMap: corev1.ConfigMap{
					ObjectMeta: v1.ObjectMeta{Name: "test-configmap", Namespace: "test"},
					Data:       map[string]string{"test-file": "updated-test-data-2"},
				},
				pod: corev1.Pod{
					ObjectMeta: v1.ObjectMeta{
						Name: "test-pod", Namespace: "test",
						Annotations: map[string]string{"lm-config-reloader/test-file-configHash": "990d186237476b3d199c1953f02975b6e6a9861834311c5dbaa5937773304643"}},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.fields.configSyncer.Sync(test.args.fetcherResponse)
			if err == nil && test.wantErr {
				t.Errorf("Sync() returned nil, instead of error")
			}
			if err != nil && !test.wantErr {
				t.Errorf("Sync() returned an unexpected error: %+v", err)
			}

			if configMapSyncer, ok := test.fields.configSyncer.(configMapConfigSyncer); ok {
				cfg, err := k8sClient.Clientset.CoreV1().ConfigMaps(os.Getenv("POD_NAMESPACE")).Get(context.TODO(), configMapSyncer.Resource.Name, v1.GetOptions{})
				if err != nil {
					t.Errorf("error occured in getting the configmap: %+v", err)
				}
				if !cmp.Equal(test.wantPayload.configMap.Data, cfg.Data) {
					t.Errorf("expected config map after sync: %+v, but got: %+v", test.wantPayload.configMap.Data, cfg.Data)
				}

				// Check for pod annotation
				pod, err := k8sClient.Clientset.CoreV1().Pods(os.Getenv("POD_NAMESPACE")).Get(context.TODO(), os.Getenv("POD_NAME"), v1.GetOptions{})
				if err != nil {
					t.Errorf("error occured in getting the pod: %+v", err)
				}
				for annotationName, annotationValue := range test.wantPayload.pod.Annotations {
					if _, ok := pod.Annotations[annotationName]; ok {
						if pod.Annotations[annotationName] != annotationValue {
							t.Errorf("expected annotation is: %s=%s, but got: %s=%s", annotationName, annotationValue, annotationName, pod.Annotations[annotationName])
						}
					} else {
						t.Errorf("expected annotation %s=%s not found", annotationName, annotationValue)
					}
				}
			} else if mutatingWebhookConfigSyncer, ok := test.fields.configSyncer.(mutatingWebhookConfigSyncer); ok {
				mutatingWebhookConfiguration, err := k8sClient.Clientset.AdmissionregistrationV1().MutatingWebhookConfigurations().Get(context.TODO(), mutatingWebhookConfigSyncer.Resource.Name, v1.GetOptions{})
				if err != nil {
					t.Errorf("error occured in getting the configmap: %+v", err)
				}
				if !cmp.Equal(test.wantPayload.mutatingWebhookConfiguration.Webhooks[0].ObjectSelector, mutatingWebhookConfiguration.Webhooks[0].ObjectSelector) ||
					!cmp.Equal(test.wantPayload.mutatingWebhookConfiguration.Webhooks[0].NamespaceSelector, mutatingWebhookConfiguration.Webhooks[0].NamespaceSelector) {
					t.Errorf("expected mutating webhook configurations after sync: %+v, but got: %+v", test.wantPayload.mutatingWebhookConfiguration, mutatingWebhookConfiguration)
				}
			}
		})
	}
}
