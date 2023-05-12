import datetime
import logging
import os
import subprocess

TRACE_DIR = "/tmp/tracefiles"

BASE_TRACE_FILENAME = "sys-trace"

SYSDIG_RECORD_DURATION = 1000

ARCHIVE_DIR = "/tmp/tracefiles/archive"

STORE_FILE = "/tmp/app/store.txt"

# Number of files for the sysdig circular files system
NUM_OF_FILES = 15

tracefiles_list = []

format = "%(asctime)s %(levelname)s %(name)s: %(message)s"
logging.basicConfig(format=format, level=logging.INFO,
                    filename="/tmp/logs/sysdig_monitor_logs.log")

while True:
    store_file_fd = open(STORE_FILE, "r")
    store = store_file_fd.read()
    store_file_fd.close()
    logging.info("Read file Store.txt. Data found: %s", store)
    # Get Timestamp
    date = datetime.datetime.now()
    date_str = date.strftime("%Y-%m-%d_%H-%M-%S")
    # File name
    filepath = os.path.join(TRACE_DIR, BASE_TRACE_FILENAME)
    filename = "{}_{}.scap.gz".format(filepath, date_str)
    command = "sysdig -pk -w {} -z -M {}".format(
                                            filename,
                                            SYSDIG_RECORD_DURATION
                                        )
    docker_command = "sudo docker exec sysdig /bin/bash -c \'{}\'".format(
                                                                    command
                                                                    )

    if len(tracefiles_list) == NUM_OF_FILES:
        old_file = tracefiles_list.pop(0)
        if old_file is not None:
            if os.path.exists(old_file):
                logging.info("Deleting file: {}".format(old_file))
                os.remove(old_file)
            else:
                logging.error("File not found: {}".format(old_file))

    tracefiles_list.append(filename)

    # returns the exit code in unix
    logging.info("Executing Command: %s", command)
    returned_value = subprocess.call(docker_command, shell=True)
    logging.debug("Value returned by subprocess: %d", returned_value)

    store_file_fd = open(STORE_FILE, "r")
    store_new = store_file_fd.read()
    store_file_fd.close()
    logging.info("Read file Store.txt. Data found: %s", store_new)

    if store.strip() == "True" or store_new == "True":
        src_path = os.path.join("/tmp/tracefiles/", filename)
        logging.info("Store value found to be 'True'. Copying file %s, to %s",
                     src_path,
                     ARCHIVE_DIR)
        copy_command = "sudo cp {} {}".format(src_path, ARCHIVE_DIR)
        returned_value = subprocess.call(copy_command, shell=True)
        logging.debug("File moved. Value returned by subprocess: %d",
                      returned_value)
