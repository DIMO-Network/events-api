import confluent_kafka
import json
import random
import datetime
import time
import string

p = confluent_kafka.Producer({"bootstrap.servers": "localhost:9092"})

types = [
	"com.dimo.zone.user.create",
	"com.dimo.zone.device.create",
	"com.dimo.zone.device.delete",
	"com.dimo.zone.device.integration.create",
	"com.dimo.zone.device.integration.delete",
	"com.dimo.zone.device.odometer.update",
	"com.dimo.zone.user.token.issue",
	"com.dimo.zone.user.referral.complete",
]

def make_id():
    return ''.join(random.choice(string.ascii_lowercase + string.ascii_uppercase + string.digits) for _ in range(27))

while True:
    msg = {
        "specversion": "1.0",
        "type": random.choice(types),
        "source": "devices-api-dev-7975b4fbbc-6tl9r",
        "subject": "37f7e69d-9530-4f96-a05e-539a726d6f7e",
        "id": make_id(),
        "time": datetime.datetime.now().isoformat() + "Z",
        "data": {
            "userId": "CioweDEzYkVFNjRmRDY4MDA3NjFiRTA0RmE2YTZlM0U3OGMzNjA1RGJjQjUSBHdlYjM",
        }
    }
    print(msg)
    p.produce("table.event", json.dumps(msg))
    time.sleep(random.randint(1, 10))
