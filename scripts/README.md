# Scripts

Shell scripts to make working with Kafka a bit easier in the local environment.

## Scripts

Brief overview of each script.

| Script            | Description                                                            |
| ----------------- | ---------------------------------------------------------------------- |
| start-kafka.sh    | Runs docker-compose to start Zookeeper and Kafka                       |
| stop-kafka.sh     | Runs docker-compose to stop Zookeeper and Kafka                        |
| create-topic.sh   | Creates a topic with a default of 7 days retention (user configurable) |
| describe-topic.sh | Describes the topic with `kafka-topics --describe`                     |
| dump-last.sh      | not really working correctly do not use                                |
| dump-messages.sh  | Dumps the last 10 messages to the console                              |
| publish.sh        | publishes message from CLI or file to topic                            |
| setup-env.sh      | Creates my local test environment                                      |

## Environment

Scripts expect a .env to define the environment.

Right now there are only two variables required.

| Variable     | Description                                                                    |
| ------------ | ------------------------------------------------------------------------------ |
| COMPOSE_FILE | The docker compose filename to use. Expected to be in projects' root directory |
| PORT         | The port that the kafka container is listeningin on                            |
