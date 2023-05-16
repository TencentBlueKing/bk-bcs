# BcsNodegroupManager Service

This is the BcsNodegroupManager service

Generated with

```
micro new bcs-nodegroup-manager
```

## Usage

Generate the proto code

```
make proto
```

Run the service

```
micro run .
```

```
// This can be removed once etcd becomes go gettable, version 3.4 and 3.5 is not,
// see https://github.com/etcd-io/etcd/issues/11154 and https://github.com/etcd-io/etcd/issues/11931.
replace google.golang.org/grpc => google.golang.org/grpc v1.27.1
```