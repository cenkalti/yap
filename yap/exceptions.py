class YapError(Exception):
    pass


class TodoNotFoundError(YapError):
    def __init__(self, tid):
        self.id = tid
        super(TodoNotFoundError, self).__init__("todo id not found: %s" % tid)


class TodoImportError(YapError):
    def __init__(self):
        super(TodoImportError, self).__init__("import completed with errors")
