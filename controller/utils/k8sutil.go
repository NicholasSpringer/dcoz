package utils

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const DCOZ_AGENT_NAME = "dcoz-agent"
const DCOZ_AGENT_NS = "default"

func getAgentIps() []string {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	dset, err := clientset.AppsV1().DaemonSets(DCOZ_AGENT_NS).Get(context.TODO(), DCOZ_AGENT_NAME, metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}
	dsetSelector := dset.Spec.Selector

	agentPods, err := clientset.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{LabelSelector: dsetSelector.String()})
	if err != nil {
		panic(err.Error())
	}
	agentIps := []string{}
	for _, agentPod := range agentPods.Items {
		podIp := agentPod.Status.PodIP
		agentIps = append(agentIps, podIp)
	}
	return agentIps
}

func getContainerIds(deploymentNames []string) [][]string {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	containerIds := make([][]string, len(deploymentNames))
	for i, deploymentName := range deploymentNames {
		dep, err := clientset.AppsV1().Deployments("default").Get(context.TODO(), deploymentName, metav1.GetOptions{})
		if err != nil {
			panic(err.Error())
		}
		depSelector := dep.Spec.Selector

		depPods, err := clientset.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{LabelSelector: depSelector.String()})
		if err != nil {
			panic(err.Error())
		}
		containerIds[i] = []string{}
		for _, depPod := range depPods.Items {
			containerStatuses := depPod.Status.ContainerStatuses
			for _, cs := range containerStatuses {
				// TODO: ADD REGEXP ??? probably need to
				containerIds[i] = append(containerIds[i], cs.ContainerID)
			}
		}
	}

	return containerIds
}
