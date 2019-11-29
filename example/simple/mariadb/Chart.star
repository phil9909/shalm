
def init(self):
    self.databases = {}


def create_database(self,db="db",username="",password=""):
    self.databases[db] = """
    CREATE OR REPLACE USER '{username}' IDENTIFIED BY '{password}';
    CREATE DATABASE IF NOT EXISTS `{db}`;
    GRANT ALL PRIVILEGES ON `{db}`.* TO '{username}'@'%' WITH GRANT OPTION;
    FLUSH PRIVILEGES;
    """.format(username=username, password=password,db=db)
