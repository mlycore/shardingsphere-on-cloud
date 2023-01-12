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
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// ComputeNodeList contains a list of ComputeNode
type ComputeNodeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ComputeNode `json:"items"`
}

// +kubebuilder:printcolumn:JSONPath=".status.readyInstances",name=ReadyInstances,type=integer
// +kubebuilder:printcolumn:JSONPath=".status.phase",name=Phase,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name="Age",type="date"
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// ComputeNode is the Schema for the ShardingSphere Proxy API
type ComputeNode struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ComputeNodeSpec `json:"spec,omitempty"`
	// +optional
	Status ComputeNodeStatus `json:"status,omitempty"`
}

type PrivilegeType string

const (
	AllPermitted PrivilegeType = "ALL_PERMITTED"
)

// ComputeNodePrivilege for storage node, the default value is ALL_PERMITTED
type ComputeNodePrivilege struct {
	Type PrivilegeType `json:"type"`
}

// ComputeNodeUser is a slice about authorized host and password for compute node.
// Format:
// user:<username>@<hostname>,hostname is % or empty string means do not care about authorized host
// password:<password>
type ComputeNodeUser struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

// ComputeNodeAuth  is used to set up initial user to login compute node, and authority data of storage node.
type ComputeNodeAuthority struct {
	Users []ComputeNodeUser `json:"users"`
	// +optional
	Privilege ComputeNodePrivilege `json:"privilege"`
}

type RepositoryType string

const (
	RepositoryTypeZookeeper RepositoryType = "ZooKeeper"
	RepositoryTypeEtcd      RepositoryType = "Etcd"
)

// Repository is the metadata persistent store for ShardingSphere
type Repository struct {
	// +kubebuilder:validation:Enum=ZooKeeper;Etcd
	// type of metadata repository
	Type RepositoryType `json:"type"`
	// properties of metadata repository
	// +optional
	Props ClusterProps `json:"props"`
}

// ComputeNodeClustersProps is the properties of a ShardingSphere Cluster
type ComputeNodeClusterProps struct {
	// namespace of registry center
	Namespace string `json:"namespace" yaml:"namespace"`
	// server lists of registry center
	ServerLists string `json:"server-lists" yaml:"server-lists"`
	// retryIntervalMilliseconds Milliseconds of retry interval. default: 500
	// +optional
	RetryIntervalMilliseconds int `json:"retryIntervalMilliseconds,omitempty" yaml:"retryIntervalMilliseconds,omitempty"`
	// the max retries of client connection. default: 3
	// +optional
	MaxRetries int `json:"maxRetries,omitempty" yaml:"maxRetries,omitempty"`
	// the seconds of ephemeral data live. default: 60
	// +optional
	TimeToLiveSeconds int `json:"timeToLiveSeconds,omitempty" yaml:"timeToLiveSeconds,omitempty"`
	// the milliseconds of operation timeout. default: 500
	// +optional
	OperationTimeoutMilliseconds int `json:"operationTimeoutMilliseconds,omitempty" yaml:"operationTimeoutMilliseconds,omitempty"`
	// password of login
	// +optional
	Digest string `json:"digest,omitempty" yaml:"digest,omitempty"`
}

type ModeType string

const (
	ModeTypeCluster    ModeType = "Cluster"
	ModeTypeStandalone ModeType = "Standalone"
)

// ComputeNodeProps is which Apache ShardingSphere provides the way of property configuration to configure system level configuration.
type ComputeNodeProps struct {
	// the max thread size of worker group to execute SQL. One ShardingSphereDataSource will use a independent thread pool, it does not share thread pool even different data source in same JVM.
	// +optional
	KernelExecutorSize int `json:"kernel-executor-size,omitempty" yaml:"kernel-executor-size,omitempty"`
	// whether validate table meta data consistency when application startup or updated.
	// +optional
	CheckTableMetadataEnabled bool `json:"check-table-metadata-enabled,omitempty" yaml:"check-table-metadata-enabled,omitempty"`
	// ShardingSphere Proxy backend query fetch size. A larger value may increase the memory usage of ShardingSphere ShardingSphereProxy. The default value is -1, which means set the minimum value for different JDBC drivers.
	// +optional
	ProxyBackendQueryFetchSize int `json:"proxy-backend-query-fetch-size,omitempty" yaml:"proxy-backend-query-fetch-size,omitempty"`
	// whether validate duplicate table when application startup or updated.
	// +optional
	CheckDuplicateTableEnabled bool `json:"check-duplicate-table-enabled,omitempty" yaml:"check-duplicate-table-enabled,omitempty"`
	// ShardingSphere Proxy frontend Netty executor size. The default value is 0, which means let Netty decide.
	// +optional
	ProxyFrontendExecutorSize int `json:"proxy-frontend-executor-size,omitempty" yaml:"proxy-frontend-executor-size,omitempty"`
	// available options of proxy backend executor suitable: OLAP(default), OLTP. The OLTP option may reduce time cost of writing packets to client, but it may increase the latency of SQL execution and block other clients if client connections are more than proxy-frontend-executor-size, especially executing slow SQL.
	// +optional
	ProxyBackendExecutorSuitable string `json:"proxy-backend-executor-suitable,omitempty" yaml:"proxy-backend-executor-suitable,omitempty"`
	// +optional
	ProxyBackendDriverType string `json:"proxy-backend-driver-type,omitempty" yaml:"proxy-backend-driver-type,omitempty"`
	// +optional
	ProxyFrontendDatabaseProtocolType string `json:"proxy-frontend-database-protocol-type" yaml:"proxy-frontend-database-protocol-type,omitempty"`
}

// ComputeNodeServerMode is the mode for ShardingSphere Proxy
type ComputeNodeServerMode struct {
	// +optional
	Repository Repository `json:"repository"`
	Type       ModeType   `json:"type"`
}

// ServerConfig defines the bootstrap config for a ShardingSphere Proxy
type ServerConfig struct {
	Authority ComputeNodeAuthority  `json:"authority"`
	Mode      ComputeNodeServerMode `json:"mode"`
	//+optional
	Props *ComputeNodeProps `json:"props"`
}

// LogbackConfig contains contents of the expected logback.xml
type LogbackConfig string

// ServiceType defines the Service in Kubernetes of ShardingSphere-Proxy
type Service struct {
	Ports []corev1.ServicePort `json:"ports,omitempty"`
	// +kubebuilder:validation:Enum=ClusterIP;NodePort;LoadBalancer;ExternalName
	Type corev1.ServiceType `json:"type"`
}

// ProxyProbe defines the probe actions for LivenesProbe, ReadinessProbe and StartupProbe
type ProxyProbe struct {
	// Probes are not allowed for ephemeral containers.
	// +optional
	LivenessProbe *corev1.Probe `json:"livenessProbe,omitempty"`
	// Probes are not allowed for ephemeral containers.
	// +optional
	ReadinessProbe *corev1.Probe `json:"readinessProbe,omitempty" `
	// Probes are not allowed for ephemeral containers.
	// +optional
	StartupProbe *corev1.Probe `json:"startupProbe,omitempty"`
}

// ConnectorType defines the frontend protocol for ShardingSphere Proxy
type ConnectorType string

const (
	ConnectorTypeMySQL      ConnectorType = "mysql"
	ConnectorTypePostgreSQL ConnectorType = "postgresql"
)

// MySQLDriver Defines the mysql-driven version in ShardingSphere-proxy
type Connector struct {
	Type ConnectorType `json:"type"`
	// +kubebuilder:validation:Pattern=`^([1-9]\d|[1-9])(\.([1-9]\d|\d)){2}$`
	// mysql-driven version,must be x.y.z
	Version string `json:"version"`
}

// BootstrapConfig is used for any ShardingSphere Proxy startup
type BootstrapConfig struct {
	// +optional
	ServerConfig ServerConfig `json:"serverConfig,omitempty"`
	// +optional
	LogbackConfig LogbackConfig `json:"logbackConfig,omitempty"`
}

// ProxySpec defines the desired state of ShardingSphereProxy
type ComputeNodeSpec struct {
	// +optional
	Bootstrap BootstrapConfig `json:"bootstrap,omitempty"`
	// replicas is the expected number of replicas of ShardingSphere-Proxy
	// +optional
	Replicas int32 `json:"replicas,omitempty"`
	// +optional
	Probes *ProxyProbe `json:"probes,omitempty"`
	// +optional
	Service Service `json:"service,omitempty"`
	// +optional
	Connector *Connector `json:"connector,omitempty"`
	// version  is the version of ShardingSphere-Proxy
	Version string `json:"version,omitempty"`
	// +optional
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	// port is ShardingSphere-Proxy startup port
	// +optional
	Ports []corev1.ContainerPort `json:"ports,omitempty"`
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`
	// +optional
	Resources v1.ResourceRequirements `json:"resources,omitempty"`
}

// ComputeNodeStatus defines the observed state of ShardingSphere Proxy
type ComputeNodeStatus struct {
	// ShardingSphere-Proxy phase are a brief summary of the ShardingSphere-Proxy life cycle
	// There are two possible phase values:
	// Ready: ShardingSphere-Proxy can already provide external services
	// NotReady: ShardingSphere-Proxy cannot provide external services
	// +optional
	Phase ComputeNodePhaseStatus `json:"phase"`

	// Conditions The conditions array, the reason and message fields
	// +optional
	Conditions ComputeNodeConditions `json:"conditions"`
	// ReadyInstances shows the number of replicas that ShardingSphere-Proxy is running normally
	// +optional
	ReadyInstances int32 `json:"readyInstances"`
}

type ComputeNodePhaseStatus string

const (
	ComputeNodeStatusReady    ComputeNodePhaseStatus = "Ready"
	ComputeNodeStatusNotReady ComputeNodePhaseStatus = "NotReady"
)

type ComputeNodeConditionType string

// ComputeNodeConditionType shows some states during the startup process of ShardingSphere-Proxy
const (
	ComputeNodeConditionInitialized ComputeNodeConditionType = "Initialized"
	ComputeNodeConditionStarted     ComputeNodeConditionType = "Started"
	ComputeNodeConditionReady       ComputeNodeConditionType = "Ready"
	ComputeNodeConditionUnknown     ComputeNodeConditionType = "Unknown"
	ComputeNodeConditionDeployed    ComputeNodeConditionType = "Deployed"
	ComputeNodeConditionFailed      ComputeNodeConditionType = "Failed"
)

type ComputeNodeConditions []ComputeNodeCondition

// ComputeNodeCondition
// | **phase** | **condition**  | **descriptions**|
// | ------------- | ---------- | ---------------------------------------------------- |
// | NotReady      | Deployed   | pods are deployed but are not created or currently pending|
// | NotReady      | Started    | pods are started but not satisfy ready requirements|
// | Ready         | Ready      | minimum pods satisfy ready requirements|
// | NotReady      | Unknown    | can not locate the status of pods |
// | NotReady      | Failed     | ShardingSphere-Proxy failed to start correctly due to some problems|
type ComputeNodeCondition struct {
	Type           ComputeNodeConditionType `json:"type"`
	Status         corev1.ConditionStatus   `json:"status"`
	LastUpdateTime metav1.Time              `json:"lastUpdateTime,omitempty"`
}

func init() {
	SchemeBuilder.Register(&ComputeNode{}, &ComputeNodeList{})
}
