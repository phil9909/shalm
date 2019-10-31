
def init(self):
  self.name = "test"
  return self


def attach_database(self,database):
  database.create_database(db="uaa",username="uaa",password="87612349234")