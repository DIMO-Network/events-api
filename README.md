# DIMO Events API

[Technical documentation](https://docs.dimo.zone/docs/overview/intro)

This little service reads system-wide events from the Kafka `topic.event`. Each partition reader commits a batch of events to the database every 10 seconds.

Based on the CloudEvent `type`, events are classified by a type and sub-type. As an example, `com.dimo.zone.user.create` gets turned into type `User` and sub-type `Create`. Events have an attached `userId`, and many have a `userDeviceId`.

A simple API that exposes a user's events.
