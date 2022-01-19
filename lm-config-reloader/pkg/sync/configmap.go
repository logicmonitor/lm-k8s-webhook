package sync

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/pkg/config"
	"github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/pkg/fetcher"
	"github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/pkg/logger"

	"crypto/sha256"

	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConfigMapResource represents Config map resource
type ConfigMapResource struct {
	Name     string
	FileName string
}

type configMapConfigSyncer struct {
	Resource         ConfigMapResource
	ReloaderEndpoint string
	k8sClient        *config.K8sClient
	HttpClient       *http.Client
}

// BuildConfigMapResource builds configmap resource
func BuildConfigMapResource(resource config.Resource) (ConfigMapResource, error) {
	var cmResource ConfigMapResource
	err := mapstructure.Decode(resource, &cmResource)
	if err != nil {
		return cmResource, err
	}
	return cmResource, nil
}

// Sync compares the configmap content with the config from the config provider
// and if not matched then updates the configmap content
func (configMapConfigSyncer configMapConfigSyncer) Sync(response *fetcher.Response) error {
	cm, err := configMapConfigSyncer.k8sClient.Clientset.CoreV1().ConfigMaps(os.Getenv("POD_NAMESPACE")).Get(context.Background(), configMapConfigSyncer.Resource.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	// compares config map data
	if !reflect.DeepEqual(string(response.FileData), cm.Data[configMapConfigSyncer.Resource.FileName]) {
		logger.Logger().Info("config content mismatch found", zap.String("config name", configMapConfigSyncer.Resource.FileName))

		// apply patch
		cm.Data[configMapConfigSyncer.Resource.FileName] = string(response.FileData)
		_, err := configMapConfigSyncer.k8sClient.Clientset.CoreV1().ConfigMaps(os.Getenv("POD_NAMESPACE")).Update(context.Background(), cm, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
		logger.Logger().Info("configmap updated", zap.String("configmap name", cm.Name))

		// Calculate config hash
		h := sha256.New()
		h.Write([]byte(cm.Data[configMapConfigSyncer.Resource.FileName]))
		configHash := h.Sum(nil)

		logger.Logger().Debug(fmt.Sprintf("config hash for %s is %x", configMapConfigSyncer.Resource.FileName, configHash))

		// annotate pod with the config hash
		if err := annotatePodWithConfigHash(os.Getenv("POD_NAME"), configMapConfigSyncer.Resource.FileName, configHash, configMapConfigSyncer.k8sClient); err != nil {
			return err
		}

		// If reload trigger is required
		if len(strings.TrimSpace(configMapConfigSyncer.ReloaderEndpoint)) > 0 {
			err := triggerReload(configMapConfigSyncer.ReloaderEndpoint, configMapConfigSyncer.HttpClient)
			if err != nil {
				return err
			}
		}
	} else {
		logger.Logger().Info("config content matched, no change detected", zap.String("config name", configMapConfigSyncer.Resource.FileName))
	}
	return nil
}
