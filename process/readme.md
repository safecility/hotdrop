## Process raw hotdrop messages

We access a stream of messages from the Vutility webhook broker
process the payloads to HotdropReadings and pass them to a Hotdrop Topic 

### Metadata

In addition to standard device meta-data we also add PowerFactor and Voltage as the Vutility Hotdrop does not measure 
these and they need to estimated at source.
