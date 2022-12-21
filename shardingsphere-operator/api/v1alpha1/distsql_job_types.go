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

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type Credential struct {
	Host     string `json:"host"`
	Port     int32  `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Schema   string `json:"schema,omitempty"`
}

type DistSQLJobSpec struct {
	Credential Credential `json:"credential"`
	Scripts    []string   `json:"scripts"`
}

type DistSQLJobStatus struct{}

// +kubebuilder:object:root=true
type DistSQLJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DistSQLJobSpec   `json:"spec,omitempty"`
	Status DistSQLJobStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type DistSQLJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DistSQLJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DistSQLJob{}, &DistSQLJobList{})
}
