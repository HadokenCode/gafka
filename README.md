# gafka
kafka clusters management tool in golang

### Terms

- zone

  Zone is groups of kafka cluster.

  - zkAddrs

    Zookeeper connection string.

- cluster

  Kafka cluster.

  - name

    Human readable kafka cluster name

  - path

    The chroot path in zk

### TODO

- [ ] topic name regex
- [ ] show each partition replication lag
- [ ] housekeep of clusters info automatically find stale info
- [ ] topology of each host
- [ ] topics show consumer groups
