/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package v1alpha1

import (
	autoscalingv2beta2 "k8s.io/api/autoscaling/v2beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AutomaticScaling HPA configuration
type AutomaticScaling struct {
	// +optional
	Enable bool `json:"enable,omitempty"`
	// +optional
	ScaleUpWindows int32 `json:"scaleUpWindows,omitempty"`
	// +optional
	ScaleDownWindows int32 `json:"scaleDownWindows,omitempty"`
	// +optional
	Target int32 `json:"target,omitempty"`
	// +optional
	MaxInstance int32 `json:"maxInstance,omitempty"`
	// +optional
	MinInstance int32 `json:"minInstance,omitempty"`
	// +optional
	CustomMetrics []autoscalingv2beta2.MetricSpec `json:"customMetrics,omitempty"`
}

// ProxySpec defines the desired state of ShardingSphereProxy
type ClusterSpec struct {
}

//+kubebuilder:printcolumn:JSONPath=".status.readyNodes",name=ReadyNodes,type=integer
//+kubebuilder:printcolumn:JSONPath=".status.phase",name=Phase,type=string
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Cluster is the Schema for the proxies API
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterSpec   `json:"spec,omitempty"`
	Status ClusterStatus `json:"status,omitempty"`
}

type ClusterStatus struct{}

//+kubebuilder:object:root=true

// ClusterList contains a list of Cluster
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cluster{}, &ClusterList{})
}
