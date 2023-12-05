package main

import (
	"fmt"
	"log"
	"os"

	"github.com/erikgeiser/promptkit/confirmation"
	"github.com/erikgeiser/promptkit/selection"
	"github.com/erikgeiser/promptkit/textinput"
)

type WordPress struct {
	namespace             string
	deploymentName        string
	uploadsVolSize        string
	containerImage        string
	containerRegistryUser string
	containerRegistryPass string
	hostname              string
	ingressTls            bool
	databaseHost          string
	databaseName          string
	databaseUser          string
	databasePass          string
}

func (w WordPress) CreateNamespace() string {
	return fmt.Sprintf("kubectl create namespace %s", w.namespace)
}

func main() {
	appType := selectInput(
		"What type of app do you want to deploy?",
		[]string{"WordPress", "Simple web app"})

	switch appType {
	case "WordPress":
		config := WordPress{
			namespace:             textInput("Namespace in Rancher?", "myproject-test", ""),
			deploymentName:        textInput("Deployment name?", "", "wordpress"),
			uploadsVolSize:        textInput("WordPress uploads volume size(GB)?", "", "2"),
			containerImage:        textInput("Container image to deploy?", "nginx:latest", ""),
			containerRegistryUser: textInput("Container registry user(leave blank if hub.docker.com)?", "", ""),
			containerRegistryPass: passWordInput("Container registry password(leave blank if hub.docker.com)?", ""),
			hostname:              textInput("Hostname that the web app should be exposed on?", "myproject.example.com", ""),
			ingressTls:            confirmationInput("Do you want to configure TLS for the app?", confirmation.No),
			databaseHost:          textInput("Database host?", "", "mysql.mysql-shared.svc.cluster.local"),
			databaseName:          textInput("Database name?", "my_project", ""),
			databaseUser:          textInput("Database user?", "", "root"),
			databasePass:          passWordInput("Database password?", ""),
		}

		confirm := confirmationInput(
			fmt.Sprintf(
				"Are you sure you want to continue deploying resources to cluster: %s?",
				"my-awesome-cluster"),
			confirmation.No)

		if confirm {
			fmt.Println("Deploying resources to kubernetes cluster...")
			fmt.Println(config.CreateNamespace())
		} else {
			fmt.Println("Aborting...")
		}
	}
}

func textInput(label string, placeholder string, initialValue string) string {
	input := textinput.New(label)
	if placeholder != "" {
		input.Placeholder = placeholder
	}
	if initialValue != "" {
		input.InitialValue = initialValue
	}

	value, err := input.RunPrompt()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	return value
}

func selectInput(label string, options []string) string {
	sp := selection.New(label, options)
	sp.Filter = nil

	value, err := sp.RunPrompt()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	return value
}

func confirmationInput(label string, initialValue confirmation.Value) bool {
	input := confirmation.New(label, initialValue)

	value, err := input.RunPrompt()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	return value
}

func passWordInput(label string, placeholder string) string {
	input := textinput.New(label)
	input.Placeholder = placeholder
	input.Hidden = true

	value, err := input.RunPrompt()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	return value
}
