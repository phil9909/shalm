
def init(self):
  self.uaa = chart("uaa")
  self.mariadb = chart("mariadb")
  self.uaa.attach_database(self.mariadb)
  self.HA = True
  self.use_istio = True
  self.name = "my-first-chart"
  return self


def secret_name(self):
  return "mysecret"