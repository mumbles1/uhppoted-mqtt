### `set-firstcard`

Sets the 'first card' configuration for a controller door. Note that the 'first card' configuration is not activated
until `refresh-tasks` is invoked.

```
Request:

topic: <root>/<requests>/device/door/firstcard:set

message:
{
    "message": {
        "request": {
            "request-id": "<request-id>",
            "client-id": "<client-id>",
            "reply-to": "<topic>",
            "device-id": "<controller-id>",
            "door": "<door-id>",
            "first-card": {
               "start-time": "<start-time>",
               "end-time": "<end-time>",
               "active-mode": "<mode>",
               "inactive-mode": "<mode>",
               "weekdays": "[<weekday>]"
            }
        }
    }
}

request-id     (optional) message ID, returned in the response
client-id      (required) client ID for authentication and authorisation (if enabled)
reply-to       (optional) topic for reply message. Defaults to uhppoted/gateway/replies (or the
                          configured reply topic) if not provided.
device-id      (required) controller serial number
door           (required) door (1..4) to open
start-time     (required) time (HH:mm) from which first-card mode is activated
end-time       (required) time (HH:mm) after which first-card mode is no longer activated
active-mode    (required) door control mode after first-card swipe (controlled, normally open or normally closed)
inactive-mode  (required) door control mode after first-card is deactivated (controlled, normally open, normally closed or firstcard)
weekdays       (required) list of days on which first-card mode can be activated.
```

```
Response:
{
  "message": {
    "reply": {
      "request-id": <request-id>,
      "client-id": <client-id>,
      "method": "set-firstcard",
      "response": {
            "device-id": <controller-id>,
            "door": <door-id>,
            "ok": <true/false>,
      },
      ...
    }
  },
  ...
}

request-id   message ID from the request
client-id    client ID from the request
device-id    controller serial number
door         door (1..4) from the request
ok           true if the first card configuration was accepted
```


Example:
```
topic: uhppoted/gateway/requests/device/door/firstcard:set

{
  "message": {
    "request": {
      "client-id": "QWERTY",
      "request-id": "AH173635G3",
      "reply-to": "uhppoted/reply/97531",
      "device-id": 405419896,
      "door": 3,
      "first-card": {
        "start-time": "08:30",
        "end-time": "16:45",
        "active-mode": "normally open",
        "inactive-mode": "firstcard",
        "weekdays": "[Mon, Tue, Thurs, Fri]"
      }
    }
  }
}

{
  "message": {
    "reply": {
      "server-id": "uhppoted"
      "client-id": "QWERTY",
      "request-id": "AH173635G3",
      "method": "set-firstcard",
      "response": {
        "device-id": 405419896,
        "door": 3,
        "ok": true
      }
    }
  }
}
```
