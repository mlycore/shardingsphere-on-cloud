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

	"github.com/apache/shardingsphere-on-cloud/shardingsphere-operator/api/v1alpha1"
	"github.com/apache/shardingsphere-on-cloud/shardingsphere-operator/pkg/reconcile"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
)

type DistSQLJobReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *DistSQLJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logger.FromContext(ctx)

	rt, err := r.getRuntimeDistSQLJob(ctx, req.NamespacedName)
	if apierrors.IsNotFound(err) {
		log.Info("Resource in work queue no longer exists!")
		return ctrl.Result{}, nil
	} else if err != nil {
		log.Error(err, "Error getting CRD resource")
		return ctrl.Result{}, err
	}

	return r.reconcile(ctx, req, rt)
}

func (r *DistSQLJobReconciler) reconcile(ctx context.Context, req ctrl.Request, rt *v1alpha1.DistSQLJob) (ctrl.Result, error) {
	namespacedName := types.NamespacedName{
		Namespace: rt.Namespace,
		Name:      rt.Name,
	}

	job := &batchv1.Job{}
	if err := r.Get(ctx, namespacedName, job); err != nil {
		if !apierrors.IsNotFound(err) {
			return ctrl.Result{}, err
		} else {
			exp := reconcile.NewJob(ctx, rt)
			if err := r.Create(ctx, exp); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{}, nil
}

func (r *DistSQLJobReconciler) getRuntimeDistSQLJob(ctx context.Context, namespacedName types.NamespacedName) (*v1alpha1.DistSQLJob, error) {
	rt := &v1alpha1.DistSQLJob{}
	err := r.Get(ctx, namespacedName, rt)
	return rt, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *DistSQLJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.DistSQLJob{}).
		Owns(&batchv1.Job{}).
		Owns(&v1.Pod{}).
		Complete(r)
}
