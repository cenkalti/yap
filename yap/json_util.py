import datetime

DATE_FMT = '%Y-%m-%d'
ISO8601_FMT = '%Y-%m-%dT%H:%M:%SZ'


def datetime_encoder(obj):
    if isinstance(obj, datetime.datetime):
        return obj.strftime(ISO8601_FMT)
    elif isinstance(obj, datetime.date):
        return obj.strftime(DATE_FMT)

    raise TypeError


def datetime_decoder(d):
    for key, value in d.iteritems():
        try:
            datetime_obj = datetime.datetime.strptime(value, ISO8601_FMT)
            d[key] = datetime_obj
        except (ValueError, TypeError):
            try:
                date_obj = datetime.datetime.strptime(value, DATE_FMT)
                d[key] = date_obj.date()
            except (ValueError, TypeError):
                continue

    return d
