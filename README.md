# ceph-hskp-example
Example usage of [ceph-getkeys](https://github.com/laimis9133/ceph-getkeys) to setup S3 bucket housekeeping.  
Two containers. Initial one retreives Ceph bucket keys from RadosGW admin API and stores them in a shared volume.  
Second container picks up the keys and cleans up the bucket using MinIO client.  
Retention and bucket names and subdirectories are set through _docker-compose.yaml_ which can be translated to a Kubernetes deployment.
  
Usage: docker-compose up --build -d  
Monitor logging/logs for output.

