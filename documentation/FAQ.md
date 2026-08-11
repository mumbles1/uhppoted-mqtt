# FAQ

1. _My MQTT broker is on an secure network - how do I disable signing and encryption and HMAC and ..?_

    In _uhppoted.conf_:
    ```
    # MQTT
    ...
    mqtt.security.authentication = NONE
    mqtt.security.HMAC.required = false
    mqtt.security.nonce.required = false
    mqtt.security.outgoing.sign = false
    mqtt.security.outgoing.encrypt = false
    ...
    ```

2. _Why am I still getting old events even though I have set `mqtt.alerts.retained = false` ?_

    The MQTT broker probably still has retained events from when `mqtt.alerts.retained`
    was set to `true`. You need to manually clear the events - normally by sending an
    empty retained message to the topics:

    ```
    mqtt publish --topic 'uhppoted/gateway/events' --message '' -r
    mqtt publish --topic 'uhppoted/gateway/system' --message '' -r
    ```

3. _Why does uhppoted-mqtt report 'multiple ACL files in tar.gz' ?_

    Because .. Apple. 

    `bsdtar` on MacOS automatically searches for extended attributes (like com.apple.provenance or resource forks) and
    prepends a hidden header file into the archive stream to preserve them. Use the `--disable-copyfile` command line
    option, e.g.:

    ```
     tar --disable-copyfile --owner=QWERTY54 --group=QWERTY54 -czvf hogwarts.tar.gz hogwarts.acl signature
    ```

    Or alternatively set the `COPYFILE_DISABLE` environment variable:

    ```
    COPYFILE_DISABLE=1 tar --owner=QWERTY54 --group=QWERTY54 -czvf hogwarts.tar.gz hogwarts.acl signature
    ```
