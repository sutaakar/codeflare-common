# codeflare-common
Common packages for use with CodeFlare Distributed Workload stack.

## GitHub actions collection
Folder github-actions contains collection of GitHub composite actions which can be reused in other repositories.
This approach is used to keep all the usable actions on one place, unlike from default usage when every action is located in dedicated repository.

Usage:
- Clone CodeFlare common repository
- Refer to the action file using `uses` step parameter pointing to the local folder path with action.yml

### KinD GitHub action
GitHub action which spins up KinD cluster with Ingress and local registry available in the cluster.

Action creates environment variables:
- `TEMP_DIR` - helper env variable, pointing to local temporary directory
- `REGISTRY_ADDRESS` - hostname of the locally available insecure registry
- `KIND_CONFIG_FILE` - helper env variable, pointing to the KinD config file
- `CLUSTER_TYPE` - type of cluster, hardcoded to `KIND`
- `CLUSTER_HOSTNAME` - hostname of the KinD cluster
