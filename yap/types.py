import operator
from datetime import datetime, date, timedelta, time
from functools import wraps

import isodate


delete_option = object()


def _delete_with_empty_string(f):
    @wraps(f)
    def inner(s):
        if s == '':
            return delete_option
        return f(s)
    return inner


def date_time_or_datetime(s, default_day, default_time):
    d = _parse_day_name(s)
    if d:
        return datetime.combine(d, default_time)
    if s == 'now':
        return datetime.now()
    if s == 'today':
        return datetime.combine(date.today(), default_time)
    if s == 'tomorrow':
        return datetime.combine(date.today() + timedelta(days=1), default_time)
    try:
        if s.startswith('-'):
            op = operator.sub
            s = s[1:]
        else:
            op = operator.add
        if s.startswith('P'):  # ISO 8601 duration
            td = isodate.parse_duration(s)
            try:
                td = td.tdelta
            except AttributeError:
                pass
            if td.seconds == 0:  # resolution is day
                now = datetime.combine(datetime.today(), default_time)
            else:
                now = datetime.now()
            return op(now, isodate.parse_duration(s))
        if 'T' in s:  # ISO 8601 combined date and time
            return isodate.parse_datetime(s)
        if '-' in s:
            return datetime.combine(isodate.parse_date(s), default_time)
        if ':' in s:
            return datetime.combine(default_day, isodate.parse_time(s))
        raise ValueError
    except ValueError:
        raise ValueError("%r is not an ISO 8601 date, time or datetime" % s)


def _parse_day_name(s):
    days = ['monday', 'tuesday', 'wednesday', 'thursday',
            'friday', 'saturday', 'sunday']
    try:
        i = days.index(s.lower())
    except ValueError:
        pass
    else:
        delta_days = (i - date.today().weekday()) % 7
        return date.today() + timedelta(days=delta_days)


@_delete_with_empty_string
def due_date(s):
    return date_time_or_datetime(s, date.today(), time.max)


@_delete_with_empty_string
def wait_date(s):
    return date_time_or_datetime(s, date.today(), time.min)


@_delete_with_empty_string
def on_date(s):
    return date_time_or_datetime(s, date.today(), time.min).date()


@_delete_with_empty_string
def duration(s):
    return isodate.parse_duration(s)
