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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// StorageNodeList contains a list of StorageNode
type StorageNodeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StorageNode `json:"items"`
}

// +kubebuilder:printcolumn:JSONPath=".status.endpoints[*].address",name="Address",type=string
// +kubebuilder:printcolumn:JSONPath=".status.endpoints[*].port",name="Port",type=integer
// +kubebuilder:printcolumn:JSONPath=".status.endpoints[*].status",name="Status",type=string
// +kubebuilder:printcolumn:JSONPath=".status.endpoints[*].arn",name="ARN",type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// StorageNode is the Schema for the ShardingSphere Proxy API
type StorageNode struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec StorageNodeSpec `json:"spec,omitempty"`
	// +optional
	Status StorageNodeStatus `json:"status,omitempty"`
}

type StorageNodeSpec struct {
	DatabaseClassName string `json:"databaseClassName,omitempty"`
}

type StorageNodeStatus struct {
	// +optional
	Endpoints []Endpoint `json:"endpoints,omitempty"`
}

type Endpoint struct {
	Arn      string `json:"arn"`
	Protocol string `json:"protocol"`
	Address  string `json:"address"`
	Port     uint32 `json:"port"`
	Status   string `json:"status"`
	User     string `json:"user"`
	//FIXME
	Pass string `json:"pass"`
}

func init() {
	SchemeBuilder.Register(&StorageNode{}, &StorageNodeList{})
}
