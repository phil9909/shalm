
# Scriptable helm charts


This project brings the starlark scripting language to helm charts.

## Features

* Define APIs for helm charts
* Ease composition of charts
* Control deployment by overriding methods
* Compatible with helm
* Share a common service like a database manager or an ingress between a set of sub charts
* Use starlark methods in templates (replacement for `_helpers.tpl`)
* Interact with kubernetes during installation
* Manage user credentials
* Act as glue code between helm charts
* Rendering of [ytt templates](https://get-ytt.io/)
* Also available as kubernetes controller

## Installation


### Prerequisite 
* Install `kubectl` e.g. using `brew install kubernetes-cli`

### Install binary

* Download `shalm` (e.g. for mac os)
```bash
curl https://github.com/kramerul/shalm/releases/download/0.1.0/shalm-binary-darwin.tgz | tar xzf
```

### Build `shalm` from source

* Install `go` e.g. using `brew install go`
* Install `shalm`
```bash
go get github.com/kramerul/shalm
```

## Usage

```bash
shalm template <chart>
shalm apply <chart>
shalm delete <chart>
shalm package <chart>
```

A set of example charts can be found in the `charts/examples` folder.

Charts can be given by path or by url. In case of an url, the chart must be packaged using `shalm package`.

## Writing charts

Just follow the rules of helm to write charts. Additionally, you can put a `Chart.star` file in the charts folder

```bash
<chart>/
├── Chart.yaml
├── values.yaml
├── Chart.star
└── templates/
└── ytt/
```

### Using ytt yaml templates

You can use ytt yaml templates to render kubernetes artifacts. You simpy put them in the `ytt` folder inside a chart.
There is currently no support for `data`, `star` or `text` files. The only value supplied to the templates is `self`, 
which is the current chart. You can access all values and methods within your chart.

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: #@ self.namespace
```

## Examples

### Share database

The following example shows how a database manager could be shared.

1. Define an API for a database manager (e.g. mariadb)

```python
def create_database(self,db="db",username="",password=""):
   ...
```


2. Define a constructor for a service, which requires a database

```python
def init(self,database=None):
  if database:
    database.create_database(db="uaa",username="uaa",password="randompass")
```


3. Use the API within another chart

```python
def init(self):
  self.mariadb = chart("mariadb")
  self.uaa = chart("uaa",database = self.mariadb)
```

### Override apply

With `shalm` it's possible to override the `apply` and `delete` methods. The following example illustrates how this could be done

```python
def init(self):
  self.mariadb = chart("mariadb")
  self.uaa = chart("uaa",database = self.mariadb)

def apply(self,k8s):
  self.mariadb.apply(k8s) # Apply mariadb stuff (recursive)
  k8s.rollout_status("statefulset","mariadb-master")  # Interact with kubernetes
  self.uaa.apply(k8s)     # Apply uaa stuff (recursive)
  self.__apply(k8s)       # Apply everthing defined in this chart (not recursive)
```

### Create User Credentials

User credentials are used to manage username and password pairs. They are mapped to kubernets `Secrets`. 
If the secret doesn't exist, the username and password are created with random content, otherwise the fields are
read from the secret. The keys used to store the username and password inside the secret can be modified.

The content of username and password can only be accessed after the call to `__apply`. 
Therefore, you need to override the `apply` method.

All user credentials created inside a `Chart.star` file are automatically applied to kubernetes.
If you run `shalm template`, the content of the username and password is undefined.

```python
def init(self):
   self.nats = chart("https://charts.bitnami.com/bitnami/nats-4.2.6.tgz")
   self.auth = user_credential("nats-auth")

def apply(self,k8s):
  self.__apply(k8s)
  self.nats.auth["user"] = self.auth.username
  self.nats.auth["password"] = self.auth.password
  self.nats.apply(k8s)
```

## Using shalm controller

Charts can be also applied (in parts) using the shalm controller.

### Install shalm controller

```bash
shalm apply charts/shalm
```

### Install a shalm chart using the controller

```bash
shalm apply --proxy <chart>
```

or from inside another shalm chart

```python
def init(self):
  self.mariadb = chart("mariadb",proxy=true)
````



## Comparison

|                                | shalm | helm  | ytt/kapp | kustomize |
|--------------------------------|-------|-------|----------|-----------|
| Scripting                      |   +   | (3.1) |  +       |    -      |
| API definition                 |   +   |   -   |  -       |    -      |
| Reuse of existing charts       |   +   |   +   |  -       |    ?      |
| Only simple logic in templates |   +   |   +   |  -       |    +      |
| Interaction with k8s           |   +   |   -   |  -       |    -      |
| Repository                     |   +   |   +   |  -       |    -      |
| Mature technology              |   -   |   +   |  +       |    +      |
| Manage user credentials        |   +   |   -   |  -       |    -      |
| Controller based installation  |   +   |   -   |  +       |    -      |
| Remove outdated objects        | +<sup>(1)</sup> |   +   |  +       |    -      |
| Migrate existing objects       | +<sup>(1)</sup> |   -   |  -       |    -      |

<sup>(1)</sup>: Must be implemented inside `apply` method

## Reference

The following section describes the available methods inside `Chart.star`

### Chart

#### `chart("<url>",namespace=namespace,proxy=false,...)`

An new chart is created.  
If no namespace is given, the namespace is inherited from the parent chart.

| Parameter | Description |
|-----------|-------------|
| `url`       |  The chart is loaded from the given url. The url can be relative.  In this case the chart is loaded from a path relative to the current chart location.  |
| `namespace` |  If no namespace is given, the namespace is inherited from the parent chart. |
| `suffix`    |  This suffix is appended to each chart name. The suffix is inhertied from the parent if no value is given|
| `proxy`     |  If true, a proxy for the chart is returned. Applying or deleting a proxy chart is done by applying a `CustomerResource` to kubernetes. The installation process is then performed by the `shalm-controller` in the background |
| `...`       |  Additional parameters are passed to the `init` method of the corresponding chart. |


#### `chart.apply(k8s)`

Applies the chart recursive to k8s. This method can be overwritten.

| Parameter | Description |
|-----------|-------------|
| `8s`       |  See below  |

#### `self.__apply(k8s,timeout=0,glob=pattern)`

Applies the chart to k8s without recursion. This should only be used within `apply`

| Parameter | Description |
|-----------|-------------|
| `k8s`       |  See below  |
| `timeout`   |  Timeout passed to `kubectl apply`. A timeout of zero means wait forever.  |
| `glob`      |  Pattern used to find the templates. Default is "*.yaml"  |


#### `chart.delete(k8s)`

Deletes the chart recursive from k8s. This method can be overwritten.

| Parameter | Description |
|-----------|-------------|
| `k8s`       |  See below  |


#### `self.__delete(k8s,timeout=0,glob=pattern)`

Deletes the chart from k8s without recursion. This should only be used within `delete`

| Parameter | Description |
|-----------|-------------|
| `k8s`       |  See below  |
| `timeout`   |  Timeout passed to `kubectl apply`, A timeout of zero means wait forever.  |
| `glob`      |  Pattern used to find the templates. Default is "*.yaml"  |

#### Attributes

| Name | Description |
|------------|-------------|
| `name`       |  Name of the chart. Defaults to `self.__class__.name`  |
| `namespace`  |  Default namespace of the chart given via command line |
| `__class__`  |  Class of the chart. See `chart_class` for details  |

### K8s

#### `k8s(kube_config_content)`

Create a new k8s object 

| Parameter | Description |
|-----------|-------------|
| `kube_config_content`  |  Content of kube config   |

#### `k8s.delete(kind,name,namespaced=false,timeout=0)`

Deletes one kubernetes object

| Parameter | Description |
|-----------|-------------|
| `kind`      |  k8s kind   |
| `name`      |  name of k8s object   |
| `timeout`   |  Timeout passed to `kubectl apply`. A timeout of zero means wait forever.  |
| `namespaced` |  If true object in the current namespace are deleted. Otherwise object in cluster scope will be deleted. Default is `true`  |

#### `k8s.get(kind,name,namespaced=false,timeout=0)`

Get one kubernetes object. The value is returned as a `dict`.

| Parameter | Description |
|-----------|-------------|
| `kind`      |  k8s kind   |
| `name`      |  name of k8s object   |
| `timeout`   |  Timeout passed to `kubectl get`. A timeout of zero means wait forever.  |
| `namespaced` |  If true object in the current namespace are listed. Otherwise object in cluster scope will be listed. Default is `true`  |

#### `k8s.watch(kind,name,namespaced=false,timeout=0)`

Watch one kubernetes object. The value is returned as a `iterator`.

| Parameter | Description |
|-----------|-------------|
| `kind`      |  k8s kind   |
| `name`      |  name of k8s object   |
| `timeout`   |  Timeout passed to `kubectl watch`. A timeout of zero means wait forever.  |
| `namespaced` |  If true object in the current namespace are listed. Otherwise object in cluster scope will be listed. Default is `true`  |

#### `k8s.rollout_status(kind,name,timeout=0)`

Wait for rollout status of one kubernetes object

| Parameter | Description |
|-----------|-------------|
| `kind`      |  k8s kind   |
| `name`      |  name of k8s object   |
| `timeout`   |  Timeout passed to `kubectl apply`. A timeout of zero means wait forever.  |

#### `k8s.wait(kind,name,condition, timeout=0)`

Wait for condition of one kubernetes object

| Parameter | Description |
|-----------|-------------|
| `kind`      |  k8s kind   |
| `name`      |  name of k8s object   |
| `condition`  |  condition   |
| `timeout`   |  Timeout passed to `kubectl apply`. A timeout of zero means wait forever.  |

### user_credential



#### `user_credential(name,username='',password='',username_key='username',password_key='password')`

Creates a new user credential. All user credentials created inside a `Chart.star` file are automatically applied to kubernetes.

| Parameter | Description |
|-----------|-------------|
| `name`      |  The name of the kubernetes secret used to hold the information   |
| `username`  |  Username. If it's empty it's either read from the secret or created with a random content.  |
| `password`  |  Password. If it's empty it's either read from the secret or created with a random content.  |
| `username_key` |  The name of the key used to store the username inside the secret  |
| `password_key` |  The name of the key used to store the password inside the secret  |


#### Attributes

| Name | Description |
|------------|-------------|
| `username` | Returns the content of the username attribute. It is only valid after calling `chart.__apply(k8s)` or it was set in the constructor. |
| `password` | Returns the content of the password attribute. It is only valid after calling `chart.__apply(k8s)` or it was set in the constructor. |


### struct

See [bazel documentation](https://docs.bazel.build/versions/master/skylark/lib/struct.html). `to_proto` and `to_json` are not yet supported.

### chart_class

The `chart_class` represents the values read from the `Chart.yaml` file

#### Attributes

| Name | Description |
|------------|-------------|
| `api_version` | API version |
| `name`        | Name |
| `version`     | Version |
| `description` | Description |
| `keywords`    | Keywords |
| `home`        | Home |
| `sources`     | Sources |
| `icon`        | Icon |


## Difference to helm

* Subcharts are not loaded automatically. They must be loaded using the `chart` command
* Global variables are not supported.
* The `--set` command line parameters are passed to the `init` method of the corresponding chart. 
It's not possible to set values (from `values.yaml`) directly. 
If you would like to set a lot of values, it's more convenient to write a separate shalm chart.
* `shalm` doesn't track installed charts on a kubernetes cluster. It works more like `kubectl apply`
* The `.Release.Name` value is build as follows: `<chart.name>-<chart.suffix>`. If no suffix is given, the hyphen is also ommited.

