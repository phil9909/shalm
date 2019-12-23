def init(self):
  pass

def apply(self,k8s):
  self.__apply(k8s,glob="crd.yaml")
  k8s.wait("customresourcedefinition.apiextensions.k8s.io", "shalmcharts.kramerul.github.com","condition=established")
  self.__apply(k8s,glob="[^c][^r][^d]*.yaml")
