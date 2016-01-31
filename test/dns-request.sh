#!/bin/bash

while [ 1 ];
do
  dig kafka.service.consul +short
  dig cassandra.service.consul +short;
  dig goshe.service.consul +short;
  dig vagrant.service.consul +short;
  dig datadog.service.consul +short;
done
