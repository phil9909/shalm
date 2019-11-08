
def init(self,database):
  self.name = "test"
  database.create_database(db="uaa",username="uaa",password="87612349234")
  return self

