[![Build Status](https://travis-ci.org/kevholditch/gokong.svg?branch=master)](https://travis-ci.org/kevholditch/gokong)

GoKong
======
A kong go client fully tested with no mocks!!

## IMPORTANT
GoKong has now been updated to support kong v1.0.0.  This is a breaking change release is not compatible with any versions <1.0.0.
 The good news is the guys over at Kong have stated that they are not going to make any breaking changes now (following semver).  If you need a version of gokong
 that supports Kong <1.0.0 then use the branch `kong-pre-1.0.0`.


## GoKong
GoKong is a easy to use api client for [kong](https://getkong.org/).  The difference with the gokong library is all of its tests are written against a real running kong running inside a docker container, yep that's right you won't see a horrible mock anywhere!!

## Supported Kong Versions
As per [travis build](https://travis-ci.org/kevholditch/gokong):
```
KONG_VERSION=1.0.0
```

## Importing

To add gokong via `go get`:
```
go get github.com/kevholditch/gokong
```

## Usage

Import gokong
```go
import (
  "github.com/kevholditch/gokong"
)
```

To create a default config for use with the client:
```go
config := gokong.NewDefaultConfig()
```

`NewDefaultConfig` creates a config with the host address set to the value of the env variable `KONG_ADMIN_ADDR`.
If the env variable is not set then the address is defaulted to `http://localhost:8001`.

There are a number of options you can set via config either by explicitly setting them when creating a config instance or
 by simply using the `NewDefaultConfig` method and using env variables.  Below is a table of the fields, the env variables that can be used
 to set them and their default values if you do not provide one via an env variable:

| Config property       | Env variable         | Default if not set    | Use                                                                             |
|:----------------------|:---------------------|:----------------------|:--------------------------------------------------------------------------------|
| HostAddress           | KONG_ADMIN_ADDR      | http://localhost:8001 | The url of the kong admin api                                                   |
| Username              | KONG_ADMIN_USERNAME  | not set               | Username for the kong admin api                                                 |
| Password              | KONG_ADMIN_PASSWORD  | not set               | Password for the kong admin api                                                 |
| InsecureSkipVerify    | TLS_SKIP_VERIFY      | false                 | Whether to skip tls certificate verification for the kong api when using https  |
| ApiKey                | KONG_API_KEY         | not set               | The api key you have used to lock down the kong admin api (via key-auth plugin) |
| AdminToken            | KONG_ADMIN_TOKEN     | not set               | The api key you have used to lock down the kong admin api (Enterprise Edition ) |


You can of course create your own config with the address set to whatever you want:
```go
config := gokong.Config{HostAddress:"http://localhost:1234"}
```

Also you can apply Username and Password for admin-api Basic Auth:
```go
config := gokong.Config{HostAddress:"http://localhost:1234",Username:"adminuser",Password:"yoursecret"}
```

If you need to ignore TLS verification, you can set InsecureSkipVerify:
```go
config := gokong.Config{InsecureSkipVerify: true}
```
This might be needed if your Kong installation is using a self-signed certificate, or if you are proxying to the Kong admin port.

Getting the status of the kong server:
```go
kongClient := gokong.NewClient(gokong.NewDefaultConfig())
status, err := kongClient.Status().Get()
```

Gokong is fluent so we can combine the above two lines into one:
```go
status, err := gokong.NewClient(gokong.NewDefaultConfig()).Status().Get()
```

## Consumers
Create a new Consumer ([for more information on the Consumer Fields see the Kong documentation](https://getkong.org/docs/0.13.x/admin-api/#consumer-object)):
```go
consumerRequest := &gokong.ConsumerRequest{
  Username: "User1",
  CustomId: "SomeId",
}

consumer, err := gokong.NewClient(gokong.NewDefaultConfig()).Consumers().Create(consumerRequest)
```

Get a Consumer by id:
```go
consumer, err := gokong.NewClient(gokong.NewDefaultConfig()).Consumers().GetById("e8ccbf13-a662-45be-9b6a-b549cc739c18")
```

Get a Consumer by username:
```go
consumer, err := gokong.NewClient(gokong.NewDefaultConfig()).Consumers().GetByUsername("User1")
```

List all Consumers:
```go
consumers, err := gokong.NewClient(gokong.NewDefaultConfig()).Consumers().List()
```

Delete a Consumer by id:
```go
err := gokong.NewClient(gokong.NewDefaultConfig()).Consumers().DeleteById("7c8741b7-3cf5-4d90-8674-b34153efbcd6")
```

Delete a Consumer by username:
```go
err := gokong.NewClient(gokong.NewDefaultConfig()).Consumers().DeleteByUsername("User1")
```

Update a Consumer by id:
```go
consumerRequest := &gokong.ConsumerRequest{
  Username: "User1",
  CustomId: "SomeId",
}

updatedConsumer, err := gokong.NewClient(gokong.NewDefaultConfig()).Consumers().UpdateById("44a37c3d-a252-4968-ab55-58c41b0289c2", consumerRequest)
```

Update a Consumer by username:
```go
consumerRequest := &gokong.ConsumerRequest{
  Username: "User2",
  CustomId: "SomeId",
}

updatedConsumer, err := gokong.NewClient(gokong.NewDefaultConfig()).Consumers().UpdateByUsername("User2", consumerRequest)
```

## Plugins
Create a new Plugin to be applied to all Services, Routes and Consumers do not set `ServiceId`, `RouteId` or `ConsumerId`.  Not all plugins can be configured in this way
 ([for more information on the Plugin Fields see the Kong documentation](https://getkong.org/docs/0.13.x/admin-api/#add-plugin)):

```go
pluginRequest := &gokong.PluginRequest{
  Name: "request-size-limiting",
  Config: map[string]interface{}{
    "allowed_payload_size": 128,
  },
}

createdPlugin, err := gokong.NewClient(gokong.NewDefaultConfig()).Plugins().Create(pluginRequest)
```

Create a new Plugin for a single Consumer (only set `ConsumerId`), Not all plugins can be configured in this way ([for more information on the Plugin Fields see the Kong documentation](https://getkong.org/docs/0.13.x/admin-api/#plugin-object)):
```go
client := gokong.NewClient(gokong.NewDefaultConfig())

consumerRequest := &gokong.ConsumerRequest{
  Username: "User1",
  CustomId: "test",
}

createdConsumer, err := client.Consumers().Create(consumerRequest)

pluginRequest := &gokong.PluginRequest{
  Name: "response-ratelimiting",
  ConsumerId: createdConsumer.Id,
  Config: map[string]interface{}{
    "limits.sms.minute": 20,
  },
}

createdPlugin, err := client.Plugins().Create(pluginRequest)
```

Create a new Plugin for a single Service (only set `ServiceId`), Not all plugins can be configured in this way ([for more information on the Plugin Fields see the Kong documentation](https://getkong.org/docs/0.13.x/admin-api/#plugin-object)):
```go
client := gokong.NewClient(gokong.NewDefaultConfig())

serviceRequest := &gokong.ServiceRequest{
  Name:     String("service"),
  Protocol: String("http"),
  Host:     String("example.com"),
}

createdService, err := client.Services().Create(serviceRequest)

pluginRequest := &gokong.PluginRequest{
  Name: "response-ratelimiting",
  ServiceId: gokong.ToId(*createdService.Id),
  Config: map[string]interface{}{
    "limits.sms.minute": 20,
  },
}

createdPlugin, err := client.Plugins().Create(pluginRequest)
```

Create a new Plugin for a single Route (only set `RouteId`), Not all plugins can be configured in this way ([for more information on the Plugin Fields see the Kong documentation](https://getkong.org/docs/0.13.x/admin-api/#plugin-object)):
```go
client := gokong.NewClient(gokong.NewDefaultConfig())

serviceRequest := &gokong.ServiceRequest{
  Name:     String("service"),
  Protocol: String("http"),
  Host:     String("example.com"),
}

createdService, err := client.Services().Create(serviceRequest)

routeRequest := &gokong.RouteRequest{
  Protocols:    StringSlice([]string{"http"}),
  Methods:      StringSlice([]string{"GET"}),
  Hosts:        StringSlice([]string{"example.com"}),
  Paths:        StringSlice([]string{"/"}),
  StripPath:    Bool(true),
  PreserveHost: Bool(false),
  Service:      &gokong.RouteServiceObject{Id: *createdService.Id},
}

createdRoute, err := client.Routes().Create(routeRequest)

pluginRequest := &gokong.PluginRequest{
  Name: "response-ratelimiting",
  RouteId: gokong.ToId(*createdRoute.Id),
  Config: map[string]interface{}{
    "limits.sms.minute": 20,
  },
}

createdPlugin, err := client.Plugins().Create(pluginRequest)
```

Get a plugin by id:
```go
plugin, err := gokong.NewClient(gokong.NewDefaultConfig()).Plugins().GetById("04bda233-d035-4b8a-8cf2-a53f3dd990f3")
```

List all plugins:
```go
plugins, err := gokong.NewClient(gokong.NewDefaultConfig()).Plugins().List(&PluginQueryString{})
```

Delete a plugin by id:
```go
err := gokong.NewClient(gokong.NewDefaultConfig()).Plugins().DeleteById("f2bbbab8-3e6f-4d9d-bada-d486600b3b4c")
```

Update a plugin by id:
```go
updatePluginRequest := &gokong.PluginRequest{
  Name:       "response-ratelimiting",
  ConsumerId: createdConsumer.Id,
  Config: map[string]interface{}{
    "limits.sms.minute": 20,
  },
}

updatedPlugin, err := gokong.NewClient(gokong.NewDefaultConfig()).Plugins().UpdateById("70692eed-2293-486d-b992-db44a6459360", updatePluginRequest)
```
## Configure a plugin for a Consumer
To configure a plugin for a consumer you can use the `CreatePluginConfig`, `GetPluginConfig` and `DeletePluginConfig` methods on the `Consumers` endpoint.
  Some plugins require configuration for a consumer for example the [jwt plugin[(https://getkong.org/plugins/jwt/#create-a-jwt-credential).

Create a plugin config for a consumer:
```go
createdPluginConfig, err := gokong.NewClient(gokong.NewDefaultConfig()).Consumers().CreatePluginConfig("f6539872-d8c5-4d6c-a2f2-923760329e4e", "jwt", "{\"key\": \"a36c3049b36249a3c9f8891cb127243c\"}")
```

Get a plugin config for a consumer by plugin config id:
```
pluginConfig, err := gokong.NewClient(gokong.NewDefaultConfig()).Consumers().GetPluginConfig("58c5229-dc92-4632-91c1-f34d9b84db0b", "jwt", "22700b52-ba59-428e-b03b-ba429b1e775e")
```

Delete a plugin config for a consumer by plugin config id:
```
err := gokong.NewClient(gokong.NewDefaultConfig()).Consumers().DeletePluginConfig("3958a860-ceac-4a6c-9bbb-ff8d69a585d2", "jwt", "bde04c3a-46bb-45c9-9006-e8af20d04342")
```

## Certificates
Create a Certificate ([for more information on the Certificate Fields see the Kong documentation](https://getkong.org/docs/0.13.x/admin-api/#certificate-object)):

```go
certificateRequest := &gokong.CertificateRequest{
  Cert: gokong.String("public key --- 123"),
  Key:  gokong.String("private key --- 456"),
}

createdCertificate, err := gokong.NewClient(gokong.NewDefaultConfig()).Certificates().Create(certificateRequest)
```

Get a Certificate by id:
```go
certificate, err := gokong.NewClient(gokong.NewDefaultConfig()).Certificates().GetById("0408cbd4-e856-4565-bc11-066326de9231")
```

List all certificates:
```go
certificates, err := gokong.NewClient(gokong.NewDefaultConfig()).Certificates().List()
```

Delete a Certificate:
```go
err := gokong.NewClient(gokong.NewDefaultConfig()).Certificates().DeleteById("db884cf2-9dd7-4e33-9ef5-628165076a42")
```

Update a Certificate:
```go
updateCertificateRequest := &gokong.CertificateRequest{
  Cert: gokong.String("public key --- 789"),
  Key:  gokong.String("private key --- 111"),
}

updatedCertificate, err := gokong.NewClient(gokong.NewDefaultConfig()).Certificates().UpdateById("1dc11281-30a6-4fb9-aec2-c6ff33445375", updateCertificateRequest)
```

# Routes

Create a Route ([for more information on the Route Fields see the Kong documentation](https://getkong.org/docs/0.13.x/admin-api/#route-object)):
```go
serviceRequest := &gokong.ServiceRequest{
  Name:     gokong.String("service-name" + uuid.NewV4().String()),
  Protocol: gokong.String("http"),
  Host:     gokong.String("foo.com"),
}

client := gokong.NewClient(NewDefaultConfig())

createdService, err := client.Services().Create(serviceRequest)

routeRequest := &gokong.RouteRequest{
  Protocols:    gokong.StringSlice([]string{"http"}),
  Methods:      gokong.StringSlice([]string{"GET"}),
  Hosts:        gokong.StringSlice([]string{"foo.com"}),
  StripPath:    gokong.Bool(true),
  PreserveHost: gokong.Bool(true),
  Service:      gokong.ToId(*createdService.Id),
  Paths:        gokong.StringSlice([]string{"/bar"})
}

createdRoute, err := client.Routes().Create(routeRequest)
```

To create a tcp route:
```go
routeRequest := &gokong.RouteRequest{
		Protocols:    gokong.StringSlice([]string{"tcp"}),
		StripPath:    gokong.Bool(true),
		PreserveHost: gokong.Bool(true),
		Snis:         gokong.StringSlice([]string{"example.com"}),
		Sources:      gokong.IpPortSliceSlice([]gokong.IpPort{{Ip: gokong.String("192.168.1.1"), Port: gokong.Int(80)}, {Ip: gokong.String("192.168.1.2"), Port: gokong.Int(81)}}),
		Destinations: gokong.IpPortSliceSlice([]gokong.IpPort{{Ip: gokong.String("172.10.1.1"), Port: gokong.Int(83)}, {Ip: gokong.String("172.10.1.2"), Port: nil}}),
		Service:      gokong.ToId(*createdService.Id),
	}
```

Get a route by ID:
```go
result, err := gokong.NewClient(gokong.NewDefaultConfig()).Routes().GetById(createdRoute.Id)
```

Get a route by Name:
```go
result, err := gokong.NewClient(gokong.NewDefaultConfig()).Routes().GetByName(createdRoute.Name)
```

List all routes:
```go
result, err := gokong.NewClient(gokong.NewDefaultConfig()).Routes().List(&gokong.RouteQueryString{})
```

Get routes from service ID or Name:
```go
result, err := gokong.NewClient(gokong.NewDefaultConfig()).Routes().GetRoutesFromServiceId(createdService.Id)
```

Update a route:
```go
routeRequest := &gokong.RouteRequest{
  Protocols:    gokong.StringSlice([]string{"http"}),
  Methods:      gokong.StringSlice([]string{"GET"}),
  Hosts:        gokong.StringSlice([]string{"foo.com"}),
  Paths:        gokong.StringSlice([]string{"/bar"}),
  StripPath:    gokong.Bool(true),
  PreserveHost: gokong.Bool(true),
  Service:      gokong.ToId(*createdService.Id),
}

createdRoute, err := gokong.NewClient(gokong.NewDefaultConfig()).Routes().Create(routeRequest)

routeRequest.Paths = gokong.StringSlice([]string{"/qux"})
updatedRoute, err := gokong.NewClient(gokong.NewDefaultConfig()).Routes().UpdateRoute(*createdRoute.Id, routeRequest)
```

Delete a route by ID:
```go
client.Routes().DeleteById(createdRoute.Id)
```

Delete a route by Name:
```go
client.Routes().DeleteByName(createdRoute.Id)
```

# Services

Create an Service ([for more information on the Service Fields see the Kong documentation](https://getkong.org/docs/0.13.x/admin-api/#service-object)):
```go
serviceRequest := &gokong.ServiceRequest{
		Name:     gokong.String("service-name-0"),
		Protocol: gokong.String("http"),
		Host:     gokong.String("foo.com"),
	}

	client := gokong.NewClient(gokong.NewDefaultConfig())

	createdService, err := client.Services().Create(serviceRequest)
```

Get information about a service with the service ID or Name
```go
serviceRequest := &gokong.ServiceRequest{
		Name:     gokong.String("service-name-0"),
    Protocol: gokong.String("http"),
    Host:     gokong.String("foo.com")
	}

client := gokong.NewClient(gokong.NewDefaultConfig())

createdService, err := client.Services().Create(serviceRequest)

resultFromId, err := client.Services().GetServiceById(createdService.Id)

resultFromName, err := client.Services().GetServiceByName(createdService.Id)
```

Get information about a service with the route ID
```go
result, err := gokong.NewClient(gokong.NewDefaultConfig()).Services().GetServiceRouteId(routeInformation.Id)
```

Get many services information
```go
result, err := gokong.NewClient(gokong.NewDefaultConfig()).Services().GetServices(&gokong.ServiceQueryString{
	Size: 500
	Offset: 300
})
```

Update a service with the service ID or Name
```go
serviceRequest := &gokong.ServiceRequest{
  Name:     gokong.String("service-name-0"),
  Protocol: gokong.String("http"),
  Host:     gokong.String("foo.com"),
}

client := gokong.NewClient(gokong.NewDefaultConfig())

createdService, err := client.Services().Create(serviceRequest)

serviceRequest.Host = gokong.String("bar.io")
updatedService, err := client.Services().UpdateServiceById(createdService.Id, serviceRequest)
result, err := client.Services().GetServiceById(createdService.Id)
```

Update a service by the route ID
```go
serviceRequest := &gokong.ServiceRequest{
  Name:     gokong.String("service-name-0"),
  Protocol: gokong.String("http"),
  Host:     gokong.String("foo.com"),
}

client := gokong.NewClient(gokong.NewDefaultConfig())

createdService, err := client.Services().Create(serviceRequest)

serviceRequest.Host = "bar.io"
updatedService, err := client.Services().UpdateServiceById(createdService.Id, serviceRequest)
result, err := client.Services().UpdateServicebyRouteId(routeInformation.Id)
```

Delete a service
```go
err = client.Services().DeleteServiceById(createdService.Id)
```

## SNIs
Create an SNI ([for more information on the Sni Fields see the Kong documentation](https://getkong.org/docs/0.13.x/admin-api/#sni-objects)):
```go
client := gokong.NewClient(gokong.NewDefaultConfig())

certificateRequest := &gokong.CertificateRequest{
  Cert: "public key --- 123",
  Key:  "private key --- 111",
}

certificate, err := client.Certificates().Create(certificateRequest)

snisRequest := &gokong.SnisRequest{
  Name:             "example.com",
  SslCertificateId: certificate.Id,
}

sni, err := client.Snis().Create(snisRequest)
```

Get an SNI by name:
```go
sni, err := client.Snis().GetByName("example.com")
```

List all SNIs:
```
snis, err := client.Snis().List()
```

Delete an SNI by name:
```go
err := client.Snis().DeleteByName("example.com")
```

Update an SNI by name:
```go
updateSniRequest := &gokong.SnisRequest{
  Name:             "example.com",
  SslCertificateId: "a9797703-3ae6-44a9-9f0a-4ebb5d7f301f",
}

updatedSni, err := client.Snis().UpdateByName("example.com", updateSniRequest)
```

## Upstreams
Create an Upstream ([for more information on the Upstream Fields see the Kong documentation](https://getkong.org/docs/0.13.x/admin-api/#upstream-objects)):
```go
upstreamRequest := &gokong.UpstreamRequest{
  Name: "test-upstream",
  Slots: 10,
}

createdUpstream, err := gokong.NewClient(gokong.NewDefaultConfig()).Upstreams().Create(upstreamRequest)
```

Get an Upstream by id:
```go
upstream, err := gokong.NewClient(gokong.NewDefaultConfig()).Upstreams().GetById("3705d962-caa8-4d0b-b291-4f0e85fe227a")
```

Get an Upstream by name:
```go
upstream, err := gokong.NewClient(gokong.NewDefaultConfig()).Upstreams().GetByName("test-upstream")
```

List all Upstreams:
```go
upstreams, err := gokong.NewClient(gokong.NewDefaultConfig()).Upstreams().List()
```

List all Upstreams with a filter:
```go
upstreams, err := gokong.NewClient(gokong.NewDefaultConfig()).Upstreams().ListFiltered(&gokong.UpstreamFilter{Name:"test-upstream", Slots:10})
```

Delete an Upstream by id:
```go
err := gokong.NewClient(gokong.NewDefaultConfig()).Upstreams().DeleteById("3a46b122-47ee-4c5d-b2de-49be84a672e6")
```

Delete an Upstream by name:
```go
err := gokong.NewClient(gokong.NewDefaultConfig()).Upstreams().DeleteById("3a46b122-47ee-4c5d-b2de-49be84a672e6")
```

Delete an Upstream by id:
```go
err := gokong.NewClient(gokong.NewDefaultConfig()).Upstreams().DeleteByName("test-upstream")
```

Update an Upstream by id:
```
updateUpstreamRequest := &gokong.UpstreamRequest{
  Name: "test-upstream",
  Slots: 10,
}

updatedUpstream, err := gokong.NewClient(gokong.NewDefaultConfig()).Upstreams().UpdateById("3a46b122-47ee-4c5d-b2de-49be84a672e6", updateUpstreamRequest)
```

Update an Upstream by name:
```go
updateUpstreamRequest := &gokong.UpstreamRequest{
  Name: "test-upstream",
  Slots: 10,
}

updatedUpstream, err := gokong.NewClient(gokong.NewDefaultConfig()).Upstreams().UpdateByName("test-upstream", updateUpstreamRequest)
```

## Targets
Create a target for an upstream ([for more information on the Target Fields see the Kong documentation](https://getkong.org/docs/0.13.x/admin-api/#upstream-objects)):
```go
targetRequest := &gokong.TargetRequest{
  Target:				"foo.com:443",
  Weight:				100,
}
createdTarget, err := gokong.NewClient(gokong.NewDefaultConfig()).Targets().CreateFromUpstreamId("upstreamId", targetRequest)
```

List all targets for an upstream
```go
targets, err := gokong.NewClient(gokong.NewDefaultConfig()).Targets().GetTargetsFromUpstreamId("upstreamId")
```

Delete a target from an upstream
```go
targets, err := gokong.NewClient(gokong.NewDefaultConfig()).Targets().DeleteFromUpstreamById("upstreamId")
```

Set target as healthy
```go
targets, err := gokong.NewClient(gokong.NewDefaultConfig()).Targets().SetTargetFromUpstreamByIdAsHealthy("upstreamId")
```

Set target as unhealthy
```go
targets, err := gokong.NewClient(gokong.NewDefaultConfig()).Targets().SetTargetFromUpstreamByIdAsUnhealthy("upstreamId")
```

List all targets for an upstream (including health status)
```go
targets, err := gokong.NewClient(gokong.NewDefaultConfig()).Targets().GetTargetsWithHealthFromUpstreamName(upstreamId)
```

**Notes**: Target methods listed above are overloaded in the same fashion as other objects exposed by this library. For parameters:
 - upstream - either name of id can be used
 - target - either id or target name (host:port) can be used

# Contributing
I would love to get contributions to the project so please feel free to submit a PR.  To setup your dev station you need go and docker installed.

Once you have cloned the repository the `make` command will build the code and run all of the tests.  If they all pass then you are good to go!

If when you run the make command you get the following error:
```
gofmt needs running on the following files:
```
Then all you need to do is run `make goimports` this will reformat all of the code (I know awesome)!!
