package controllers

import (
	"context"
	"reflect"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	chatv1alpha1 "github.com/hown3d/chat-operator/api/v1alpha1"
)

// +kubebuilder:docs-gen:collapse=Imports

/*
The first step to writing a simple integration test is to actually create an instance of Rocket you can run tests against.
Note that to create a Rocket, you’ll need to create a stub Rocket struct that contains your Rocket’s specifications.
Note that when we create a stub Rocket, the Rocket also needs stubs of its required downstream objects.
Without the stubbed Job template spec and the Pod template spec below, the Kubernetes API will not be able to
create the Rocket.
*/
var _ = Describe("Rocket controller", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		RocketName      = "test-rocket"
		RocketNamespace = "default"
		PodName         = "test-pod"

		timeout  = time.Second * 20
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When updating Rocket Status", func() {
		It("Should update Rocket Status.Pods list when new Pods are created", func() {
			By("By creating a new Rocket")
			ctx := context.Background()
			rocket := &chatv1alpha1.Rocket{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "chat.accso.de/v1alpha1",
					Kind:       "Rocket",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      RocketName,
					Namespace: RocketNamespace,
				},
				Spec: chatv1alpha1.RocketSpec{Replicas: 1},
			}
			Expect(k8sClient.Create(ctx, rocket)).Should(Succeed())

			/*
				After creating this Rocket, let's check that the Rocket's Spec fields match what we passed in.
				Note that, because the k8s apiserver may not have finished creating a Rocket after our `Create()` call from earlier, we will use Gomega’s Eventually() testing function instead of Expect() to give the apiserver an opportunity to finish creating our Rocket.
				`Eventually()` will repeatedly run the function provided as an argument every interval seconds until
				(a) the function’s output matches what’s expected in the subsequent `Should()` call, or
				(b) the number of attempts * interval period exceed the provided timeout value.
				In the examples below, timeout and interval are Go Duration values of our choosing.
			*/

			rocketLookupKey := types.NamespacedName{Name: RocketName, Namespace: RocketNamespace}
			createdRocket := &chatv1alpha1.Rocket{}

			// We'll need to retry getting this newly created Rocket, given that creation may not immediately happen.
			Eventually(func() bool {
				err := k8sClient.Get(ctx, rocketLookupKey, createdRocket)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
			// Let's make sure our Schedule string value was properly converted/handled.
			Expect(createdRocket.Spec.Replicas).Should(Equal(int32(1)))
			/*
				Now that we've created a Rocket in our test cluster, the next step is to write a test that actually tests our Rocket controller’s behavior.
				Let’s test the Rocket controller’s logic responsible for updating Rocket.Status.Pods with actively running Pods.
				First, we should get the test Rocket we created earlier, and verify that it currently does not have any active Pods.
				We use Gomega's `Consistently()` check here to ensure that the active Pods remain empty over a duration of time.
			*/
			By("By checking the Rocket has zero Pods")
			Consistently(func() (int, error) {
				err := k8sClient.Get(ctx, rocketLookupKey, createdRocket)
				if err != nil {
					return -1, err
				}
				return len(createdRocket.Status.Pods), nil
			}, duration, interval).Should(Equal(0))
			/*
				Next, we actually create a stubbed Job that will belong to our Rocket, as well as its downstream template specs.
				We set the Job's status's "Active" count to 2 to simulate the Job running two pods, which means the Job is actively running.
				We then take the stubbed Job and set its owner reference to point to our test Rocket.
				This ensures that the test Job belongs to, and is tracked by, our test Rocket.
				Once that’s done, we create our new Job instance.
			*/
			By("By creating a new Job")
			labels := labelsForRocket(RocketName)
			testPod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      PodName,
					Namespace: RocketNamespace,
					Labels:    labels,
				},
				Spec: corev1.PodSpec{
					// For simplicity, we only fill out the required fields.
					Containers: []corev1.Container{
						{
							Name:  "test-container",
							Image: "test-image",
						},
					},
					RestartPolicy: corev1.RestartPolicyOnFailure,
				},
				Status: corev1.PodStatus{},
			}

			// Note that your Rocket’s GroupVersionKind is required to set up this owner reference.
			kind := reflect.TypeOf(chatv1alpha1.Rocket{}).Name()
			groupVersionKind := chatv1alpha1.GroupVersion.WithKind(kind)

			controllerRef := metav1.NewControllerRef(createdRocket, groupVersionKind)
			testPod.SetOwnerReferences([]metav1.OwnerReference{*controllerRef})
			Expect(k8sClient.Create(ctx, testPod)).Should(Succeed())
			/*
				Adding this Job to our test Rocket should trigger our controller’s reconciler logic.
				After that, we can write a test that evaluates whether our controller eventually updates our Rocket’s Status field as expected!
			*/
			By("By checking that the Rocket has one pod")
			Eventually(func() ([]string, error) {
				err := k8sClient.Get(ctx, rocketLookupKey, createdRocket)
				if err != nil {
					return nil, err
				}

				names := []string{}
				for _, pod := range createdRocket.Status.Pods {
					names = append(names, pod)
				}
				return names, nil
			}, timeout, interval).Should(ConsistOf(PodName), "should list our pod %s in the pods list in status", PodName)
		})
	})

})

/*
	After writing all this code, you can run `go test ./...` in your `controllers/` directory again to run your new test!
*/
