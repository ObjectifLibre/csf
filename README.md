# Continuous Security Framework

[![Build Status](https://travis-ci.org/ObjectifLibre/csf.svg?branch=master)](https://travis-ci.org/ObjectifLibre/csf) [![FOSSA Status](https://app.fossa.io/api/projects/custom%2B4963%2Fgit%40github.com%3AObjectifLibre%2Fcsf.git.svg?type=shield)](https://app.fossa.io/projects/custom%2B4963%2Fgit%40github.com%3AObjectifLibre%2Fcsf.git?ref=badge_shield) [![codebeat badge](https://codebeat.co/badges/a3974bbc-b9e2-4a52-a260-af70cc06034b)](https://codebeat.co/projects/github-com-objectiflibre-csf-master)

Continuous Security Framework (CSF for short) is an open-source project aiming at enabling continous security in cloud infrastructures (but not only).
You can see it as IFTTT for the cloud. Similar to a typical continuous integration, CSF can be used to build pipelines composed of different tasks
with simple javascript that anyone can use to build powerful automatic decision-making scripts.

## Getting started

### Terminology

 - *Eventsource* - a module that will send events to CSF (ex: a new vulnerability has been found by clair)
 - *Action module* - a module that contains one or more actions (ex: send a mail)

### Installation

The best way to run csf is to use the docker image `objectiflibre/csf`. You can also download the binary or build CSF yourself.
Take a look at [this sample config](https://github.com/ObjectifLibre/csf/blob/master/csf_config/config_sample.yaml) and modify it if needed.

```bash
docker run -d \
  -v $PWD/csf_config:/csf_config \
  -v $PWD/csf_data:/db \
  -p 8888:8888 \
  objectiflibre/csf
```

### Use cases

Events trigger pipelines that can dynamically respond to events using scripts. Currently implemented events are:

- [Clair notification](https://github.com/coreos/clair) about a new vulnerability in a docker image
- A new pod is spawned in kubernetes
- An [ImagePolicyWebhook review request](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#what-does-each-admission-controller-do) from kubernetes (can allow or deny the use of an image in k8s)
- Notifications from Openstack Nova (via AMQP)

Currently implemented actions are:

- Send a mail
- Check if an image is in a kubernetes pod or deployment
- Respond to an ImagePolicyWebhook image review request
- Scan a docker image using an external clair server
- Scan an instance / virtual machine / host via ssh using [vuls.io](https://vuls.io) and docker

Need something else ? Open an issue or [write your own module](https://github.com/ObjectifLibre/csf/blob/master/docs/write_modules.md) !


### Pipelines

You can use multiple actions to easily build complex pipelines. Here is a simple example:

![example](https://raw.githubusercontent.com/ObjectifLibre/csf/master/docs/csf_example.png)

Another use case is [on the fly docker images scanning](https://github.com/ObjectifLibre/csf/blob/master/docs/k8s_imagereviewWebhook_clair.md) with kubernetes.

You can find different sample json files in the `samples` folder.

