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
	"github.com/apache/shardingsphere-on-cloud/shardingsphere-operator/api/v1alpha1"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func ComputeNodeNewConfigMap(cn *v1alpha1.ComputeNode) *v1.ConfigMap {
	cm := ComputeNodeDefaultConfigMap(cn.GetObjectMeta(), cn.GroupVersionKind())

	cm.Name = cn.Name
	cm.Namespace = cn.Namespace
	cm.Labels = cn.Labels
	cn.Spec.LogbackConfig = logback
	cm.Data["logback.xml"] = logback

	if y, err := yaml.Marshal(cn.Spec.ServerConfig); err == nil {
		cm.Data["server.yaml"] = string(y)
	}

	return cm
}

func ComputeNodeDefaultConfigMap(meta metav1.Object, gvk schema.GroupVersionKind) *v1.ConfigMap {
	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "shardingsphere-proxy",
			Namespace: "default",
			Labels:    map[string]string{},
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(meta, gvk),
			},
		},
		Data: map[string]string{},
	}
}

// FIXME: check if changed first, then decide if need to respawn the Pods
func ComputeNodeUpdateConfigMap(cn *v1alpha1.ComputeNode, cur *v1.ConfigMap) *v1.ConfigMap {
	exp := &v1.ConfigMap{}
	exp.ObjectMeta = cur.ObjectMeta
	exp.Labels = cur.Labels
	exp.Annotations = cur.Annotations
	exp.Data = ComputeNodeNewConfigMap(cn).Data
	return exp
}
