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

package controllers

import (
	"context"
	"fmt"

	"github.com/apache/shardingsphere-on-cloud/shardingsphere-operator/api/v1alpha1"
	"github.com/apache/shardingsphere-on-cloud/shardingsphere-operator/pkg/kubernetes/computenode"
	"github.com/apache/shardingsphere-on-cloud/shardingsphere-operator/pkg/kubernetes/storagenode"
	reconcile "github.com/apache/shardingsphere-on-cloud/shardingsphere-operator/pkg/reconcile/cluster"
	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	clusterControllerName = "cluster-controller"
)

type ClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger

	ComputeNode computenode.ComputeNode
	StorageNode storagenode.StorageNode
	//TODO: add a sql driver
}

func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Cluster{}).
		Owns(&v1alpha1.ComputeNode{}).
		Owns(&v1alpha1.StorageNode{}).
		Complete(r)
}

func (r *ClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.Log.WithValues(clusterControllerName, req.NamespacedName)

	clu := &v1alpha1.Cluster{}

	if err := r.Get(ctx, req.NamespacedName, clu); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{RequeueAfter: defaultRequeueTime}, nil
		} else {
			logger.Error(err, "get cluster")
			return ctrl.Result{Requeue: true}, err
		}
	}

	errors := []error{}
	if err := r.reconcileComputeNode(ctx, clu); err != nil {
		logger.Error(err, "reconcile computenode")
		errors = append(errors, err)
	}
	if err := r.reconcileStorageNode(ctx, clu); err != nil {
		logger.Error(err, "reconcile storagenode")
		errors = append(errors, err)
	}

	if err := r.executeDistSQL(ctx, clu); err != nil {
		logger.Error(err, "execute DistSQL")
		errors = append(errors, err)
	}

	if len(errors) != 0 {
		return ctrl.Result{Requeue: true}, errors[0]
	}

	// if err := r.reconcileStatus(ctx, clu); err != nil {
	// 	logger.Error(err, "reconcile status")
	// }

	return ctrl.Result{RequeueAfter: defaultRequeueTime}, nil
}

func (r *ClusterReconciler) reconcileComputeNode(ctx context.Context, cluster *v1alpha1.Cluster) error {
	for _, s := range cluster.Spec.Schemas {
		cn, found, err := r.getComputeNodeByNamespacedName(ctx, types.NamespacedName{
			Namespace: cluster.Namespace,
			Name:      fmt.Sprintf("%s-%s", cluster.Name, s.Name),
		})
		if found {
			if err := r.updateComputeNode(ctx, cluster, cn, s.Name); err != nil {
				return err
			}
		} else {
			if err != nil {
				return err
			} else {
				if err := r.createComputeNode(ctx, cluster, s.Name); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (r *ClusterReconciler) getComputeNodeByNamespacedName(ctx context.Context, namespacedName types.NamespacedName) (*v1alpha1.ComputeNode, bool, error) {
	dp, err := r.ComputeNode.GetByNamespacedName(ctx, namespacedName)
	// found
	if dp != nil {
		return dp, true, nil
	}
	// error
	if err != nil {
		return nil, false, err
	} else {
		// not found
		return nil, false, nil
	}
}

func (r *ClusterReconciler) updateComputeNode(ctx context.Context, cluster *v1alpha1.Cluster, cn *v1alpha1.ComputeNode, schema string) error {
	exp := reconcile.UpdateComputeNode(cluster, cn, schema)
	if err := r.Update(ctx, exp); err != nil {
		return err
	}
	return nil
}

func (r *ClusterReconciler) createComputeNode(ctx context.Context, cluster *v1alpha1.Cluster, schema string) error {
	cn := reconcile.NewComputeNode(cluster, schema)
	if err := r.Create(ctx, cn); err != nil {
		if apierrors.IsAlreadyExists(err) {
			return nil
		}
		return err
	}
	return nil
}

func (r *ClusterReconciler) reconcileStorageNode(ctx context.Context, cluster *v1alpha1.Cluster) error {
	for _, s := range cluster.Spec.Schemas {
		sn, found, err := r.getStorageNodeByNamespacedName(ctx, types.NamespacedName{
			Namespace: cluster.Namespace,
			Name:      fmt.Sprintf("%s-%s", cluster.Name, s.Name),
		})
		if found {
			if err := r.updateStorageNode(ctx, cluster, sn, s.Name); err != nil {
				return err
			}
		} else {
			if err != nil {
				return err
			} else {
				if err := r.createStorageNode(ctx, cluster, s.Name); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (r *ClusterReconciler) getStorageNodeByNamespacedName(ctx context.Context, namespacedName types.NamespacedName) (*v1alpha1.StorageNode, bool, error) {
	sn, err := r.StorageNode.GetByNamespacedName(ctx, namespacedName)
	// found
	if sn != nil {
		return sn, true, nil
	}
	// error
	if err != nil {
		return nil, false, err
	} else {
		// not found
		return nil, false, nil
	}
}

func (r *ClusterReconciler) updateStorageNode(ctx context.Context, cluster *v1alpha1.Cluster, sn *v1alpha1.StorageNode, schema string) error {
	exp := reconcile.UpdateStorageNode(cluster, sn, schema)
	if err := r.Update(ctx, exp); err != nil {
		return err
	}
	return nil
}

func (r *ClusterReconciler) createStorageNode(ctx context.Context, cluster *v1alpha1.Cluster, schema string) error {
	sn := reconcile.NewStorageNode(cluster, schema)
	if err := r.Create(ctx, sn); err != nil {
		if apierrors.IsAlreadyExists(err) {
			return nil
		}
		return err
	}
	return nil
}

func (r *ClusterReconciler) executeDistSQL(ctx context.Context, schema *v1alpha1.Schema) error {

}
