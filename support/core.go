/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package support

import (
	"encoding/json"
	"io"

	"github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func CreateConfigMap(t Test, namespace string, content map[string][]byte) *corev1.ConfigMap {
	t.T().Helper()

	configMap := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "config-",
			Namespace:    namespace,
		},
		BinaryData: content,
		Immutable:  Ptr(true),
	}

	configMap, err := t.Client().Core().CoreV1().ConfigMaps(namespace).Create(t.Ctx(), configMap, metav1.CreateOptions{})
	t.Expect(err).NotTo(gomega.HaveOccurred())
	t.T().Logf("Created ConfigMap %s/%s successfully", configMap.Namespace, configMap.Name)

	return configMap
}

func Raw(t Test, obj runtime.Object) runtime.RawExtension {
	t.T().Helper()
	data, err := json.Marshal(obj)
	t.Expect(err).NotTo(gomega.HaveOccurred())
	return runtime.RawExtension{
		Raw: data,
	}
}

func GetPods(t Test, namespace string, options metav1.ListOptions) []corev1.Pod {
	t.T().Helper()
	pods, err := t.Client().Core().CoreV1().Pods(namespace).List(t.Ctx(), options)
	t.Expect(err).NotTo(gomega.HaveOccurred())
	return pods.Items
}

func GetPodLogs(t Test, pod *corev1.Pod, options corev1.PodLogOptions) []byte {
	t.T().Helper()
	stream, err := t.Client().Core().CoreV1().Pods(pod.GetNamespace()).GetLogs(pod.GetName(), &options).Stream(t.Ctx())
	t.Expect(err).NotTo(gomega.HaveOccurred())

	defer func() {
		t.Expect(stream.Close()).To(gomega.Succeed())
	}()

	bytes, err := io.ReadAll(stream)
	t.Expect(err).NotTo(gomega.HaveOccurred())

	return bytes
}

func storeAllPodLogs(t Test, namespace *corev1.Namespace) {
	t.T().Helper()

	pods, err := t.Client().Core().CoreV1().Pods(namespace.Name).List(t.Ctx(), metav1.ListOptions{})
	t.Expect(err).NotTo(gomega.HaveOccurred())

	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			t.T().Logf("Retrieving Pod Container %s/%s/%s logs", pod.Namespace, pod.Name, container.Name)
			storeContainerLog(t, namespace, pod.Name, container.Name)
		}
	}
}

func storeContainerLog(t Test, namespace *corev1.Namespace, podName, containerName string) {
	t.T().Helper()

	options := corev1.PodLogOptions{Container: containerName}
	stream, err := t.Client().Core().CoreV1().Pods(namespace.Name).GetLogs(podName, &options).Stream(t.Ctx())
	if err != nil {
		t.T().Logf("Failed to retrieve logs for Pod Container %s/%s/%s, logs cannot be stored", namespace.Name, podName, containerName)
		return
	}

	defer func() {
		t.Expect(stream.Close()).To(gomega.Succeed())
	}()

	bytes, err := io.ReadAll(stream)
	t.Expect(err).NotTo(gomega.HaveOccurred())

	containerLogFileName := "pod-" + podName + "-" + containerName
	WriteToOutputDir(t, containerLogFileName, Log, bytes)
}

func CreateServiceAccount(t Test, namespace string) *corev1.ServiceAccount {
	t.T().Helper()

	serviceAccount := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "sa-",
			Namespace:    namespace,
		},
	}
	serviceAccount, err := t.Client().Core().CoreV1().ServiceAccounts(namespace).Create(t.Ctx(), serviceAccount, metav1.CreateOptions{})
	t.Expect(err).NotTo(gomega.HaveOccurred())
	t.T().Logf("Created ServiceAccount %s/%s successfully", serviceAccount.Namespace, serviceAccount.Name)

	return serviceAccount
}

func CreatePersistentVolumeClaim(t Test, namespace string, storageSize string, accessMode ...corev1.PersistentVolumeAccessMode) *corev1.PersistentVolumeClaim {
	t.T().Helper()

	pvc := &corev1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "PersistentVolumeClaim",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "pvc-",
			Namespace:    namespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: accessMode,
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(storageSize),
				},
			},
		},
	}
	pvc, err := t.Client().Core().CoreV1().PersistentVolumeClaims(namespace).Create(t.Ctx(), pvc, metav1.CreateOptions{})
	t.Expect(err).NotTo(gomega.HaveOccurred())
	t.T().Logf("Created PersistentVolumeClaim %s/%s successfully", pvc.Namespace, pvc.Name)

	return pvc
}

func GetNodes(t Test) []corev1.Node {
	t.T().Helper()
	nodes, err := t.Client().Core().CoreV1().Nodes().List(t.Ctx(), metav1.ListOptions{})
	t.Expect(err).NotTo(gomega.HaveOccurred())
	return nodes.Items
}

func GetNodeInternalIP(t Test, node corev1.Node) (IP string) {
	t.T().Helper()

	for _, address := range node.Status.Addresses {
		if address.Type == "InternalIP" {
			IP = address.Address
		}
	}
	t.Expect(IP).Should(gomega.Not(gomega.BeEmpty()), "Node internal IP address not found")

	return
}
