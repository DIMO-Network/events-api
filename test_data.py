import confluent_kafka
import json
import random
import datetime
import time
import uuid

p = confluent_kafka.Producer({"bootstrap.servers": "localhost:9092"})

while True:
    msg = {
        "specversion": "1.0",
        "type": "zone.dimo.device.added",
        "source": "devices-api-dev-7975b4fbbc-6tl9r",
        "subject": "37f7e69d-9530-4f96-a05e-539a726d6f7e",
        "id": str(uuid.uuid4()),
        "time": datetime.datetime.now().isoformat() + "Z",
        "data": {}
    }
    print(msg)
    p.produce("events", json.dumps(msg))
    time.sleep(random.randint(1, 10))
