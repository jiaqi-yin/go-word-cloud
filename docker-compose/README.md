### Producer cli

docker run -it --rm --network word-cloud_app-tier confluentinc/cp-kafka:7.0.1 /bin/kafka-console-producer --bootstrap-server kafka-broker-1:9092 --topic test_topic

### Consumer cli

docker run -it --rm --network word-cloud_app-tier confluentinc/cp-kafka:7.0.1 /bin/kafka-console-consumer --bootstrap-server kafka-broker-1:9092 --topic test_topic
