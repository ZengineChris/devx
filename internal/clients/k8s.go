package clients

import (
	"context"
	"fmt"
	"os"

	"github.com/zenginechris/devx/internal/projects"
    corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func UpdateDeployment(project projects.Project, imageTag string) {
	var useContext string
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}

	// Set context if specified
	if useContext != "" {
		configOverrides.CurrentContext = useContext
	}

	// Set context if specified
	if useContext != "" {
		configOverrides.CurrentContext = useContext
	}

	// Get the clientConfig respecting the current context
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	config, err := clientConfig.ClientConfig()
	if err != nil {
		fmt.Printf("Error building kubeconfig: %s\n", err.Error())
		os.Exit(1)
	}

	// Create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("Error creating kubernetes client: %s\n", err.Error())
		os.Exit(1)
	}

	// Get current context name
	rawConfig, err := clientConfig.RawConfig()
	if err != nil {
		fmt.Printf("Warning: couldn't get current context name: %s\n", err.Error())
	} else {
		fmt.Printf("Using Kubernetes context: %s\n", rawConfig.CurrentContext)
	}

	// Get the deployment
	deployment, err := clientset.AppsV1().Deployments(project.Namespace).Get(context.TODO(), project.DeploymentName, metav1.GetOptions{})
	if err != nil {
		fmt.Printf("Error getting deployment: %s\n", err.Error())
		os.Exit(1)
	}
	containerUpdated := false
	fmt.Printf("Updating deployment %s containers:\n", project.DeploymentName)
	for i := range deployment.Spec.Template.Spec.Containers {
		container := &deployment.Spec.Template.Spec.Containers[i]
		fmt.Printf("  - Container %s: %s -> %s\n",
			container.Name,
			container.Image,
			imageTag)

		container.Image = imageTag


        	// Set image pull policy to Never
		pullPolicyBefore := container.ImagePullPolicy
		container.ImagePullPolicy = corev1.PullNever
		fmt.Printf("    Pull Policy: %s -> %s\n", pullPolicyBefore, container.ImagePullPolicy)

		containerUpdated = true
	}

	if !containerUpdated {
		fmt.Println("No containers found in the deployment")
		os.Exit(1)
	}
	_, err = clientset.AppsV1().Deployments(project.Namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
	if err != nil {
		fmt.Printf("Error updating deployment: %s\n", err.Error())
		os.Exit(1)
	}

}
