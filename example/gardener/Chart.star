def init(self,shootname="test"):
  self.shootname = shootname
  self.domain = shootname + ".istio.shoot.canary.k8s-hana.ondemand.com"
  self.c21s = chart("https://storage.googleapis.com/c21s-helm/tested/20191210_0637-c21s.tgz",namespace="kubecf",domain=self.domain)

def apply(self,garden_k8s):
  self.__apply(garden_k8s)
  for shoot in garden_k8s.watch("shoot",self.shootname):
    lastOp = shoot.status.lastOperation
    if lastOp.type == "Create" and lastOp.progress == 100:
      break
  kubeconfig = garden_k8s.get("secret",self.shootname + ".kubeconfig").data.kubeconfig
  shoot_k8s = k8s(kubeconfig)
  self.c21s.apply(shoot_k8s)

def delete(self,garden_k8s):
  kubeconfig = garden_k8s.get("secret",self.shootname + ".kubeconfig").data.kubeconfig
  shoot_k8s = k8s(kubeconfig)
  self.c21s.delete(shoot_k8s)
  self.__delete(garden_k8s)
