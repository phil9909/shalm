def init(self):
   self.database = chart("../database")
   self.messagebroker = chart("../messagebroker")
   self.logging = chart("../logging")
   self.database = chart("../database")
   self.blobstore = chart("../blobstore")
   self.ingress = chart("../ingress")
   self.uaa = chart("../uaa",database = self.database, logging= self.logging, ingress=self.ingress)
   self.capi = chart("../capi", database = self.database, logging= self.logging, uaa=self.uaa, blobstore=self.blobstore, ingress=self.ingress)

def create_or_update_database(self,db="db",username="",password=""):
  pass
