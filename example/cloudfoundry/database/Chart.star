def init(self):
   self.postgres = chart("https://charts.bitnami.com/bitnami/postgresql-ha-1.1.0.tgz")

def create_or_update_database(self,db="db",username="",password=""):
  print("Create or update database " + db)
