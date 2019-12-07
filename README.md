# Brigade VSTS Gateway

## Deploy the gateway

First you have to clone this repo: `git clone https://github.com/radu-matei/brigade-vsts-gateway`.

As VSTS only requires a public endpoint to send events to, the way you deploy this gateway depends on what you have available on your cluster (it can be a service exposed as `NodePort` or `LoadBalancer`, or an ingress). For simplicity, the default service type is `LoadBalancer`, which should provision a new public IP address.

```
helm install -n brigade-vsts-gateway ./charts/brigade-vsts-gateway
```

> By default, the chart assumes your cluster has RBAC enabled. If you don't, either modify the `rbac.enabled` value in values.yaml to `false`, or pass `--set rbac.enabled=false` to the `helm install` command.

> If your deployed Brigade in another namespace than `default`, pass the `--set namespace=<your-namespace>`

> If you want to create an ingress for this gateway, change `serviceType` to `ClusterIP`, set `ingress.enabled` to `true` and provide a host. Also make sure to check the ingress annotations in `ingress.yaml`.

## Create an SSH key and use it in VSTS

- create an SSH key (if you want to reuse an existing key you can skip this step):

```
$ ssh-keygen -f ~/.ssh/vsts_rsa -C "<your-email>"
```

- [add the public key to VSTS following the instructions in the documentation][vsts-doc-key]

- now you're ready to create the Brigade project(note that you can find the clone URL from VSTS): 

```
project: "my/brigade-project"
repository: "<org>.visualstudio.com/<project>"

# This is an SSH clone URL (this can be found in VSTS)
cloneURL: "ssh://<org>@vs-ssh.visualstudio.com:22/<project>/_ssh/<project>"

# This is a string you generate and use when creating the webhook in VSTS
vstsToken: "super-secret-token-you-choose"

# paste your entire key here:
sshKey: |-
  -----BEGIN RSA PRIVATE KEY-----
  <PRIVATE KEY HERE>
  -----END RSA PRIVATE KEY-----
# The rest of your config...
```

> Pay attention at formatting when pasting your SSH key in `values.yaml` - wrong formatting will result in failure to clone the repo!

## Create the webhook to the gateway

Depending on how you exposed this gateway (service or ingress), you need to get the endpoint to it:

```
$ kubectl get svc
NAME                    TYPE           CLUSTER-IP     EXTERNAL-IP        PORT(S)        AGE
brigade-vsts-gateway    LoadBalancer   10.0.122.230   <your-public-IP>   80:30057/TCP   12m
```

Now you need to [setup the webhook][vsts-webhooks] using your endpoint:

`http://<your-public-ip-or-ingress>/vets/<brigade-project>/<secret-token>`

> Note that VSTS requries individial setup for each event type - you can reuse the same URL for all event types

Assuming you registered a webhook for `Code Pushed`, which generates a `git.push` event and you configured `brigade.js` as below:

```
const { events } = require('brigadier')

events.on("git.push", (e, p) => {
  console.log(e)
})
```

Whenever code is pushed to this repository, the associated `brigade.js` file will be executed - in this case it only logs the received event to the console, but you can write any pipeline you need.

```
==========[  brigade-worker-01cjt2tj91apbypsvcavwh19mm  ]==========
prestart: no dependencies file found
prestart: src/brigade.js written
[brigade] brigade-worker version: 0.15.0
[brigade:k8s] Creating PVC named brigade-worker-01cjt2tj91apbypsvcavwh19mm
{ buildID: '01cjt2tj91apbypsvcavwh19mm',
  workerID: 'brigade-worker-01cjt2tj91apbypsvcavwh19mm',
  type: 'git.push',
  provider: 'vsts',
  revision: { commit: 'HEAD', ref: 'master' },
  logLevel: 1,
  payload: '{"id":"a2344768-ac49-4979-b609-a5f3dd3086af","eventType":"git.push","publisherId":"tfs","scope":"","message":{"text":"Radu Matei pushed updates to :master\\r\\n(https://.visualstudio.com//_git//#version=GBmaster)","html":"Radu Matei pushed updates to \\u003ca href=\\"https://.visualstudio.com//_git//\\"\\u003e\\u003c/a\\u003e:\\u003ca href=\\"https://.visualstudio.com//_git//#version=GBmaster\\"\\u003emaster\\u003c/a\\u003e","markdown":"Radu Matei pushed updates to [](https://.visualstudio.com//_git//):[master](https://.visualstudio.com//_git//#version=GBmaster)"},"detailedMessage":{"text":"Radu Matei pushed a commit to :master\\r\\n - Added file brigade.js 5ad5571e (https://.visualstudio.com//_git//commit/5ad5571e2d4a87ea2c07597f22dcf55fce6bf791)","html":"Radu Matei pushed a commit to \\u003ca href=\\"https://.visualstudio.com//_git//\\"\\u003e\\u003c/a\\u003e:\\u003ca href=\\"https://.visualstudio.com//_git//#version=GBmaster\\"\\u003emaster\\u003c/a\\u003e\\r\\n\\u003cul\\u003e\\r\\n\\u003cli\\u003eAdded file brigade.js \\u003ca href=\\"https://.visualstudio.com//_git//commit/5ad5571e2d4a87ea2c07597f22dcf55fce6bf791\\"\\u003e5ad5571e\\u003c/a\\u003e\\u003c/li\\u003e\\r\\n\\u003c/ul\\u003e","markdown":"Radu Matei pushed a commit to [](https://.visualstudio.com//_git//):[master](https://.visualstudio.com//_git//#version=GBmaster)\\r\\n* Added file brigade.js [5ad5571e](https://.visualstudio.com//_git//commit/5ad5571e2d4a87ea2c07597f22dcf55fce6bf791)"},"resource":{"_links":{"commits":{"href":"https://.visualstudio.com/_apis/git/repositories//pushes/2/commits"},"pusher":{"href":""},"refs":{"href":"https://.visualstudio.com//_apis/git/repositories//refs/heads/master"},"repository":{"href":"https://.visualstudio.com//_apis/git/repositories/"},"self":{"href":"https://.visualstudio.com/_apis/git/repositories//pushes/2"}},"commits":[{"author":{"date":"2018-07-19T20:13:37Z","email":"","name":"Radu Matei"},"comment":"Added file brigade.js","commitId":"5ad5571e2d4a87ea2c07597f22dcf55fce6bf791","committer":{"date":"2018-07-19T20:13:37Z","email":"","name":"Radu Matei"},"url":"https://.visualstudio.com/_apis/git/repositories//commits/5ad5571e2d4a87ea2c07597f22dcf55fce6bf791"}],"date":"2018-07-19T20:13:37.9380287Z","pushId":2,"pushedBy":{"_links":{"avatar":{"href":"https://.visualstudio.com/_apis/GraphProfile/MemberAvatars/aad.ZTJjYjM1YWYtYzBlYi03YWIyLTk5MDYtM2IwYzI1N2IwMDFh"}},"descriptor":"aad.ZTJjYjM1YWYtYzBlYi03YWIyLTk5MDYtM2IwYzI1N2IwMDFh","displayName":"Radu Matei","id":"","imageUrl":"https://.visualstudio.com/_api/_common/identityImage?id=","uniqueName":"","url":""},"refUpdates":[{"name":"refs/heads/master","newObjectId":"5ad5571e2d4a87ea2c07597f22dcf55fce6bf791","oldObjectId":"66dd97a81a661523cad103253e3d4b2f37097c36"}],"repository":{"defaultBranch":"refs/heads/master","id":"","name":"","project":{"id":"","name":"","state":"wellFormed","url":"https://.visualstudio.com/_apis/projects/","visibility":"unchanged"},"remoteUrl":"https://.visualstudio.com//_git/","url":"https://.visualstudio.com/_apis/git/repositories/"},"url":"https://.visualstudio.com/_apis/git/repositories//pushes/2"},"resourceVersion":"1.0","resourceContainers":{"collection":{"id":"16c41dd0-3170-4e27-a8e9-ceefcddb13d7"},"account":{"id":"38a4f22b-1e27-40c9-9e83-15953e7dab69"},"project":{"id":""}},"createdDate":"2018-07-19T20:13:43.2693201Z"}' }
[brigade:app] after: default event handler fired
[brigade:app] beforeExit(2): destroying storage
[brigade:k8s] Destroying PVC named brigade-worker-01cjt2tj91apbypsvcavwh19mm
```

## Troubleshooting

If your builds don't get executed, first check if the Brigade worker has access to the repository:

```
$ kubectl logs <brigade-worker> -c vcs-sidecar

+ : HEAD
+ : master
+ : /vcs
+ refspec=master
+ git ls-remote --exit-code ssh://<org>@vs-ssh.visualstudio.com:22/<project>/_ssh/<project> master
+ cut -f2
Host key verification failed.
 fatal: Could not read from remote repository.
 Please make sure you have the correct access rights
and the repository exists.
+ full_ref=
+ git init -q /vcs
+ cd /vcs
+ git fetch -q --force --update-head-ok ssh://<org>@vs-ssh.visualstudio.com:22/<project>/_ssh/<project> master
Host key verification failed.
 fatal: Could not read from remote repository.
 Please make sure you have the correct access rights
and the repository exists.
```

This means the key is wrong, or passed incorrectly to the Brigade project.

## Building from source and running locally

Prerequisites:
- [Docker][docker]
- `make` (optional)

To update vendored dependencies:
- `make dep`

To verify vendored dependencies:
- `make verify-vendored-code`

To build from source:
- `make build`

Running locally:
- if running locally, you should provide an environment variable for the Kubernetes configuration file:
  - on Linux (including Windows Subsystem for Linux) and macOS: `export KUBECONFIG=<path-to-config>`
  - on Windows: `$env:KUBECONFIG="<path-to-config>"` 

- starting the binary will start listening on port 8080

- at this point, your server should be able to start accepting incoming requests to `localhost:8080`
- you can test the server locally, using Postman (POST requests with your desired JSON payload - see the `testdata` folders used for testing)
- please note that running locally with a Kubernetes config file set is equivalent to running privileged inside the cluster, and any Brigade builds created will get executed!

# Contributing

This Brigade project accepts contributions via GitHub pull requests. This document outlines the process to help get your contribution accepted.

## Signed commits

A DCO sign-off is required for contributions to repos in the brigadecore org.  See the documentation in
[Brigade's Contributing guide](https://github.com/brigadecore/brigade/blob/master/CONTRIBUTING.md#signed-commits)
for how this is done.


[vsts-doc-key]: https://docs.microsoft.com/en-us/vsts/git/use-ssh-keys-to-authenticate?view=vsts#step-2--add-the-public-key-to-vststfs
[vsts-webhooks]: https://docs.microsoft.com/en-us/vsts/service-hooks/services/webhooks
[go]: https://golang.org/doc/install
[dep]: https://github.com/golang/dep
