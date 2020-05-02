# kubectl-wrapper-tekton-step
Send `cloudevent v2` to `slink`. it could work as `step` of [Tekton-pipeline](https://github.com/tektoncd/pipeline), the goal is to send `cloudevent` in step of tekton. 
There is sample of `receiver` in `cmd/receive`, could used for testing.

## Take a try
1. Build image  

    `make image TAG=V0.0.1`  

2. Push image to your favourite repo registry  

3. Run `yaml`s in `./deploy` in a `tekton` ready environment, don't forget to replace the image in `sender-deploy.yaml`  

## Install the Task

```
kubectl apply -f deploy/sender-deploy.yaml
```

## Inputs 

### Parameters

* **event-type**: Type of Event.
* **event-id**: ID of Event.
* **source**: Source of Event.
* **data**: Data of Event, support `Json` only, for example: '{"hello": "world!"}'
* **target**: Address of target, for example, http://helloworld-go.default.svc.cluster.local
* **slink**: Slink of target object. Actually it's a `corev1.ObjectReference`:
```
      apiVersion: serving.knative.dev/v1alpha1
      kind: Service
      name: helloworld-go
      namespace: default
```


## Usage

This TaskRun runs the Task to deploy the given Kubernetes resource.

```
apiVersion: tekton.dev/v1beta1
kind: TaskRun
metadata:
  name: sender
spec:
  taskRef:
    name: sender
  params:
  - name: slink
    value: |
      apiVersion: serving.knative.dev/v1alpha1
      kind: Service
      name: helloworld-go
      namespace: default
```


