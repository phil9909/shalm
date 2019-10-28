
def init(self):
  self.uaa = chart("uaa")
  self.database = chart("database")
  self.HA = True
  self.use_istio = True
  return self

