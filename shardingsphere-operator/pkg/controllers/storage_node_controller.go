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
	"errors"
	"strings"

	"github.com/apache/shardingsphere-on-cloud/shardingsphere-operator/api/v1alpha1"
	"github.com/database-mesh/golang-sdk/aws/client/rds"
	meshapi "github.com/database-mesh/golang-sdk/kubernetes/api/v1alpha1"
	"github.com/database-mesh/golang-sdk/pkg/random"
	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	storageNodeControllerName = "storage-node-controller"
)

type StorageNodeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger

	AWSRds rds.RDS
	// DatabaseClass v1alpha1.StorageNode
}

func (r *StorageNodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.StorageNode{}).
		Complete(r)
}

func (r *StorageNodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.Log.WithValues(storageNodeControllerName, req.NamespacedName)

	sn := &v1alpha1.StorageNode{}
	if err := r.Get(ctx, req.NamespacedName, sn); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{RequeueAfter: defaultRequeueTime}, nil
		} else {
			logger.Error(err, "get cluster")
			return ctrl.Result{Requeue: true}, err
		}
	}

	logger.Info("storagenode found")
	if err := r.reconcileDatabaseClass(ctx, sn); err != nil {
		return ctrl.Result{Requeue: true}, err
	}

	return ctrl.Result{RequeueAfter: defaultRequeueTime}, nil
}

func (r *StorageNodeReconciler) reconcileDatabaseClass(ctx context.Context, sn *v1alpha1.StorageNode) error {
	if sn.Spec.DatabaseClassName != "" && sn.DeletionTimestamp.IsZero() {
		class := &meshapi.DatabaseClass{}
		err := r.Get(ctx, types.NamespacedName{
			Name: sn.Spec.DatabaseClassName,
		}, class)
		if err != nil {
			return err
		}

		switch class.Spec.Provisioner {
		case meshapi.DatabaseProvisionerAWSRdsInstance:
			err = r.reconcileAWSRdsInstance(ctx, class, sn)
		case meshapi.DatabaseProvisionerAWSRdsCluster:
			err = r.reconcileAWSRdsCluster(ctx, class, sn)
		case meshapi.DatabaseProvisionerAWSRdsAurora:
			err = r.reconcileAWSRdsAurora(ctx, class, sn)
		default:
			err = errors.New("unknown DatabaseClass provisioner")
		}
		return err
	}

	return nil
}

func (r *StorageNodeReconciler) reconcileAWSRdsInstance(ctx context.Context, class *meshapi.DatabaseClass, sn *v1alpha1.StorageNode) error {
	return nil
}

func (r *StorageNodeReconciler) reconcileAWSRdsCluster(ctx context.Context, class *meshapi.DatabaseClass, sn *v1alpha1.StorageNode) error {
	subnetGroupName := sn.Annotations[meshapi.AnnotationsSubnetGroupName]
	vpcSecurityGroupIds := sn.Annotations[meshapi.AnnotationsVPCSecurityGroupIds]
	availabilityZones := sn.Annotations[meshapi.AnnotationsAvailabilityZones]
	randompass := random.StringN(8)
	desc, err := r.AWSRds.Cluster().SetDBClusterIdentifier(sn.Name).Describe(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "DBClusterNotFound") {
			if err := r.AWSRds.Cluster().
				SetEngine(class.Spec.Engine.Name).
				SetEngineVersion(class.Spec.Engine.Version).
				SetDBClusterIdentifier(sn.Name).
				SetMasterUsername(class.Spec.DefaultMasterUsername).
				SetMasterUserPassword(randompass).
				SetDBClusterInstanceClass(class.Spec.Instance.Class).
				//NOTE
				// SetDatabaseName(svc.DatabaseMySQL.DB).
				SetDatabaseName("ssdb").
				SetAllocatedStorage(100).
				SetVpcSecurityGroupIds(strings.Split(vpcSecurityGroupIds, ",")).
				SetStorageType("io1").
				SetIOPS(1000).
				SetDBSubnetGroupName(subnetGroupName).
				SetAvailabilityZones(strings.Split(availabilityZones, ",")).
				SetPublicAccessible(class.Spec.PubliclyAccessible).
				// SetAllocatedStorage(class.Spec.Storage.AllocatedStorage).
				Create(ctx); err != nil {
				r.Log.Error(err, "desc rds cluster creat")
				return err
			}
			return nil
		}
		r.Log.Error(err, "desc rds cluster other")
		return err
	}

	//TODO: check available then distsql
	// 1. make sure it is available and set storage node status
	if desc != nil {
		pe := v1alpha1.Endpoint{
			Protocol: "MySQL",
			Address:  desc.PrimaryEndpoint,
			Port:     uint32(desc.Port),
			Status:   desc.Status,
			Arn:      desc.DBClusterArn,
		}

		re := v1alpha1.Endpoint{
			Protocol: "MySQL",
			Address:  desc.ReaderEndpoint,
			Port:     uint32(desc.Port),
			Status:   desc.Status,
			Arn:      desc.DBClusterArn,
		}

		rt := &v1alpha1.StorageNode{}
		if err := r.Get(ctx, types.NamespacedName{
			Namespace: sn.Namespace,
			Name:      sn.Name,
		}, rt); err != nil {
			return err
		}

		if rt.Status.Endpoints == nil {
			rt.Status.Endpoints = []v1alpha1.Endpoint{}
		}
		rt.Status.Endpoints = []v1alpha1.Endpoint{pe, re}

		if err := r.Status().Update(ctx, rt); err != nil {
			r.Log.Error(err, "update storagenode status")
			return err
		}
	}

	return nil
}

func (r *StorageNodeReconciler) reconcileAWSRdsAurora(ctx context.Context, class *meshapi.DatabaseClass, sn *v1alpha1.StorageNode) error {
	return nil
}
