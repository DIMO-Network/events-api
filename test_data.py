import confluent_kafka
import json
import random
import datetime
import time

p = confluent_kafka.Producer({"bootstrap.servers": "localhost:9092"})

while True:
    msg = {
        "specversion": "1.0",
        "time": datetime.datetime.now().isoformat() + "Z",
        "data": {}
    }
    p.produce("events", json.dumps(msg))
    time.sleep(random.randint(1, 10))
