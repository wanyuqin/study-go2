package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/kubernetes"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	deploymentutil "k8s.io/kubectl/pkg/util/deployment"
)

const (
	ChangeCauseAnnotation = "kubernetes.io/change-cause"
)

var annotationsToSkip = map[string]bool{
	corev1.LastAppliedConfigAnnotation:       true,
	deploymentutil.RevisionAnnotation:        true,
	deploymentutil.RevisionHistoryAnnotation: true,
	deploymentutil.DesiredReplicasAnnotation: true,
	deploymentutil.MaxReplicasAnnotation:     true,
	v1.DeprecatedRollbackTo:                  true,
}

func main() {
	// var kubeconfig *string
	// if home := homedir.HomeDir(); home != "" {
	// 	kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	// } else {
	// 	kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	// }
	// flag.Parse()
	//
	// config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	// if err != nil {
	// 	panic(err)
	// }
	// clientset, err := kubernetes.NewForConfig(config)
	// apps := clientset.AppsV1()
	// if err != nil {
	// 	panic(err)
	// }
	//
	// deploymentsClient := clientset.AppsV1().Deployments(apiv1.NamespaceDefault)
	//
	// name := "demo-deployment"
	// deployment, err := deploymentsClient.Get(context.Background(), name, metav1.GetOptions{
	// 	ResourceVersion: "25341",
	// })
	// if err != nil {
	// 	log.Fatal(err)
	// 	return
	// }
	//
	// deploymentsClient.List(context.Background(), metav1.ListOptions{
	// 	FieldSelector: name,
	// })
	//
	// _, oldRSs, newRS, err := deploymentutil.GetAllReplicaSets(deployment, apps)
	// if err != nil {
	// 	log.Fatal(err)
	// 	return
	// }
	//
	// allRs := append(oldRSs, newRS)
	// historyInfo := make(map[int64]*corev1.PodTemplateSpec)
	// for _, rs := range allRs {
	// 	v, err := deploymentutil.Revision(rs)
	// 	if err != nil {
	// 		klog.Warningf("unable to get revision from replicaset %s for deployment %s in namespace %s: %v", rs.Name, name, apiv1.NamespaceDefault, err)
	// 		continue
	// 	}
	// 	historyInfo[v] = &rs.Spec.Template
	// 	changeCause := getChangeCause(rs)
	// 	if historyInfo[v].Annotations == nil {
	// 		historyInfo[v].Annotations = make(map[string]string)
	// 	}
	// 	if len(changeCause) > 0 {
	// 		historyInfo[v].Annotations[ChangeCauseAnnotation] = changeCause
	// 	}
	// }
	//
	// if len(historyInfo) == 0 {
	// 	fmt.Printf("No rollout history found.")
	// 	return
	// }
	//
	// revisions := make([]int64, 0, len(historyInfo))
	// for r := range historyInfo {
	// 	revisions = append(revisions, r)
	// }
	//
	// fmt.Println(revisions)
	WorkQueueStart()
}

func prompt() {
	fmt.Printf("-> Press Return key to continue.")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		break
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	fmt.Println()
}

func getChangeCause(obj runtime.Object) string {
	accessor, err := meta.Accessor(obj)
	if err != nil {
		return ""
	}
	return accessor.GetAnnotations()[ChangeCauseAnnotation]
}

func DeploymentRollback(deployment *v1.Deployment, toRevision int64, dryRunStrategy cmdutil.DryRunStrategy) (string, error) {
	if toRevision < 0 {
		return "", errors.New("")
	}

	return "", nil
}

func DeploymentRevision(deployment *v1.Deployment, c kubernetes.Interface, toRevision int64) (revision *v1.ReplicaSet, err error) {

	_, allOldRSs, newRS, err := deploymentutil.GetAllReplicaSets(deployment, c.AppsV1())
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve replica sets from deployment %s: %v", deployment.Name, err)
	}
	allRSs := allOldRSs
	if newRS != nil {
		allRSs = append(allRSs, newRS)
	}

	var (
		latestReplicaSet   *v1.ReplicaSet
		latestRevision     = int64(-1)
		previousReplicaSet *v1.ReplicaSet
		previousRevision   = int64(-1)
	)
	for _, rs := range allRSs {
		if v, err := deploymentutil.Revision(rs); err == nil {
			if toRevision == 0 {
				if latestRevision < v {
					// newest one we've seen so far
					previousRevision = latestRevision
					previousReplicaSet = latestReplicaSet
					latestRevision = v
					latestReplicaSet = rs
				} else if previousRevision < v {
					// second newest one we've seen so far
					previousRevision = v
					previousReplicaSet = rs
				}
			} else if toRevision == v {
				return rs, nil
			}
		}
	}

	if toRevision > 0 {
		return nil, errors.New("revision not found")
	}

	if previousReplicaSet == nil {
		return nil, fmt.Errorf("no rollout history found for deployment %q", deployment.Name)
	}
	return previousReplicaSet, nil
}

func getDeploymentPatch(podTemplate *corev1.PodTemplateSpec, annotations map[string]string) (types.PatchType, []byte, error) {
	// Create a patch of the Deployment that replaces spec.template
	patch, err := json.Marshal([]interface{}{
		map[string]interface{}{
			"op":    "replace",
			"path":  "/spec/template",
			"value": podTemplate,
		},
		map[string]interface{}{
			"op":    "replace",
			"path":  "/metadata/annotations",
			"value": annotations,
		},
	})
	return types.JSONPatchType, patch, err
}
