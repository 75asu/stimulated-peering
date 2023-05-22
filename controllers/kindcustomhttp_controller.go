package controllers

import (
	"context"
	"fmt"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	groupspv1 "github.com/measutosh/http-go-operator/api/v1"
)

// KindCustomHttpReconciler reconciles a KindCustomHttp object
type KindCustomHttpReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *KindCustomHttpReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)
	customHttp := &groupspv1.KindCustomHttp{}
	existingHttpDeployment := &appsv1.Deployment{}
	existingHttpService := &corev1.Service{}
	existingConfigMap := &corev1.ConfigMap{}

	log.Info(fmt.Sprintf("⚡️ Event received! ⚡️"))

	if err := r.Get(ctx, req.NamespacedName, customHttp); err != nil {
		if errors.IsNotFound(err) {
			log.Info("KindCustomHttp resource not found")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get KindCustomHttp resource")
		return ctrl.Result{}, err
	}

	// CR deleted : check if  the Deployment and the Service must be deleted
	err := r.Get(ctx, req.NamespacedName, customHttp)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("KindCustomHttp resource not found, check if a Deployment Or Service must be deleted.")
			// }

			// Delete Deployment
			err = r.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, existingHttpDeployment)
			if err != nil {
				if errors.IsNotFound(err) {
					log.Info("Nothing to do, no deployment found.")
					return ctrl.Result{}, nil
				} else {
					log.Error(err, "❌ Failed to get Deployment")
					return ctrl.Result{}, err
				}
			} else {
				log.Info("☠️ Deployment exists: delete it. ☠️")
				r.Delete(ctx, existingHttpDeployment)
			}

			// Delete Service
			err = r.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, existingHttpService)
			if err != nil {
				if errors.IsNotFound(err) {
					log.Info("Nothing to do, no service found.")
					return ctrl.Result{}, nil
				} else {
					log.Error(err, "❌ Failed to get Service")
					return ctrl.Result{}, err
				}
			} else {
				log.Info("☠️ Service exists: delete it. ☠️")
				r.Delete(ctx, existingHttpService)
				return ctrl.Result{}, nil
			}

		}
	} else {

		// Check if the deployment already exists, if not: create a new one.
		err = r.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, existingHttpDeployment)
		if err != nil && errors.IsNotFound(err) {
			// Define a new deployment
			newHttpDeployment := r.createDeployment(customHttp)
			log.Info("✨ Creating a new Deployment", "\nDeployment.Namespace", newHttpDeployment.Namespace, "\nDeployment.Name", newHttpDeployment.Name)

			err = r.Create(ctx, newHttpDeployment)
			if err != nil {
				log.Error(err, "❌ Failed to create new Deployment", "\nDeployment.Namespace", newHttpDeployment.Namespace, "\nDeployment.Name", newHttpDeployment.Name)
				return ctrl.Result{}, err
			}
		} else if err == nil {

			// Deployment exists, check if the Deployment must be updated
			var replicaCount int32 = customHttp.Spec.ReplicaCount
			if *existingHttpDeployment.Spec.Replicas != replicaCount {
				log.Info("🔁 Number of replicas changes, update the deployment! 🔁")
				existingHttpDeployment.Spec.Replicas = &replicaCount
				err = r.Update(ctx, existingHttpDeployment)
				if err != nil {
					log.Error(err, "❌ Failed to update Deployment", "\nDeployment.Namespace", existingHttpDeployment.Namespace, "\nDeployment.Name", existingHttpDeployment.Name)
					return ctrl.Result{}, err
				}
			}
		} else if err != nil {
			log.Error(err, "Failed to get Deployment")
			return ctrl.Result{}, err
		}

		// Check if the service already exists, if not: create a new one
		err = r.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, existingHttpService)
		if err != nil && errors.IsNotFound(err) {
			// Create the Service
			newHttpService := r.createService(customHttp)
			log.Info("✨ Creating a new Service", "\nService.Namespace", newHttpService.Namespace, "\nService.Name", newHttpService.Name)
			err = r.Create(ctx, newHttpService)
			if err != nil {
				log.Error(err, "❌ Failed to create new Service", "\nService.Namespace", newHttpService.Namespace, "\nService.Name", newHttpService.Name)
				return ctrl.Result{}, err
			}
		} else if err == nil {
			// Service exists, check if the port has to be updated.
			var port int32 = customHttp.Spec.Port
			if existingHttpService.Spec.Ports[0].Port != port {
				log.Info("🔁 Port number changes, update the service! 🔁")
				existingHttpService.Spec.Ports[0].Port = port
				err = r.Update(ctx, existingHttpService)
				if err != nil {
					log.Error(err, "❌ Failed to update Service", "\nService.Namespace", existingHttpService.Namespace, "\nService.Name", existingHttpService.Name)
					return ctrl.Result{}, err
				}
			}
		} else if err != nil {
			log.Error(err, "Failed to get Service")
			return ctrl.Result{}, err
		}
	}

	// existingConfigMap := &corev1.ConfigMap{}
	err = r.Get(ctx, types.NamespacedName{Name: req.Name + "-configmap", Namespace: req.Namespace}, existingConfigMap)
	if err != nil && errors.IsNotFound(err) {
		// Create a new configmap
		newConfigMap := r.createConfigMap(customHttp)
		log.Info("✨ Creating a new ConfigMap", "\nConfigMap.Namespace", newConfigMap.Namespace, "\nConfigMap.Name", newConfigMap.Name)
		err = r.Create(ctx, newConfigMap)
		if err != nil {
			log.Error(err, "❌ Failed to create new ConfigMap", "\nConfigMap.Namespace", newConfigMap.Namespace, "\nConfigMap.Name", newConfigMap.Name)
			return ctrl.Result{}, err
		}
	} else if err == nil {
		// ConfigMap exists, update the data if it has changed
		if !reflect.DeepEqual(existingConfigMap.Data, customHttp.Spec.ConfigMapData) {
			log.Info("🔄 ConfigMap data changes, update the ConfigMap! 🔄")
			existingConfigMap.Data = customHttp.Spec.ConfigMapData
			err = r.Update(ctx, existingConfigMap)
			if err != nil {
				log.Error(err, "❌ Failed to update ConfigMap", "\nConfigMap.Namespace", existingConfigMap.Namespace, "\nConfigMap.Name", existingConfigMap.Name)
				return ctrl.Result{}, err
			}
		}
	} else if err != nil {
		log.Error(err, "Failed to get ConfigMap")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *KindCustomHttpReconciler) createConfigMap(customHttpCR *groupspv1.KindCustomHttp) *corev1.ConfigMap {
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      customHttpCR.Name + "-configmap",
			Namespace: customHttpCR.Namespace,
		},
		Data: customHttpCR.Spec.ConfigMapData,
	}
	return configMap
}

func (r *KindCustomHttpReconciler) createDeployment(customHttpCR *groupspv1.KindCustomHttp) *appsv1.Deployment {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      customHttpCR.Name,
			Namespace: customHttpCR.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &customHttpCR.Spec.ReplicaCount,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "measutosh-http-server"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "measutosh-http-server"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: "measutosh/http-server",
						Name:  "http-go-server-pod",
						Ports: []corev1.ContainerPort{{
							ContainerPort: 80,
							Name:          "http",
							Protocol:      "TCP",
						}},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "config-volume",
								MountPath: "/config",
							},
						},
					}},
					Volumes: []corev1.Volume{
						{
							Name: "config-volume",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: customHttpCR.Name + "-configmap",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	return deployment
}

// Create a Service for the HTTP server.
func (r *KindCustomHttpReconciler) createService(customHttpCR *groupspv1.KindCustomHttp) *corev1.Service {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      customHttpCR.Name,
			Namespace: customHttpCR.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": "measutosh-http-server",
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Protocol:   corev1.ProtocolTCP,
					Port:       customHttpCR.Spec.Port,
					TargetPort: intstr.FromInt(80),
				},
			},
			ClusterIP: corev1.ClusterIPNone,
		},
	}

	return service
}

func (r *KindCustomHttpReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&groupspv1.KindCustomHttp{}).
		Watches(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForObject{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(e event.CreateEvent) bool {
				return false
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				return false
			},
			UpdateFunc: func(e event.UpdateEvent) bool {
				return true
			},
		})).
		Owns(&corev1.ConfigMap{}). // Add this line to watch and own ConfigMaps
		Complete(r)
}
