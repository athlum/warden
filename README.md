# Warden [![Build Status](https://travis-ci.org/athlum/warden.png)](https://travis-ci.org/athlum/warden)

<!--
create time: 2016-01-18 22:54:02
Author: Athlum
-->

Warden is a service discovery tool for [docker](https://www.docker.com/) + [mesos](http://mesos.apache.org/). 

Warden uses [zookeeper](http://zookeeper.apache.org/) to store service network data(ip address and port). Tasks launched in container would be took by warden agent and registered after pass the user-defined health check. Agent will keep watching on those containers and unregister any container failed in health check.

There are three modules in warden project.

* **warden-agent** - Warden agent watches on docker daemon and take created containers as ephemeral node on zookeeper. Agent do the health check periodically to guarantee all watched containers are healthy.
* **warden-guardian** - Warden guardian is a service running on your HAProxy/Nginx server that provide the upstream automatically, as soon as any watched app znode updated.
* **warden-template** - It's a commandline upstream/frontend render tool for service registration.