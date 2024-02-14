#!/bin/bash
# Requires MinIO client

while [[ ! -s "/tmp/config.json" ]]; do sleep 2; done  # Not needed if ceph-getkeys used as initContainer in Kubernetes
mv /tmp/config.json /root/.mc/config.json
mc find s3/${ceph_bucket}/${dir} --name "*backup*" --older-than ${days} --exec "mc rm {}" >> /logging/log
