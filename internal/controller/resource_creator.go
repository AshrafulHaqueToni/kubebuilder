package controller

import (
	ashappsv1 "github.com/AshrafulHaqueToni/kubebuilder/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func newDeployment(ash *ashappsv1.Ash, deploymentName string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: ash.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(ash, ashappsv1.GroupVersion.WithKind("Ash")),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: ash.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":  ash.Name,
					"Kind": "Ash",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":  ash.Name,
						"Kind": "Ash",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  ash.Name,
							Image: ash.Spec.Container.Image,
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: ash.Spec.Container.Port,
								},
							},
						},
					},
				},
			},
		},
	}
}

func newService(ash *ashappsv1.Ash, serviceName string) *corev1.Service {
	labels := map[string]string{
		"app":  ash.Name,
		"Kind": "Ash",
	}
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind: "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: ash.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(ash, ashappsv1.GroupVersion.WithKind("Ash")),
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Port:       ash.Spec.Container.Port,
					TargetPort: intstr.FromInt(int(ash.Spec.Container.Port)),
					Protocol:   "TCP",
					NodePort:   ash.Spec.Service.ServiceNodePort,
				},
			},
			Type: func() corev1.ServiceType {
				if ash.Spec.Service.ServiceType == "NodePort" {
					return corev1.ServiceTypeNodePort
				} else {
					return corev1.ServiceTypeClusterIP
				}
			}(),
		},
	}
}
