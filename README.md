
# Scriptable helm charts


This project brings the starlark scripting language to helm charts.

## Features

* Define APIs for helm charts
* Ease composition of charts
* Control deployment by overriding methods
* Compatible with helm
* Share a common service like a database manager or an ingress between a set of sub charts
* Use starlark methods in templates (replacement for `_helpers.tpl`)

## Installation

go get 

## Examples

### Share database

Define an API for a database manager (e.g. mariadb)

```python
def create_database(self,db="db",username="",password=""):
   ...
```


Define an API for a service, which requires a database

```python
def init(self,database=None):
  if database:
    database.create_database(db="uaa",username="uaa",password="randompass")
```


Use the API within another chart

```python
def init(self):
  self.mariadb = chart("mariadb")
  self.uaa = chart("uaa",database = self.mariadb)
```

### Override apply

```python
def init(self):
  self.mariadb = chart("mariadb")
  self.uaa = chart("uaa",database = self.mariadb)

def apply(self,k8s,release):
  self.mariadb.apply(release)
  self.uaa.apply(release)
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

## TODO

* Allow access to kubernetes during apply or delete
  * Read existing secrets (e.g.`load_or_create_secret()`)
  * Read ClusterIP of service
  * Wait for deployment‚
  * Wait for
    * CRDs
    * Deployments
    * Statefulsets
* Implement Push and Pull with [OCI registry](https://github.com/opencontainers/distribution-spec/blob/master/spec.md)
* Add tags to helm charts `chart("mariadb:3.6.5")`
