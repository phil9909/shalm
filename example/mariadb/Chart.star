
def init(self):
    self.databases = {}


def create_database(self,db="db",username="",password=""):
    self.databases[db] = """
    CREATE USER '{username}' IDENTIFIED BY '{password}';
    CREATE DATABASE `{db}`;
    GRANT ALL PRIVILEGES ON `{db}`.* TO '{username}'@'%' WITH GRANT OPTION;
    FLUSH PRIVILEGES;
    """.format(username=username, password=password,db=db)
