# filestore
----

## Running filestore locally


#### Start the server

```
cd server
go run main.go
```

The filestore server will start listening on port 8090


#### Creating the client binary

>Note: Just run the `make` command which will create a binary named `store`


Once the server is listing on port 8090 you can start running different command using store binary


1. Add file to store
``` 
$ ./store add first.txt second.txt new.txt

File first.txt uploaded successfully
File second.txt uploaded successfully
File new.txt uploaded successfully

```
2. List files in the store
```
$ ./store ls

first.txt
new.txt
second.txt
```

3. Remove a file
```
$ ./store rm first.txt

Remove file name is first.txt
```
4. Update contents to a file in the store
```
$ ./store update first.txt

File second.txt updated successfully
```
4. Returns number of words in the file
```
$ ./store wc first.txt second.txt

 41 , new.txt

 9 , second.txt
```



## Running server on the k8s cluster

> Note: Here I have used kind cluster

- Create the deployement and service available inside the deployment/server directory

Now port-forward you service using below command
```
kubectl port-forward svc/filestoreserver-service 8090:8090
```
Once the portforward starts working you can create the client binary and run all the above command it will work as expected.