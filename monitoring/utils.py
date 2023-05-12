#! /bin/env python
import os


def clean_dict(log_dict):
    """ Cleaning the dictionary and removing "." from the
    key values, so it can be pushed

    Args:
        log_dict (dict): dictionary that needs to be cleaned.

    Returns:
        log_dict: cleaned dictionary
    """
    new_dict = {}
    for key in log_dict.keys():
        new_key = key.replace(".", "_")
        new_dict[new_key] = log_dict[key]
    log_dict = new_dict.copy()
    new_dict.clear()
    return log_dict
