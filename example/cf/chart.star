
def init(self):
  self.uaa = chart("uaa")
  self.database = chart("database")
  self.uaa.set_database(self.database)
  self.helm = helm("helm")
  self.helm.default["data"]["mybool"] = 1234
  self.HA = True
  self.use_istio = True
  self.name = "my-first-chart"
  return self

