
def init(self):
  self.mariadb = chart("../mariadb")
  self.mariadb.slave['replicas'] = 2
  self.uaa = chart("../uaa",self.mariadb)
  self.name = "my-first-chart"
  return self


def __secret_name(self):
  return "mysecret"

def apply(self, k8s, release):
  self.mariadb.apply(k8s,release)
  k8s.rollout_status(release.namespace,"statefulset","mariadb-master")
  self.uaa.apply(k8s,release)
  self.__apply(k8s,release)