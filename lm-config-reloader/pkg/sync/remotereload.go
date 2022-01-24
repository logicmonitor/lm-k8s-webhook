package sync

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/pkg/logger"

	"go.uber.org/zap"
)

// triggerReload triggers config reload by sending the request to the specified endpoint
func triggerReload(reloadUrl string, client *http.Client) error {
	logger.Logger().Debug("Reloading configuration gracefully via POST request", zap.String("URL", reloadUrl))
	req, err := http.NewRequest(http.MethodPost, reloadUrl, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("%s endpoint returned statuscode %v; response: %v", reloadUrl, resp.StatusCode, string(body))
	}
	logger.Logger().Info("config reload triggered")
	return nil
}
