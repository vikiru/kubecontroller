/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	webappv1 "clusterscan.api.io/kubecontroller/api/v1"
)

// ClusterScanReconciler reconciles a ClusterScan object
type ClusterScanReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=webapp.clusterscan.api.io,resources=clusterscans,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=webapp.clusterscan.api.io,resources=clusterscans/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=webapp.clusterscan.api.io,resources=clusterscans/finalizers,verbs=update
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch,resources=jobs/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ClusterScan object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.3/pkg/reconcile
func (r *ClusterScanReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Retrieve the ClusterScan instance, output error if not found
	var clusterScan webappv1.ClusterScan
	if err := r.Get(ctx, req.NamespacedName, &clusterScan); err != nil {
		log.Error(err, "unable to fetch ClusterScan")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Obtain the list of all child jobs
	var childJobs batchv1.JobList
	if err := r.List(ctx, &childJobs, client.InNamespace(req.Namespace)); err != nil {
		log.Error(err, "unable to list child Jobs")
		return ctrl.Result{}, err
	}

	// Keep track of all active, successful and failed jobs then output via logger

	var activeJobs []*batchv1.Job
	var successfulJobs []*batchv1.Job
	var failedJobs []*batchv1.Job

	for i, job := range childJobs.Items {
		_, finishType := returnJobStatus(&job)
		switch finishType {
		case "":
			activeJobs = append(activeJobs, &childJobs.Items[i])
		case batchv1.JobFailed:
			failedJobs = append(failedJobs, &childJobs.Items[i])
		case batchv1.JobComplete:
			successfulJobs = append(successfulJobs, &childJobs.Items[i])
		}
	}

	log.V(1).Info("job count", "active jobs", len(activeJobs), "successful jobs", len(successfulJobs), "failed jobs", len(failedJobs))

	// Determine the type of job that the ClusterScan instance represents
	jobType, _ := determineJobType(clusterScan)
	uniqueName := createUniqueName(jobType, clusterScan)

	// Handle recurring CronJobs
	if jobType == webappv1.ClusterJobType("CronJob") {
		// Create a CronJob and make the controller the owner
		cronJob, _ := createCronJob(clusterScan, uniqueName)
		if err := ctrl.SetControllerReference(&clusterScan, cronJob, r.Scheme); err != nil {
			log.Error(err, "unable to set controller reference for cron job")
			return ctrl.Result{}, err
		}

		// Check if the cron job already exists and if not, create the job
		var cronJobInstance batchv1.CronJob
		if err := r.Get(ctx, client.ObjectKey{Namespace: cronJob.Namespace, Name: cronJob.Name}, &cronJobInstance); err != nil {
			if errors.IsNotFound(err) {
				logMessage := fmt.Sprintf("Creating a new %s [%s] within %s namespace", jobType, cronJob.Name, cronJob.Namespace)
				log.Info(logMessage)
				if err := r.Create(ctx, cronJob); err != nil {
					log.Error(err, "unable to create CronJob")
					return ctrl.Result{}, err
				}
			}
		}
	} else {
		// Handle non-recurring Jobs

		// Create a Job and make the controller the owner
		job, _ := createJob(clusterScan, uniqueName)
		if err := ctrl.SetControllerReference(&clusterScan, job, r.Scheme); err != nil {
			log.Error(err, "unable to set controller reference for job")
			return ctrl.Result{}, err
		}

		// Check if the job already exists and if not, create the job
		var jobInstance batchv1.Job
		if err := r.Get(ctx, client.ObjectKey{Namespace: job.Namespace, Name: job.Name}, &jobInstance); err != nil {
			if errors.IsNotFound(err) {
				logMessage := fmt.Sprintf("Creating a new %s [%s] within %s namespace", jobType, job.Name, job.Namespace)
				log.Info(logMessage)
				if err := r.Create(ctx, job); err != nil {
					log.Error(err, "unable to create Job")
					return ctrl.Result{}, err
				}
			}
		}
	}

	clusterScan.Status.LastScheduleTime = &metav1.Time{Time: time.Now()}
	clusterScan.Status.StatusMessage = fmt.Sprintf("A %s with the name [%s] belonging to the namespace %s has been successfully scheduled", jobType, uniqueName, clusterScan.Namespace)

	if err := r.Update(ctx, &clusterScan); err != nil {
		log.Error(err, "Unable to update cluster scan status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// determineJobType returns the type of job that the ClusterScan instance represents.
func determineJobType(clusterScan webappv1.ClusterScan) (webappv1.ClusterJobType, error) {
	if clusterScan.Spec.Recurring {
		return webappv1.ClusterJobType("CronJob"), nil
	}
	return webappv1.ClusterJobType("Job"), nil
}

// createCronJob creates a CronJob matching the metadata and spec provided by the ClusterScan instance.
func createCronJob(clusterScan webappv1.ClusterScan, name string) (*batchv1.CronJob, error) {
	cronJob := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: clusterScan.Namespace,
		},
		Spec: batchv1.CronJobSpec{
			Schedule: clusterScan.Spec.Schedule,
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: clusterScan.Spec.JobTemplate.Spec,
			},
		},
	}
	return cronJob, nil
}

// createJob creates a Job matching the metadata and spec provided by the ClusterScan instance.
func createJob(clusterScan webappv1.ClusterScan, name string) (*batchv1.Job, error) {
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: clusterScan.Namespace,
		},
		Spec: clusterScan.Spec.JobTemplate.Spec,
	}
	return job, nil
}

// createUniqueName creates a unique name for the Job and CronJob matching the metadata and spec provided by the ClusterScan instance.
func createUniqueName(jobType webappv1.ClusterJobType, clusterScan webappv1.ClusterScan) string {
	return fmt.Sprintf("%s-%s", clusterScan.Name, strings.ToLower(string(jobType)))
}

// returnJobStatus returns the status of the Job, marking it as completed if it is completed or failed.
func returnJobStatus(job *batchv1.Job) (bool, batchv1.JobConditionType) {
	for _, c := range job.Status.Conditions {
		if (c.Type == batchv1.JobComplete || c.Type == batchv1.JobFailed) && c.Status == corev1.ConditionTrue {
			return true, c.Type
		}
	}
	return false, ""
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterScanReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&webappv1.ClusterScan{}).
		Owns(&batchv1.Job{}).
		Owns(&batchv1.CronJob{}).
		Complete(r)
}
