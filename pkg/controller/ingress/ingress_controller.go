package ingress

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type DeploymentValues struct {
	Name      string `json:"name"`
	NameSpace string `json:"namespace"`
	Hosts     string `json:"hosts"`
}

type GatewayValues struct {
	DeploymentValues
	CredentialName string `json:"credentialName"`
}

type VSValues struct {
	DeploymentValues
	Number   []int    `json:"number"`
	Host     []string `json:"host"`
	Prefix   []string `json:"prefix"`
	Gateways string   `json:"gateways"`
}

var log = logf.Log.WithName("controller_ingress")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Ingress Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileIngress{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	fmt.Printf("123456")
	// Create a new controller
	c, err := controller.New("ingress-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Ingress
	err = c.Watch(&source.Kind{Type: &extensionsv1beta1.Ingress{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Ingress
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &extensionsv1beta1.Ingress{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileIngress implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileIngress{}

// ReconcileIngress reconciles a Ingress object
type ReconcileIngress struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Ingress object and makes changes based on the state read
// and what is in the Ingress.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileIngress) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Ingress")
	// var tmpIngress string
	// var ingress *unstructured.Unstructured

	// Fetch the Ingress instance
	instance := &extensionsv1beta1.Ingress{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	// ---------------------------------------
	var gw GatewayValues
	var vs VSValues
	ingress := &unstructured.Unstructured{}
	ingress.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "extensions",
		Kind:    "ingresses",
		Version: "v1beta1",
	})
	ingressList := &unstructured.UnstructuredList{}
	ingressList.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "extensions",
		Kind:    "Ingress",
		Version: "v1beta1",
	})
	_ = r.client.List(context.Background(), ingressList)
	// reqLogger.Info(ingressList, "TESTE 1")
	// fmt.Printf("TESTE8 %s", ingressList)
	for i := 0; i < len(ingressList.Items); i++ {
		gw.Name = fmt.Sprintf("%s", ingressList.Items[i].GetName())
		gw.NameSpace = fmt.Sprintf("%s", ingressList.Items[i].GetNamespace())
		tmpSpec := ingressList.Items[i].Object["spec"]
		tmpRules := tmpSpec.(map[string]interface{})["rules"]
		qtdRules := len(tmpRules.([]interface{}))
		fmt.Printf("TESTE_INGRESS_NAME %s\n", gw.Name)
		fmt.Printf("TESTE_INGRESS_NAMESPACE %s\n", gw.NameSpace)
		fmt.Printf("TESTE_INGRESS_QTD_RULES %d\n", qtdRules)
		_ = r.setGateway(gw)

		for j := 0; j < qtdRules; j++ {
			gw.Hosts = fmt.Sprintf("%s", tmpRules.([]interface{})[j].(map[string]interface{})["host"])
			tmpPaths := tmpRules.([]interface{})[j].(map[string]interface{})["http"].(map[string]interface{})["paths"]
			qtdPaths := len(tmpPaths.([]interface{}))
			fmt.Printf("TESTE_HOST %s\n", gw.Hosts)
			fmt.Printf("TESTE_INGRESS_QTD_PATHS %d\n", qtdPaths)
			vs.Host = make([]string, qtdPaths)
			vs.Prefix = make([]string, qtdPaths)
			vs.Number = make([]int, qtdPaths)
			vs.Prefix = make([]string, qtdPaths)
			vs.Name = gw.Name
			vs.NameSpace = gw.NameSpace
			vs.Gateways = gw.Name
			for z := 0; z < qtdPaths; z++ {
				vs.Prefix[z] = fmt.Sprintf("%s", tmpPaths.([]interface{})[z].(map[string]interface{})["path"])
				tmpBackend := tmpPaths.([]interface{})[z].(map[string]interface{})["backend"]
				vs.Host[z] = fmt.Sprintf("%s", tmpBackend.(map[string]interface{})["serviceName"])
				vs.Number[z], _ = strconv.Atoi(fmt.Sprintf("%d", tmpBackend.(map[string]interface{})["servicePort"]))
				fmt.Printf("TESTE_PATH %s\n", vs.Prefix[z])
				fmt.Printf("TESTE_SERVICENAME %s\n", vs.Host[z])
				fmt.Printf("TESTE_SERVICEPORT %d\n", vs.Number[z])
			}
		}
		r.setVirtualService(vs)
	}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, ingress)
	reqLogger.Error(err, "TESTE 1")
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Error(err, "TESTE")
		// Deployment created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	}
	// ---------------------------------------
	// Define a new Pod object
	pod := newPodForCR(instance)

	// Set Ingress instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, pod, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Pod already exists
	found := &corev1.Pod{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
		err = r.client.Create(context.TODO(), pod)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Pod created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Pod already exists - don't requeue
	reqLogger.Info("Skip reconcile: Pod already exists", "Pod.Namespace", found.Namespace, "Pod.Name", found.Name)
	return reconcile.Result{}, nil
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *extensionsv1beta1.Ingress) *corev1.Pod {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-pod",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "busybox",
					Image:   "busybox",
					Command: []string{"sleep", "3600"},
				},
			},
		},
	}
}

func (r *ReconcileIngress) setGateway(gw GatewayValues) error {
	dep := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "networking.istio.io/v1alpha3",
			"kind":       "Gateway",
			"metadata": map[string]interface{}{
				"name":      gw.Name,
				"namespace": gw.NameSpace,
			},
			"spec": map[string]interface{}{
				"selector": map[string]interface{}{
					"istio": "ingressgateway",
				},
				"servers": []map[string]interface{}{
					{},
				},
			},
		},
	}

	// 	dep.Object["spec"].(map[string]interface{})["servers"].([]map[string]interface{})[0]["port"].(map[string]interface{})["number"] = 80
	// 	tmpTT := dep.Object["spec"].(map[string]interface{})["servers"].([]map[string]interface{})[0]
	// 	dep.Object["spec"].(map[string]interface{})["servers"] = []map[string]interface{}{
	// 		map[string]interface{}{
	// 			"port": map[string]interface{}{
	// 				"number":   443,
	// 				"name":     "https-" + gw.Name,
	// 				"protocol": "HTTPS",
	// 			},

	// 			"tls": map[string]interface{}{
	// 				"mode":           "SIMPLE",
	// 				"credentialName": gw.CredentialName,
	// 			},
	// 			"hosts": []interface{}{
	// 				gw.Hosts,
	// 			},
	// 		},
	// 		map[string]interface{}{
	// 			"port": map[string]interface{}{
	// 				"number":   443,
	// 				"name":     "https-" + gw.Name,
	// 				"protocol": "HTTPS",
	// 			},

	// 			"tls": map[string]interface{}{
	// 				"mode":           "SIMPLE",
	// 				"credentialName": gw.CredentialName,
	// 			},
	// 			"hosts": []interface{}{
	// 				gw.Hosts,
	// 			},
	// 		},
	// 	}
	// 	dep.Object["spec"].(map[string]interface{})["servers"] = tmpTT

	serverHTTPS := map[string]interface{}{
		"port": map[string]interface{}{
			"number":   443,
			"name":     "https-" + gw.Name,
			"protocol": "HTTPS",
		},
		"tls": map[string]interface{}{
			"mode":           "SIMPLE",
			"credentialName": gw.CredentialName,
		},
		"hosts": []interface{}{
			gw.Hosts,
		},
	}
	serverHTTP := map[string]interface{}{
		"port": map[string]interface{}{
			"number":   80,
			"name":     "http-" + gw.Name,
			"protocol": "HTTP",
		},
		"hosts": []interface{}{
			gw.Hosts,
		},
	}
	if gw.CredentialName != "" {
		dep.Object["spec"].(map[string]interface{})["servers"] = serverHTTPS
	} else {
		dep.Object["spec"].(map[string]interface{})["servers"] = serverHTTP
	}
	// result := []map[string]interface{}{}
	// result = []map[string]interface{}{tmpTT, tmpTT}
	fmt.Printf("TESTE_TT %s\n", serverHTTP)
	fmt.Printf("TESTE_TT2 %s\n", serverHTTPS)
	// fmt.Printf("TESTE_MP3 %s\n", result)
	// dep.Object["spec"].(map[string]interface{})["servers"] = result
	depOut, _ := json.MarshalIndent(dep.Object, "", "  ")
	fmt.Printf("\n\n----------------------------------------------------------\n")
	fmt.Printf("%s\n", depOut)
	return nil
}

func (r *ReconcileIngress) setVirtualService(vs VSValues) error {
	dep := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "networking.istio.io/v1alpha3",
			"kind":       "VirtualService",
			"metadata": map[string]interface{}{
				"name":      vs.Name,
				"namespace": vs.NameSpace,
			},
			"spec": map[string]interface{}{
				"hosts": []interface{}{
					vs.Hosts,
				},
				"gateways": []interface{}{
					vs.Gateways,
				},
				"http": []map[string]interface{}{
					{
						"match": []map[string]interface{}{
							{
								"uri": map[string]interface{}{
									"prefix": vs.Prefix[0],
								},
							},
						},
						"route": []map[string]interface{}{
							{
								"destination": map[string]interface{}{
									"port": map[string]interface{}{
										"number": vs.Number[0],
									},
									"host": vs.Host[0],
								},
							},
						},
					},
				},
			},
		},
	}
	result := []map[string]interface{}{}
	match := map[string]interface{}{}
	match = map[string]interface{}{
		"uri": map[string]interface{}{
			"prefix": vs.Prefix[0],
		},
	}
	for i := 0; i < len(vs.Prefix); i++ {
		match = map[string]interface{}{
			"uri": map[string]interface{}{
				"prefix": vs.Prefix[i],
			},
		}
		result = append(result, match)
		// result[i] = match
		// matchOut, _ := json.MarshalIndent(match, "", "  ")

		// tmpResult := result[0]
		// fmt.Printf("TESTE_TMP_RESULT %s\n", tmpResult)
		// result = []map[string]interface{}{tmpResult, match}
	}
	fmt.Printf("TESTE_RESULT %s\n", result)
	dep.Object["spec"].(map[string]interface{})["http"].([]map[string]interface{})[0]["match"] = result
	depOut, _ := json.MarshalIndent(dep.Object, "", "  ")
	fmt.Printf("\n\n----------------------------------------------------------\n")
	fmt.Printf("%s\n", depOut)
	return nil
}
