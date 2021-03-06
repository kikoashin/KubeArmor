# Docker and Kubernetes Installation

- Requirements

    You can install Docker and Kubernetes in any Ubuntu platforms (e.g., 16.04, 18.04, 20.04, and 20.10).

- Prerequisites

    You need to disable the swap partition in advance for Kubernetes setup.
    
    ```
    $ sudo vi /etc/fstab
    (comment out the line for swap)
    $ sudo reboot
    ```

- Docker Installation

    You can simply install Docker through the following command.
    
    ```
    $ cd KubeArmor/contributions/bare-metal/docker
    (docker) $ ./install_docker.sh
    (docker) $ exit
    ```

- Kubernetes Installation (single machine)

    Now, you are ready to install Kubernetes. Please run the following command.
    
    ```
    $ cd KubeArmor/contributions/bare-metal/k8s
    (k8s) $ ./install_kubernetes.sh
    (k8s) $ ./initialize_kubernetes.sh flannel master
    ```

    Instead of Flannel, you can use other CNIs too.
    
    ```
    (k8s) $ ./initialize_kubernetes.sh [ weave | calico | cilium ] master
    ```

- Kubernetes Installation (multiple machines)

    If you use multiple machines to set up a multi-node environment, Please run the following command.

    - Master Node
    
    ```
    $ cd KubeArmor/contributions/bare-metal/k8s
    (k8s) $ ./install_kubernetes.sh
    (k8s) $ ./initialize_kubernetes.sh [ flannel | weave | calico | cilium ]
    ```

    <font color=green>Here, the master node will only serve Kubernetes services since you do not put "master" at the above command end. However, if you also want to use the master node to deploy containers, you can put "master" at the above command end.</font>

    - Worker Node
    
    ```
    $ sudo kubeadm ... (the command that you get from the master node)
    ```
