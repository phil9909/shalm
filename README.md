
# Scriptable helm charts


This project brings the starlark scripting language to helm charts.

## Features

* Calculate values read from `values.yaml`
* Calculate values to configure sub charts
* Use starlark methods in templates (replacement for `_helpers.tpl`)
* Ability to define an API for each helm chart
* Ease the configuration of sub charts
* Share a common service like a database manager or an ingress between a set of subcharts


## Example

Define an API for a database manager (e.g. mariadb)

```python
def create_database(self,db="db",username="",password=""):
   ...
```


Define an API for a service, which requires a database

```python
def attach_database(self,database):
  database.create_database(db="uaa",username="uaa",password="randompass")
```


Use the API within another chart

```python
def init(self):
  self.uaa = chart("uaa")
  self.mariadb = chart("mariadb")
  self.uaa.attach_database(self.mariadb)
```


## TODO

* Allow access to kubernetes during templating
  * Read existing secrets (e.g.`load_or_create_secret()`)
  * Read ClusterIP of service
* Implement Push and Pull with [OCI registry](https://github.com/opencontainers/distribution-spec/blob/master/spec.md)
* Add tags to helm charts `chart("mariadb:3.6.5")`
* Support passing parameters to `chart("mariadb",instances=5,rootpassword='2324234')`
* Add cobra command line interface
* Support `template`, `apply` and `delete` as cobra command