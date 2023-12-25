package app

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/adde/kade/internal/prompts"
	"github.com/adde/kade/internal/svc"
	"github.com/briandowns/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/erikgeiser/promptkit/confirmation"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/util/homedir"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	APP_VERSION   = "v0.1.0"
	SEPARATOR     = "\n*****************************************************************\n\n"
	SEPARATOR_NNL = "*****************************************************************"
)

func CheckAppVersion() {
	var checkVersion bool

	flag.BoolVar(&checkVersion, "version", false, "display the current version")
	flag.BoolVar(&checkVersion, "v", false, "alias for display the current version")

	flag.Parse()

	if checkVersion {
		fmt.Println(APP_VERSION)
		os.Exit(0)
	}
}

func Create() {
	PrintHeader()

	clientset, rawConfig := InitKubernetesConnection()
	appType := GetAppType()

	CreateAppByType(clientset, rawConfig, appType)
}

func CreateAppByType(clientset *kubernetes.Clientset, rawConfig api.Config, appType string) {
	switch appType {
	case "WordPress":
		wp := svc.WordPress{
			Namespace:             prompts.TextInput("Namespace in Rancher/Kubernetes?", svc.WP_PLACEHOLDER_NAMESPACE, "", true),
			DeploymentName:        prompts.TextInput("Deployment name?", "", svc.WP_PLACEHOLDER_DEPLOYMENT, true),
			UploadsVolSize:        prompts.TextInput("WordPress uploads volume size(Gi)?", "", svc.WP_PLACEHOLDER_WP_UPLOADS, true),
			ContainerImage:        prompts.TextInput("Container image to deploy?", "", "wordpress:6.4.2", true),
			ContainerRegistryUri:  prompts.TextInput("Container registry URI(leave blank if docker.com)?", "", "", false),
			ContainerRegistryUser: prompts.TextInput("Container registry user(leave blank if docker.com)?", "", "", false),
			ContainerRegistryPass: prompts.PassWordInput("Container registry password(leave blank if docker.com)?", "", false),
			Hostname:              prompts.TextInput("Hostname that the web app should be exposed on?", svc.WP_PLACEHOLDER_HOSTNAME, "", true),
			IngressTls:            prompts.ConfirmationInput("Do you want to configure TLS for the app?", confirmation.No),
			DatabaseHost:          prompts.TextInput("Database host?", svc.WP_PLACEHOLDER_DB_HOST, "", true),
			DatabaseName:          prompts.TextInput("Database name?", svc.WP_PLACEHOLDER_DB_NAME, "", true),
			DatabaseUser:          prompts.TextInput("Database user?", "", svc.WP_PLACEHOLDER_DB_USER, true),
			DatabasePass:          prompts.PassWordInput("Database password?", "", true),
			Clientset:             clientset,
		}

		confirm := prompts.ConfirmationInput(
			fmt.Sprintf(
				"Are you sure you want to continue deploying to cluster: %s?",
				rawConfig.Contexts[rawConfig.CurrentContext].Cluster),
			confirmation.No)

		if confirm {
			fmt.Print(SEPARATOR)
			fmt.Print("Deploying resources to cluster...\n\n")

			// Create namespace
			wp.CreateNamespace(
				"✔ Namespace %s created\n\n",
				"⚠ Namespace already exists, continuing...\n\n")

			// Create PVC
			wp.CreatePvc(
				"✔ PVC %s created\n\n",
				"⚠ PVC already exists, continuing...\n\n")

			wp.CreateDbPasswordSecret(
				"✔ Database password secret %s created\n\n",
				"⚠ Database password secret already exists, continuing...\n\n",
			)

			wp.CreateRegistryAuthSecret(
				"✔ Container registry auth secret %s created\n\n",
				"⚠ Container registry auth secret already exists, continuing...\n\n",
				"⚠ No container registry credentials provided, skipping...\n\n",
			)

			wp.CreateDeployment(
				"✔ WordPress deployment %s created\n\n",
				"⚠ WordPress deployment already exists, continuing...\n\n",
			)

			wp.CreateService(
				"✔ WordPress service %s created\n\n",
				"⚠ WordPress service already exists, continuing...\n\n",
			)

			wp.CreateIngress(
				"✔ WordPress ingress %s created\n\n",
				"⚠ WordPress ingress already exists, continuing...\n\n",
			)

			PrintPreparingEnvironment()
			PrintEnvironmentReady(wp.GetDeploymentUrl())
		} else {
			fmt.Println("Aborting...")
		}
		break
	case "Simple web app":
		PrintNotImplemented()
		break
	}
}

func InitKubernetesConnection() (*kubernetes.Clientset, api.Config) {
	kubeconfig := GetKubeconfig()

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Using config from path: ~/.kube/config")
	fmt.Println("Connecting to Kubernetes cluster... ")

	s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	s.Start()

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	loadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig}
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	rawConfig, err := kubeConfig.RawConfig()
	if err != nil {
		log.Fatal(err)
	}

	_, err = clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		fmt.Println("Failed to connect to the cluster:", err)
		os.Exit(1)
	}

	time.Sleep(500 * time.Millisecond)
	s.Stop()
	fmt.Printf("Successfully connected to cluster: %s\n\n", rawConfig.Contexts[rawConfig.CurrentContext].Cluster)

	return clientset, rawConfig
}

func GetKubeconfig() string {
	var kubeconfig string

	if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	} else {
		kubeconfig = ""
	}

	return kubeconfig
}

func GetAppType() string {
	appType := prompts.SelectInput(
		"What type of app do you want to deploy?",
		[]string{"WordPress", "Simple web app"})

	fmt.Println()

	return appType
}

func PrintHeader() {
	style := getContainerStyle()
	fmt.Println(style.Render("KADE\nKubernetes Application Deployment Engine"))
}

func PrintNotImplemented() {
	style := getContainerStyle()
	fmt.Println(style.Render("Not implemented yet, come back later!"))
}

func PrintPreparingEnvironment() {
	p := spinner.New(spinner.CharSets[26], 250*time.Millisecond)
	p.Prefix = "Preparing the environment "
	p.Start()
	time.Sleep(25 * time.Second)
	p.Stop()
}

func PrintEnvironmentReady(url string) {
	style := getContainerStyle()
	fmt.Println(style.Render("✔ Environment ready:\n\n" + url))
}

func getContainerStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Border(lipgloss.RoundedBorder()).
		PaddingTop(1).
		PaddingBottom(1).
		PaddingLeft(4).
		PaddingRight(4).
		Width(64).Align(lipgloss.Center)
}
