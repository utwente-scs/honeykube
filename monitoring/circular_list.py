import datetime
import os


class TraceFile:

    filename = None
    is_active = False
    store = False

    def __init__(self, filename, timestamp, store=False, is_active=True):
        self.filename = filename
        self.timestamp = timestamp
        self.store = store
        self.is_active = is_active


class CircularList:

    clist = list()
    size = 0
    last_index = 0

    # For permanent storage of trace files
    start_store = False
    store_time = None

    def __init__(self, size):
        """ Initialise the circular list

        Args:
            size (int): Max size of the list
        """
        # Set the size of the list
        self.size = size

        # Initialise the list of the given size
        self.clist = [TraceFile(None, None)] * self.size

    def set_isactive(self, is_active):
        """ Set is_active flag for the file at the
            last index

        Args:
            is_active (bool): [description]
        """
        self.clist[self.last_index].is_active = is_active

    def get_active(self):
        """ Return the active file info

        Returns:
            [type]: [description]
        """
        return self.clist[self.last_index]

    def get_last_active(self):
        """
        Return the the
        """
        index = (self.last_index - 1) % self.size
        return self.clist[index]

    def start_storage(self, time):
        """ Start file storage

        Args:
            time: time of the intrusion
        """
        self.start_store = True
        self.store_time = time
        self.clist[self.last_index].store = True

    def stop_storage(self):
        """ Stop file storage
        """
        self.start_store = False
        self.store_time = None

    def push_to_database(self, db_handle, tracefile):
        """ Push tracefiles to the database
        """
        return db_handle.put(
            open(tracefile.filename, "rb"),
            filename=os.path.split(tracefile.filename)[-1],
            time=tracefile.timestamp
        )

    def add(self, new_file, timestamp, start_store):
        """ Add new item to the list

        Args:
            new_item

        Returns:
            old_item: The value replaced by the new_item
        """
        # Set is_active to false
        self.set_isactive(False)

        # Increment Index
        self.last_index += 1
        self.last_index = self.last_index % self.size

        # Get the old value
        old_item = self.clist[self.last_index]

        # Get the index to insert the value
        self.clist[self.last_index] = TraceFile(
            new_file,
            timestamp,
            start_store
        )

        return old_item
