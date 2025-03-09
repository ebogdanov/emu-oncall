# Emu-OnCall
Template of REST API service which is recognized by on-premise Grafana On-Call installation as Cloud instance
Implemented in Golang

Using this engine you can ~~install~~ create own on-premise instance and create connections for sending alerts into your desired IMs or any corporate systems, whatever you like and want use to

Article in Russian with explanation and example of usage: https://habr.com/ru/articles/739948/ 

Main concepts: 
* If phone number is filled for user in database, you'll get "verified" flag in OnCall on-premise, and can set up rules of notifications with usual OnCall's IRM flow and you own flexibility
* You can route alerts to HTTP APIs, VoIP, whatever you like and prefer
* plugin/ folder is used for engines, which implementation is your own implementation, you can combine any logic, based on your needs
* Doesn't have UI, since it's not clear how it should look like. Right now this is only REST API

### Application configuration
Based on Viper library, so that can read command line arguments, env variables, configration files (YAML format)
Example can be found in config/config.yaml file

| Option name                               | Default value     | Description                                                                                                                                                                                                                  |  
|-------------------------------------------|-------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| debug                                     | false             | Option enables pprof in http handlers                                                                                                                                                                                        |
| config                                    | config/config.yml | Can be used as CLI option to specify config file to read with                                                                                                                                                                |
| app.port                                  | :8880             | Port on which applicaiton should listen for connections                                                                                                                                                                      |
| app.hostname                              | (blank)           | Hostname which will be sent to OnCall. This should be specified if service is proxied by something                                                                                                                           |
| app.log_level                             | debug             | Logging level (by zerolog naming)                                                                                                                                                                                            |
| app.log_path                              | (blank)           | path of log file there records should be appended. If not specified - logs will be written in JSON format to stdout                                                                                                          | 
| app.version                               | 1.0               | Application version, can be used for debugging, etc                                                                                                                                                                          |
| app.auth_token                            | (blank)           | This is token which should be set in OnCall UI. if not specified - any string will be accepted as correct                                                                                                                    |
| grafana.url                               | (blank)           | If you want more details more details about incident to be loaded from Grafana, you need to setup connection between them                                                                                                    |
| grafana.header_token                      | (blank)           | Technical account credentials, for some reason Service account doesn't works for me                                                                                                                                          |
| grafana.oncall.url                        | (blank)           | If you want to load some more data from OnCall (like track schedules), you need to specify it                                                                                                                                |
| grafana.oncall.header_token               | (blank)           | Reserved for future usage                                                                                                                                                                                                    | 
| grafana.oncall.incident_details           | false             | Using this option you can load more details about incident                                                                                                                                                                   |
| grafana.oncall.schedules                  | (blank)           | Map of schedules to be tracked and notification sent to current on duty person                                                                                                                                               | 
| grafana.oncall.schedules.[NAME].name      | (blank)           | Name of schedule, to be used internally. Can be overwritten if ICal will show it's own name                                                                                                                                  |
| grafana.oncall.schedules.[NAME].transport | "local"           | Can be "local" or "callback". How to notify about change of state in schedule. local - means send SMS via plugin implementation. Callback can call any URL if specified. It's good to use if escalation chain should be used |
| grafana.oncall.schedules.[NAME].ical_url  | (blank)           | ICal export URL from OnCall UI, so that engine can track changes in schedules and send notifications                                                                                                                         |
| grafana.oncall.[NAME].callback_url        | (blank)           | URL to be sent data about schedule change. Bet way to send notification using it as alerts with native On-Call workflow (different channels, history, escalation, etc)                                                       |
| db.addr                                   | 127.0.0.1         | Database Hostname                                                                                                                                                                                                            |
| db.port                                   | 5432              | Database Port (PostgreSQL-tested, but if requited you can try to use other driver)                                                                                                                                           |
| db.user                                   | admin             | Database username                                                                                                                                                                                                            |
| db.password                               | admin             | Database password                                                                                                                                                                                                            | 
| db.dbname                                 | emu_oncall        | Database name                                                                                                                                                                                                                |
| db.driver                                 | postgres          | Database driver name, if you want to use different rather than PostgreSQL                                                                                                                                                    | 
| plugin.*                                  | Map               | If you need to send some configuration to your plugin you can use it via plugin map of interfaces                                                                                                                            |

### Start 
1. Install service (depends on your environment)
2. Check it's service (look at scripts/info.sh for example)
3. Setup for your OnCall instance environment variable GRAFANA_CLOUD_ONCALL_API_URL pointing to installed URL, like ```GRAFANA_CLOUD_ONCALL_API_URL: "http://192.168.0.205:8880"```
4. Restart Grafana OnCall Instance
5. In "your local" OnCall UI go to Settings -> Cloud -> Connect Open Source OnCall and Grafana Cloud OnCall -> enter API key -> press "Save key and connect"
6. If key is correct and there is network connectivity you should see Success message
7. To get user ability to send notification to this service you should add it in oncall_users table in the database: fill phone_number, and assign one of roles: admin/user. Also OnCall for some reason requires local user to have "Editor" role to be synced
8. That is it. Now you can route alerts to this instance and forward anywhere you like, for testing purposes messages will be sent to stdout

### Schedules tracking
If you want to track schedules you can copy URL of ICal for schedule, and configure it via config file. 
So that service every minute will track changes, and if any - send notification via local plugin or callback

### Plans
* Mobile application support
* Stateless implementation, so that service can be hosted in k8s with many pods
