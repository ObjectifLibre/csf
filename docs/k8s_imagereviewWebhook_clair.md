# On-the-fly scanning of docker images used by kubernetes

Kubernetes has a Webhook that allow a third party software to allow or deny each docker image used by kubernetes. CSF implements this through the `k8s_imagevalidator`
event source and the `k8s_imagevalidate` action module.

## Configuration

### Generate certificates

```bash
echo "subjectAltName = IP:123.123.123.123" > extfile.cnf
openssl genrsa -out ca.key 2048
openssl req -x509 -new -nodes -key ca.key \
 -subj "/C=US/ST=CA/O=k8s-test" \
 -sha256 -days 1024 -out ca.crt
openssl genrsa -out client.key 2048
openssl genrsa -out webhook.key 2048
openssl req -new -key client.key -out client.csr \
 -subj "/C=US/ST=CA/O=Acme, Inc./CN=123.123.123.123"
openssl req -new -key webhook.key -out webhook.csr \
 -subj "/C=US/ST=CA/O=Acme, Inc./CN=123.123.123.123"
openssl x509 -req -in client.csr -CA ca.crt -extfile \
 extfile.cnf -CAkey ca.key -CAcreateserial \
 -out client.crt -days 500 -sha256
openssl x509 -req -in webhook.csr -CA ca.crt -extfile \
 extfile.cnf -CAkey ca.key -CAcreateserial \
 -out webhook.crt -days 500 -sha256
```

This is required because kubernetes will not accept certificates without SANs (alternative names). Do not forget to change 123.123.123.123 with your local ip address.

`webhook.crt` / `webhook.key` will be used by CSF and `client.key` / `client.crt` will be used by kubernetes.

### k8s configuration

You'll need a few files, I'm assuming you are using minikube to test this webhook.
Please ssh to your minkube using `minikube ssh` and create the folowing files:

admission_configuration.json:
```json
{
  "imagePolicy": {
     "kubeConfigFile": "/var/lib/localkube/certs/imagepolicy/kube-imagepolicy.yml",
     "allowTTL": 50,
     "denyTTL": 50,
     "retryBackoff": 500,
     "defaultAllow": false
  }
}
```

kube-imagepolicy.yml:
```
apiVersion: v1
kind: Config
clusters:
- cluster:
    certificate-authority: /var/lib/localkube/certs/imagepolicy/ca.crt
    server: https://123.123.123.123:1323/image_policy
  name: csf_webhook
contexts:
- context:
    cluster: csf_webhook
    user: api-server
  name: csf_validator
current-context: csf_validator
preferences: {}
users:
- name: api-server
  user:
    client-certificate: /var/lib/localkube/certs/imagepolicy/client.crt
    client-key:  /var/lib/localkube/certs/imagepolicy/client.key
```

Using minikube to test the webhook, it needs access to the certificates and config files when it starts, but minikube custom mounts folders are mounted after minikube startup.
So we use a folder not meant for this usage, `/var/lib/localkube/certs/`, which is mounted by default because the k8s api-server needs it.

You will also need `client.key`, `client.crt` and `ca.crt` on your minikube.

Create the `/var/lib/localkube/certs/imagepolicy` folder and copy all the files mentioned above in it.

### CSF configuration

Nothing fancy here, juste create a `k8s_imagevalidator.yml` file in your config folder (`config` by default):

```yaml
webhook_endpoint: "[::]:1323"
certificate: "/path/to/webhook.crt"
key: "/path/to/webhook.key"
```

Enable `k8s_imagevalidator` event source and `k8s_imagevalidate` action module in CSF config, then add the folowing reaction:

```json
{
    "name": "k8s_scan_images",
    "event": "k8s_image_validation_request",
    "script": "result = {'images': []}; for (i = 0; i < event['image_review']['spec']['containers'].length; i++) { container = event['image_review']['spec']['containers'][i]; if (container['image'].substring(0, 20) != 'gcr.io/k8s-minikube'){result['images'].push(container['image']);console.log(container['image'])}}; if (result['images'].length > 0) {nextAction = 'scan_images_clair'} else {nextAction = 'validate_image'; result = {'validate': true, 'uuid': event['uuid']};};",
    "actions": {
	"scan_images_clair": {
	    "action": "scan_images",
	    "module": "klar",
	    "script": "nextAction = 'validate_image'; result = {'uuid': event['uuid'], 'validate': true}; for (i = 0; i < event['image_review']['spec']['containers'].length; i++) { container = event['image_review']['spec']['containers'][i]; if (actionData[container['image']].length > 0) {result['validate'] = false; break;}}"
	},
	"validate_image": {
	    "action": "validate_image",
	    "module": "k8s_imagevalidate",
	    "script": ""
	}
    }
}
```

You can easily customize the script to add a severity threshold, whitelist registries or anything else you could need.

You can use httpie to easily use the API, assuming csf runs on your localhost:

```bash
cat k8s_imagepolicy_clair.json | http POST http://localhost:8888/v1/events/k8s_image_validation_request/reactions
```


### Clair

Put this clair config in a `clair_config` folder:

clair.yml:
```yaml
clair:
  database:
    type: pgsql
    options:
      source: host=localhost port=5432 user=postgres sslmode=disable statement_timeout=60000
      cachesize: 16384
      paginationkey:

  api:
    addr: "0.0.0.0:6060"
    healthaddr: "0.0.0.0:6061"
    timeout: 900s
    servername:
    cafile:
    keyfile:
    certfile:

  worker:
    namespace_detectors:
      - os-release
      - lsb-release
      - apt-sources
      - alpine-release
      - redhat-release

    feature_listers:
      - apk
      - dpkg
      - rpm

  updater:
    interval: 10m
    enabledupdaters:
      - debian
      - ubuntu
      - rhel
      - oracle
      - alpine

  notifier:
    attempts: 1

    renotifyinterval: 10m

    http:
      endpoint:
      servername:
      cafile:
      keyfile:
      certfile:

      proxy:
```


## Start things

Launch a postgresql database for clair:

```bash
docker run -d  -p 5432:5432  --name=pg  postgres:9.6
```

Start clair:

```bash
docker run --net=host -it -v $PWD/clair_config:/config quay.io/coreos/clair-git \
  -config=/config/config.yml -log-level=debug
```

You'll need to wait for clair to download all the vulnerabilities.


```bash
minikube start --extra-config=apiserver.admission-control="Initializers,NamespaceLifecycle,LimitRanger,ServiceAccount,DefaultStorageClass,DefaultTolerationSeconds,NodeRestriction,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,ResourceQuota,ImagePolicyWebhook" --extra-config=apiserver.admission-control-config-file=/var/lib/localkube/certs/imagepolicy/admission_configuration.json
```

Note that if you `minikube delete` you will need to copy the files in the certs folder again.

Launch CSF and deploy something on your minikube, you should see requests for docker images validations from kubernetes and scans in clair logs.
