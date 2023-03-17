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
	"strings"

	"github.com/apache/shardingsphere-on-cloud/shardingsphere-operator/api/v1alpha1"
	"github.com/apache/shardingsphere-on-cloud/shardingsphere-operator/pkg/kubernetes/computenode"
	"github.com/apache/shardingsphere-on-cloud/shardingsphere-operator/pkg/kubernetes/storagenode"
	reconcile "github.com/apache/shardingsphere-on-cloud/shardingsphere-operator/pkg/reconcile/cluster"
	"github.com/go-logr/logr"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
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

	if len(errors) != 0 {
		return ctrl.Result{Requeue: true}, errors[0]
	}

	if err := r.executeDistSQL(ctx, clu, &clu.Spec.Schemas[0]); err != nil {
		logger.Error(err, "execute DistSQL")
		return ctrl.Result{RequeueAfter: 30}, err
	}

	if err := r.reconcileStatus(ctx, clu); err != nil {
		logger.Error(err, "reconcile status")
	}

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

func (r *ClusterReconciler) executeDistSQL(ctx context.Context, cluster *v1alpha1.Cluster, schema *v1alpha1.Schema) error {
	sn, ok, err := r.getStorageNodeByNamespacedName(ctx, types.NamespacedName{
		Namespace: cluster.Namespace,
		Name:      fmt.Sprintf("%s-%s", cluster.Name, schema.Name),
	})
	if err != nil {
		return err
	}
	if ok {
		var available bool
		for _, ep := range sn.Status.Endpoints {
			if strings.ToLower(ep.Status) != "available" {
				available = false
				return nil
			}
		}
		if available {
			r.Log.Info("available")
			if err := r.createSchema(ctx, cluster, schema.Name); err != nil {
				r.Log.Error(err, "create schema")
				return err
			}
			if err := r.registerStorageUnits(ctx, cluster, schema.Name, sn); err != nil {
				r.Log.Error(err, "register storage units")
				return err
			}
			if err := r.createEncryption(ctx, cluster, schema.Name); err != nil {
				r.Log.Error(err, "create encryption")
				return err
			}

		}
	}

	// following steps move to cluster controller
	// 2. create logical database if not exists
	// 3. using logical database
	// 4. register storage node
	// 5. create encrypt rule
	return nil
}

func (r *ClusterReconciler) createSchema(ctx context.Context, cluster *v1alpha1.Cluster, schema string) error {
	var (
		dialects string = "mysql"
		user     string = "root"
		pass     string = "root"
		host     string = fmt.Sprintf("%s:3307", cluster.Status.Service)
		database string = schema
	)
	dbconn, err := gorm.Open(dialects, fmt.Sprintf("%s:%s@(%s)/%s?charset=utf8&parseTime=True&loc=Local", user, pass, host, database))
	if err != nil {
		return err
	}
	dbconn.LogMode(true)
	if dbconn.Raw(fmt.Sprintf("CREATE DATABASE %s", schema)).Error != nil {
		return err
	}

	return nil
}

func (r *ClusterReconciler) registerStorageUnits(ctx context.Context, cluster *v1alpha1.Cluster, schema string, sn *v1alpha1.StorageNode) error {
	var (
		dialects string = "mysql"
		user     string = "root"
		pass     string = "root"
		host     string = fmt.Sprintf("%s:3307", cluster.Status.Service)
		database string = schema
	)
	dbconn, err := gorm.Open(dialects, fmt.Sprintf("%s:%s@(%s)/%s?charset=utf8&parseTime=True&loc=Local", user, pass, host, database))
	if err != nil {
		return err
	}
	dbconn.LogMode(true)
	ds0 := fmt.Sprintf("HOST=%s,PORT=%d,DB=%s,USER=%s,PASSWORD=%s", sn.Status.Endpoints[0].Address, sn.Status.Endpoints[0].Port, schema, sn.Status.Endpoints[0].User, sn.Status.Endpoints[0].Pass)
	ds1 := fmt.Sprintf("HOST=%s,PORT=%d,DB=%s,USER=%s,PASSWORD=%s", sn.Status.Endpoints[1].Address, sn.Status.Endpoints[1].Port, schema, sn.Status.Endpoints[1].User, sn.Status.Endpoints[1].Pass)
	if dbconn.Raw(fmt.Sprintf("USE %s; REGISTER STORAGE UNITS IF NOT EXISTS ds_0(%s),ds_1(%s);", schema, ds0, ds1)).Error != nil {
		return err
	}

	return nil
}

func (r *ClusterReconciler) createEncryption(ctx context.Context, cluster *v1alpha1.Cluster, schema string) error {
	var (
		dialects string = "mysql"
		user     string = "root"
		pass     string = "root"
		host     string = fmt.Sprintf("%s:3307", cluster.Status.Service)
		database string = schema
	)
	distsql := cluster.Spec.Schemas[0].Encryption.ToDistSQL()
	dbconn, err := gorm.Open(dialects, fmt.Sprintf("%s:%s@(%s)/%s?charset=utf8&parseTime=True&loc=Local", user, pass, host, database))
	if err != nil {
		return err
	}
	dbconn.LogMode(true)
	if dbconn.Raw(fmt.Sprintf("USE %s; %s", schema, distsql)).Error != nil {
		return err
	}

	return nil
}

func (r *ClusterReconciler) reconcileStatus(ctx context.Context, cluster *v1alpha1.Cluster) error {
	cn, ok, err := r.getComputeNodeByNamespacedName(ctx, types.NamespacedName{
		Namespace: cluster.Namespace,
		Name:      fmt.Sprintf("%s-%s", cluster.Name, cluster.Spec.Schemas[0].Name),
	})
	if err != nil {
		r.Log.Error(err, "get computenode")
		return err
	}

	if ok {
		rt := &v1alpha1.Cluster{}
		if err := r.Get(ctx, types.NamespacedName{
			Namespace: cluster.Namespace,
			Name:      cluster.Name,
		}, rt); err != nil {
			r.Log.Error(err, "get runtime cluster")
			return err
		}

		rt.Status.Service = fmt.Sprintf("%s.%s", cn.Name, cn.Namespace)
		if err := r.Status().Update(ctx, rt); err != nil {
			r.Log.Error(err, "update cluster status")
			return err
		}
	}
	return nil
}
