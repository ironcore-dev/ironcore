# `IP` allocation to Subnets with `IPv4`

This example allocates IPs of type `IPv4` to child subnets with the specified prefix length referring to the parent prefix. 
The following artifacts will be deployed in your namespace:   
- 1 IronCore parent `Prefix`, and 2 child `Prefixes`

## Usage
1. Adapt the `namespace` in `kustomization.yaml`

2. Run (`kubectl apply -k ./`)