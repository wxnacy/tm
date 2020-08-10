#!/usr/bin/env python
# -*- coding:utf-8 -*-
# Author: wxnacy(wxnacy@gmail.com)
# Description: 日志


import logging
import logging.handlers


def create_logger():
    """创建日志"""
    logger = logging.getLogger('tmsql')
    logger.setLevel(logging.DEBUG)

    file_handler = logging.handlers.RotatingFileHandler(
        '/tmp/tmsql.log', maxBytes=104857600, backupCount=20
    )
    #  error_file_handler.setLevel(logging.ERROR)
    logger.addHandler(file_handler)

    return logger

logger = create_logger()
