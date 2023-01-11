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

package reconcile

import (
	"testing"

	autoscalingv2beta2 "k8s.io/api/autoscaling/v2beta2"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/pointer"

	"github.com/apache/shardingsphere-on-cloud/shardingsphere-operator/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_IsRunning(t *testing.T) {
	ts := metav1.Time{}
	cases := []struct {
		podlist *v1.PodList
		exp     bool
		message string
	}{
		{
			podlist: &v1.PodList{
				Items: []v1.Pod{},
			},
			exp:     false,
			message: "Empty PodList should be false",
		},
		{
			podlist: &v1.PodList{
				Items: []v1.Pod{
					{
						Status: v1.PodStatus{
							Phase: v1.PodRunning,
						},
					},
					{
						Status: v1.PodStatus{
							Phase: v1.PodFailed,
						},
					},
				},
			},
			// At least one Pod is Running considered the Cluster be availabe
			exp:     true,
			message: "First Pod is Running and second Pod is not Running should be false",
		},
		{
			podlist: &v1.PodList{
				Items: []v1.Pod{
					{
						Status: v1.PodStatus{
							Phase: v1.PodFailed,
						},
					},
					{
						Status: v1.PodStatus{
							Phase: v1.PodRunning,
						},
					},
				},
			},
			// At least one Pod is Running considered the Cluster be availabe
			exp:     true,
			message: "First Pod is not Running and second Pod is Running should be false",
		},
		{
			podlist: &v1.PodList{
				Items: []v1.Pod{
					{
						Status: v1.PodStatus{
							Phase: v1.PodRunning,
						},
					},
					{
						Status: v1.PodStatus{
							Phase: v1.PodRunning,
						},
					},
				},
			},
			exp:     true,
			message: "All Pods are running should be true",
		},
		{
			podlist: &v1.PodList{
				Items: []v1.Pod{
					{
						Status: v1.PodStatus{
							Phase: v1.PodFailed,
						},
					},
					{
						Status: v1.PodStatus{
							Phase: v1.PodFailed,
						},
					},
				},
			},
			exp:     false,
			message: "All Pods are not running should be false",
		},
		{
			podlist: &v1.PodList{
				Items: []v1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{
							DeletionTimestamp: &ts,
						},
						Status: v1.PodStatus{
							Phase: v1.PodRunning,
						},
					},
				},
			},
			exp:     false,
			message: "Pod is running with deletion timestamp should be false",
		},
	}

	for _, c := range cases {
		act := IsRunning(c.podlist)
		if len(c.podlist.Items) != 0 {
			assert.Equal(t, c.exp, act, c.message)
		}
	}
}

func Test_CountingReadyPods(t *testing.T) {
	ts := metav1.Time{}
	cases := []struct {
		podlist *v1.PodList
		exp     int32
		message string
	}{
		{
			podlist: &v1.PodList{
				Items: []v1.Pod{},
			},
			exp:     0,
			message: "Empty PodList should be 0",
		},
		{
			podlist: &v1.PodList{
				Items: []v1.Pod{
					{
						Status: v1.PodStatus{
							ContainerStatuses: []v1.ContainerStatus{},
						},
					},
				},
			},
			exp:     0,
			message: "Only one Pod without any container statuses should be 0",
		},
		{
			podlist: &v1.PodList{
				Items: []v1.Pod{
					{
						Status: v1.PodStatus{
							ContainerStatuses: []v1.ContainerStatus{
								{
									Ready: true,
								},
							},
						},
					},
				},
			},
			exp:     1,
			message: "Only one Pod is running should be 1",
		},
		{
			podlist: &v1.PodList{
				Items: []v1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{
							DeletionTimestamp: &ts,
						},
						Status: v1.PodStatus{
							ContainerStatuses: []v1.ContainerStatus{
								{
									Ready: true,
								},
							},
						},
					},
				},
			},
			exp:     0,
			message: "Pod has ready container but with deletion timestamp should be 0",
		},
		{
			podlist: &v1.PodList{
				Items: []v1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{
							DeletionTimestamp: &ts,
						},
						Status: v1.PodStatus{
							ContainerStatuses: []v1.ContainerStatus{
								{
									Ready: true,
								},
							},
						},
					},
					{
						Status: v1.PodStatus{
							ContainerStatuses: []v1.ContainerStatus{
								{
									Ready: true,
								},
							},
						},
					},
				},
			},
			exp:     1,
			message: "First Pod has ready container, second Pod has ready container but with deletion timestamp should be 0",
		},
		{
			podlist: &v1.PodList{
				Items: []v1.Pod{
					{
						Status: v1.PodStatus{
							ContainerStatuses: []v1.ContainerStatus{
								{
									Ready: true,
								},
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							DeletionTimestamp: &ts,
						},
						Status: v1.PodStatus{
							ContainerStatuses: []v1.ContainerStatus{
								{
									Ready: true,
								},
							},
						},
					},
				},
			},
			exp:     1,
			message: "First Pod has ready container but with deletion timestamp, second Pod has ready container should be 0",
		},
	}

	for _, c := range cases {
		act := CountingReadyPods(c.podlist)
		assert.Equal(t, c.exp, act, c.message)
	}
}

func Test_ConstructCascadingService(t *testing.T) {
	cases := []struct {
		proxy   *v1alpha1.ShardingSphereProxy
		exp     *v1.Service
		message string
	}{
		{
			exp:     &v1.Service{},
			message: "Nil ShardingSphereProxy definition should lead to empty Service",
		},
		{
			proxy:   &v1alpha1.ShardingSphereProxy{},
			exp:     &v1.Service{},
			message: "Empty ShardingSphereProxy definition should lead to empty Service",
		},
		{
			proxy: &v1alpha1.ShardingSphereProxy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testname",
					Namespace: "testnamespace",
				},
				Spec: v1alpha1.ProxySpec{
					ServiceType: v1alpha1.ServiceType{
						Type:     v1.ServiceTypeNodePort,
						NodePort: 33007,
					},
					Port: 3307,
				},
			},
			exp: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testname",
					Namespace: "testnamespace",
				},
				Spec: v1.ServiceSpec{
					Selector: map[string]string{
						"apps": "testname",
					},
					Type: v1.ServiceTypeNodePort,
					Ports: []v1.ServicePort{
						{
							Name:       "proxy-port",
							TargetPort: fromInt32(3307),
							Port:       3307,
							NodePort:   33007,
						},
					},
				},
			},
			message: "Normal ShardingSphereProxy definition should lead to normal Service",
		},
	}

	for _, c := range cases {
		act := ConstructCascadingService(c.proxy)
		assert.Equal(t, c.exp.ObjectMeta.Name, act.ObjectMeta.Name, c.message)
		assert.Equal(t, c.exp.ObjectMeta.Namespace, act.ObjectMeta.Namespace, c.message)
		if c.proxy != nil {
			assert.EqualValues(t, c.exp.Spec.Selector, act.Spec.Selector, c.message)
			assert.Equal(t, c.exp.Spec.Type, act.Spec.Type, c.message)
			assert.Equal(t, c.exp.Spec.Ports, act.Spec.Ports, c.message)
		}
	}
}

func Test_ConstructCascadingConfigmap(t *testing.T) {
	/*
	   cases := []struct{

	   }{
	       {

	       },
	   }

	   for _, c := range cases {

	   }
	*/
}

func Test_ConstructHPA(t *testing.T) {
	/*
	   cases := []struct{

	   }{
	       {

	       },
	   }

	   for _, c := range cases {

	   }
	*/
}

func Test_ToYAML(t *testing.T) {
	cases := []struct {
		proxyCfg *v1alpha1.ShardingSphereProxyServerConfig
		exp      string
		message  string
	}{
		{
			proxyCfg: &v1alpha1.ShardingSphereProxyServerConfig{},
			exp: `mode:
  type: ""
  repository:
    type: ""
    props:
      namespace: ""
      server-lists: ""
authority:
  users: []
  privilege: null
`,
			message: "Empty proxy server config should generate yaml with empty value",
		},
		{
			proxyCfg: &v1alpha1.ShardingSphereProxyServerConfig{
				Spec: v1alpha1.ProxyConfigSpec{
					ClusterConfig: v1alpha1.ClusterConfig{
						Type: "Cluster",
						Repository: v1alpha1.RepositoryConfig{
							Type: "Zookeeper",
							Props: v1alpha1.ClusterProps{
								Namespace:                    "cluster-sharding-mode",
								ServerLists:                  "host1:2181,host2:2181",
								RetryIntervalMilliseconds:    500,
								MaxRetries:                   3,
								TimeToLiveSeconds:            60,
								OperationTimeoutMilliseconds: 500,
							},
						},
					},
					Authority: v1alpha1.Auth{
						Users: []v1alpha1.User{
							{
								User:     "John",
								Password: "123456",
							},
						},
						Privilege: &v1alpha1.Privilege{
							Type: "Admin",
						},
					},
					Props: &v1alpha1.Props{
						KernelExecutorSize:                16,
						CheckTableMetadataEnabled:         false,
						ProxyBackendQueryFetchSize:        -1,
						CheckDuplicateTableEnabled:        false,
						ProxyFrontendExecutorSize:         0,
						ProxyBackendExecutorSuitable:      "OLAP",
						ProxyBackendDriverType:            "JDBC",
						ProxyFrontendDatabaseProtocolType: "",
					},
				},
			},
			exp: `mode:
  type: Cluster
  repository:
    type: Zookeeper
    props:
      namespace: cluster-sharding-mode
      server-lists: host1:2181,host2:2181
      retryIntervalMilliseconds: 500
      maxRetries: 3
      timeToLiveSeconds: 60
      operationTimeoutMilliseconds: 500
authority:
  users:
  - user: John
    password: "123456"
  privilege:
    type: Admin
props:
  kernel-executor-size: 16
  proxy-backend-query-fetch-size: -1
  proxy-backend-executor-suitable: OLAP
  proxy-backend-driver-type: JDBC
`,
			message: "Should generate yaml with the same value as ShardingSphereProxyServerConfig.Spec",
		},
	}

	for _, c := range cases {
		act := toYaml(c.proxyCfg)
		assert.Equal(t, c.exp, act, c.message)
	}
}

func Test_UpdateService(t *testing.T) {
	cases := []struct {
		proxy   *v1alpha1.ShardingSphereProxy
		service *v1.Service
		message string
	}{
		{
			proxy: &v1alpha1.ShardingSphereProxy{
				Spec: v1alpha1.ProxySpec{
					ServiceType: v1alpha1.ServiceType{
						Type:     "ServiceTypeNodePort",
						NodePort: 3001,
					},
					Port: 3000,
				},
			},
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{{
						Port:     3000,
						NodePort: 3001,
					}},
				},
			},
			message: "Service should be updated",
		},
		{
			proxy: &v1alpha1.ShardingSphereProxy{
				Spec: v1alpha1.ProxySpec{
					ServiceType: v1alpha1.ServiceType{
						Type:     "ServiceTypeNodePort",
						NodePort: 0,
					},
					Port: 3000,
				},
			},
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{{
						Port:     3000,
						NodePort: 3001,
					}},
				},
			},
			message: "Service NodePort should not be updated if proxy nodeport is 0",
		},
	}

	for _, c := range cases {
		UpdateService(c.proxy, c.service)
		assert.Equal(t, c.proxy.Spec.ServiceType.Type, c.service.Spec.Type, c.message)
		assert.Equal(t, c.proxy.Spec.Port, c.service.Spec.Ports[0].Port, c.message)
		assert.Equal(t, c.proxy.Spec.Port, c.service.Spec.Ports[0].TargetPort.IntVal, c.message)
		if c.proxy.Spec.ServiceType.NodePort != 0 {
			assert.Equal(t, c.proxy.Spec.ServiceType.NodePort, c.service.Spec.Ports[0].NodePort, c.message)
		}
	}
}

func Test_UpdateHPA(t *testing.T) {
	cases := []struct {
		proxy   *v1alpha1.ShardingSphereProxy
		hpa     *autoscalingv2beta2.HorizontalPodAutoscaler
		message string
	}{
		{
			proxy: &v1alpha1.ShardingSphereProxy{
				Spec: v1alpha1.ProxySpec{
					AutomaticScaling: &v1alpha1.AutomaticScaling{
						Enable:           true,
						ScaleUpWindows:   15,
						ScaleDownWindows: 30,
						MaxInstance:      5,
						MinInstance:      3,
						Target:           70,
					},
				},
			},
			hpa: &autoscalingv2beta2.HorizontalPodAutoscaler{
				Spec: autoscalingv2beta2.HorizontalPodAutoscalerSpec{
					ScaleTargetRef: autoscalingv2beta2.CrossVersionObjectReference{
						Kind:       "Deployment",
						Name:       "test",
						APIVersion: appsv1.SchemeGroupVersion.String(),
					},
					MinReplicas: pointer.Int32(3),
					MaxReplicas: 10,
					Metrics: []autoscalingv2beta2.MetricSpec{
						{
							Type: autoscalingv2beta2.ResourceMetricSourceType,
							Resource: &autoscalingv2beta2.ResourceMetricSource{
								Name: "cpu",
								Target: autoscalingv2beta2.MetricTarget{
									Type:               autoscalingv2beta2.UtilizationMetricType,
									AverageUtilization: pointer.Int32(60),
								},
							},
						},
					},
					Behavior: &autoscalingv2beta2.HorizontalPodAutoscalerBehavior{
						ScaleUp: &autoscalingv2beta2.HPAScalingRules{
							StabilizationWindowSeconds: pointer.Int32(15),
						},
						ScaleDown: &autoscalingv2beta2.HPAScalingRules{
							StabilizationWindowSeconds: pointer.Int32(60),
							Policies: []autoscalingv2beta2.HPAScalingPolicy{
								{
									Type:          autoscalingv2beta2.PodsScalingPolicy,
									Value:         1,
									PeriodSeconds: 30,
								},
							},
						},
					},
				},
			},
			message: "hpa should be updated",
		},
	}

	for _, c := range cases {
		UpdateHPA(c.proxy, c.hpa)
		assert.Equal(t, c.proxy.Spec.AutomaticScaling.Target, *c.hpa.Spec.Metrics[0].Resource.Target.AverageUtilization)
		assert.Equal(t, c.proxy.Spec.AutomaticScaling.ScaleDownWindows, *c.hpa.Spec.Behavior.ScaleDown.StabilizationWindowSeconds)
	}
}

func Test_UpdateHPAMetric(t *testing.T) {
	cases := []struct {
		proxy   *v1alpha1.ShardingSphereProxy
		hpa     *autoscalingv2beta2.HorizontalPodAutoscaler
		message string
	}{
		{
			proxy: &v1alpha1.ShardingSphereProxy{
				Spec: v1alpha1.ProxySpec{
					AutomaticScaling: &v1alpha1.AutomaticScaling{
						Enable:           true,
						ScaleUpWindows:   15,
						ScaleDownWindows: 30,
						MaxInstance:      5,
						MinInstance:      3,
						Target:           70,
						CustomMetrics: []autoscalingv2beta2.MetricSpec{
							{
								Type: "pods",
								Pods: &autoscalingv2beta2.PodsMetricSource{
									Metric: autoscalingv2beta2.MetricIdentifier{
										Name: "latency_seconds_test_avg",
									},
									Target: autoscalingv2beta2.MetricTarget{
										Type:         autoscalingv2beta2.AverageValueMetricType,
										AverageValue: resource.NewQuantity(12, resource.DecimalSI),
									},
								},
							},
						},
					},
				},
			},
			hpa: &autoscalingv2beta2.HorizontalPodAutoscaler{
				Spec: autoscalingv2beta2.HorizontalPodAutoscalerSpec{
					ScaleTargetRef: autoscalingv2beta2.CrossVersionObjectReference{
						Kind:       "Deployment",
						Name:       "test",
						APIVersion: appsv1.SchemeGroupVersion.String(),
					},
					MinReplicas: pointer.Int32(3),
					MaxReplicas: 10,
					Metrics: []autoscalingv2beta2.MetricSpec{
						{
							Type: autoscalingv2beta2.ResourceMetricSourceType,
							Resource: &autoscalingv2beta2.ResourceMetricSource{
								Name: "cpu",
								Target: autoscalingv2beta2.MetricTarget{
									Type:               autoscalingv2beta2.UtilizationMetricType,
									AverageUtilization: pointer.Int32(60),
								},
							},
						},
					},
					Behavior: &autoscalingv2beta2.HorizontalPodAutoscalerBehavior{
						ScaleUp: &autoscalingv2beta2.HPAScalingRules{
							StabilizationWindowSeconds: pointer.Int32(15),
						},
						ScaleDown: &autoscalingv2beta2.HPAScalingRules{
							StabilizationWindowSeconds: pointer.Int32(60),
							Policies: []autoscalingv2beta2.HPAScalingPolicy{
								{
									Type:          autoscalingv2beta2.PodsScalingPolicy,
									Value:         1,
									PeriodSeconds: 30,
								},
							},
						},
					},
				},
			},
			message: "hpa should be updated",
		},
	}

	for _, c := range cases {
		UpdateHPA(c.proxy, c.hpa)
		assert.Equal(t, c.proxy.Spec.AutomaticScaling.CustomMetrics, c.hpa.Spec.Metrics)
	}
}
