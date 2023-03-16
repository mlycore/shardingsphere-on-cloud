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

package cluster

import (
	"fmt"

	"github.com/apache/shardingsphere-on-cloud/shardingsphere-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewStorageNode(cluster *v1alpha1.Cluster, schema string) *v1alpha1.StorageNode {
	var spec v1alpha1.StorageNodeSpec
	for _, s := range cluster.Spec.Schemas {
		if s.Name == schema {
			spec = s.Topology.StorageNode
			break
		}
	}

	exp := &v1alpha1.StorageNode{
		ObjectMeta: metav1.ObjectMeta{
			Name:        fmt.Sprintf("%s-%s", cluster.Name, schema),
			Namespace:   cluster.Namespace,
			Labels:      cluster.Labels,
			Annotations: cluster.Annotations,
		},
		Spec: spec,
	}
	return exp
}

func UpdateStorageNode(cluster *v1alpha1.Cluster, cur *v1alpha1.StorageNode, schema string) *v1alpha1.StorageNode {
	exp := &v1alpha1.StorageNode{}
	exp.ObjectMeta = cur.ObjectMeta
	// exp.ObjectMeta.ResourceVersion = ""
	exp.Labels = cur.Labels
	exp.Annotations = cur.Annotations
	//FIXME: this should b a immutable parameter
	exp.Spec = NewStorageNode(cluster, schema).Spec
	return exp
}
