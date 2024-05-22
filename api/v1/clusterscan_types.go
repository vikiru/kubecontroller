/*
Copyright 2024.

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

package v1

import (
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ClusterJobType defines the type of job
type ClusterJobType string

const (
	// CronJob indicates that the ClusterScan (CRD) is a CronJob
	CronJob ClusterJobType = "CronJob"

	// Job indicates that the ClusterScan(CRD) is a Job
	Job ClusterJobType = "Job"
)

// ClusterScanSpec defines the desired state of ClusterScan
type ClusterScanSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Specifies whether this is a recurring job or not
	Recurring bool `json:"recurring"`

	// The schedule in Cron format, see https://en.wikipedia.org/wiki/Cron.
	Schedule string `json:"schedule"`

	// Specifies the job that will be created when executing a CronJob.
	JobTemplate batchv1.JobTemplateSpec `json:"jobTemplate"`
}

// ClusterScanStatus defines the observed state of ClusterScan
type ClusterScanStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Information when was the last time the job was successfully scheduled.
	// +optional
	LastScheduleTime *metav1.Time `json:"lastScheduleTime,omitempty"`

	// A detailed status message indicating the progress/completion status of the job.
	StatusMessage string `json:"statusMessage,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ClusterScan is the Schema for the clusterscans API
type ClusterScan struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterScanSpec   `json:"spec,omitempty"`
	Status ClusterScanStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ClusterScanList contains a list of ClusterScan
type ClusterScanList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterScan `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterScan{}, &ClusterScanList{})
}

//+kubebuilder:docs-gen:collapse=Root Object Definitions
