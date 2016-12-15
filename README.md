<!--[metadata]>
+++
title = "QingCloud driver for docker machine"
description = "QingCloud driver for docker machine"
keywords = ["machine, qingcloud, driver"]
+++
<![end-metadata]-->

# Docker Machine Driver of QingCloud
Create machines on [QingCloud](https://qingcloud.com/).  You will need an Access Key ID, Secret Access Key and a Zone ID.

creates docker instances on QingCloud.

```bash
docker-machine create -d qingcloud docker-machine
```

## Installation

The easiest way to install the QingCloud docker-machine driver is to:

```
go get github.com/yunify/docker-machine-driver-qingcloud
```

binaries also available,you can download from [releases](https://github.com/yunify/docker-machine-driver-qingcloud/releases)


## Example Usage
eg. Export your credentials into your shell environment

```bash
export QINGCLOUD_ACCESS_KEY_ID='<Your access key ID>'
export QINGCLOUD_SECRET_ACCESS_KEY='<Your secret access key>'
export QINGCLOUD_ZONE='<The QingCloud zone id>'
export QINGCLOUD_VXNET_ID='<Vxnet id>'
export QINGCLOUD_LOGIN_KEYPAIR='<Login ssh keypiar name>'
export QINGCLOUD_SSH_KEYPATH='<Ssh key local path>'


docker-machine create -d qingcloud <machine-name>
```

or  pass as cmdline flags

```bash
docker-machine create --driver qingcloud --qingcloud-access-key-id <Access key id> --qingcloud-secret-access-key <Secret access key> --qingcloud-vxnet-id <VxNet id> --qingcloud-login-keypair <Ssh key pair name> --qingcloud-ssh-keypath <Ssh key local path> <machine-name>
```

## Options

```bash
docker-machine create -d qingcloud --help
```

| CLI option                       | Environment variable        | Default      |Description                                      |
|----------------------------------|-----------------------------|--------------|-------------------------------------------------|
|--qingcloud-access-key-id 		   |QINGCLOUD_ACCESS_KEY_ID		 |				|QingCloud access key id
|--qingcloud-secret-access-key     |QINGCLOUD_SECRET_ACCESS_KEY	 | 				|QingCloud secret access key
|--qingcloud-cpu			       |							 |1             |QingCloud cpu count
|--qingcloud-memory     		   | 							 |1024	        |QingCloud memory size in MB
|--qingcloud-image          	   |QINGCLOUD_IMAGE  			 |xenialx64b	|Instance image ID,default is ubuntu16.4
|--qingcloud-login-keypair 		   |QINGCLOUD_LOGIN_KEYPAIR		 |				|Login keypair id
|--qingcloud-ssh-keypath 		   |QINGCLOUD_SSH_KEYPATH		 |~/.ssh/id_rsa	|SSH Key for Instance
|--qingcloud-vxnet-id 			   |QINGCLOUD_VXNET_ID			 |				|Vxnet id
|--qingcloud-zone       		   |QINGCLOUD_ZONE				 |pek3a 		|QingCloud zone

## Note
1. Only support create docker machine in vpc currently.
2. If run on your local machine, you must connect the vpc by vpn, this driver does not automatically assign public ip.
3. The qingcloud-ssh-keypath should match with qingcloud-login-keypair.

## Related links

- **Docker Machine**: https://docs.docker.com/machine/
- **Report bugs**: https://github.com/yunify/docker-machine-driver-qingcloud/issues

## License

Apache 2.0
