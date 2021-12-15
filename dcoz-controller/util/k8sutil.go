package util

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const DCOZ_AGENT_NAME = "dcoz-agent"
const DCOZ_AGENT_NS = "default"

func GetAgentIps() ([]string, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	dset, err := clientset.AppsV1().DaemonSets(DCOZ_AGENT_NS).Get(context.TODO(), DCOZ_AGENT_NAME, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	dsetSelector := dset.Spec.Selector

	agentPods, err := clientset.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{LabelSelector: dsetSelector.String()})
	if err != nil {
		return nil, err
	}
	agentIps := []string{}
	for _, agentPod := range agentPods.Items {
		podIp := agentPod.Status.PodIP
		agentIps = append(agentIps, podIp)
	}
	return agentIps, nil
}

func GetContainerIds(depName string) ([]string, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	dep, err := clientset.AppsV1().Deployments("default").Get(context.TODO(), depName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	depSelector := dep.Spec.Selector
	depPods, err := clientset.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{LabelSelector: depSelector.String()})
	if err != nil {
		return nil, err
	}

	containerIds := []string{}
	for _, depPod := range depPods.Items {
		containerStatuses := depPod.Status.ContainerStatuses
		for _, cs := range containerStatuses {
			// TODO: ADD REGEXP ??? probably need to
			containerIds = append(containerIds, cs.ContainerID)
		}
	}
	return containerIds, nil
}
