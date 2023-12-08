#### Pre-initialization

1. ##### The position of pre-initialization in the process of adding nodes

Submit the task of adding nodes ---> Reinstall the operating system on the node ---> **Execute pre-initialization** ---> Deploy the kubelet service ---> Execute post-initialization ---> Enable node scheduling --- > Add node completed

2. ##### Parameter Description

- The pre-initialization script will be executed on each server in the added node list
- Built-in variables can be used in pre-initialization scripts. For example, to add a node list, you can use built-in variables {{ .NodeIPList }}

3. ##### Scenes to be used

- Before the DaemonSet process starts, it depends on the special directory on the Woker node, which needs to be created in advance
- The DaemonSet process requires special kernel parameters, which need to be configured on the Woker node before the process starts
- DaemonSet starts in hostnetwork mode, inherits /etc/resolv.conf of Woker node, and needs to configure a special nameserver
- Other special configurations for business

4. ##### Example

- Create a business directory on the Woker node

```bash
mkdir /data/logs
```

- Woker node adjusts kernel parameters

```bash
echo 'net.core.rmem_default = 52428800
net.core.rmem_max = 52428800
net.core.wmem_default = 52428800
net.core.wmem_max = 52428800' >>/etc/sysctl.conf
sysctl -p
```

- Modify /etc/resolv.conf of the node host

```bash
mkdir -p /data/backup
cp -a /etc/resolv.conf /data/backup/resolv.conf_$(date +%Y%m%d%H%M%S)
echo 'nameserver 8.8.8.8
nameserver 114.114.114.114' >/etc/resolv.conf
```
#### Post-initialization

1. ##### The position of post-initialization in the process of adding nodes

Submit the task of adding a node ---> Reinstall the operating system on the node ---> Execute pre-initialization ---> Deploy the kubelet service ---> **Execute post-initialization** ---> Enable node scheduling --- > Add node completed

2. ##### Parameter Description

- simple script will be executed on every server in the list of added nodes
- Built-in variables can be used in simple scripts. For example, to add a node list, you can use the built-in variable {{ .NodeIPList }}
- The parameter type of stdops only supports parameters of the input box type, do not use other types of parameters
- Built-in variables can be used in the parameters of stdops, for example, to add a node list, you can use the built-in variable {{ .NodeIPList }}
- In stdops, if you want to use the default value of the parameter, just leave the parameter blank in the node template configuration

3. ##### Scenes to be used

- There is no fixed scene for post-initialization, and the business implements custom operations according to its own scene

- The simple script is the same as the pre-initialization, used for simple initialization scenarios, it is recommended not to be too complicated

- stdops process, using blueking stdops as the process engine for complex initialization scenarios

#### Clean up before node recycling

1. ##### Clean up the position in the node shrinking task before recycling the node
Submit the shrink task ---> set the node to be unschedulable ---> expel the pod on the deleted node ---> **execute node cleanup before recycling** ---> remove the node from the cluster ---> transfer the node or recycle the node ---> complete the shrink task

2. ##### Parameter Description

- simple script will be executed on every server in the added node list
- Built-in variables can be used in simple scripts. For example, to add a node list, you can use the built-in variable {{ .NodeIPList }}
- The stdops parameter type only supports parameters of the input box type, do not use other types of parameters
- Built-in variables can be used in the stdops parameter, for example, to add a node list, you can use the built-in variable {{ .NodeIPList }}
- If you want to use the default value of the stdops parameter, set the parameter to empty in the node template configuration

3. ##### Scenes to be used
- Clear data or logs before node scaling