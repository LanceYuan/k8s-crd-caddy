package controllers

import (
	devopsv1 "k8s-crd-caddy/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
)

func NewDeployment(app *devopsv1.Static) *appsv1.Deployment {
	selector := map[string]string{"app": controllerName}
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      controllerName,
			Namespace: app.Namespace,
			Labels:    selector,
		},
		Spec: appsv1.DeploymentSpec{
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
			},
			Replicas: pointer.Int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: selector,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      controllerName,
					Namespace: app.Namespace,
					Labels:    selector,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  controllerName,
							Image: "lanceyuan/caddy:v2.4.6",
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
							LivenessProbe: &corev1.Probe{
								Handler:             corev1.Handler{HTTPGet: &corev1.HTTPGetAction{Port: intstr.IntOrString{Type: intstr.Int, IntVal: 8080}, Path: "/health", Scheme: corev1.URISchemeHTTP}},
								InitialDelaySeconds: 10,
								TimeoutSeconds:      10,
								PeriodSeconds:       10,
							},
							ReadinessProbe: &corev1.Probe{
								Handler:             corev1.Handler{HTTPGet: &corev1.HTTPGetAction{Port: intstr.IntOrString{Type: intstr.Int, IntVal: 8080}, Path: "/health", Scheme: corev1.URISchemeHTTP}},
								InitialDelaySeconds: 10,
								TimeoutSeconds:      10,
								PeriodSeconds:       10,
							},
						},
					},
				},
			},
		},
	}
}

func NewService(app *devopsv1.Static) *corev1.Service {
	selector := map[string]string{"app": controllerName}
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      controllerName,
			Namespace: app.Namespace,
			Labels:    selector,
		},
		Spec: corev1.ServiceSpec{
			Selector: selector,
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Protocol:   corev1.ProtocolTCP,
					Port:       80,
					TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: 8080},
				},
				{
					Name:       "admin",
					Protocol:   corev1.ProtocolTCP,
					Port:       2019,
					TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: 2019},
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}
}

func NewIngress(app *devopsv1.Static) *networkingv1.Ingress {
	ingObj := &networkingv1.Ingress{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "networking.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: app.Namespace,
			Labels: map[string]string{
				"app":  "static",
				"name": app.Name,
			},
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: app.Spec.IngressClassName,
		},
	}
	pathType := networkingv1.PathTypeImplementationSpecific
	for _, host := range app.Spec.Hosts {
		ingObj.Spec.Rules = append(ingObj.Spec.Rules, networkingv1.IngressRule{
			Host: host,
			IngressRuleValue: networkingv1.IngressRuleValue{
				HTTP: &networkingv1.HTTPIngressRuleValue{
					Paths: []networkingv1.HTTPIngressPath{
						{
							Path:     app.Spec.Path,
							PathType: &pathType,
							Backend: networkingv1.IngressBackend{
								ServiceName: controllerName,
								ServicePort: intstr.IntOrString{Type: intstr.Int, IntVal: 80},
								//Service: &networkingv1.IngressServiceBackend{
								//	Name: controllerName,
								//	Port: networkingv1.ServiceBackendPort{
								//		Name:   "http",
								//		Number: 80,
								//	},
								//},
							},
						},
					},
				},
			},
		})
	}
	return ingObj
}

func UpdateExistIngress(ingress *networkingv1.Ingress, instance *devopsv1.Static) (*networkingv1.Ingress, error) {
	var (
		ruleIdx   []interface{}
		updateIdx [][2]int
		noHost    []string
	)
	for idx, rule := range ingress.Spec.Rules {
		for _, host := range instance.Spec.Hosts {
			if rule.Host == host {
				ruleIdx = append(ruleIdx, idx)
			} else {
				ruleIdx = append(ruleIdx, host)
			}
		}
	}
	for _, idx := range ruleIdx {
		if idxNum, ok := idx.(int); ok {
			for pathIdx, path := range ingress.Spec.Rules[idxNum].HTTP.Paths {
				if path.Path == instance.Spec.Path {
					updateIdx = append(updateIdx, [2]int{idxNum, pathIdx})
				} else {
					updateIdx = append(updateIdx, [2]int{idxNum, -1})
				}
			}
		} else {
			noHost = append(noHost, idx.(string))
		}
	}
	pathType := networkingv1.PathTypeImplementationSpecific
	for _, rulePathIdx := range updateIdx {
		if rulePathIdx[0] >= 0 && rulePathIdx[1] >= 0 {
			ingress.Spec.Rules[rulePathIdx[0]].HTTP.Paths[rulePathIdx[1]].Backend.ServiceName = controllerName
			ingress.Spec.Rules[rulePathIdx[0]].HTTP.Paths[rulePathIdx[1]].Backend.ServicePort = intstr.IntOrString{Type: intstr.Int, IntVal: 80}
		} else if rulePathIdx[0] >= 0 && rulePathIdx[1] < 0 {
			ingress.Spec.Rules[rulePathIdx[0]].HTTP.Paths = append(ingress.Spec.Rules[rulePathIdx[0]].HTTP.Paths, networkingv1.HTTPIngressPath{
				Path:     instance.Spec.Path,
				PathType: &pathType,
				Backend: networkingv1.IngressBackend{
					ServiceName: controllerName,
					ServicePort: intstr.IntOrString{Type: intstr.Int, IntVal: 80},
				},
			})
		}
		for _, host := range noHost {
			ingress.Spec.Rules = append(ingress.Spec.Rules, networkingv1.IngressRule{
				Host: host,
				IngressRuleValue: networkingv1.IngressRuleValue{
					HTTP: &networkingv1.HTTPIngressRuleValue{
						Paths: []networkingv1.HTTPIngressPath{
							{
								Path:     instance.Spec.Path,
								PathType: &pathType,
								Backend: networkingv1.IngressBackend{
									ServiceName: controllerName,
									ServicePort: intstr.IntOrString{Type: intstr.Int, IntVal: 80},
								},
							},
						},
					},
				},
			})
		}
	}
	return ingress, nil
}
