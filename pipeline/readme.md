## Pipelines for Hotdrop 

Once Vutility Hotdrop messages are brokered we provide a number of pipelines for the output

* messagestore: add the messages to google datastore
* bigquery: place the messages in a bigquery topic to view messages in bigquery
* usage: create data usage messages for further processing

### Related pipelines

Usage messages are intended for usage pipelines which store data and create time windowed data of usage for analysis

