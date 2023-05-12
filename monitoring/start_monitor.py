import datetime
import json
import logging
import os
import re
import shlex
import subprocess
import threading
import time

import gridfs
import pymongo

from circular_list import CircularList
from config import *
from utils import clean_dict

# Tracefiles
tracefiles_list = CircularList(NUM_OF_FILES)
# MongoDB Client
client = pymongo.MongoClient(
    MONGODB_URL,
    serverSelectionTimeoutMS=2000
)
database = client[DATABASE_NAME]
# Define DB Collection connections
falco_logs = database.falco_logs
sysdig_trace_gridfs = gridfs.GridFS(database)
# Logs List
falco_logs_list = []


def start_file_store(time, tracefiles_list):
    tracefiles_list.start_storage(time)
    start_time_file = tracefiles_list.get_active()

    if time - datetime.timedelta(seconds=INTRUSION_RECORD_DURATION) < start_time_file.timestamp:
        last_file_info = tracefiles_list.get_last_active()
        if last_file_info.filename is not None:
            tracefiles_list.push_to_database(
                sysdig_trace_gridfs,
                last_file_info
            )


def push_logs_list_to_database(falco_logs_list):
    """ Push the list to the database

    Args:
        falco_logs_list (list): list of logs
    """
    if falco_logs_list:
        for log in falco_logs_list:
            try:
                falco_logs.insert_one(log)
            except pymongo.errors.DuplicateKeyError:
                continue
        falco_logs_list = []
    return


def push_logs(falco_logs_list):
    """Start the push logs thread and call the push

    Args:
        falco_logs_list (list): list of logs to push to mongodb
    """
    while True:
        push_logs_list_to_database(falco_logs_list)
        time.sleep(5)

    return


def poll_falco_logs(falco_logs_list, tracefiles_list):
    """Poll logs from falco pods using `kubetail`
    system call.

    Args:
        falco_logs_list (list): List of falco logs

    Returns:
        rc: return value from the process poll
    """
    command = "sudo -u padfoot kubetail falco --namespace falco &"
    logging.info("Thread falco_thread : starting command %s", command)
    process = subprocess.Popen(
                    shlex.split(command),
                    shell=False,
                    stdout=subprocess.PIPE,
                    stderr=subprocess.PIPE
                )
    pattern = re.compile(r'(\[falco-.*\]) ({.*})')

    logging.info("Thread falco_thread : starting polling falco logs.")
    while True:
        output = process.stdout.readline()
        if process.poll() is not None:
            logging.error("Thread falco_thread : Unable to poll the logs.")
            break
        if output:
            line = output.strip().decode()
            matches = pattern.search(line)
            if matches is not None:
                log = json.loads(matches.group(2))
                log["output_fields"] = clean_dict(log["output_fields"])
                falco_logs_list.append(log)
                if log["priority"].upper() in ["WARNING", "ERROR"]:
                    logging.info("Thread falco_thread : \
                        Identified interaction in the cluster")
                    if not tracefiles_list.start_store:
                        print(log)
                        start_file_store(
                            datetime.datetime.now(),
                            tracefiles_list
                        )
    rc = process.poll()
    return rc


def start_sysdig(tracefiles_list):
    """ Start the sysdig circular monitoring system
    """
    logging.info("Thread sysdig_thread : starting sysdig loop...")
    while True:
        # Set file store to False
        start_store = False
        # Get Timestamp
        date = datetime.datetime.now()
        date_str = date.strftime("%Y-%m-%d_%H-%M-%S")
        # File name
        filepath = os.path.join(TRACE_DIR, BASE_TRACE_FILENAME)
        filename = "{}_{}.scap.gz".format(filepath, date_str)
        command = "sudo sysdig -pk -w {} -z -M {}".format(
                                                filename,
                                                SYSDIG_RECORD_DURATION
                                            )
        # Check if the store flag is set in the last file.
        last_file_info = tracefiles_list.get_active()
        if last_file_info.store:
            tracefiles_list.push_to_database(
                sysdig_trace_gridfs,
                last_file_info
            )
        # Check if storage is started
        if tracefiles_list.start_store:
            # Check if the record duration, required next file to be stored.
            print(tracefiles_list.store_time + datetime.timedelta(seconds=INTRUSION_RECORD_DURATION))
            if tracefiles_list.store_time + datetime.timedelta(seconds=INTRUSION_RECORD_DURATION) > date:
                start_store = True
                print("Record Continued!")
            else:
                print("Completed record!")
                tracefiles_list.stop_storage()
                start_store = False
        print("Start store", start_store)
        old_file = tracefiles_list.add(filename, date, start_store)

        if old_file.filename is not None:
            if os.path.exists(old_file.filename):
                logging.info("Deleting file: {}".format(old_file.filename))
                os.remove(old_file.filename)
            else:
                logging.error("File not found: {}".format(old_file))

        # returns the exit code in unix
        logging.info("Executing Command: %s", command)
        returned_value = subprocess.call(command, shell=True)
        logging.debug("Value returned by subprocess: %d", returned_value)


if __name__ == "__main__":
    format = "%(asctime)s: %(message)s"
    logging.basicConfig(format=format, level=logging.INFO,
                        datefmt="%H:%M:%S")

    logging.info("Main : starting sysdig_thread thread")
    sysdig_thread = threading.Thread(
        target=start_sysdig,
        args=(tracefiles_list,),
        name="sysdig_thread"
    )
    sysdig_thread.start()

    logging.info("Main : starting falco_thread thread")
    falco_thread = threading.Thread(
        target=poll_falco_logs,
        args=(falco_logs_list, tracefiles_list,),
        name="falco_thread",
        daemon=True
    )
    falco_thread.start()

    logging.info("Main : starting push_logs_thread thread")
    push_logs = threading.Thread(
        target=push_logs,
        args=(falco_logs_list,),
        name="push_logs_thread"
    )
    push_logs.start()

    logging.info("Main : Joining the threads")
    sysdig_thread.join()
    falco_thread.join()
    push_logs.join()
