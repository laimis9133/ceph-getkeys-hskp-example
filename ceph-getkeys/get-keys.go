// Script queries radosgw as admin and retrieves access and secret keys for a specified Ceph S3 bucket
// Requires a radosgw user with read permissions. Create one in Ceph node having radosgw-admin available:
// radosgw-admin user create --uid your-name --display-name "your-name" --caps "buckets=*;users=*;usage=read;metadata=read;zone=read --access-key=your-access-key --secret-key=your-secret-key

package main

import (
  "os"
  "fmt"
  "sync"
  "strings"
  "context"
  "io/ioutil"
  "text/template"
  "encoding/json"
  "github.com/ceph/go-ceph/rgw/admin"
)

func main() {
  co, err := connectCeph()
  ch := make(chan string)
  wg := sync.WaitGroup{}

  args := os.Args[1:]
  envArgs := os.Getenv("ceph_bucket")

  var givenBucket string
  var buckets []string

  if len(args) > 0 {
    givenBucket = args[0]
    buckets = []string{givenBucket}
  } else if envArgs != "" {
    givenBucket = envArgs
    buckets = []string{givenBucket}
  } else {
    buckets, err = co.ListBuckets(context.Background())
    if err != nil { panic(fmt.Sprintf("Unable to list S3 buckets.", err))}
  }

  for _, bucket := range buckets {
    wg.Add(1)
    access_key, secret_key := getBucketKeys(bucket, co)

    go func(bucket string, access_key string, secret_key string) {
      ch <- fmt.Sprintf("%s %s %s", bucket, access_key, secret_key)
      defer wg.Done()
    }(bucket, access_key, secret_key)
  }

  go func() {
    wg.Wait()
    close(ch)
  }()

  templateFileAWS, err := os.Open("template-aws.txt")
  if err != nil { panic(fmt.Sprintf("Cannot open AWS credentials template file.", err)) }
  defer templateFileAWS.Close()

  templateFileMinIO, err := os.Open("template-minio.txt")
  if err != nil { panic(fmt.Sprintf("Cannot open MinIO credentials template file.", err)) }
  defer templateFileMinIO.Close()

  templateContentAWS, err := ioutil.ReadAll(templateFileAWS)
  tmpl, err := template.New("template").Parse(string(templateContentAWS))
  if err != nil { panic(fmt.Sprintf("Cannot parse template.", err)) }

  templateContentMinIO, err := ioutil.ReadAll(templateFileMinIO)
  tmplMinIO, err := template.New("template").Parse(string(templateContentMinIO))
  if err != nil { panic(fmt.Sprintf("Cannot parse template.", err)) }

  credentials, err := os.Create("/tmp/credentials")  
  defer credentials.Close()

  configjson, err := os.Create("/tmp/config.json")  
  defer configjson.Close()

  for data := range ch {
    data := data

    parts := strings.Fields(data)
    entry := credentialsStruct{
    Bucket: parts[0],
    AccessKey: parts[1],
    SecretKey: parts[2],
    }

    err := tmpl.Execute(credentials, entry)
    if err != nil { panic(fmt.Sprintf("Cannot parse AWS credentials.", err)) }
    fmt.Fprintln(credentials)

    err = tmplMinIO.Execute(configjson, entry)
    if err != nil { panic(fmt.Sprintf("Cannot parse credentials.", err)) }
    fmt.Fprintln(configjson)
  }

}

func getBucketKeys(bucket string, co *admin.API) (access_key string, secret_key string) {

  bucketStruct := admin.Bucket{Bucket: bucket}

  bucketInfo, err := co.GetBucketInfo(context.Background(), bucketStruct)
  if err != nil {
    panic(fmt.Sprintf("Unable to reach Ceph RGW API.", err))
  }

  decodedBucketInfo := admin.Bucket{}
  codedBucketInfo, err := json.Marshal(bucketInfo)
  err = json.Unmarshal(codedBucketInfo, &decodedBucketInfo)

  ownerUserID := decodedBucketInfo.Owner

  userStruct := admin.User{ID: ownerUserID}
  userInfo, err := co.GetUser(context.Background(), userStruct)
  if err != nil {
    panic(fmt.Sprintf("Unable to reach Ceph RGW API.", err))
  }

  decodedUserInfo := admin.User{}
  codedUserInfo, err := json.Marshal(userInfo)
  err = json.Unmarshal(codedUserInfo, &decodedUserInfo)

  keys := decodedUserInfo.Keys

  decodedKeys := []admin.UserKeySpec{}
  keysJSON, err := json.Marshal(keys)
  err = json.Unmarshal(keysJSON, &decodedKeys)
  readDecodedKey := decodedKeys[0]

  User := readDecodedKey.User
  access_key = readDecodedKey.AccessKey
  secret_key = readDecodedKey.SecretKey

  fmt.Println(bucket, User, access_key, secret_key)
  return access_key, secret_key
}

func connectCeph() (*admin.API, error) {
  co, err := admin.New("https://your.ceph.endpoint", "radosgw-admin-accesskey", "radosgw-admin-secretkey", nil)
    if err != nil {
        panic(fmt.Sprintf("Cannot connect to Ceph RGW.", err))
    }

  return co, nil
}

type credentialsStruct struct {
  Bucket    string
  AccessKey string
  SecretKey string
}
