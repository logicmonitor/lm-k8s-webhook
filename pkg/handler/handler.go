package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/logicmonitor/lm-k8s-webhook/pkg/config"
	"github.com/logicmonitor/lm-k8s-webhook/pkg/mutation"

	"net/http"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	corev1 "k8s.io/api/core/v1"
)

type LMPodMutationHandler struct {
	Client   *config.K8sClient
	decoder  *admission.Decoder
	Log      logr.Logger
	LMConfig *config.Config
}

// Handle is called internally to handle the admission request
func (podMutationHandler *LMPodMutationHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	logger := podMutationHandler.Log.WithValues("lm-podmutator-webhook", fmt.Sprintf("%s/%s", req.Namespace, req.Name))
	pod := &corev1.Pod{}

	logger.Info("Received admission request:", req.Namespace, req.Name)

	err := podMutationHandler.decoder.Decode(req, pod)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	logger.Info("Calling mutation")

	params := NewParams(pod, podMutationHandler, req.Namespace)

	err = mutation.RunMutations(ctx, params)

	if err != nil {
		logger.Error(err, "Error occured in mutating the k8s resource")
		return admission.Errored(http.StatusInternalServerError, err)
	}

	logger.Info("End mutation")

	// End Mutation
	marshaledPod, err := json.Marshal(pod)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledPod)
}

// InjectDecoder injects the decoder.
func (a *LMPodMutationHandler) InjectDecoder(d *admission.Decoder) error {
	a.decoder = d
	return nil
}

func NewParams(pod *corev1.Pod, mutationHandler *LMPodMutationHandler, namespace string) *mutation.Params {
	return &mutation.Params{
		Client:    mutationHandler.Client,
		Pod:       pod,
		LMConfig:  mutationHandler.LMConfig,
		Mutations: mutation.Mutations,
		Namespace: namespace,
		Log:       mutationHandler.Log,
	}
}
