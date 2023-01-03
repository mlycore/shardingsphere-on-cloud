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

//+kubebuilder:printcolumn:JSONPath=".status.readyNodes",name=ReadyNodes,type=integer
//+kubebuilder:printcolumn:JSONPath=".status.phase",name=Phase,type=string
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ComputeNode is the Schema for the ShardingSphere Proxy API
type ComputeNode struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ComputeNodeSpec   `json:"spec,omitempty"`
	Status ComputeNodeStatus `json:"status,omitempty"`
}

type PrivilegeType string

const (
	AllPermitted PrivilegeType = "ALL_PERMITTED"
)

type Privilege struct {
	Type PrivilegeType `json:"type"`
}

type User struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

type Authority struct {
	Privilege Privilege `json:"privilege"`
	Users     []User    `json:"users"`
}

type RepositoryType string

const (
	RepositoryTypeZookeeper RepositoryType = "ZooKeeper"
	RepositoryTypeEtcd      RepositoryType = "Etcd"
)

type Repository struct {
	Type  RepositoryType `json:"type"`
	Props ClusterProps   `json:"props"`
}

// Use these consts for verification
// const (
// 	PropsKeyKernelExecutorSize           = "kernel-executor-size"
// 	PropsKeyCheckTableMetadataEnabled    = "check-table-metadata-enabled"
// 	PropsKeyProxyBackendQueryFetchSize   = "proxy-backend-query-fetch-size"
// 	PropsKeyCheckDuplicateTableEnabled   = "check-duplicate-table-enabled"
// 	PropsKeyFrontendExecutorSize         = "proxy-frontend-executor-size"
// 	PropsKeyBackendExecutorSuitable      = "proxy-backend-executor-suitable"
// 	PropsKeyBackendDriverType            = "proxy-backend-driver-type"
// 	PropsKeyFrontendDatabaseProtocolType = "proxy-frontend-database-protocol-type"
// )

// const (
// 	ClusterPropsKeyNamespace                    = "namespace"
// 	ClusterPropsKeyServerLists                  = "server-lists"
// 	ClusterPropsKeyRetryIntervalMilliseconds    = "retryIntervalMilliseconds"
// 	ClusterPropsKeyMaxRetries                   = "maxRetries"
// 	ClusterPropsKeyTimeToLiveSeconds            = "timeToLiveSeconds"
// 	ClusterPropsKeyOperationTimeoutMilliseconds = "operationTimeoutMilliseconds"
// 	ClusterPropsKeyDigest                       = "digest"
// )

type ClusterProps struct {
	// Namespace of registry center
	Namespace string `json:"namespace" yaml:"namespace"`
	//Server lists of registry center
	ServerLists string `json:"server-lists" yaml:"server-lists"`
	//RetryIntervalMilliseconds Milliseconds of retry interval. default: 500
	// +optional
	RetryIntervalMilliseconds int `json:"retryIntervalMilliseconds,omitempty" yaml:"retryIntervalMilliseconds,omitempty"`
	// MaxRetries Max retries of client connection. default: 3
	// +optional
	MaxRetries int `json:"maxRetries,omitempty" yaml:"maxRetries,omitempty"`
	// TimeToLiveSeconds Seconds of ephemeral data live.default: 60
	// +optional
	TimeToLiveSeconds int `json:"timeToLiveSeconds,omitempty" yaml:"timeToLiveSeconds,omitempty"`
	// OperationTimeoutMilliseconds Milliseconds of operation timeout. default: 500
	// +optional
	OperationTimeoutMilliseconds int `json:"operationTimeoutMilliseconds,omitempty" yaml:"operationTimeoutMilliseconds,omitempty"`
	// Password of login
	// +optional
	Digest string `json:"digest,omitempty" yaml:"digest,omitempty"`
}

type ModeType string

const (
	ModeTypeCluster    ModeType = "cluster"
	ModeTypeStandalone ModeType = "memory"
)

// Props Apache ShardingSphere provides the way of property configuration to configure system level configuration.
type Props struct {
	// The max thread size of worker group to execute SQL. One ShardingSphereDataSource will use a independent thread pool, it does not share thread pool even different data source in same JVM.
	// +optional
	KernelExecutorSize int `json:"kernel-executor-size,omitempty" yaml:"kernel-executor-size,omitempty"`
	// Whether validate table meta data consistency when application startup or updated.
	// +optional
	CheckTableMetadataEnabled bool `json:"check-table-metadata-enabled,omitempty" yaml:"check-table-metadata-enabled,omitempty"`
	// ShardingSphereProxy backend query fetch size. A larger value may increase the memory usage of ShardingSphere ShardingSphereProxy. The default value is -1, which means set the minimum value for different JDBC drivers.
	// +optional
	ProxyBackendQueryFetchSize int `json:"proxy-backend-query-fetch-size,omitempty" yaml:"proxy-backend-query-fetch-size,omitempty"`
	// Whether validate duplicate table when application startup or updated.
	// +optional
	CheckDuplicateTableEnabled bool `json:"check-duplicate-table-enabled,omitempty" yaml:"check-duplicate-table-enabled,omitempty"`
	// ShardingSphereProxy frontend Netty executor size. The default value is 0, which means let Netty decide.
	// +optional
	ProxyFrontendExecutorSize int `json:"proxy-frontend-executor-size,omitempty" yaml:"proxy-frontend-executor-size,omitempty"`
	// Available options of proxy backend executor suitable: OLAP(default), OLTP. The OLTP option may reduce time cost of writing packets to client, but it may increase the latency of SQL execution and block other clients if client connections are more than proxy-frontend-executor-size, especially executing slow SQL.
	// +optional
	ProxyBackendExecutorSuitable string `json:"proxy-backend-executor-suitable,omitempty" yaml:"proxy-backend-executor-suitable,omitempty"`
	// +optional
	ProxyBackendDriverType string `json:"proxy-backend-driver-type,omitempty" yaml:"proxy-backend-driver-type,omitempty"`
	// +optional
	ProxyFrontendDatabaseProtocolType string `json:"proxy-frontend-database-protocol-type" yaml:"proxy-frontend-database-protocol-type,omitempty"`
}

type ServerMode struct {
	Repository Repository `json:"repository"`
	Type       ModeType   `json:"type"`
}

type ServerConfig struct {
	Authority Authority  `json:"authority"`
	Mode      ServerMode `json:"mode"`
	Props     *Props     `json:"props"`
}

type LogbackConfig string

// ServiceType defines the Service in Kubernetes of ShardingSphere-Proxy
type Service struct {
	Ports []corev1.ServicePort `json:"ports,omitempty"`

	// +kubebuilder:validation:Enum=ClusterIP;NodePort;LoadBalancer;ExternalName
	Type corev1.ServiceType `json:"type"`
}

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

// ProxySpec defines the desired state of ShardingSphereProxy
type ComputeNodeSpec struct {
	ServerConfig ServerConfig `json:"serverConfig,omitempty"`
	// +optional
	LogbackConfig LogbackConfig `json:"logbackConfig,omitempty"`

	// AutomaticScaling *AutomaticScaling `json:"automaticScaling,omitempty"`
	//Replicas is the expected number of replicas of ShardingSphere-Proxy
	Replicas int32 `json:"replicas"`

	// +optional
	Probes *ProxyProbe `json:"probes"`

	// +optional
	Service Service `json:"service"`

	// +optional
	Connector *Connector `json:"connector,omitempty"`

	// Version  is the version of ShardingSphere-Proxy
	Version string `json:"version"`
	// +optional
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`

	//Port is ShardingSphere-Proxy startup port
	// +optional
	Ports []corev1.ContainerPort `json:"ports"`

	// +optional
	Env []corev1.EnvVar `json:"env"`

	// +optional
	Resources v1.ResourceRequirements `json:"resources,omitempty"`
}

// ComputeNodeStatus defines the observed state of ShardingSphereProxy
type ComputeNodeStatus struct {
	//ShardingSphere-Proxy phase are a brief summary of the ShardingSphere-Proxy life cycle
	//There are two possible phase values:
	//Ready: ShardingSphere-Proxy can already provide external services
	//NotReady: ShardingSphere-Proxy cannot provide external services
	Phase PhaseStatus `json:"phase"`

	//Conditions The conditions array, the reason and message fields
	Conditions Conditions `json:"conditions"`
	//ReadyNodes shows the number of replicas that ShardingSphere-Proxy is running normally
	ReadyNodes int32 `json:"readyNodes"`
}

type PhaseStatus string

const (
	StatusReady    PhaseStatus = "Ready"
	StatusNotReady PhaseStatus = "NotReady"
)

type ConditionType string

// ConditionType shows some states during the startup process of ShardingSphere-Proxy
const (
	ConditionInitialized ConditionType = "Initialized"
	ConditionStarted     ConditionType = "Started"
	ConditionReady       ConditionType = "Ready"
	ConditionUnknown     ConditionType = "Unknown"
)

type Conditions []Condition

// Condition
// | **condition** | **status** | **directions**|
// | ------------- | ---------- | ---------------------------------------------------- |
// | Initialized   | true       | Initialization successful|
// | Initialized   | false      | initialization failed|
// | Started       | true       | pod started successfully but not ready|
// | Started       | false      | pod started failed|
// | Ready         | true       | The pod is ready and can provide external services|
// | Unknown       | true       | ShardingSphere-Proxy failed to start correctly due to some problems |
type Condition struct {
	Type           ConditionType          `json:"type"`
	Status         corev1.ConditionStatus `json:"status"`
	LastUpdateTime metav1.Time            `json:"lastUpdateTime,omitempty"`
}

func (p *ComputeNode) SetInitialized() {
	p.Status.Phase = StatusNotReady
	p.Status.Conditions = append([]Condition{}, Condition{
		Type:           ConditionInitialized,
		Status:         corev1.ConditionTrue,
		LastUpdateTime: metav1.Now(),
	})
}

func (p *ComputeNode) SetInitializationFailed() {
	p.Status.Phase = StatusNotReady
	p.Status.Conditions = append([]Condition{}, Condition{
		Type:           ConditionInitialized,
		Status:         corev1.ConditionFalse,
		LastUpdateTime: metav1.Now(),
	})
}

func (p *ComputeNode) SetPodStarted(readyNodes int32) {
	p.Status.Phase = StatusNotReady
	p.Status.Conditions = append([]Condition{}, Condition{
		Type:           ConditionStarted,
		Status:         corev1.ConditionTrue,
		LastUpdateTime: metav1.Now(),
	})
	p.Status.ReadyNodes = readyNodes
}

func (p *ComputeNode) SetPodNotStarted(readyNodes int32) {
	p.Status.Phase = StatusNotReady
	p.Status.Conditions = append([]Condition{}, Condition{
		Type:           ConditionStarted,
		Status:         corev1.ConditionFalse,
		LastUpdateTime: metav1.Now(),
	})
	p.Status.ReadyNodes = readyNodes
}

func (p *ComputeNode) SetReady(readyNodes int32) {
	p.Status.Phase = StatusReady
	p.Status.Conditions = append([]Condition{}, Condition{
		Type:           ConditionReady,
		Status:         corev1.ConditionTrue,
		LastUpdateTime: metav1.Now(),
	})
	p.Status.ReadyNodes = readyNodes

}

func (p *ComputeNode) SetFailed() {
	p.Status.Phase = StatusNotReady
	p.Status.Conditions = append([]Condition{}, Condition{
		Type:           ConditionUnknown,
		Status:         corev1.ConditionTrue,
		LastUpdateTime: metav1.Now(),
	})
}
func (p *ComputeNode) UpdateReadyNodes(readyNodes int32) {
	p.Status.ReadyNodes = readyNodes
}

func init() {
	SchemeBuilder.Register(&ComputeNode{}, &ComputeNodeList{})
}
