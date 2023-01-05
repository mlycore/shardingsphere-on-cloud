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
	"github.com/apache/shardingsphere-on-cloud/shardingsphere-operator/pkg/reconcile"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
)

const defaultRequeueTime = 10

type ComputeNodeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// SetupWithManager sets up the controller with the Manager.
func (r *ComputeNodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.ComputeNode{}).
		Owns(&appsv1.Deployment{}).
		Owns(&v1.Service{}).
		Owns(&v1.ConfigMap{}).
		Complete(r)
}

func (r *ComputeNodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logger.FromContext(ctx)

	cn := &v1alpha1.ComputeNode{}
	if err := r.Get(ctx, req.NamespacedName, cn); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{RequeueAfter: defaultRequeueTime}, nil
		} else {
			log.Error(err, "get computenode")
			return ctrl.Result{Requeue: true}, err
		}
	}

	errors := []error{}
	if err := r.reconcileDeployment(ctx, cn); err != nil {
		log.Error(err, "Reconcile Deployment Error")
		errors = append(errors, err)
	}
	if err := r.reconcileService(ctx, cn); err != nil {
		log.Error(err, "Reconcile Service Error")
		errors = append(errors, err)
	}
	if err := r.reconcileConfigMap(ctx, cn); err != nil {
		log.Error(err, "Reconcile ConfigMap Error")
		errors = append(errors, err)
	}
	if err := r.reconcilePodList(ctx, cn); err != nil {
		log.Error(err, "Reconcile PodList Error")
		errors = append(errors, err)
	}

	if len(errors) != 0 {
		return ctrl.Result{Requeue: true}, errors[0]
	}

	return ctrl.Result{RequeueAfter: defaultRequeueTime}, nil
}

func (r *ComputeNodeReconciler) reconcileDeployment(ctx context.Context, cn *v1alpha1.ComputeNode) error {
	cur := &appsv1.Deployment{}
	if err := r.Get(ctx, types.NamespacedName{
		Namespace: cn.Namespace,
		Name:      cn.Name,
	}, cur); err != nil {
		if apierrors.IsNotFound(err) {
			// create
			exp := reconcile.ComputeNodeNewDeployment(cn)
			if err := r.Create(ctx, exp); err != nil {
				return err
			}
			return nil
		} else {
			return err
		}
	}
	// update
	exp := reconcile.ComputeNodeUpdateDeployment(cn, cur)
	if err := r.Update(ctx, exp); err != nil {
		return err
	}

	return nil
}

func (r *ComputeNodeReconciler) reconcileService(ctx context.Context, cn *v1alpha1.ComputeNode) error {
	cur := &v1.Service{}
	if err := r.Get(ctx, types.NamespacedName{
		Namespace: cn.Namespace,
		Name:      cn.Name,
	}, cur); err != nil {
		if apierrors.IsNotFound(err) {
			// create
			exp := reconcile.ComputeNodeNewService(cn)
			if err := r.Create(ctx, exp); err != nil {
				return err
			}
			return nil
		} else {
			return err
		}
	}
	// update
	exp := reconcile.ComputeNodeUpdateService(cn, cur)
	if err := r.Update(ctx, exp); err != nil {
		return err
	}

	return nil
}

func (r *ComputeNodeReconciler) reconcileConfigMap(ctx context.Context, cn *v1alpha1.ComputeNode) error {
	cur := &v1.ConfigMap{}
	if err := r.Get(ctx, types.NamespacedName{
		Namespace: cn.Namespace,
		Name:      cn.Name,
	}, cur); err != nil {
		if apierrors.IsNotFound(err) {
			// create
			exp := reconcile.ComputeNodeNewConfigMap(cn)
			if err := r.Create(ctx, exp); err != nil {
				return err
			}
			return nil
		} else {
			return err
		}
	}

	// update
	exp := reconcile.ComputeNodeUpdateConfigMap(cn, cur)
	if err := r.Update(ctx, exp); err != nil {
		return err
	}

	return nil
}

func (r *ComputeNodeReconciler) reconcilePodList(ctx context.Context, cn *v1alpha1.ComputeNode) error {
	list := &v1.PodList{}
	cur := &v1alpha1.ComputeNode{}
	if err := r.Get(ctx, types.NamespacedName{Namespace: cn.Namespace, Name: cn.Name}, cur); err != nil {
		return err
	}

	lbs := labels.Set{}
	lbs = cn.Labels
	opts := &client.ListOptions{
		LabelSelector: lbs.AsSelector(),
	}
	if err := r.List(ctx, list, client.InNamespace(cn.Namespace), opts); err != nil {
		return err
	}

	// the number of ready Pods
	readyNodes := reconcile.CountingReadyPods(list)

	// status not running, need reconcile
	if readyNodes != cur.Spec.Replicas {
		// check status
		if cur.Status.Phase != v1alpha1.ComputeNodeStatusNotReady {
			cur.Status.Phase = v1alpha1.ComputeNodeStatusNotReady
			// check condition
			if readyNodes < miniReadyCount {
				// initialization
				cur.Status.Conditions = append([]v1alpha1.ComputeNodeCondition{}, v1alpha1.ComputeNodeCondition{
					Type:           v1alpha1.ComputeNodeConditionInitialized,
					Status:         v1.ConditionTrue,
					LastUpdateTime: metav1.Now(),
				})
			} else {
				cur.Status.Conditions = append([]v1alpha1.ComputeNodeCondition{}, v1alpha1.ComputeNodeCondition{
					Type:           v1alpha1.ComputeNodeConditionStarted,
					Status:         v1.ConditionTrue,
					LastUpdateTime: metav1.Now(),
				})
			}
			//FIXME: how about status false and unknown?
		}
	} else {
		// status running
		if cur.Status.Phase != v1alpha1.ComputeNodeStatusReady {
			cur.Status.Phase = v1alpha1.ComputeNodeStatusReady
			cur.Status.Conditions = append([]v1alpha1.ComputeNodeCondition{}, v1alpha1.ComputeNodeCondition{
				Type:           v1alpha1.ComputeNodeConditionReady,
				Status:         v1.ConditionTrue,
				LastUpdateTime: metav1.Now(),
			})
		}
	}
	cur.Status.ReadyNodes = readyNodes

	// TODO: Compare Status with or without modification
	if err := r.Status().Update(ctx, cur); err != nil {
		fmt.Printf("update status error: %+v\n", err)
		return err
	}

	return nil
}
