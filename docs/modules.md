# CSF Modules

## Event sources

```bash
http GET http://$CSF_ADDR:8888/v1/events
```

### Clair

Clair is a [vulnerability scanner](https://github.com/coreos/clair) for docker images.
This event source listens for [clair notifications](https://github.com/coreos/clair/blob/master/Documentation/notifications.md) and retrieves notification details as a CSF event.

Configuration:

```yaml
notification_endpoint: http://localhost:6060/notifications
webhook_listen_addr: :4242
```

Event:
```json
{
  "data": [
    {
      "name": "new",
      "type": "Object containing new vuln details"
    },
    {
      "name": "old",
      "type": "Object containing old vuln details"
    }
  ],
  "name": "new_vuln"
}
```

See [clair API swagger](https://app.swaggerhub.com/apis/coreos/clair/3.0#/NotificationService/GetNotification) for more details.

### Kubernetes events

Currently the only event implemented is `new_pod` which send an event for each new pod created. You can easily add other k8s ressources to watch using the `NewListWatchFromClient` function.

The event data is simply a [k8s pod object](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11/#pod-v1-core).

Event:
```json
{
            "data": [
                {
                    "name": "pod",
                    "type": "k8s pod struct k8s.io/api/core/v1.Pod"
                }
            ],
            "name": "new_pod"
}
```

Configuration:

A working kube client config (located by default in `$HOME/.kube/config`).

### Kubernetes ImagePolicyWebhook

Implements the [ImagePolicyWebhook](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#what-does-each-admission-controller-do) of kuberenetes as an event and an associated action. This allows to accept or deny any image that your k8s cluster wants to use.

The event data is an uuid need by the `k8s_imagevalidate` action module to reply to the request on the webhook, and the name of the image.

Configuration:

```yaml
webhook_endpoint: "[::]:1323"
certificate: "/path/to/webhook.crt"
key: "/path/to/webhook.key"
```

Event:
```json
{
            "data": [
                {
                    "name": "image_review",
                    "type": "k8s object ImageReview"
                },
                {
                    "name": "uuid",
                    "type": "uuid of the request"
                }
            ],
            "name": "k8s_image_validation_request"
}
```

### Openstack

This event source connects to RabbitMQ via AMQP and listen for compute-related notifications. Every notification is then sent as a CSF event.

Configuration:

```yaml
rabbit_url: "amqp://user:password@host:5672//"
```

Event:

```json
{
            "data": [
                {
                    "name": "instance",
                    "type": "instance object"
                }
            ],
            "name": "compute_event"
        }
```

## Action modules

```bash
http GET http://$CSF_ADDR:8888/v1/actions
```

### Klar

This action module uses an external clair server to scan docker images and the codebase from [klar](https://github.com/optiopay/klar), an open-source cli client for clair.

Configuration:

```yaml
clairAddress: localhost
clairOutput: Low
threshold: 0
clairTimeout: 60
dockerConfig:
  user: ""
  password: ""
  token: ""
  insecureTLS: false
  insecureRegistry: false
  timeout: 60
```

The action `scan_image` will send a scan request to clair and fetch the results. `scan_images` can scan multiples images at once.

Actions:

```json
"klar": [
        {
            "data_in": [
                {
                    "name": "image",
                    "type": "string, docker image name"
                }
            ],
            "data_out": [
                {
                    "name": "vulns",
                    "type": "array of vulnerabilities"
                }
            ],
            "name": "scan_image"
        },
        {
            "data_in": [
                {
                    "name": "images",
                    "type": "[]string, array of docker image names"
                }
            ],
            "data_out": [
                {
                    "name": "{{name of the image scanned}}",
                    "type": "array of vulnerabilities"
                }
            ],
            "name": "scan_images"
        }
    ]
```

### Kubernetes

Actions implemented in the `k8s` action modules allows to check if an image is deployed in a deployement or in a pod.

Configuration:

A working kube client config (located by default in `$HOME/.kube/config`).

Actions:

```json
"k8s": [
        {
            "data_in": [
                {
                    "name": "image",
                    "type": "string"
                },
                {
                    "name": "namespace",
                    "type": "string"
                }
            ],
            "data_out": [
                {
                    "name": "deployments",
                    "type": "array of strings (names of deployments)"
                }
            ],
            "name": "is_image_deployed"
        },
        {
            "data_in": [
                {
                    "name": "image",
                    "type": "string"
                },
                {
                    "name": "namespace",
                    "type": "string"
                }
            ],
            "data_out": [
                {
                    "name": "pods",
                    "type": "array of strings (names of pods)"
                }
            ],
            "name": "is_image_in_pods"
        }
    ]
```

### Send a mail

Intended for alerting purposes (but can be used for anything), you'll need an SMTP server and your credentials.

Configuration:

```yaml
server: smtp.gmail.com
user: user@gmail.com
password: yourSuperSecurePassword
port: 587
```

Action:

```json
"mail": [
        {
            "data_in": [
                {
                    "name": "to",
                    "type": "string"
                },
                {
                    "name": "subject",
                    "type": "string"
                },
                {
                    "name": "content",
                    "type": "string"
                }
            ],
            "data_out": [],
            "name": "send_mail"
        }
    ]
```

### Vuls.io: scan a Linux/FreeBSD instance / host

This action modules uses docker (usually the host socket) to launch a container that will perform a scan on a given host via ssh. It does not handle vulnerabilities download and update, see [vuls.io docs](https://vuls.io/docs/en/tutorial-docker.html).

Configuration:

```yaml
db_path: /path/to/vuls/db
logs_path: /path/to/vuls/logs
ssh_key_path: /path/to/.ssh/id_rsa
use_host_docker_socket: true
docker_endpoint: unix:///var/run/docker.sock
```

Actions:

```json
"vuls": [
        {
            "data_in": [
                {
                    "name": "host",
                    "type": "string"
                },
                {
                    "name": "user",
                    "type": "string"
                },
                {
                    "name": "port",
                    "type": "string"
                },
                {
                    "name": "deep_scan",
                    "type": "bool"
                }
            ],
            "data_out": [
                {
                    "name": "result",
                    "type": "string (json of vuls result)"
                }
            ],
            "name": "scan_instance"
        }
    ]
```