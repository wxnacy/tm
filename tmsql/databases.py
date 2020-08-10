#!/usr/bin/env python
# -*- coding:utf-8 -*-
# Author: wxnacy(wxnacy@gmail.com)
# Description:

import timeit

from tmsql.loggers import logger

import records
import pymysql.cursors

from urllib.parse import urlparse

URL_CONFIG = urlparse('mysql+pymysql://root:wxnacy@127.0.0.1:3306/study?charset=utf8mb4')

db = records.Database('mysql+pymysql://root:wxnacy@127.0.0.1:3306/study?charset=utf8mb4')

def db_query(sql):
    t0 = timeit.default_timer()
    rows = db.query(sql)
    logger.debug(dir(rows))

    t1 = timeit.default_timer()
    time_used = f'{(t1-t0):0.2f}'
    print(rows.dataset)
    print(rows.as_dict())
    print(f'{len(rows.all())} rows in set ({time_used} sec)')
    print('')


class BaseDB(object):
    @classmethod
    def create_conn(cls):
        '''创建mysql链接'''
        return pymysql.connect(
            host=URL_CONFIG.hostname,
            port=URL_CONFIG.port,
            user=URL_CONFIG.username,
            password=URL_CONFIG.password,
            db=URL_CONFIG.path[1:],
            charset='utf8mb4',
            cursorclass=pymysql.cursors.DictCursor
        )

    @classmethod
    def query(cls, sql, params):
        """
        查询操作
        :param sql:
        :param params:
        :return:
        """
        conn = cls.create_conn()
        try:
            cursor = conn.cursor()

            res = cursor.execute(sql, params)
            print(res)
            conn.commit()
            print(dir(cursor))
            result = cursor.fetchall()
            cursor.close()
            return result
        except BaseException as e:
            logger.error(traceback.format_exc())
            return []
        finally:
            conn.close()

    @classmethod
    def execute(cls, sql, params):
        """
        更新操作
        :param sql:
        :param params:
        :return:
        """
        conn = cls.create_conn()
        try:
            cursor = conn.cursor()

            result = cursor.execute(sql, params)
            conn.commit()
            cursor.close()
            return result
        except BaseException as e:
            logger.error(traceback.format_exc())
            return False
        finally:
            conn.close()

    @classmethod
    def query_db(cls, sql, **kwargs):
        res = db.engine.execute(text(sql), **kwargs)
        keys = res.keys()
        Record = namedtuple('Record', res.keys())
        records = {Record(*r) for r in res.fetchall()}
        return records
        #  res = [r for r in records]

        #  def _fmt_i(k, v):
            #  return k ,v

        #  def _fmt(o):
            #  r = list(map(_fmt_i, keys, o))
            #  return {k: v for k, v in r}
        #  res = [_fmt(o) for o in res.fetchall()]
        #  return res

    @classmethod
    def execute_db(cls, sql, **kwargs):
        res = db.engine.execute(text(sql), **kwargs)
        db.session.commit()
        return res

if __name__ == "__main__":
    db = BaseDB()
    res = db.query("select * from user;", [])
    print(res)
