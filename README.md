
# Scriptable helm charts


This project brings the starlark scripting language to helm charts.

## Features

* Calculate values read from `values.yaml`
* Calculate values to configure sub charts
* Use starlark methods in templates (replacement for `_helpers.tpl`)
* Ability to define an API for each helm chart


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
