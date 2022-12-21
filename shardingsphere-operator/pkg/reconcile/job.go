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
	"context"
	"fmt"
	"strings"

	"github.com/apache/shardingsphere-on-cloud/shardingsphere-operator/api/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewJob(ctx context.Context, rt *v1alpha1.DistSQLJob) *batchv1.Job {
	var (
		defaultCompletion  int32 = 1
		defaultParallelism int32 = 1
	)

	command := []string{"mysql"}
	args := []string{
		fmt.Sprintf("-h%s", rt.Spec.Credential.Host),
		fmt.Sprintf("-P%d", rt.Spec.Credential.Port),
		fmt.Sprintf("-u%s", rt.Spec.Credential.Username),
		fmt.Sprintf("-p%s", rt.Spec.Credential.Password),
		fmt.Sprintf("-e%s", strings.Join(rt.Spec.Scripts, ";")),
	}
	env := []corev1.EnvVar{
		{
			Name:  "MYSQL_ROOT_PASSWORD",
			Value: "root",
		},
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rt.Name,
			Namespace: rt.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(rt.GetObjectMeta(), rt.GroupVersionKind()),
			},
		},
		Spec: batchv1.JobSpec{
			Completions: &defaultCompletion,
			Parallelism: &defaultParallelism,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"apps": rt.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            "executor",
							Image:           "mysql:5.7",
							ImagePullPolicy: corev1.PullIfNotPresent,
							Env:             env,
							Command:         command,
							Args:            args,
						},
					},
					RestartPolicy: corev1.RestartPolicyOnFailure,
				},
			},
		},
	}
	return job
}
