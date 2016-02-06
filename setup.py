# coding=utf-8
import os
import re
from setuptools import setup


def read(*fname):
    with open(os.path.join(os.path.dirname(__file__), *fname)) as f:
        return f.read()


def get_version():
    for line in read('yap', '__init__.py').splitlines():
        m = re.match(r"""__version__\s*=\s*['"](.*)['"]""", line)
        if m:
            return m.groups()[0].strip()
    raise Exception('Cannot find version')


setup(
    name='yap',
    version=get_version(),
    author=u'Cenk Altı',
    author_email='cenkalti@gmail.com',
    keywords='todo task',
    url='https://github.com/cenkalti/yap',
    packages=['yap'],
    install_requires=[
        'SQLAlchemy>=1.0,<2',
        'tabulate>=0.7,<2',
        'isodate>=0.5,<2',
    ],
    description='Command line todo app',
    zip_safe=True,
    entry_points={
        'console_scripts': [
            'yap = yap.__main__:main',
        ],
    },
    classifiers=[
        'Development Status :: 3 - Alpha',
        'Environment :: Console',
        'Intended Audience :: End Users/Desktop',
        'License :: OSI Approved :: MIT License',
        'Natural Language :: English',
        'Operating System :: OS Independent',
        'Programming Language :: Python :: 2.7',
        'Topic :: Utilities',
    ],
)
