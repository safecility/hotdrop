### Publish Bigquery compatible messages from Hotdrop

To store data in Bigquery via pubsub we need 
* a bigQuery table with a schema
* a pubsub.Topic using the same schema
* a pubsub.Subscription set to handle bigQuery

(see: https://cloud.google.com/pubsub/docs/create-bigquery-subscription)

protobuf messages passed to the bigQuery topic will be stored automatically