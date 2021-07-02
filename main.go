package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/k8s-autoops/autoops"
	admissionv1 "k8s.io/api/admission/v1"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func exit(err *error) {
	if *err != nil {
		log.Println("exited with error:", (*err).Error())
		os.Exit(1)
	} else {
		log.Println("exited")
	}
}

func main() {
	var err error
	defer exit(&err)

	log.SetFlags(0)
	log.SetOutput(os.Stdout)

	var maxPods int
	if maxPods, err = strconv.Atoi(os.Getenv("MAX_PODS")); err != nil {
		return
	}

	if maxPods < 10 {
		maxPods = 10
	}

	var client *kubernetes.Clientset
	if client, err = autoops.InClusterClient(); err != nil {
		return
	}

	s := &http.Server{
		Addr: ":443",
		Handler: autoops.NewMutatingAdmissionHTTPHandler(
			func(ctx context.Context, request *admissionv1.AdmissionRequest, patches *[]map[string]interface{}) (deny string, err error) {
				var buf []byte
				if buf, err = request.Object.MarshalJSON(); err != nil {
					return
				}
				var pod corev1.Pod
				if err = json.Unmarshal(buf, &pod); err != nil {
					return
				}
				log.Println("Try to Create Pod:", pod.Name, "in", pod.Namespace)
				var replicaSetName string
				for _, ref := range pod.OwnerReferences {
					if ref.Kind == "ReplicaSet" {
						replicaSetName = ref.Name
					}
				}
				if replicaSetName == "" {
					return
				}
				log.Println("Found ReplicaSet:", replicaSetName)
				var replicaSet *v1.ReplicaSet
				if replicaSet, err = client.AppsV1().ReplicaSets(request.Namespace).Get(ctx, replicaSetName, metav1.GetOptions{}); err != nil {
					return
				}
				var deploymentName string
				for _, ref := range replicaSet.OwnerReferences {
					if ref.Kind == "Deployment" {
						deploymentName = ref.Name
					}
				}
				if deploymentName == "" {
					return
				}
				log.Println("Found Deployment:", deploymentName)
				var deployment *v1.Deployment
				if deployment, err = client.AppsV1().Deployments(request.Namespace).Get(ctx, deploymentName, metav1.GetOptions{}); err != nil {
					return
				}
				labels := deployment.Spec.Template.Labels
				var pods *corev1.PodList
				if pods, err = client.CoreV1().Pods(request.Namespace).List(ctx, metav1.ListOptions{
					LabelSelector: Labels2Selector(labels),
				}); err != nil {
					return
				}
				log.Println("Current Pods:", len(pods.Items))
				if len(pods.Items) > maxPods {
					deny = fmt.Sprintf("Max Pods Exceeded by autoops.enforce-deployment-max-pods")
					return
				}
				return
			},
		),
	}

	if err = autoops.RunAdmissionServer(s); err != nil {
		return
	}
}

func Labels2Selector(labels map[string]string) string {
	var items []string
	for k, v := range labels {
		items = append(items, k+"="+v)
	}
	return strings.Join(items, ",")
}
