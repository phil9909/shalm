
def init(self):
  self.uaa = chart("uaa")
  self.mariadb = chart("mariadb")
  self.mariadb.slave['replicas'] = 2
  self.uaa.attach_database(self.mariadb)
  self.HA = True
  self.use_istio = True
  self.name = "my-first-chart"
  return self


def __secret_name(self):
  return "mysecret"