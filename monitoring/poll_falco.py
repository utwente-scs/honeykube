import datetime
import json
import logging
import os
import re
import shlex
import subprocess

from utils import clean_dict

STORE_FILE = "store.txt"
COPY_SCRIPT = "copy_to_nodes.sh"
KUBECONFIG = "~/.kube/config_hk"
INTRUSION_RECORD_DURATION = 900


def poll_falco_logs():
    """Poll logs from falco pods using `kubetail`
    system call.

    Args:
        falco_logs_list (list): List of falco logs

    Returns:
        rc: return value from the process poll
    """
    command = "kubetail falco --namespace falco -t gke_research-gcp-credits_europe-west4_honeykube &"
    process = subprocess.Popen(
                    shlex.split(command),
                    shell=False,
                    stdout=subprocess.PIPE,
                    stderr=subprocess.PIPE
                )
    pattern = re.compile(r'(\[falco-.*\]) ({.*})')
    start_store = False
    start_time = None

    logging.info("Starting polling falco logs.")
    while True:
        output = process.stdout.readline()
        if start_store:
            time_now = datetime.datetime.now()
            if start_time + datetime.timedelta(seconds=INTRUSION_RECORD_DURATION) <= time_now:
                logging.info("Completed Recording! Stop storing trace files.")
                start_store = False
                start_time = None
                store_file_fd = open(STORE_FILE, "w")
                store_file_fd.write("False")
                store_file_fd.close()
                copy_command = "./{} {} {} /tmp/app".format(
                                                    COPY_SCRIPT,
                                                    KUBECONFIG,
                                                    STORE_FILE
                                                )
                logging.info("Copying store.txt with 'False' to all nodes")
                returned_value = subprocess.call(copy_command, shell=True)
                logging.debug(
                    "Value returned by subprocess: %d",
                    returned_value
                )

        if process.poll() is not None:
            logging.error("Unable to poll the logs.")
            break
        if output:
            line = output.strip().decode()
            matches = pattern.search(line)
            if matches is not None:
                log = json.loads(matches.group(2))
                log["output_fields"] = clean_dict(log["output_fields"])
                if log["priority"].upper() in ["WARNING", "ERROR", "CRITICAL", "ALERT", "EMERGENCY"]:
                    # check if store hasn't already started
                    if not start_store:
                        logging.info(
                            "Identified interaction in the cluster. Priority: %s",
                            log["priority"]
                        )
                        logging.info(json.dumps(log))
                        logging.info("Start storing trace files.")
                        start_store = True
                        start_time = datetime.datetime.now()
                        store_file_fd = open(STORE_FILE, "w")
                        store_file_fd.write("True")
                        store_file_fd.close()
                        copy_command = "./{} {} {} /tmp/app".format(
                                                    COPY_SCRIPT,
                                                    KUBECONFIG,
                                                    STORE_FILE
                                                )
                        logging.info(
                            "Copying store.txt with 'True' to all nodes with\
                                start_time {}".format(start_time)
                            )
                        returned_value = subprocess.call(
                            copy_command,
                            shell=True
                        )
                        logging.debug(
                            "Value returned by subprocess: %d",
                            returned_value
                        )

    rc = process.poll()
    return rc


if __name__ == "__main__":
    format = "%(asctime)s %(levelname)s %(name)s: %(message)s"
    logging.basicConfig(format=format, level=logging.INFO,
                        filename="/var/log/falco2/falco_poll_logs.log")
    poll_falco_logs()
