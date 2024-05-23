<h1 align="center">Kubecontroller </h1>

**Kubecontroller** is a custom kubernetes controller built utilizing [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) with the purpose of using a **Custom Resource Definition** (CRD) known as `ClusterScan` to reconcile one-off [Jobs](https://kubernetes.io/docs/concepts/workloads/controllers/job/) and re-occuring [CronJobs](https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/).

A `ClusterScan` is defined as follows:

```golang
type ClusterScan struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterScanSpec   `json:"spec,omitempty"`
	Status ClusterScanStatus `json:"status,omitempty"`
}
```

Looking closely at the struct definition, the major areas to focus on are the `ClusterScanSpec` and `ClusterScanStatus` which define what the `ClusterScan` resource is and what information is being updated throughout its lifetime.

The `ClusterScanSpec` is defined as follows:

```golang
type ClusterScanSpec struct {
	// Specifies whether this is a recurring job or not
	Recurring bool `json:"recurring"`

	// The schedule in Cron format, see https://en.wikipedia.org/wiki/Cron.
	Schedule string `json:"schedule"`

	// Specifies the job that will be created when executing a CronJob.
	JobTemplate batchv1.JobTemplateSpec `json:"jobTemplate"`
}
```

`recurring`: Used to differentiate between jobs and cron jobs. This is a mandatory parameter.

`Schedule`: Used to specify the schedule at which the cronjob will be executed. This is a mandatory parameter.

`JobTemplateSpec`: Used to specify the details of the job/cronjob that will be executed. This is a mandatory parameter.

Finally, the `ClusterScanStatus` is defined as follows:

```golang
type ClusterScanStatus struct {
	// Information when was the last time the job was successfully scheduled.
	// +optional
	LastScheduleTime *metav1.Time `json:"lastScheduleTime,omitempty"`

	// A detailed status message indicating the progress/completion status of the job.
	StatusMessage string `json:"statusMessage,omitempty"`
}
```

`LastScheduleTime`: Used to keep track of the time at which a job was scheduled for the specific ClusterScan instance.

`StatusMessage`: Used to describe what is going on with the ClusterScan instance, such as creating the specified jobs/cronjobs.

## Table of Contents
- [Table of Contents](#table-of-contents)
- [Prerequisites](#prerequisites)
- [Setup](#setup)
- [Cluster Deployment](#cluster-deployment)
  - [To Uninstall](#to-uninstall)
- [Project Distribution](#project-distribution)
- [Contributing](#contributing)
- [License](#license)


## Prerequisites

- [Go](https://go.dev/dl/) version v1.21.0+
- [Docker](https://docs.docker.com/get-docker/) version 17.03+.
- [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl) version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.
  - [kind](https://kind.sigs.k8s.io/)
  - [minikube](https://minikube.sigs.k8s.io/docs/)

## Setup

1. Clone this repository to your local machine.

```bash
git clone https://github.com/vikiru/kubecontroller.git
cd kubecontroller
```

2. Setup a local cluster using either kind or minikube.

   1. Setup a local cluster using kind
    ```bash
    kind create cluster
    ```
   2. Setup a local cluster using minikube
    ```bash
    minikube start
    ```
3. Install custom resource definition, `ClusterScan` onto the local cluster

```bash
make install
```

4. In a separate terminal, create custom resources and apply them to the cluster

    1. For example, to create a ClusterScan resource that is set to run a cron job
    ```bash
    kubectl apply -f config/samples/cronjob_sample.yaml
    ```
    2. Alernatively, create a ClusterScan resource that is set to run a job
    ```bash
    kubectl apply -f config/samples/job_sample.yaml
    ```

5. Observe the newly created resources.

```bash
# To observe all jobs within the cluster 
kubectl get jobs

# To observe all cronjobs within the cluster
kubectl get cronjobs

# To observe all clusterscans within the cluster
kubectl get clusterscans
```

6. (Optional) Delete resources with kubectl.

```bash
kubectl delete <custom resource name> <job name>
```


## Cluster Deployment

**Build and push your image to the location specified by `IMG`:**

```sh
make docker-build docker-push IMG=<some-registry>/kubecontroller:tag
```

**NOTE:** This image ought to be published in the personal registry you specified.
And it is required to have access to pull the image from the working environment.
Make sure you have the proper permission to the registry if the above commands don’t work.

**Install the `ClusterScan` api into the cluster:**

```sh
make install
```

**Deploy the Manager to the cluster with the image specified by `IMG`:**

```sh
make deploy IMG=<some-registry>/kubecontroller:tag
```

> **NOTE**: If you encounter RBAC errors, you may need to grant yourself cluster-admin
privileges or be logged in as admin.

**Create instances of your solution**
You can apply the samples (examples) from the config/sample:

```sh
kubectl apply -k config/samples/
```

>**NOTE**: Ensure that the samples has default values to test it out.

### To Uninstall
**Delete the instances (CRs) from the cluster:**

```sh
kubectl delete -k config/samples/
```

**Delete the APIs(CRDs) from the cluster:**

```sh
make uninstall
```

**UnDeploy the controller from the cluster:**

```sh
make undeploy
```

## Project Distribution

Following are the steps to build the installer and distribute this project to users.

1. Build the installer for the image built and published in the registry:

```sh
make build-installer IMG=<some-registry>/kubecontroller:tag
```

NOTE: The makefile target mentioned above generates an 'install.yaml'
file in the dist directory. This file contains all the resources built
with Kustomize, which are necessary to install this project without
its dependencies.

2. Using the installer

Users can just run kubectl apply -f <URL for YAML BUNDLE> to install the project, i.e.:

```sh
kubectl apply -f https://raw.githubusercontent.com/<org>/kubecontroller/<tag or branch>/dist/install.yaml
```

## Contributing
// TODO(user): Add detailed information on how you would like others to contribute to this project

**NOTE:** Run `make help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

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

