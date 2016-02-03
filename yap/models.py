import os
from datetime import datetime, timedelta, time

from sqlalchemy import Column, Integer, String, DateTime, and_, create_engine
from sqlalchemy.orm import sessionmaker
from sqlalchemy.ext.hybrid import hybrid_property
from sqlalchemy.ext.declarative import declarative_base

import yap

_sql_echo = bool(os.environ.get('YAP_SQL_ECHO'))
engine = create_engine('sqlite:///%s' % yap.DB_PATH, echo=_sql_echo)
Session = sessionmaker(bind=engine)

Base = declarative_base()


class Todo(Base):
    __tablename__ = 'todo'

    id = Column(Integer, primary_key=True)
    title = Column(String, nullable=False)
    due_date = Column(DateTime)
    wait_date = Column(DateTime)
    created_at = Column(DateTime, nullable=False, default=datetime.utcnow)
    done_at = Column(DateTime)

    def __repr__(self):
        return "<%s id=%i>" % (self.__class__.__name__, self.id)

    @hybrid_property
    def done(self):
        return self.done_at != None

    @property
    def overdue(self):
        if self.due_date is None:
            return False
        return self.due_date < datetime.utcnow()

    @property
    def remaining(self):
        if self.due_date is None:
            return None
        return self.due_date - datetime.utcnow()

    @hybrid_property
    def waiting(self):
        if self.wait_date is None:
            return False
        return self.wait_date > datetime.utcnow()

    @waiting.expression
    def waiting(cls):
        return and_(cls.wait_date != None, cls.wait_date > datetime.utcnow())

    @property
    def str_due_date(self):
        if self.due_date:
            due_date = human_datetime(self.due_date)
            if self.overdue:
                return red(due_date)
            if self.remaining < timedelta(days=1):
                return yellow(due_date)
            return due_date

    @property
    def str_done_at(self):
        return human_datetime(self.done_at)

    @property
    def str_wait_date(self):
        return human_datetime(self.wait_date)


def human_datetime(d):
    if d.time() == time.min:
        fmt = yap.DATE_FORMAT
    else:
        fmt = yap.DATETIME_FORMAT
    return d.strftime(fmt)


def red(s):
    return "\033[91m {}\033[00m".format(s)


def yellow(s):
    return "\033[93m {}\033[00m".format(s)
