```conf
ectl set /bcshealth/zone-01/TCP::127.0.0.12:8001/zone-01 '{
    "action": "",
    "zone": "zone-01",
    "protocol": "TCP",
    "Url": "127.0.0.12:8001",
    "status": {
        "slaveInfo": {
            "slaveClusterName": "zone-01",
            "zones": [
                "zone-01"
            ],
            "ip": "127.0.0.12",
            "port": 8001,
            "hostname": "mesos-master-2",
            "scheme": "http",
            "version": "Version :1.4.2",
            "cluster": "master",
            "pid": 24213
        },
        "success": false,
        "message": "this is new health test message",
        "finishedAt": 1521812251
    }
}'

ectl rm /bcshealth/zone-01/TCP::127.0.0.12:8001/zone-01

curl -H "Content-Type:application/json" -X POST -d '{
    "action": "",
    "zone": "zone-01",
    "protocol": "TCP",
    "Url": "127.0.0.12:8001",
    "status": {
        "slaveInfo": {
            "slaveClusterName": "zone-01",
            "zones": [
                "zone-01"
            ],
            "ip": "127.0.0.12",
            "port": 8001,
            "hostname": "mesos-master-2",
            "scheme": "http",
            "version": "Version :1.4.2",
            "cluster": "master",
            "pid": 24213
        },
        "success": true,
        "message": "this is new health test message",
        "finishedAt": 1521818173
    }
}' http://127.0.0.12:8001/bcshealth/v1/reportjobs

```