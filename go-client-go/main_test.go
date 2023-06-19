package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"testing"

	v1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	deploymentutil "k8s.io/kubectl/pkg/util/deployment"
)

func setup() (*kubernetes.Clientset, error) {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)

	return clientset, err
}

func Test_DeploymentRevision(t *testing.T) {

	clientset, err := setup()
	if err != nil {
		log.Fatal(err)
		return

	}
	name := "demo-deployment"
	deployment, err := clientset.AppsV1().Deployments(apiv1.NamespaceDefault).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		log.Fatal(err)
		return

	}

	toRevision := int64(1)
	rsForRevision, err := DeploymentRevision(deployment, clientset, toRevision)
	if err != nil {
		log.Fatal(err)
		return
	}

	delete(rsForRevision.Spec.Template.Labels, v1.DefaultDeploymentUniqueLabelKey)

	annotations := map[string]string{}
	for k := range annotationsToSkip {
		if v, ok := deployment.Annotations[k]; ok {
			annotations[k] = v
		}
	}
	for k, v := range rsForRevision.Annotations {
		if !annotationsToSkip[k] {
			annotations[k] = v
		}
	}

	patchType, patch, err := getDeploymentPatch(&rsForRevision.Spec.Template, annotations)
	if err != nil {
		log.Fatal(err)
		return
	}

	patchOptions := metav1.PatchOptions{}
	if _, err = clientset.AppsV1().Deployments(apiv1.NamespaceDefault).Patch(context.TODO(), name, patchType, patch, patchOptions); err != nil {
		log.Fatal(fmt.Errorf("failed restoring revision %d: %v", toRevision, err))
		return
	}
}

func Test_GetAllRs(t *testing.T) {
	clientset, err := setup()
	if err != nil {
		log.Fatal(err)
		return
	}

	name := "demo-deployment"
	deployment, err := clientset.AppsV1().Deployments(apiv1.NamespaceDefault).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		log.Fatal(err)
		return

	}

	_, allOldRSs, newRS, err := deploymentutil.GetAllReplicaSets(deployment, clientset.AppsV1())
	if err != nil {
		log.Fatal(err)
		return
	}
	allRSs := allOldRSs
	if newRS != nil {
		allRSs = append(allRSs, newRS)
	}

	for _, v := range allRSs {
		rv, err := deploymentutil.Revision(v)
		if err != nil {
			log.Fatal(err)
			return
		}

		fmt.Printf("rsName %s , revision %d, time:%s \n\n", v.Name, rv, v.CreationTimestamp)
	}

}

func Test_FindByLabel(t *testing.T) {
	clientset, err := setup()
	if err != nil {
		return
	}
	lp, err := labels.Parse("serivce-id=123")
	if err != nil {
		log.Fatal(lp)
		return
	}

	dl, err := clientset.AppsV1().Deployments(metav1.NamespaceDefault).List(context.Background(), metav1.ListOptions{
		LabelSelector: "serivce-id=123",
	})
	if err != nil {
		log.Fatal(err)
		return
	}

	for _, deploy := range dl.Items {
		fmt.Println(deploy.Name)
	}
}
