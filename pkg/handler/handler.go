package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"lm-webhook/pkg/config"
	"lm-webhook/pkg/mutation"
	"net/http"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	corev1 "k8s.io/api/core/v1"
)

type LMPodMutator struct {
	Client   client.Client
	decoder  *admission.Decoder
	Log      logr.Logger
	LMConfig *config.Config
}

// Handle is called internally to handle the admission request
func (podMutator *LMPodMutator) Handle(ctx context.Context, req admission.Request) admission.Response {
	logger := podMutator.Log.WithValues("lm-podmutator-webhook", fmt.Sprintf("%s/%s", req.Namespace, req.Name))
	pod := &corev1.Pod{}

	logger.Info("Receieved admission request:", req.Namespace, req.Name)

	err := podMutator.decoder.Decode(req, pod)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	logger.Info("Calling mutation")

	err = mutation.MutatePod(pod, podMutator.LMConfig, podMutator.Client, req.Namespace)

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
func (a *LMPodMutator) InjectDecoder(d *admission.Decoder) error {
	a.decoder = d
	return nil
}
