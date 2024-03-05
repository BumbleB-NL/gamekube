# GameKube
Enterprise grade gameserver deployment on self hosted Kubernetes

# Prerequisites
- Basic knowledge about Kubernetes deployments, services and PV's/PVC's.
- Basic installation of Kubernetes Cluster (using rancher) [tested on a 7 node cluster, 3 control-plane & etcd nodes, 4 worker nodes].
- Helm installed.
- Preconfigured NFS server that can be accessed by the cluster.

# Components
The project consists of multiple components:
- MetalLB loadbalancer for providing access to the services.
- NFS CSI provisioner for dynamically creating PV's on a NFS server when they are needed.
- Deployments for multiple gameservers including factorio and minecraft servers.
- Options for configuring installations using configmaps and secretmaps.

![Functional](Designs/GameKubeFunctionalDesignV2.drawio.png)

Network design:

![Network](Designs/GameKubeNetworkV2.drawio.png)

# Installation
## NFS CSI provisioner
Required to dynamically provision each pod with a Persistant Volume.


# ToDo
-[ ] Userfriendly portal
-[ ] Container/pod isolation
-[ ] Deploy multiple of the same server within a users namespace without manually changing deployments.
-[ ] Configureable options like CPU, memory, and storage for users.
