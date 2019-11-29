
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

## Installation

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

A set of example charts can be found in the `examples` folder.

Charts can be given by path or by url. In case of an url, the chart must be packaged using `shalm package`.

## Writing chars

Just follow the rules of helm to write charts. Additionally, you can put a `Chart.star` file in the charts folder

```bash
<chart>/
├── Chart.yaml
├── values.yaml
├── Chart.star
└── templates/
```

## Examples

### Share database

The following example shows how a database manager could be shared.

1. Define an API for a database manager (e.g. mariadb)

```python
def create_database(self,db="db",username="",password=""):
   ...
```


2. Define an constructor for a service, which requires a database

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
  k8s.wait???                     # Interact with kubernetes (not defined yet)
  self.uaa.apply(k8s)     # Apply uaa stuff (recursive)
  self.__apply(k8s)       # Apply everthing defined in this chart (not recursive)
```


## Comparison

|                                | shalm | helm  | ytt | kustomize |
|--------------------------------|-------|-------|-----|-----------|
| Scripting                      |   +   | (3.1) |  +  |    -      |
| API definition                 |   +   |   -   |  -  |    -      |
| Reuse of existing charts       |   +   |   +   |  -  |    ?      |
| Only simple logic in templates |   +   |   +   |  -  |    +      |
| Interaction with k8s           |   +   |   -   |  -  |    -      |
| Repository                     |   +   |   +   |  -  |    -      |
| Mature technology              |   -   |   +   |  ?  |    +      |


## Reference

The following section describes the avaiable methods inside `Chart.star`

### Chart

#### `chart("<url>",namespace=namespace,...)`

An new chart is created.  
If no namespace is given, the namespace is inherited from the parent chart.

| Parameter | Description |
|-----------|-------------|
| url       |  The chart is loaded from the given url. The url can be relative.  In this case the chart is loaded from a path relative to the current chart location.  |
| namespace |  If no namespace is given, the namespace is inherited from the parent chart. |


#### `chart.apply(k8s)`

Applies the chart recursive to k8s. This method can be overwritten.

| Parameter | Description |
|-----------|-------------|
| k8s       |  See below  |

#### `self.__apply(k8s,timeout=0,glob=pattern)`

Applies the chart to k8s without recursion. This should only be used within `apply`

| Parameter | Description |
|-----------|-------------|
| k8s       |  See below  |
| timeout   |  Timeout passed to `kubectl apply`. A timeout of zero means wait forever.  |
| glob      |  Pattern used to find the templates. Default is "*.yaml"  |


#### `chart.delete(k8s)`

Deletes the chart recursive from k8s. This method can be overwritten.

| Parameter | Description |
|-----------|-------------|
| k8s       |  See below  |


#### `self.__delete(k8s,timeout=0,glob=pattern)`

Deletes the chart from k8s without recursion. This should only be used within `delete`

| Parameter | Description |
|-----------|-------------|
| k8s       |  See below  |
| timeout   |  Timeout passed to `kubectl apply`, A timeout of zero means wait forever.  |
| glob      |  Pattern used to find the templates. Default is "*.yaml"  |

### K8s

#### `k8s.delete(kind,name,namespaced=false,timeout=0,)`

Deletes one kubernetes object

| Parameter | Description |
|-----------|-------------|
| kind      |  k8s kind   |
| name      |  name of k8s object   |
| timeout   |  Timeout passed to `kubectl apply`. A timeout of zero means wait forever.  |
| namespaced |  If true object in the current namespace are deleted. Otherwise object in cluster scope will be deleted. Default is `true`  |

#### `k8s.rollout_status(kind,name,timeout=0,)`

Deletes one kubernetes object

| Parameter | Description |
|-----------|-------------|
| kind      |  k8s kind   |
| name      |  name of k8s object   |
| timeout   |  Timeout passed to `kubectl apply`. A timeout of zero means wait forever.  |


## Difference to helm

* Subcharts are not loaded automatically. They must be loaded using the `chart` command
* Global variables are not supported.

## TODO

* Allow access to kubernetes during apply or delete
  * Read existing secrets (e.g.`load_or_create_secret()`)
  * Read ClusterIP of service
