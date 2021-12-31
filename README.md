#Kube Drift
## A controller for detecting drift(s) in a Kubernetes clusters
### What is it a drift?
- A drift is a situation where an object in a cluster has changed in a way that is not expected - desired state is not met.
- The controller is designed to detect changes and calculate the drift - difference between the desired state and the current state.
- The focus of this controller is to detect and prevent such a situation.


## Build and Deployment

```bash
 make docker-build docker-push IMG="hugomatus/kube-drift:v1alpha1
```

```bash
make deploy IMG="hugomatus/kube-drift:v1alpha1"
```

```bash
kubectl expose deployment kube-drift-controller-manager -n kube-drift-system --type=NodePort --name=kube-drift --port=8001 --target-port=8001```
```
## Drift API

![demo](assets/kube-drift-api.gif)