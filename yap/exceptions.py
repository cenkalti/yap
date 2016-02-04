class YapError(Exception):
    pass


class TaskNotFoundError(YapError):
    def __init__(self, tid):
        self.id = tid
        super(TaskNotFoundError, self).__init__("task id not found: %s" % tid)


class TaskImportError(YapError):
    def __init__(self):
        super(TaskImportError, self).__init__("import completed with errors")
