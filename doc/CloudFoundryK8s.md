# Cloudfoundry on Kubernetes

## Deployment problem

A setup of cloudfoundry in kubernetes could look like the following diagram.

![](CloudFoundryK8s.png)


Setting up this scenario with a normal `helm` could be challanging because it is quite difficult to share a database instance between different helm charts.

A look at the existing `helm` charts for `capi`, `eirini` or `uaa` also showed that sharing somthing like a database instance is not in focus. It seems that every `helm` chart tries
to provide its own database. The same is also true for logging or ingress routes.

But maintaining a set of multiple database instances within one CloudFounry installation seems to be a lot of effort.

## Possible solution

This problem could be solved quite easy using [shalm](https://github.com/kramerul/shalm.git)

Shalm uses helm rendering of templates and additionally allows you to [define API](https://github.com/kramerul/shalm/blob/b195a681148171ed208a4ff314e1c1c5b0a7f376/example/cloudfoundry/database/Chart.star#L4) between charts.

These API can be used to share instances. Instances of `shalm` charts can be simply passed to the [constructor](https://github.com/kramerul/shalm/blob/b195a681148171ed208a4ff314e1c1c5b0a7f376/example/cloudfoundry/capi/Chart.star#L1) of other `shalm` charts. 
In addition, existing `helm` charts (e.g. [postgresql](https://github.com/kramerul/shalm/blob/b195a681148171ed208a4ff314e1c1c5b0a7f376/example/cloudfoundry/database/Chart.star#L2)) can be reused.

### Example

The [cloudfoundry example](https://github.com/kramerul/shalm/tree/master/example/cloudfoundry) illustrates in a simple way how `shalm` could be used to solve this problem. There is a lot of stuff missing in the example, but hopefully it shows how it could work.

The example can be run using the follwing command

```bash
go run github.com/kramerul/shalm template example/cloudfoundry/cloudfoundry
```