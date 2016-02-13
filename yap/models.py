from datetime import datetime, timedelta, time

import isodate
from sqlalchemy import Column, Integer, String, DateTime, Boolean, and_, func
from sqlalchemy.orm import sessionmaker
from sqlalchemy.types import TypeDecorator
from sqlalchemy.ext.hybrid import hybrid_property
from sqlalchemy.ext.declarative import declarative_base

Session = sessionmaker()


class Base(object):
    def to_dict(self):
        d = {}
        keys = self.__table__.columns.keys()
        for columnName in keys:
            d[columnName] = getattr(self, columnName)
        return d

    @classmethod
    def from_dict(cls, d):
        obj = cls()
        for name in d.keys():
            setattr(obj, name, d[name])
        return obj
Base = declarative_base(cls=Base)


class Duration(TypeDecorator):

    impl = String

    def process_bind_param(self, value, dialect):
        if value is None:
            return None
        return isodate.duration_isoformat(value)

    def process_result_value(self, value, dialect):
        if value is None:
            return None
        return isodate.parse_duration(value)


class Task(Base):
    __tablename__ = 'task'

    id = Column(Integer, primary_key=True)
    title = Column(String, nullable=False)
    due_date = Column(DateTime)
    wait_date = Column(DateTime)
    created_at = Column(DateTime, nullable=False, default=datetime.now)
    done_at = Column(DateTime)
    context = Column(String)
    recur = Column(Duration)
    shift = Column(Boolean)
    order = Column(Integer)

    def __repr__(self):
        return "<%s id=%i>" % (self.__class__.__name__, self.id)

    @hybrid_property
    def done(self):
        return self.done_at != None

    @hybrid_property
    def archived(self):
        return self.id < 0 and self.done_at == None

    @archived.expression
    def archived(self):
        return and_(self.id < 0, self.done_at == None)

    @hybrid_property
    def recurring(self):
        return self.recur != None

    @property
    def overdue(self):
        if self.due_date is None:
            return False
        return self.due_date < datetime.now()

    @property
    def remaining(self):
        if self.due_date is None:
            return None
        return self.due_date - datetime.now()

    @hybrid_property
    def waiting(self):
        if self.wait_date is None:
            return False
        return self.wait_date > datetime.now()

    @waiting.expression
    def waiting(cls):
        return and_(cls.wait_date != None, cls.wait_date > datetime.now())

    @property
    def str_due_date(self):
        if self.due_date:
            due_date = str_datetime(self.due_date)
            if self.overdue:
                return red(due_date)
            yellow_after = self.due_date - timedelta(days=1)
            if self.wait_date and self.wait_date > yellow_after:
                yellow_after = self.wait_date
            if datetime.now() > yellow_after:
                return yellow(due_date)
            return due_date

    @property
    def str_done_at(self):
        if self.done_at:
            return str_datetime(self.done_at)

    @property
    def str_wait_date(self):
        if self.wait_date:
            return str_datetime(self.wait_date)

    @classmethod
    def find_next_order(cls, session):
        return (session.query(func.max(cls.order)).scalar() or 0) + 1


def str_datetime(dt):
    if dt.time() == time.max:
        fmt = isodate.DATE_EXT_COMPLETE
    else:
        fmt = isodate.DATE_EXT_COMPLETE + 'T' + isodate.TIME_EXT_COMPLETE
    return isodate.strftime(dt, fmt)


def red(s):
    return "\033[91m{}\033[00m".format(s)


def yellow(s):
    return "\033[93m{}\033[00m".format(s)
