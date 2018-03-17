# Distributed Tracing Example

This is the sample code and example application for something I wrote on my blog: [Distributed Tracing and You](https://www.rayashmanjr.com/post/171945496692/distributed-logging-and-you).  It's a bunch of toy go web services added on to a fork of [deviantony's ELK stack docker-compose repo](https://github.com/deviantony/docker-elk).  If you'd like to play with it or follow along, you should just be able to do the following:

1.  Clone this repository
2.  docker-compose up

Once that's finished (and it'll take a while...), you have the following:

* A web front-end that accepts a street address on port 8083
* A geolocation service that takes in a street address on port 8081
* A temperature service that accepts latitude and longitude on 8082
* Kibana (for searching the logs) on port 5601