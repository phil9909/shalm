
def init(self,database=None):
  self.name = "test"
  if database:
    database.create_database(db="uaa",username="uaa",password="87612349234")
  return self

