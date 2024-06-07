package app

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/adde/kade/internal/config"
	"github.com/adde/kade/internal/prompts"
	"github.com/adde/kade/internal/svc"
	"github.com/adde/kade/internal/version"
	"github.com/briandowns/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/erikgeiser/promptkit/confirmation"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/util/homedir"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var skipVersionCheck bool

func ParseFlags() {
	var checkVersion bool
	var createConfig bool

	flag.BoolVar(&checkVersion, "version", false, "display the current version")
	flag.BoolVar(&checkVersion, "v", false, "alias for display the current version")

	flag.BoolVar(&skipVersionCheck, "skip-version-check", false, "skip checking for latest version of the app")

	flag.BoolVar(&createConfig, "create-config", false, "create config file")
	flag.BoolVar(&createConfig, "cc", false, "alias for create config file")

	flag.Parse()

	if checkVersion {
		fmt.Println(version.CurrentVersion)
		os.Exit(0)
	}

	if createConfig {
		config.CreateConfig()
		os.Exit(0)
	}
}

func Create() {
	if !skipVersionCheck {
		PrintVersionInfo()
	}

	PrintHeader()

	clientset, rawConfig := InitKubernetesConnection()
	appConfig := config.GetConfig()
	appType := GetAppType()

	CreateAppByType(clientset, rawConfig, appConfig, appType)
}

func CreateAppByType(clientset *kubernetes.Clientset, rawConfig api.Config, appConfig *config.Config, appType string) {
	switch appType {
	case "WordPress":
		wp := svc.WordPress{
			Namespace:             prompts.TextInput("Namespace in Rancher/Kubernetes?", svc.WP_PLACEHOLDER_NAMESPACE, "", true),
			DeploymentName:        prompts.TextInput("Deployment name?", "", svc.WP_PLACEHOLDER_DEPLOYMENT, true),
			UploadsVolSize:        prompts.TextInput("WordPress uploads volume size(Gi)?", "", svc.WP_PLACEHOLDER_WP_UPLOADS, true),
			ContainerImage:        prompts.TextInput("Container image to deploy?", "", "wordpress:6.4.2", true),
			ContainerRegistryUri:  prompts.TextInput("Container registry URI(leave blank if docker.com)?", "", appConfig.Global.ContainerRegistry.Uri, false),
			ContainerRegistryUser: prompts.TextInput("Container registry user(leave blank if docker.com)?", "", appConfig.Global.ContainerRegistry.User, false),
			ContainerRegistryPass: prompts.PassWordInput("Container registry password(leave blank if docker.com)?", "", appConfig.Global.ContainerRegistry.Pass, false),
			Hostname:              prompts.TextInput("Hostname that the web app should be exposed on?", svc.WP_PLACEHOLDER_HOSTNAME, "", true),
			IngressTls:            prompts.ConfirmationInput("Do you want to configure TLS for the app?", confirmation.No),
			DatabaseHost:          prompts.TextInput("Database host?", "", appConfig.Global.Database.Host, true),
			DatabaseName:          prompts.TextInput("Database name?", svc.WP_PLACEHOLDER_DB_NAME, "", true),
			DatabaseUser:          prompts.TextInput("Database user?", "", appConfig.Global.Database.User, true),
			DatabasePass:          prompts.PassWordInput("Database password?", "", appConfig.Global.Database.Pass, true),
			Clientset:             clientset,
		}

		confirm := prompts.ConfirmationInput(
			fmt.Sprintf(
				"Are you sure you want to continue deploying to cluster: %s?",
				rawConfig.Contexts[rawConfig.CurrentContext].Cluster),
			confirmation.No)

		if confirm {
			sepStyle := getSeparatorStyle()
			fmt.Println(sepStyle.Render(""))
			fmt.Print("Deploying resources to cluster...\n\n")

			// Create namespace
			wp.CreateNamespace(
				"✔ Namespace %s created\n",
				"⚠ Namespace already exists, continuing...\n")

			// Create PVC
			wp.CreatePvc(
				"✔ PVC %s created\n",
				"⚠ PVC already exists, continuing...\n")

			wp.CreateDbPasswordSecret(
				"✔ Database password secret %s created\n",
				"⚠ Database password secret already exists, continuing...\n",
			)

			wp.CreateRegistryAuthSecret(
				"✔ Container registry auth secret %s created\n",
				"⚠ Container registry auth secret already exists, continuing...\n",
				"⚠ No container registry credentials provided, skipping...\n",
			)

			wp.CreateDeployment(
				"✔ WordPress deployment %s created\n",
				"⚠ WordPress deployment already exists, continuing...\n",
			)

			wp.CreateService(
				"✔ WordPress service %s created\n",
				"⚠ WordPress service already exists, continuing...\n",
			)

			wp.CreateIngress(
				"✔ WordPress ingress %s created\n",
				"⚠ WordPress ingress already exists, continuing...\n",
			)

			fmt.Println()
			PrintPreparingEnvironment(wp)
			PrintEnvironmentReady(wp.GetDeploymentUrl())
		} else {
			fmt.Println("Aborting...")
		}
	case "Simple web app":
		PrintNotImplemented()
	}
}

func InitKubernetesConnection() (*kubernetes.Clientset, api.Config) {
	kubeconfig := GetKubeconfig()

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Using kube config from path: ~/.kube/config")
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
	fmt.Println(style.Render("KADE\nKubernetes Application Deployment Engine\n" + version.CurrentVersion))
	fmt.Println()
}

func PrintVersionInfo() {
	if !version.IsLatestVersion() {
		style := getContainerStyle()
		fmt.Println(style.Render(
			"A new version of KADE is available, please update to the latest version.\n\n" +
				"Current version: " + version.CurrentVersion + "\nLatest version: " + version.LatestVersion))
		os.Exit(0)
	}
}

func PrintNotImplemented() {
	style := getContainerStyle()
	fmt.Println(style.Render("Not implemented yet, come back later!"))
}

func PrintPreparingEnvironment(wp svc.WordPress) {
	p := spinner.New(spinner.CharSets[26], 250*time.Millisecond)
	p.Prefix = "Preparing the environment "
	p.Start()

	count := 0
	for !wp.IsDeploymentReady() {
		time.Sleep(5 * time.Second)
		count++
		if count == 11 {
			break
		}
	}

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
		Align(lipgloss.Center).
		PaddingTop(1).
		PaddingBottom(1).
		PaddingLeft(4).
		PaddingRight(4).
		Width(64)
}

func getSeparatorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderTop(false).
		BorderRight(false).
		BorderLeft(false).
		MarginBottom(1).
		Width(66)
}
