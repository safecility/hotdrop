## usage

usage pipelines create a simple metering message with kWh used at a given time

usage is one of those microservices that is really micro, simply producing a kWh, time reading from
the incoming message

### general pipeline

usage is sent to general topics that handle time windows, billing etc...