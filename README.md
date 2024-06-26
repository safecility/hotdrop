# hotdrop
a safecility microservice device repo for the vutility hotdrop

microservices configured are

webhook -> process -> pipelines: 
    *  bigquery
    *  messagestore
    *  usage

the usage pipeline outputs to a microservice usage topic that can be accessed by generalized microservices based on the
MeterReading type
