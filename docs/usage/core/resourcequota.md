# ResourceQuota

A `ResourceQuota` in `Ironcore` provides a mechanism to manage and limit the usage of resources across multiple requesting entities. This allows to protect a system from usage spikes and services can be kept responsive. With the help of `ResourceQuota` user can define a hard limit with list of resources along with `ScopeSelector`. The `ResourcequotaController` reconciler leverages this information to create a ResourceQuota in Ironcore infrastructure.
(`Note`: ResourceQuota is a namespaced resource and it can only limit resource count/accumulated resource usage within deifned namespace)
	
## Example ResourceQuota Resource
An example of how to define a `ResourceQuota` in `Ironcore`
```
apiVersion: core.ironcore.dev/v1alpha1
kind: ResourceQuota
metadata:
  name: resource-quota-sample
spec:
  hard: # Hard is the mapping of strictly enforced resource limits.
    requests.cpu: "10"
    requests.memory: 100Gi
    requests.storage: 10Ti
```

# Key Fields:
- `hard`(`ResourceList`): hard is a `ResourceList` of the strictly enforced number of resources. `ResourceList` is a list of ResourceName alongside their resource quantity.
- `scopeSelector`(`ResourceScopeSelector`): scopeSelector selects the resources that are subject to this quota. (`Note`: By using scopeSelectors, only certain resources may be tracked.)

(`Note`: Refer to <a href="https://github.com/ironcore-dev/ironcore/blob/main/docs/api-reference/core.md">API Reference</a> for more detailed description of `ResourceList` and `ResourceScopeSelector`.)

# Reconciliation Process:

- **Gathering matching evaluators**: ResourcequotaController retrieves all the matching evaluators from the registry for the specified resources in hard spec. Each resource evaluator implements set of Evaluator interface methods which helps in retrieving the current usage of that perticular resource type.
- **Calculating resource usage**: Resource usage is calculated by iterating over each evaluator and listing the namespace resource of that particular type. Listed resources are then filtred out by matching specified scope selector and accumulated usage is calculated. 
- **Status update**: Once usage data is available resource quota status is updated with the enforced hard resource limits and currently used resources.
- **Resource quota handling**: On request of create/update resources whether to allow create/update based on resource quota usage is handled via admission controller. Resources that would exceed the quota will fail with the HTTP status code 403 Forbidden.
