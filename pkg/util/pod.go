/*
 Copyright 2023 The Kapacity Authors.
 Copyright 2014 The Kubernetes Authors.

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

package util

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetPodNames generate a list of pod names from a list of pod objects.
func GetPodNames(pods []*corev1.Pod) []string {
	result := make([]string, 0, len(pods))
	for _, pod := range pods {
		result = append(result, pod.Name)
	}
	return result
}

// IsPodRunning returns if the given pod's phase is running and is not being deleted.
func IsPodRunning(pod *corev1.Pod) bool {
	return pod.DeletionTimestamp.IsZero() && pod.Status.Phase == corev1.PodRunning
}

// IsPodActive returns if the given pod has not terminated.
func IsPodActive(pod *corev1.Pod) bool {
	return pod.Status.Phase != corev1.PodSucceeded &&
		pod.Status.Phase != corev1.PodFailed &&
		pod.DeletionTimestamp.IsZero()
}

// IsPodReady returns true if a pod is ready; false otherwise.
func IsPodReady(pod *corev1.Pod) bool {
	return IsPodReadyConditionTrue(pod.Status)
}

// IsPodReadyConditionTrue returns true if a pod is ready; false otherwise.
func IsPodReadyConditionTrue(status corev1.PodStatus) bool {
	condition := GetPodReadyCondition(status)
	return condition != nil && condition.Status == corev1.ConditionTrue
}

// GetPodReadyCondition extracts the pod ready condition from the given status and returns that.
// Returns nil if the condition is not present.
func GetPodReadyCondition(status corev1.PodStatus) *corev1.PodCondition {
	_, condition := GetPodCondition(&status, corev1.PodReady)
	return condition
}

// GetPodCondition extracts the provided condition from the given status and returns that.
// Returns nil and -1 if the condition is not present, and the index of the located condition.
func GetPodCondition(status *corev1.PodStatus, conditionType corev1.PodConditionType) (int, *corev1.PodCondition) {
	if status == nil {
		return -1, nil
	}
	return GetPodConditionFromList(status.Conditions, conditionType)
}

// GetPodConditionFromList extracts the provided condition from the given list of condition and
// returns the index of the condition and the condition. Returns -1 and nil if the condition is not present.
func GetPodConditionFromList(conditions []corev1.PodCondition, conditionType corev1.PodConditionType) (int, *corev1.PodCondition) {
	if conditions == nil {
		return -1, nil
	}
	for i := range conditions {
		if conditions[i].Type == conditionType {
			return i, &conditions[i]
		}
	}
	return -1, nil
}

// UpdatePodCondition updates existing pod condition or creates a new one. Sets LastTransitionTime to now if the
// status has changed.
// Returns true if pod condition has changed or has been added.
func UpdatePodCondition(status *corev1.PodStatus, condition *corev1.PodCondition) bool {
	condition.LastTransitionTime = metav1.Now()
	// Try to find this pod condition.
	conditionIndex, oldCondition := GetPodCondition(status, condition.Type)

	if oldCondition == nil {
		// We are adding new pod condition.
		status.Conditions = append(status.Conditions, *condition)
		return true
	}
	// We are updating an existing condition, so we need to check if it has changed.
	if condition.Status == oldCondition.Status {
		condition.LastTransitionTime = oldCondition.LastTransitionTime
	}

	isEqual := condition.Status == oldCondition.Status &&
		condition.Reason == oldCondition.Reason &&
		condition.Message == oldCondition.Message &&
		condition.LastProbeTime.Equal(&oldCondition.LastProbeTime) &&
		condition.LastTransitionTime.Equal(&oldCondition.LastTransitionTime)

	status.Conditions[conditionIndex] = *condition
	// Return true if one of the fields have changed.
	return !isEqual
}

// AddPodCondition adds a pod condition if not exists. Sets LastTransitionTime to now if not exists.
// Returns true if pod condition has been added.
func AddPodCondition(status *corev1.PodStatus, condition *corev1.PodCondition) bool {
	if _, oldCondition := GetPodCondition(status, condition.Type); oldCondition != nil {
		return false
	}
	condition.LastTransitionTime = metav1.Now()
	status.Conditions = append(status.Conditions, *condition)
	return true
}

// AddPodReadinessGate adds the provided condition to the pod's readiness gates.
// Returns true if the readiness gate has been added.
func AddPodReadinessGate(spec *corev1.PodSpec, conditionType corev1.PodConditionType) bool {
	for _, rg := range spec.ReadinessGates {
		if rg.ConditionType == conditionType {
			return false
		}
	}
	spec.ReadinessGates = append(spec.ReadinessGates, corev1.PodReadinessGate{ConditionType: conditionType})
	return true
}
