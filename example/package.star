
def init():
  return struct(
    uaa = chart("uaa"),
    database = chart("database"),
    HA = False,
    memory = struct(
      requests = False,
      limits = False
    ),
    cpu = struct(
      requests= False,
      limits= False
    ),
    use_istio= False
  )