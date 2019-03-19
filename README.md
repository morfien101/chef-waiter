# Chef Waiter

[![Build Status](https://travis-ci.org/morfien101/chef-waiter.svg?branch=master)](https://travis-ci.org/morfien101/chef-waiter)

A simple HTTP(S) API wrapper around chef client.

[What is the Chef Waiter](#what-is-the-chef-waiter)

[How do I use Chef Waiter](#how-do-i-use-chef-waiter)

[Installing](#installing)

[Running](#running)

[Configuration File](#configuration-file)

[Maintenance mode](#maintenance-mode)

[Locking the chef Waiter](#locking-the-chef-waiter)

[Chef service replacement](#chef-service-replacement)

[Example Flow](#example-flow)

[Logging](#logging)

[Metrics](#metrics)

![waiter](./README/waiter_T.png "chef waiter")

## What is the Chef Waiter

The Chef Waiter service was created to enable on demand runs of chef without the use of push jobs.
Push jobs are not available when using the managed chef service from opscode.

This leaves you in a situation that you have to run chef __very__ frequently or just wait for changes to roll out. This poses an issue on CD pipelines as the CD pipeline is non-deterministic. With out knowing if a chef run passes or fails you don't know if the deploy worked.

## I just use SSH or Winrm, so why do I need this?

Using SSH and WinRM is a valid way to do this but it opens a security hole in the process. SSH/WinRM with out very tight control can open a gate into your servers what would allow someone to do anything on it.

When using the chef waiter it can only do one thing... Run chef.

The chef waiter was built with CI/CD in mind and all responses are mainly for integration with these pipelines. This makes integration easier than running chef via SSH and WinRM.

A bonus point, chef waiter will run chef under its control and track it. Therefore if you have unstable connections you don't need to worry about it killing your chef runs.

## How do I use Chef Waiter

The following URLs are available to you.
The are designed to be used by API clients rather than Web browsers. They return JSON strings mostly or simple text.

The default port for running Chef Waiter on is 8901 TCP.

Example request:

```bash
$> curl http://127.0.0.1:8901/chefclient
```

```json
{
    "35434398-b40a-4686-ab38-38deccd4241b": {
        "status":"registered",
        "exitcode":99,
        "starttime":1542124123,
        "ondemand":true
    }
}
```

```bash
curl -v -XPOST http://localhost:8901/chefclient --data-raw "recipe[chefwaiter::test]"
```

```json
{
    "0a92d0a7-dfda-4b28-8195-0e00ff120fc5":{
        "status":"running",
        "exitcode":99,
        "starttime":1542124815,
        "ondemand":true,
        "custom_run":true,
        "custom_run_string":"recipe[chefwaiter::test]"
    }
}
```

```bash
$> curl http://127.0.0.1:8901/chefclient/35434398-b40a-4686-ab38-38deccd4241b
```

```json
{
    "35434398-b40a-4686-ab38-38deccd4241b": {
        "status":"complete",
        "exitcode":0,
        "starttime":1542124123,
        "ondemand":true
    }
}
```

```bash
$> curl http://127.0.0.1:8901/chef/lastrun
```

```json
{
    "last_run_guid":"35434398-b40a-4686-ab38-38deccd4241b"
}
```

Chefwaiter will determine if the chef run passed or failed based on the exit code of the run. If the run passed you will see a status of `complete` if it failed you will see `failed`.

Below is a table describing the API for chef waiter. Chefwaiter was built with easy understanding for humans in mind. MOST the requests are GET based. There is very little that chefwaiter needs in terms of data and these are passed in via the URL.

| URL | METHOD |Description|
|-----|--------|------------|
| /chefclient | GET | Use this to create a run. You will have a json payload returned with a guid for the run.
| /chefclient | POST | Use this to create a run with a custom recipe string. See chef -o option. The string should be like `"recipe[chefwaiter::test]"`. It is also possible to override the lock with a query parameter in the URL `force=true`.
| /chefclient/{guid} | GET | Used with the GUID that you received from /chefclient to get the status of the run.
| /cheflogs/{guid} | GET | Used with the GUID that you received from /chefclient to get the chef logs from a run.
| /chef/nextrun | GET | Used to get the time when the next run will happen. This time is the time when the server is free to start the next run and will usually happen with in a minute of this time.
|/chef/interval| GET | Used to get the time between automatic chef runs.
|/chef/interval/{i}| GET | Used to set the time between chef runs. This needs to be a positive number and represents minutes between runs.
|/chef/on| GET | Used to turn on automatic runs of chef
|/chef/off| GET | Used to turn off automatic runs of chef
|/chef/lastrun| GET | Returns the guid of the last run. It starts as blank when the service starts.
|/chef/allruns| GET | Used to get the state of all jobs in chefwaiter currently.
|/chef/enabled| GET | Used to check if chef is currently enabled to run periodically
|/chef/maintenance| GET | Shows if the chef waiter is in maintenance mode currently.
|/chef/maintenance/start/{i}| GET | Requests that chef waiter be put into maintenance mode for i number of minutes. This must be a whole number.
|/chef/maintenance/end| GET | Removes the maintenance timer allowing periodic runs to start again.
|/chef/lock| GET | Shows the status of the lock for runs.
|/chef/lock/set| GET | Turns on the lock for chef runs. Stops any runs from occurring.
|/chef/lock/remove| GET | Turns off the lock for chef runs. Enables normal operation again.
|/_status | GET | Return status information about the chef waiter.
| /healthcheck | GET | Returns a 200 OK to show that the server is online.

## Custom Runs

Chef waiter is able to do custom runs which allow you run recipes once without change the default run list.
This is useful when you want to run a subset of recipes or to bootstrap a machine then run a maintenance recipe the rest of the time.

It is important to consider the security implications of this. This effectively allows the chef waiter to run ANY recipe on your chef server once for each request.

With this in mind the configuration has a toggle that allows you to whitelist the text that you are allowed to post the chef waiter when requesting a custom run.

The text that you send needs to match exactly what you put in your whitelists. The whitelist is a list so many options can be made available.

See the [Configuration File](#configuration-file) for more details.

## Installing

### Preferred option

The chef waiter can be installed via the [chef-waiter cookbook](https://github.com/morfien101/chef-waiter-cookbook).

### Optional method

1. Download chef-waiter from the releases page.
1. Extract the binary
1. Move the binary to somewhere to run it.
1. From a terminal or prompt windows run the below:

```bash
# Linux
/usr/local/bin/chefwaiter --service install
```

```cmd
# Windows
C:\Program Files\chefwaiter\chefwaiter.exe --service install
```

Remember to allow **8901 TCP** Inbound if you choose to install manually.

## Running

The service is runs the same on Windows and Linux. The service binary itself is responsible for creating the service files needed to start and run the chef waiter as a service on which ever OS.

Make sure that the service is running after installing it as discussed in the _Installing_ section.
The service will need port **8901-TCP** open to communicate with the outside world.

### Firewall access

| Port | Protocol | Description |
|----|--------|-----------|
| 8901 | TCP | Used to host the HTTP service for Chef Waiter

### Directories of interest

|Directory|OS|Description|
|---------|--|-----------|
|/etc/chefwaiter/ | Linux | Used to store configuration and state files for Chef Waiter|
|/usr/local/bin/ | Linux | Location the binary is stored on a linux computer|
|/var/log/chefwaiter/ |Linux| Location where Chef Waiter will store the log files for chef|
|C:\Program Files\chefwaiter\ | Windows | Location of both configuration files and binary|
|C:\logs\chefwaiter\ |Windows| Location where Chef Waiter will store the log files for chef|

### Configuration file

The Chef Waiter can be configured by a configuration file in the form of json.

The file location needs to be set using an environment variable.

`CHEFWAITER_CONFIG`

It should have the value of the file path eg:

`/etc/chefwaiter/config.json`

or

`c:\Program Files\chefwaiter\config.json`

If no config file is specified Chef Waiter will start with sane defaults.

An example file is below:

```json
{
    "state_table_size": 20,
    "periodic_chef_runs": true,
    "run_interval": 10,
    "debug": false,
    "logs_location": /var/log/chefwaiter,
    "state_location": /etc/chefwaiter,
    "metrics_enabled": true,
    "metrics_host": "statsd-client.local:8125",
    "metrics_default_tags": {
        "tag_name": "value",
        "tag_name": "value"
    },
    "whitelist_custom_runs": true,
    "allowed_custom_runs": [
        "role[chefwaiter]",
        "recipe[deploy_new_app]"
    ]
}
```

Default Configuration settings:

| Setting | Windows | Linux | Description |
---|---|---|---
|state_table_size| 20 | 20 | Chefwaiter will keep a log of the past x number of run. This setting dictates that value. |
| periodic_chef_runs | true | true | This setting will tell chef waiter to run chef runs periodically like the normal chef service. |
| run_interval | 30 | 30 | How often in minutes should chef waiter start a chef run. |
| debug | false | false | Show debug log printing. |
| logs_location | C:\logs\chefwaiter | /var/log/chefwaiter | Where should chefwaiter store the chef run logs. |
| state_location | C:\Program Files\chefwaiter | /etc/chefwaiter | Chefwaiter writes a state file to disk periodically to maintain state through reboots. This settings dictates where that file should be kept. |
| enable_tls | false | false | Should Chefwaiter us TLS on the web server. |
| certificate_path | ./cert.crt | ./cert.crt | location of the TLS certificate. |
| key_path | ./cert.key | ./cert.key | Location of the TLS certificates private key. |
metrics_enabled | false | false | Turn on the statsd metric shipper.
metrics_host | 127.0.0.1:8125 | 127.0.0.1:8125 | Location of the statsd server.
metrics_default_tags | nil | nil | Custom tags that you would like to add in key value pairs.
| whitelist_custom_runs | false | false | Turn on the whitelist for custom runs.
| allowed_custom_runs | nil | nil | A list of the text that chef waiter will accept for white listing the custom runs.

## Maintenance mode

The Chef Waiter can be put into maintenance mode.

This dictates that **no periodic runs will be allowed to be triggered during this time period**.

Therefore any periodic runs will be skipped and you would have to wait for the next time trigger to be started.

Maintenance mode has no effect to **on demand** runs.

This will allow you to control the runs but also to stop uncontrolled runs from occurring while you are doing deployments.

## Locking the chef waiter

Chef waiter has a lock out mode in it. This allows you to request that a server not run chef `on demand` or `periodically`.

This is useful for servers that are in production that you wish to protect from accidental changes.

Use `/chef/lock` for checking the status of the lock.

`/chef/lock/set` and `/chef/lock/remove` will enable and disable the lock respectively.

The lock can be overridden when running a custom job. This is because the job is already very specific, use with care.

It requires that you send a `force=true` query parameter in the URL when sending requests.

See example below:

```bash
curl "http://localhost:8901/chefclient?force=true" --data '"recipe[chefwaiter::test]"'
```

## Chef service replacement

The Chef Waiter has been written to be a replacement for the chef __service__.

This feature is turned on by default and can be turn off with the use of the "periodic_chef_runs" configuration file setting. By making use of this feature you will be able to get the logs for the periodic runs via the API.

Important:

```text
Periodic runs will run before on demand runs should there be 2 that are ready to be kicked off at the same time.

This in turn means that your on demand runs will always return the latest details.
```

The periodic runs can also be controlled via the API. You can change the interval in minutes for the runs as well as turn it off and on. You can use "/chef/lastrun" to get the GUID for the last run which will allow you to get the logs for the run.

The logs for the periodic runs are stored under the same directory as on demand runs. They are also subject to the same clean up process. This means that you do not need to rotate logs as the chef waiter will do that for you.

## Example Flow

1. /chefclient (gather the GUID from this step)
1. /chefclient/<guid from step 1>
1. /cheflogs/<guid from step 1>

If you request a run while a run is queued you will keep getting the same guid back until the run starts and a new run can be queued. This means that you can only ever have 1 chef run running and 1 queued at a time.

## Logging

The service will log to the default logging system for the OS that it is running on. Either Windows Event Viewer or Syslog for linux.

Logs for chef runs will be contained in files that have the name set to the GUID that represents the chef run.

The files will be cleared out by the chef waiter periodically. This is triggered every minute and is controlled by a flag to specify the number of log files that you want to keep. The default is 20.
This can be changed as well as the location of the logs by settings in the above configuration file.

Log file paths will look like below:

```text
# Linux
/var/log/chefwaiter/0038cf85-68a1-4b8a-8898-f56261f02d65.log

# Windows
* C:\logs\chefwaiter\0038cf85-68a1-4b8a-8898-f56261f02d65.log
```

## Metrics

Chef waiter sends out statsd metrics to an endpoint dictated by the `metrics_host` configuration value. Metrics need to be enabled by setting the `metrics_enabled` to `true` in the configuration file. If the values is not set no metrics will be sent.

All metrics will have a tag `host` which will be the host name or `not_available` if it can't be found for some reason.
The hostname can be overridden in the configuration by setting a tag called `host`.

Chef waiter will try to lookup the DNS record of the endpoint once ever 2 minutes. This allows for DNS name changes to happen with out the need to restart the chef waiter. Useful in modern distributed compute environments.

The following metrics are available.

Metric Name | Metric Tags | Description
---|---|---
chefwaiter_starting | version: [chefwaiter_version] | Event sent when starting the chef waiter.
chefwaiter_shutting_down | version: [chefwaiter_version] | Event sent when stopping the chef waiter.
chefwaiter_state_table_size | none | How large the state table is. This should be the same as the number of logs being held by the chef waiter.
chefwaiter_chef_run_time | none | How long the chef run took in Milliseconds
chefwaiter_run_starting | job_type: ["periodic", "demand"] | A chef run has started.
chefwaiter_run_finished | job_type: ["periodic", "demand"] | A chef run has finished.
