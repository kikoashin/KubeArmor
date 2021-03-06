# Example Microservice Deployment

To deploy the example microservice, please run the following commands.

```
$ cd examples/multiubuntu
(examples/multiubuntu) $ kubectl apply -f .
```

# Overview of Example Microservice

Here is the overview of the example microservice in terms of labels.

<center><img src=../../documentation/resources/multiubuntu.png></center>

To verify KubeArmor's functionalities, we provide sample security policies for the example microservice.

# Example 1 - Block a process execution

* Deploy a system policy

```
$ cd security-policies
(security-policies) $ kubectl -n multiubuntu apply -f ksp-group-1-proc-path-block.yaml
```

* Execute /bin/sleep

```
$ kubectl -n multiubuntu exec -it {pod name for ubuntu 1} -- bash
# sleep 1
(Permission Denied)
```

* Check audit logs

```
$ kubectl -n kube-system logs {KubeArmor in the node where ubuntu 1 is located}
```

# Example 2 - Block a file access

* Deploy a system policy

```
$ cd security-policies
(security-policies) $ kubectl -n multiubuntu apply -f ksp-ubuntu-5-file-dir-recursive-block.yaml
```

* Access /credentials/password

```
$ kubectl -n multiubuntu exec -it {pod name for ubuntu 5} -- bash
# cat cat /credentials/password
(Permission Denied)
```

* Check audit logs

```
$ kubectl -n kube-system logs {KubeArmor in the node where ubuntu 5 is located}
```
