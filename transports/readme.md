# Hotdrop Transports

Hotdrop runs over LoRA - however, the version we are using uses an encoded payload that must be accessed 
via the Vutility API, an unnecessary overhead.

The API allows a webhook to be configured to send uplink data - 
take the data from the webhook and place on a pubsub topic