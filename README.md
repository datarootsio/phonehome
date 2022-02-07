[![swagger](https://img.shields.io/badge/openapi-available-blue?logo=swagger)](https://api.phonehome.dev/swagger/index.html)
[![coverage](https://img.shields.io/badge/coverage-report-blueviolet?logo=go)](https://phonehome.dev/coverage.html)

# phonehome: KISS telemetry for FOSS packages

## What is it?

Sometimes you want to have a better understanding of how, how, what is using your package/tool/platform. While very good telemetry services exist for this [phonehome.dev](https://phonehome.dev) provides a free KISS service for FOSS packages with a focus of staying out of your way.

## How to use it

Using the [phonehome.dev](https://phonehome.dev) telemetry service is as easy as posting a HTTP POST request. Let's have a look at an example.

Let's assume your Github user / repo is: `foouser/barrepo`.

You can then do a POST request to `api.phonehome.dev/foouser/barrepo`. If you leave the body empty your essentially just "phoning home". This might be useful to get for example some usage stats on your CLI tool. More useful however is when you start including a structured body.

This allows you to record stuff that might be valuable to analyse retrospectively like the version of your software or the occurence of errors.

The POST endpoint `api.phonehome.dev/{organisation}/{repository}` take a body in the shape of a JSON object:

```json
{
    "version": "version of your software",
    "err": "this err occurred"
}
```

While the content needs to be a JSON object the keys and values are completely up to you to define. The main limitation is that nested objects are not allowed. Basically make sure to use a simple object with keys:values. When values that are not strings or numbers are encountered they are stripped of your payload and a warning announcing this will be added to the response.


## Guidelines

Some "please take this into consideration" guidelines.

- Be transparent to you users that you're using a telemetry service.
- If you expect exceptionally high volumes of telemetry calls, get in touch proactively.
- For now this only works for 

## Where to find more info

- [GitHub repo](https://github.com/datarootsio/phonehome)
- [Swagger docs](https://api.phonehome.dev/swagger/index.html)
- [Go coverage report](https://phonehome.dev/coverage.html)
- [questions & issues](https://github.com/datarootsio/phonehome/issues)
## Roadmap

Make sure to follow the GitHub [issues](https://github.com/datarootsio/phonehome/issues). 

Most likely Python and Go utility pacakages will be provided in the near future to ease registry telemetry calls and use them in the context of traces.
