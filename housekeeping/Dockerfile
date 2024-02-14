FROM minio/mc:latest

RUN mkdir /root/.mc
COPY ./find-and-clean.sh /find-and-clean.sh

ENV ceph_bucket=""
ENV days=""
ENV dir=""

ENTRYPOINT ["sh", "-c", "./find-and-clean.sh"]
