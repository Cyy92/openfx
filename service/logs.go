package service

import (
	"fmt"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func GetLog(functionName string, functionNamespace string, clientset *kubernetes.Clientset) (string, error) {
	var opts metav1.ListOptions
	opts.LabelSelector = fmt.Sprintf("openfx-fn=%s", functionName)
	podList, err := clientset.CoreV1().Pods(functionNamespace).List(opts)
	if err != nil {
		log.Println(err)
		return "", status.Error(codes.Internal, err.Error())
	}

	var logs string
	for _, pod := range podList.Items {
		if pod.Status.Phase == v1.PodFailed && pod.Status.Reason == "Evicted" {
			log.Printf("[%s] Skipping evicted pod.", pod.Name)
			continue
		}

		options := &v1.PodLogOptions{
			Follow:     false,
			Timestamps: false,
			Previous:   false,
		}

		body, err := clientset.CoreV1().Pods(functionNamespace).GetLogs(pod.Name, options).DoRaw()
		if err != nil {
			log.Println(err)
			return "", status.Error(codes.Internal, err.Error())
		}

		logs = logs + fmt.Sprintf("---\nName: %s\nLog:\n", functionName) + string(body)
	}

	return logs, nil
}
