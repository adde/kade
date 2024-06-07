package svc

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/adde/kade/internal/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	WP_PLACEHOLDER_NAMESPACE  = "myproject"
	WP_PLACEHOLDER_DEPLOYMENT = "wordpress"
	WP_PLACEHOLDER_WP_UPLOADS = "2"
	WP_PLACEHOLDER_HOSTNAME   = "myproject.example.com"
	WP_PLACEHOLDER_DB_HOST    = "db.namespace.svc.cluster.local"
	WP_PLACEHOLDER_DB_NAME    = "my_project"
	WP_PLACEHOLDER_DB_USER    = "root"
	K8S_PVC_NAME              = "wp-uploads"
	K8S_DB_SECRET_KEY         = "WORDPRESS_DB_PASSWORD"
	K8S_DB_SECRET_NAME        = "wp-db-password"
	K8S_REGISTRY_SECRET_NAME  = "wp-registry-auth"
	K8S_CLUSTER_ISSUER_NAME   = "letsencrypt"
)

type DockerConfig struct {
	Auths map[string]struct {
		Username string
		Password string
		Auth     string
	}
}

type WordPress struct {
	Namespace             string
	DeploymentName        string
	UploadsVolSize        string
	ContainerImage        string
	ContainerRegistryUri  string
	ContainerRegistryUser string
	ContainerRegistryPass string
	Hostname              string
	IngressTls            bool
	DatabaseHost          string
	DatabaseName          string
	DatabaseUser          string
	DatabasePass          string
	Clientset             *kubernetes.Clientset
}

func (w WordPress) GetDeploymentUrl() string {
	wordpressUrl := w.Hostname

	if w.IngressTls {
		wordpressUrl = "https://" + wordpressUrl
	} else {
		wordpressUrl = "http://" + wordpressUrl
	}

	return wordpressUrl
}

func (w WordPress) CreateNamespace(successMessage, existsMessage string) *corev1.Namespace {
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: w.Namespace,
		},
	}

	createdNamespace, err := w.Clientset.CoreV1().Namespaces().Create(
		context.Background(),
		namespace,
		metav1.CreateOptions{},
	)

	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			fmt.Print(existsMessage)
		} else {
			log.Fatal(err)
		}
	} else {
		fmt.Printf(successMessage, w.Namespace)
	}

	return createdNamespace
}

func (w WordPress) CreatePvc(successMessage, existsMessage string) *corev1.PersistentVolumeClaim {
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      K8S_PVC_NAME,
			Namespace: w.Namespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(w.UploadsVolSize + "Gi"),
				},
			},
		},
	}

	createdPvc, err := w.Clientset.
		CoreV1().
		PersistentVolumeClaims(w.Namespace).
		Create(context.Background(), pvc, metav1.CreateOptions{})

	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			fmt.Print(existsMessage)
		} else {
			log.Fatal(err)
		}
	} else {
		fmt.Printf(successMessage, createdPvc.Name)
	}

	return createdPvc
}

func (w WordPress) CreateDbPasswordSecret(successMessage, existsMessage string) *corev1.Secret {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      K8S_DB_SECRET_NAME,
			Namespace: w.Namespace,
		},
		Data: map[string][]byte{
			K8S_DB_SECRET_KEY: []byte(w.DatabasePass),
		},
		Type: corev1.SecretTypeOpaque,
	}

	createdSecret, err := w.Clientset.CoreV1().Secrets(secret.Namespace).Create(
		context.Background(),
		secret,
		metav1.CreateOptions{},
	)

	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			fmt.Print(existsMessage)
		} else {
			log.Fatal(err)
		}
	} else {
		fmt.Printf(successMessage, createdSecret.Name)
	}

	return createdSecret
}

func (w WordPress) CreateRegistryAuthSecret(successMessage, existsMessage, noContainerRegistryCredsMessage string) *corev1.Secret {
	if w.ContainerRegistryUri == "" || w.ContainerRegistryUser == "" || w.ContainerRegistryPass == "" {
		fmt.Print(noContainerRegistryCredsMessage)
		return nil
	}

	dockerConfigJson, err := json.Marshal(w.getDockerAuthConfig())
	if err != nil {
		panic(err)
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      K8S_REGISTRY_SECRET_NAME,
			Namespace: w.Namespace,
		},
		Data: map[string][]byte{
			".dockerconfigjson": dockerConfigJson,
		},
		Type: corev1.SecretTypeDockerConfigJson,
	}

	createdSecret, err := w.Clientset.CoreV1().Secrets(secret.Namespace).Create(
		context.Background(),
		secret,
		metav1.CreateOptions{},
	)

	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			fmt.Print(existsMessage)
		} else {
			log.Fatal(err)
		}
	} else {
		fmt.Printf(successMessage, createdSecret.Name)
	}

	return createdSecret
}

func (w WordPress) CreateDeployment(successMessage, existsMessage string) *appsv1.Deployment {
	uniqueID := utils.GenerateUniqueID()
	appLabel := "deployment-" + w.Namespace + "-" + w.DeploymentName + "-" + uniqueID

	wordpressUrl := w.Hostname
	if w.IngressTls {
		wordpressUrl = "https://" + wordpressUrl
	} else {
		wordpressUrl = "http://" + wordpressUrl
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      w.DeploymentName,
			Namespace: w.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: utils.Int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": appLabel,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": appLabel,
					},
				},
				Spec: corev1.PodSpec{
					ImagePullSecrets: []corev1.LocalObjectReference{},
					Containers: []corev1.Container{
						{
							Name:  w.DeploymentName,
							Image: w.ContainerImage,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 80,
									Name:          w.DeploymentName,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      K8S_PVC_NAME,
									MountPath: "/var/www/html/wp-content/uploads",
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "WORDPRESS_URL",
									Value: wordpressUrl,
								},
								{
									Name:  "WORDPRESS_DB_HOST",
									Value: w.DatabaseHost,
								},
								{
									Name:  "WORDPRESS_DB_USER",
									Value: w.DatabaseUser,
								},
								{
									Name:  "WORDPRESS_DB_NAME",
									Value: w.DatabaseName,
								},
								{
									Name: K8S_DB_SECRET_KEY,
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											Key: K8S_DB_SECRET_KEY,
											LocalObjectReference: corev1.LocalObjectReference{
												Name: K8S_DB_SECRET_NAME,
											},
										},
									},
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: K8S_PVC_NAME,
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: K8S_PVC_NAME,
								},
							},
						},
					},
				},
			},
		},
	}

	// Append pull secret if registry credentials are provided
	if w.ContainerRegistryUri != "" && w.ContainerRegistryUser != "" && w.ContainerRegistryPass != "" {
		deployment.Spec.Template.Spec.ImagePullSecrets = append(
			deployment.Spec.Template.Spec.ImagePullSecrets,
			corev1.LocalObjectReference{
				Name: K8S_REGISTRY_SECRET_NAME,
			},
		)
	}

	createdDeployment, err := w.Clientset.AppsV1().Deployments(deployment.Namespace).Create(
		context.Background(),
		deployment,
		metav1.CreateOptions{},
	)

	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			fmt.Print(existsMessage)
		} else {
			log.Fatal(err)
		}
	} else {
		fmt.Printf(successMessage, createdDeployment.Name)
	}

	return createdDeployment
}

func (w WordPress) CreateService(successMessage, existsMessage string) *corev1.Service {
	deployment, err := w.Clientset.AppsV1().Deployments(w.Namespace).Get(
		context.Background(),
		w.DeploymentName,
		metav1.GetOptions{},
	)

	if err != nil {
		log.Fatal(err)
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      w.DeploymentName,
			Namespace: w.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": deployment.Spec.Selector.MatchLabels["app"],
			},
			Ports: []corev1.ServicePort{
				{
					Port:     80,
					Name:     w.DeploymentName,
					Protocol: corev1.ProtocolTCP,
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}

	createdService, err := w.Clientset.CoreV1().Services(service.Namespace).Create(
		context.Background(),
		service,
		metav1.CreateOptions{},
	)

	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			fmt.Print(existsMessage)
		} else {
			log.Fatal(err)
		}
	} else {
		fmt.Printf(successMessage, createdService.Name)
	}

	return createdService
}

func (w WordPress) CreateIngress(successMessage, existsMessage string) *networkingv1.Ingress {
	pathType := networkingv1.PathTypeImplementationSpecific

	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      w.DeploymentName,
			Namespace: w.Namespace,
			Annotations: map[string]string{
				"nginx.ingress.kubernetes.io/proxy-body-size":       "1G",
				"nginx.ingress.kubernetes.io/proxy-connect-timeout": "30",
				"nginx.ingress.kubernetes.io/proxy-read-timeout":    "600",
				"nginx.ingress.kubernetes.io/proxy-send-timeout":    "600",
				"nginx.ingress.kubernetes.io/configuration-snippet": "more_set_headers \"X-Robots-Tag: noindex\";\n",
			},
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: w.Hostname,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									PathType: &pathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: w.DeploymentName,
											Port: networkingv1.ServiceBackendPort{
												Number: 80,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	if w.IngressTls {
		// Append annotation for letsencrypt
		ingress.ObjectMeta.Annotations["cert-manager.io/cluster-issuer"] = K8S_CLUSTER_ISSUER_NAME

		// Append TLS config
		ingress.Spec.TLS = append(ingress.Spec.TLS, networkingv1.IngressTLS{
			Hosts:      []string{w.Hostname},
			SecretName: w.Namespace + "-tls",
		})
	}

	createdIngress, err := w.Clientset.NetworkingV1().Ingresses(ingress.Namespace).Create(
		context.Background(),
		ingress,
		metav1.CreateOptions{},
	)

	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			fmt.Print(existsMessage)
		} else {
			panic(err)
		}
	} else {
		fmt.Printf(successMessage, createdIngress.Name)
	}

	return createdIngress
}

func (w WordPress) getDockerAuthConfig() DockerConfig {
	dockerConfig := DockerConfig{
		Auths: map[string]struct {
			Username string
			Password string
			Auth     string
		}{
			w.ContainerRegistryUri: {
				Username: w.ContainerRegistryUser,
				Password: w.ContainerRegistryPass,
				Auth: base64.StdEncoding.EncodeToString(
					[]byte(w.ContainerRegistryUser + ":" + w.ContainerRegistryPass)),
			},
		},
	}

	return dockerConfig
}
