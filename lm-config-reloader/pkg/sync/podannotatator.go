package sync

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/pkg/config"
	"github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/pkg/logger"

	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// annotatePodWithConfigHash adds annotation to the specified pod with value as a config hash
func annotatePodWithConfigHash(podName string, configName string, configHash []byte, k8sClient *config.K8sClient) error {
	updateOptions := metav1.UpdateOptions{}
	pod, err := k8sClient.Clientset.CoreV1().Pods(os.Getenv("POD_NAMESPACE")).Get(context.Background(), podName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	if pod.ObjectMeta.Annotations == nil {
		pod.ObjectMeta.Annotations = make(map[string]string, 1)
	}
	pod.ObjectMeta.Annotations[fmt.Sprintf("lm-config-reloader/%s-configHash", configName)] = fmt.Sprintf("%x", configHash)
	pod.ObjectMeta.Annotations[fmt.Sprintf("lm-config-reloader/%s-last-modified", configName)] = time.Now().String()

	_, err = k8sClient.Clientset.CoreV1().Pods(os.Getenv("POD_NAMESPACE")).Update(context.Background(), pod, updateOptions)
	if err != nil {
		return err
	}
	logger.Logger().Info("annotation is added to the pod", zap.String("pod name", podName))
	return nil
}
