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

package storagenode

import (
	"context"

	"github.com/apache/shardingsphere-on-cloud/shardingsphere-operator/api/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewStorageNode(c client.Client) StorageNode {
	return storageNodeClient{
		storageNodeGetter: storageNodeGetter{
			Client: c,
		},
		storageNodeSetter: storageNodeSetter{
			Client: c,
		},
	}
}

type StorageNode interface {
	StorageNodeGetter
	StorageNodeSetter
}

type storageNodeClient struct {
	storageNodeGetter
	storageNodeSetter
}

type StorageNodeGetter interface {
	GetByNamespacedName(context.Context, types.NamespacedName) (*v1alpha1.StorageNode, error)
}

type storageNodeGetter struct {
	client.Client
}

func (g storageNodeGetter) GetByNamespacedName(ctx context.Context, namespacedName types.NamespacedName) (*v1alpha1.StorageNode, error) {
	sn := &v1alpha1.StorageNode{}
	if err := g.Get(ctx, namespacedName, sn); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	} else {
		return sn, nil
	}
}

type StorageNodeSetter interface {
}

type storageNodeSetter struct {
	client.Client
}
