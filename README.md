# GameKube
Enterprise grade gameserver deployment on self hosted Kubernetes

# Prerequisites
Basic installation of Kubernetes Cluster (using rancher) [tested on a 7 node cluster, 3 control-plane & etcd nodes, 4 worker nodes].
Preconfigured NFS server that can be accessed by the cluster.

# Components
The project consists of multiple components as shown in the diagrams below.

![Functional](Designs/GameKubeFunctionalDesignV2.drawio.png)

Network design:
![Network](Designs/GameKubeNetworkV2.drawio.png)

# Installation
